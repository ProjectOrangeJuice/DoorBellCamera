package main

import (
	"context"
	fmt "fmt"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cameraStructure struct {
	prev        string
	notified    string
	ignoreTimer bool
}

var timer time.Timer
var databaseClient *mongo.Client

const imageLocation string = "../images"
const fullVideoLocation string = "../videos"
const smallVideoLocation string = "../videos/s"

func main() {
	setupDatabase()
	//go http.ListenAndServe("localhost:8080", nil)
	mstream := make(chan *Buffer)
	go recvMotionImg(mstream)
	buf := &Buffer{}
	c := setupRabbit()
	for input := range c {
		fmt.Println("Got message")
		input.Ack(true)
		// Pass the message to our video creator
		err := proto.Unmarshal(input.Body, buf)
		if err != nil {
			fmt.Printf("Failed to unmash %v\n", err)
			continue
		}
		mstream <- buf // Must block before getting next image
		// Otherwise we will change the buffer while its working
	}
}

func setupDatabase() {
	var err error
	databaseClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Printf("Failed to connect to database: %s", err)
		return
	}
}
