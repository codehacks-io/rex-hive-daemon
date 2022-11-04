package message_handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"rex-hive-daemon/hive_message"
	"rex-hive-daemon/hive_spec"
	"rex-hive-daemon/machine_meta"
	"rex-hive-daemon/rexprint"
	"rex-hive-daemon/slice_tools"
	"sync"
	"time"
)

const storeToDatabaseEverySeconds = 1
const maxMessagesToStorePerRequest = 200
const databaseTimeoutSeconds = 5
const mongoDatabaseName = "rex-hive"
const mongoCollectionHiveMessage = "message"
const mongoCollectionHiveRun = "run"

var (
	holdingMessages []*hive_message.HiveMessage
	writingMessages []string
	lockForHolding  sync.Mutex
	lockForWriting  sync.Mutex
	didStartup      = false
	machineMeta     = &machine_meta.MachineMeta{}
	// Channels helps to block execution while there are still pending messages to store in DB
	flushChan        *chan bool
	hiveRunId        interface{}
	flushRequestTime time.Time
)

const maxFlushGraceSeconds = 5

func Run(hiveSpec *hive_spec.HiveSpec) {
	if didStartup {
		return
	}
	didStartup = true

	// Get machine metadata
	machineMeta = machine_meta.GetMachineMeta()
	hiveSpec.RuntimeMachine = machineMeta
	hiveSpec.Id = genHiveRunId()
	hiveSpec.Time = time.Now()
	insertResult, err := insertOne(mongoCollectionHiveRun, hiveSpec)
	if err != nil {
		fmt.Println("cannot insert hive run in mongodb", rexprint.ErrColor(err.Error()))
	} else {
		hiveRunId = insertResult.InsertedID
		fmt.Println(rexprint.OutColor(fmt.Sprintf("hive run inserted with ID %s", hiveRunId)))
	}

	for {
		time.Sleep(storeToDatabaseEverySeconds * time.Second)
		hearBeat()
	}
}

func Flush(c *chan bool) {
	flushChan = c
	flushRequestTime = time.Now()
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
	docs := make([]interface{}, toWriteLength) // Prepare data to write into MongoDB
	for i, v := range holdingMessages {
		writingMessages[i] = v.TempId
		// Add machine meta right before sending it to DB
		v.RuntimeMachine = machineMeta
		v.HiveRunId = hiveRunId
		docs[i] = v
		i++
		if i >= toWriteLength {
			break
		}
	}

	// IMPORTANT: Unlock the 'holding' map before trying to write to DB since that op could take a few seconds.
	lockForHolding.Unlock()

	// Insert data in MongoDB
	_, err := insertMany(mongoCollectionHiveMessage, docs)

	if err != nil {
		// Reset the writing array, even if data fails to be stored in DB
		fmt.Println(rexprint.ErrColor(fmt.Sprintf("Error sotring %d messages. Held before: %d, hold now: %d. Error: %s", toWriteLength, holdingMessagesLength, len(holdingMessages), err)))
		writingMessages = []string{}
		lockForWriting.Unlock() // Unlock writing array even if DB storing fails
	} else {
		// Remove the stored messages from temp holding map. Using its own mutex to manipulate the map.
		lockForHolding.Lock()

		for _, k := range writingMessages {
			holdingMessages = *slice_tools.RemoveFirst(&holdingMessages, func(x *hive_message.HiveMessage) bool { return x.TempId == k })
		}
		fmt.Println(rexprint.Dim(fmt.Sprintf("Stored %d messages. Held before: %d, hold now: %d", toWriteLength, holdingMessagesLength, len(holdingMessages))))
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
			if time.Now().Unix() > flushRequestTime.Unix()+maxFlushGraceSeconds {
				fmt.Println(rexprint.ErrColor("Flush requested: There are pending messages to be stored in DB, but timeout was reached, will exit now"))
				*flushChan <- true
				flushChan = nil
			} else {
				fmt.Println(rexprint.Dim("Flush requested: There are pending messages to be stored in DB, will wait to flush..."))
			}
		}
	}
}

func OnHiveMessage(message *hive_message.HiveMessage) {
	lockForHolding.Lock()
	message.TempId = genHiveMessageId()
	message.Time = time.Now()
	holdingMessages = append(holdingMessages, message)
	lockForHolding.Unlock()
}

func genHiveRunId() string {
	id, _ := uuid.NewRandom()
	return id.String()
}

func genHiveMessageId() string {
	id, _ := uuid.NewRandom()
	return id.String()
}
