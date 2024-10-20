package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
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
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	collection := conn.Collection("settings")
	filter := bson.M{"_id": 0}
	doc := collection.FindOne(context.TODO(), filter)
	var settings cameraSettings
	err := doc.Decode(&settings)
	if err != nil {
		//Content not found
	}
	log.Printf("%s", settings)

}

func setConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var settings cameraSettings
	err := decoder.Decode(&settings)
	failOnError(err, "decode new settings")

	collection := conn.Collection("settings")
	filter := bson.M{"_id": 0}
	collection.FindOneAndReplace(context.TODO(), filter, settings, options.FindOneAndReplace().SetUpsert(true))
}
