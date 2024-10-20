package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type videoRecord struct {
	Code  string
	Size  int64
	Image string
	Stamp int64
}

// Get next 5 videos after the date given. If no date given, from the last record
func getNextSet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	conn := databaseClient.Database("doorbell")
	collection := conn.Collection("video")
	findOptions := options.Find()
	// Sort by
	findOptions.SetSort(bson.D{{"start", -1}})
	findOptions.SetLimit(5)

	// Skip the ones we've seen
	last, ok := params["last"]
	var filter bson.M
	if ok {
		n, err := strconv.ParseInt(last, 10, 64)
		if err != nil {
			log.Printf("Failed to convert string(%s) to int64 - %v", last, err)
		}
		filter = bson.M{
			"stamp": bson.M{"$gt": n},
		}
	}

	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		log.Printf("Failed to get video records %v", err)
	}

	var records []videoRecord
	for cur.Next(context.TODO()) {
		var record videoRecord
		cur.Decode(&record)
		records = append(records, record)
	}

	json.NewEncoder(w).Encode(records)
}

// Get all the videos between two dates
func getBetween(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	conn := databaseClient.Database("doorbell")
	collection := conn.Collection("video")
	findOptions := options.Find()
	// Sort by
	findOptions.SetSort(bson.D{{"start", -1}})
	findOptions.SetLimit(5)

	var filter bson.M

	n, err := strconv.ParseInt(params["first"], 10, 64)
	if err != nil {
		log.Printf("Failed to convert string(%s) to int64 - %v", params["first"], err)
	}
	n2, err := strconv.ParseInt(params["last"], 10, 64)
	if err != nil {
		log.Printf("Failed to convert string(%s) to int64 - %v", params["last"], err)
	}
	filter = bson.M{
		"stamp": bson.M{"$gt": n, "$lt": n2},
	}

	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		log.Printf("Failed to get video records %v", err)
	}

	var records []videoRecord
	for cur.Next(context.TODO()) {
		var record videoRecord
		cur.Decode(&record)
		records = append(records, record)
	}

	json.NewEncoder(w).Encode(records)
}

// Delete motion

func deleteMotion(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("delete video for %s, requested %s", r.RemoteAddr, params["code"])
	conn := databaseClient.Database("doorbell")
	collection := conn.Collection("video")
	filter := bson.M{"code": params["code"]}
	collection.DeleteOne(context.TODO(), filter)
	err := os.Remove(fmt.Sprintf("%s/%s.mp4", videoLoc, params["code"]))
	if err != nil {
		log.Printf("Failed to delete %s because %v", params["code"], err)
	}

}

// Get motion
const videoLoc = "../../storage/videos"

func getHQVideo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	code := strings.ReplaceAll(params["code"], "+", " ")
	log.Printf("Get video for %s, requested %s", r.RemoteAddr, fmt.Sprintf("%s/%s.mp4", videoLoc, code))
	http.ServeFile(w, r, fmt.Sprintf("%s/%s.mp4", videoLoc, code))

}

func getLQVideo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	code := strings.ReplaceAll(params["code"], "+", " ")
	log.Printf("Get video for %s, requested %s", r.RemoteAddr, fmt.Sprintf("%s/s/%s.mp4", videoLoc, code))
	http.ServeFile(w, r, fmt.Sprintf("%s/s/%s.mp4", videoLoc, code))

}
