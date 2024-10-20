package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cameraSettings struct {
	Name       string
	Connection string
	FPS        int
	Area       [][]int
	Amount     []int
	Threshold  []int
	MinCount   []int
	Motion     bool
}

var conn *mongo.Database
var setting *cameraSettings

func getSettings() {
	var err error
	conn, err = configDB(context.Background())
	failOnError(err, "Couldn't connect to database")
	grab()
	go func() {
		for {
			time.Sleep(time.Second * 60)
			grab()
		}
	}()
}

func grab() {
	collection := conn.Collection("settings")
	filter := bson.M{"_id": 0}
	doc := collection.FindOne(context.TODO(), filter)
	err := doc.Decode(&setting)
	if err != nil {
		//Content not found
	}
	log.Printf("%v", setting)
}

func configDB(ctx context.Context) (*mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s", "localhost")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to mongo: %v", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("mongo client couldn't connect with background context: %v", err)
	}
	todoDB := client.Database("camera")
	return todoDB, nil
}
