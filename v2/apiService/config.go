package main

import (
	"context"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type cameraSettings struct {
	Name       string
	Connection string
	FPS        int
	Area       [][]int
	Amount     []int
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
