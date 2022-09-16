package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	res, err := http.Get("http://169.254.169.254/latest/dynamic/instance-identity/document")
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	b, _ := io.ReadAll(res.Body)
	fmt.Println(string(b))

	data := WakaMessage{}
	data.Title = "Demo two xxx"
	data.Index = 123
	doc := &AwsEc2IdentityDoc{}
	_ = json.Unmarshal(b, doc)
	data.AwsData = *doc
	testMongo(&data)
}

func testMongo(data interface{}) {
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

	res, err := coll.InsertOne(context.TODO(), data)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("inserted document with ID %v\n", res.InsertedID)
}

type WakaMessage struct {
	Index   int
	Title   string
	AwsData AwsEc2IdentityDoc
}

type AwsEc2IdentityDoc struct {
	AccountId               string      `json:"accountId"`
	Architecture            string      `json:"architecture"`
	AvailabilityZone        string      `json:"availabilityZone"`
	BillingProducts         interface{} `json:"billingProducts"`
	DevpayProductCodes      interface{} `json:"devpayProductCodes"`
	MarketplaceProductCodes interface{} `json:"marketplaceProductCodes"`
	ImageId                 string      `json:"imageId"`
	InstanceId              string      `json:"instanceId"`
	InstanceType            string      `json:"instanceType"`
	KernelId                interface{} `json:"kernelId"`
	PendingTime             time.Time   `json:"pendingTime"`
	PrivateIp               string      `json:"privateIp"`
	RamdiskId               interface{} `json:"ramdiskId"`
	Region                  string      `json:"region"`
	Version                 string      `json:"version"`
}
