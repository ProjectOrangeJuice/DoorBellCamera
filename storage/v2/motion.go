package main

import (
	"context"
	fmt "fmt"
	"io/ioutil"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type settings struct {
	Name               string
	Connection         string
	FPS                int
	MinCount           int
	Motion             bool
	Blur               int
	Debug              bool
	BufferBefore       int
	BufferAfter        int
	NoMoveRefreshCount int
	Zones              []zone
}
type zone struct {
	X1          int
	Y1          int
	X2          int
	Y2          int
	Threshold   int
	BoxJump     int
	SmallIgnore int
	Area        int
}

func recvMotionImg(buf chan *Buffer) {
	databaseClient, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	conn := databaseClient.Database("doorbell")
	db := conn.Collection("setting")
	filter := bson.M{"_id": 0}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	doc := db.FindOne(ctx, filter)
	cancel()
	var s settings
	doc.Decode(&s)

	db = conn.Collection("motion")
	for msg := range buf {
		fmt.Printf("About to write image %s\n", msg.Code)
		// Store the image
		location := fmt.Sprintf("%s/%s-%d.jpg", imageLocation, msg.Code, msg.Count)
		err := ioutil.WriteFile(location, msg.Image, 0644)
		if err != nil {
			log.Fatalf("Failed to write image %s", err)
		}
		fmt.Println("Written image")
	}

}
