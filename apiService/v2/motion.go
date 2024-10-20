package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// Get motions ( pagination )

// Delete motion

func deleteMotion(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("delete video for %s, requested %s", r.RemoteAddr, params["code"])
	conn := databaseClient.Database("doorbell")
	collection := conn.Collection("video")
	filter := bson.M{"code": params["code"]}
	collection.DeleteOne(context.TODO(), filter)
	err := os.Remove(fmt.Sprintf("%s/%s.mp4", videoLoc, params["code"]))
	log.Printf("Failed to delete %s because %v", params["code"], err)

}

// Get motion
const videoLoc = "../storage/videos"

func getHQVideo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("Get video for %s, requested %s", r.RemoteAddr, params["code"])
	http.ServeFile(w, r, fmt.Sprintf("%s/%s.mp4", videoLoc, params["code"]))

}

func getLQVideo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("Get video for %s, requested %s", r.RemoteAddr, params["code"])
	http.ServeFile(w, r, fmt.Sprintf("%s/s/%s.mp4", videoLoc, params["code"]))

}
