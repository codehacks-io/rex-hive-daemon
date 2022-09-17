package message_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"rex-daemon/machine_meta"
	"rex-daemon/swarm_message"
	"sync"
	"time"
)

const storeToDatabaseEverySeconds = 1
const databaseTimeoutSeconds = 5

var (
	holdingMessages = map[string]*swarm_message.SwarmMessage{}
	writingMessages []string
	lockForHolding  sync.Mutex
	lockForWriting  sync.Mutex
	didStartup      = false
	machineMeta     = &machine_meta.MachineMeta{}
)

func Run() {
	if didStartup {
		return
	}
	didStartup = true

	for {
		machineMeta = machine_meta.GetMachineMeta()
		time.Sleep(storeToDatabaseEverySeconds * time.Second)
		go hearBeat()
	}
}

func hearBeat() {
	bulkStoreMessagesInMongo()
}

func connectDb() *mongo.Client {
	// Get connection string
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Println("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	// Connect to database
	t := databaseTimeoutSeconds * time.Second
	opts := options.Client().ApplyURI(uri).SetTimeout(t).SetConnectTimeout(t).SetSocketTimeout(t).SetServerSelectionTimeout(t)
	client, err := mongo.Connect(context.TODO(), opts)

	// Handle connection error
	if err != nil {
		log.Println(err)
		return nil
	}

	return client
}

func bulkStoreMessagesInMongo() {
	// If no holding messages, there's nothing to store
	if len(holdingMessages) <= 0 {
		fmt.Println("------------  nothing to store, skipping --------------")
		return
	}

	// Get DB connection
	client := connectDb()
	if client == nil {
		return
	}

	// Disconnect from DB on exit
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Get DB collection
	coll := client.Database("swarm-chan").Collection("m")

	// Get array of message IDs to store. Using mutex lock to make it goroutine-safe.
	lockForWriting.Lock()
	lockForHolding.Lock() // Lock the 'holding' map to quickly get messages to write
	writingMessages = make([]string, len(holdingMessages))
	i := 0
	for k := range holdingMessages {
		writingMessages[i] = k
		i++
	}
	fmt.Println("------------  will store --------------", len(writingMessages))

	// Prepare data to write into MongoDB
	docs := make([]interface{}, len(writingMessages))
	for index, key := range writingMessages {
		docs[index] = holdingMessages[key]
		i++
	}

	// IMPORTANT: Unlock the 'holding' map before trying to write to DB since that op could take a few seconds.
	lockForHolding.Unlock()

	// Insert data in MongoDB
	_, err := coll.InsertMany(context.TODO(), docs)

	if err != nil {
		// Reset the writing array, even if data fails to be stored in DB
		writingMessages = []string{}
		lockForWriting.Unlock() // Unlock writing array even if DB storing fails
		log.Println(err)
	} else {
		// Remove the stored messages from temp holding map. Using its own mutex to manipulate the map.
		lockForHolding.Lock()
		fmt.Println("-- Store succeeded, holding before store ", len(holdingMessages))
		for _, k := range writingMessages {
			delete(holdingMessages, k)
		}
		fmt.Println("-- Holding after store ", len(holdingMessages))
		lockForHolding.Unlock()

		// Reset the writing array, as data has been written to DB
		writingMessages = []string{}
		lockForWriting.Unlock() // Unlock writing array after DB storing completes
	}
}

func testMongo() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("sample_mflix").Collection("movies")
	title := "Back to the Future"
	var result bson.M
	err = coll.FindOne(context.TODO(), bson.D{{"title", title}}).Decode(&result)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the title %s\n", title)
		return
	}

	if err != nil {
		panic(err)
	}
	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", jsonData)
}

func ProcessSwarmMessage(message *swarm_message.SwarmMessage) {
	lockForHolding.Lock()
	idd, _ := uuid.NewRandom()
	message.Machine = machineMeta
	holdingMessages[idd.String()] = message
	lockForHolding.Unlock()
}
