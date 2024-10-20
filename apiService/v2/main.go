package main

import (
	"context"
	"log"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var databaseClient *mongo.Client

func main() {
	// Setup database
	var err error
	databaseClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Printf("Failed to connect to database: %s", err)
		return
	}

	router := mux.NewRouter()
	router.HandleFunc("/config", getConfig).Methods("GET")
	router.HandleFunc("/config", setConfig).Methods("POST")

}
