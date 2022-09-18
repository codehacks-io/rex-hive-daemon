package message_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"rex-daemon/machine_meta"
	"rex-daemon/process_swarm"
	"rex-daemon/rexprint"
	"rex-daemon/swarm_message"
	"sync"
	"time"
)

const storeToDatabaseEverySeconds = 1
const maxMessagesToStorePerRequest = 50
const databaseTimeoutSeconds = 5
const mongoDatabaseName = "swarm-chan"
const mongoCollectionMessages = "m"
const mongoCollectionSpecs = "s"

var (
	holdingMessages = map[string]*swarm_message.SwarmMessage{}
	writingMessages []string
	lockForHolding  sync.Mutex
	lockForWriting  sync.Mutex
	didStartup      = false
	machineMeta     = &machine_meta.MachineMeta{}
	// Channels helps to block execution while there are still pending messages to store in DB
	flushChan *chan bool
)

func Run(swarmSpec *process_swarm.ProcessSwarm) {
	if didStartup {
		return
	}
	didStartup = true

	// Get machine metadata
	machineMeta = machine_meta.GetMachineMeta()
	swarmSpecId, err := insertOne(mongoCollectionSpecs, swarmSpec)
	if err != nil {
		fmt.Println("cannot insert swarm spec in mongodb", rexprint.ErrColor(err.Error()))
	} else {
		fmt.Println(rexprint.OutColor(fmt.Sprintf("swarm spec id inserted as %s", swarmSpecId.InsertedID)))
	}

	for {
		time.Sleep(storeToDatabaseEverySeconds * time.Second)
		hearBeat()
	}
}

func Flush(c *chan bool) {
	flushChan = c
}

func hearBeat() {
	bulkStoreMessagesInMongo()
}

func connectDb() (*mongo.Client, error) {
	// Get connection string
	if err := godotenv.Load(); err != nil {
		return nil, errors.New("no .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, errors.New("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	// Connect to database
	t := databaseTimeoutSeconds * time.Second
	opts := options.Client().ApplyURI(uri).SetTimeout(t).SetConnectTimeout(t).SetSocketTimeout(t).SetServerSelectionTimeout(t)
	return mongo.Connect(context.TODO(), opts)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func insertOne(collection string, document interface{}) (*mongo.InsertOneResult, error) {
	// Get DB connection
	client, err := connectDb()
	if err != nil {
		return nil, err
	}

	// Disconnect from DB on exit
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			fmt.Println(rexprint.ErrColor(err.Error()))
		}
	}()

	// Get DB collection
	coll := client.Database(mongoDatabaseName).Collection(collection)
	return coll.InsertOne(context.TODO(), document)
}

func insertMany(collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	// Get DB connection
	client, err := connectDb()
	if err != nil {
		return nil, err
	}

	// Disconnect from DB on exit
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			fmt.Println(rexprint.ErrColor(err.Error()))
		}
	}()

	// Get DB collection
	coll := client.Database(mongoDatabaseName).Collection(collection)
	return coll.InsertMany(context.TODO(), documents)
}

func bulkStoreMessagesInMongo() {
	// If no holding messages, there's nothing to store
	if len(holdingMessages) <= 0 {
		fmt.Println(rexprint.Dim("Holding zero messages, skipping save"))
		if flushChan != nil {
			*flushChan <- true
			flushChan = nil
		}
		return
	}

	// Get array of message IDs to store. Using mutex lock to make it goroutine-safe.
	lockForWriting.Lock()
	lockForHolding.Lock() // Lock the 'holding' map to quickly get messages to write
	holdingMessagesLength := len(holdingMessages)
	toWriteLength := min(holdingMessagesLength, maxMessagesToStorePerRequest)
	fmt.Println(rexprint.Dim(fmt.Sprintf("Will store %d messages. Holding: %d, Max: %d", toWriteLength, holdingMessagesLength, maxMessagesToStorePerRequest)))
	writingMessages = make([]string, toWriteLength)
	i := 0
	for k, v := range holdingMessages {
		writingMessages[i] = k
		// Add machine meta right before sending it to DB
		v.Machine = machineMeta
		i++
		if i >= toWriteLength {
			break
		}
	}

	// Prepare data to write into MongoDB
	docs := make([]interface{}, len(writingMessages))
	for index, key := range writingMessages {
		docs[index] = holdingMessages[key]
		i++
	}

	// IMPORTANT: Unlock the 'holding' map before trying to write to DB since that op could take a few seconds.
	lockForHolding.Unlock()

	// Insert data in MongoDB
	_, err := insertMany(mongoCollectionMessages, docs)

	if err != nil {
		// Reset the writing array, even if data fails to be stored in DB
		writingMessages = []string{}
		lockForWriting.Unlock() // Unlock writing array even if DB storing fails
		log.Println(err)
	} else {
		// Remove the stored messages from temp holding map. Using its own mutex to manipulate the map.
		lockForHolding.Lock()
		beforeStoreSize := len(holdingMessages)
		for _, k := range writingMessages {
			delete(holdingMessages, k)
		}
		fmt.Println(rexprint.Dim(fmt.Sprintf("Stored %d messages. Held before: %d, hold now: %d", len(writingMessages), beforeStoreSize, len(holdingMessages))))
		lockForHolding.Unlock()

		// Reset the writing array, as data has been written to DB
		writingMessages = []string{}
		lockForWriting.Unlock() // Unlock writing array after DB storing completes
	}

	// If flushChan is set, flushing has been requested.
	if flushChan != nil {
		lockForHolding.Lock()
		thereAreHoldingMessages := len(holdingMessages) >= 1
		lockForHolding.Unlock()
		if !thereAreHoldingMessages {
			fmt.Println(rexprint.Dim("Flush requested: All messages stored in DB, flush complete."))
			*flushChan <- true
			flushChan = nil
		} else {
			fmt.Println(rexprint.Dim("Flush requested: There are pending messages to be stored in DB, will wait to flush..."))
		}
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
		fmt.Println(rexprint.ErrColor(err.Error()))
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			fmt.Println(rexprint.ErrColor(err.Error()))
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
		fmt.Println(rexprint.ErrColor(err.Error()))
	}
	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println(rexprint.ErrColor(err.Error()))
	}
	fmt.Printf("%s\n", jsonData)
}

func ProcessSwarmMessage(message *swarm_message.SwarmMessage) {
	lockForHolding.Lock()
	message.Time = time.Now()
	idd, _ := uuid.NewRandom()
	holdingMessages[idd.String()] = message
	lockForHolding.Unlock()
}
