package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type videoRecord struct {
	Code   string
	Name   string
	Start  string
	End    string
	Reason string
}

func getMotions(w http.ResponseWriter, r *http.Request) {
	log.Print("get motion")
	collection := conn.Collection("video")
	cur, err := collection.Find(context.TODO(), bson.M{})
	failOnError(err, "Failed to get video records")

	var records []videoRecord
	for cur.Next(context.TODO()) {
		var record videoRecord
		err := cur.Decode(&record)
		failOnError(err, "Failed to decode record")
		records = append(records, record)
	}

	json.NewEncoder(w).Encode(records)

}
