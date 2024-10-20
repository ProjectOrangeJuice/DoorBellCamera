package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
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
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
	)

	log.Fatal(http.ListenAndServe(":8000", handlers.CompressHandler(cors(router))))
}
