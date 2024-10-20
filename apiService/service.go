package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var conn *mongo.Database
var connect *amqp.Connection
var logger *log.Logger

const server = "amqp://guest:guest@localhost:5672/"

func main() {
	var err error
	//Create a database connection
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")

	conn, err = configDB(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()
	router.HandleFunc("/config", getConfig).Methods("GET")
	router.HandleFunc("/config", setConfig).Methods("POST")
	router.HandleFunc("/stream/{camera}", getVideo).Methods("GET", "OPTIONS")
	router.HandleFunc("/motion", getMotions).Methods("GET")

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	)

	log.Fatal(http.ListenAndServe(":8000", cors(router)))
	log.Print("ended")
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
