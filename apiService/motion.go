package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

type videoRecord struct {
	Code      string
	Name      string
	Start     string
	End       string
	Reason    string
	Thumbnail string
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

func getMotion24(w http.ResponseWriter, r *http.Request) {
	log.Print("get motion")
	t := time.Now()
	t = t.Add(time.Duration(-24) * time.Hour)
	stamp := t.Unix()
	strconv.FormatInt(stamp, 10)
	collection := conn.Collection("video")
	filter := bson.M{
		"start": bson.M{"$gt": strconv.FormatInt(stamp, 10)},
	}
	cur, err := collection.Find(context.TODO(), filter)
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

const videoLoc = "/storeDrive/videos"

func getMotion(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("Get video for %s, requested %s", r.RemoteAddr, params["code"])
	http.ServeFile(w, r, fmt.Sprintf("%s/%s.mp4", videoLoc, params["code"]))

}

func deleteMotion(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("delete video for %s, requested %s", r.RemoteAddr, params["code"])

	collection := conn.Collection("video")
	filter := bson.M{"code": params["code"]}
	collection.DeleteOne(context.TODO(), filter)
	err := os.Remove(fmt.Sprintf("%s/%s.mp4", videoLoc, params["code"]))
	failOnError(err, "Failed to delete")

}
