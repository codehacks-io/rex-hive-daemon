package message_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"rex-daemon/swarm_message"
	"sync"
)

var (
	messages []*swarm_message.SwarmMessage
	lock     sync.Mutex
)

var didStartup = false

func Run() {
	if didStartup {
		return
	}
	didStartup = true
	mongoAddTestData()
}

func mongoAddTestData() {
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

	coll := client.Database("swarm-chan").Collection("process-messages")
	result, err := coll.InsertOne(context.TODO(), bson.D{{"data", "yeah, it does work"}})

	if err != nil {
		panic(err)
	}

	fmt.Println("success")

	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", jsonData)
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
	lock.Lock()
	messages = append(messages, message)
	lock.Unlock()
	fmt.Println(fmt.Sprintf("(%d) Received msg from %d: %+v", len(messages), (*message).Pid, *message))
}
