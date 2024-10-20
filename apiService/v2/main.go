package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var databaseClient *mongo.Client

const server = "amqp://guest:guest@localhost:5672/"

var connect *amqp.Connection

func main() {
	// Setup database
	var err error
	databaseClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Printf("Failed to connect to database: %s", err)
		return
	}

	//rabbit
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	go streamEngine()

	router := mux.NewRouter()
	router.HandleFunc("/config", getConfig).Methods("GET")
	router.HandleFunc("/config", setConfig).Methods("POST")

	// Directly get video
	router.HandleFunc("/motion/hq/{code}", getHQVideo).Methods("GET")
	router.HandleFunc("/motion/lq/{code}", getLQVideo).Methods("GET")

	// Delete video
	router.HandleFunc("/motion/{code}", deleteMotion).Methods("DELETE")
	// videos
	router.HandleFunc("/videos/{last}", getNextSet).Methods("GET")
	router.HandleFunc("/videos/", getNextSet).Methods("GET")
	router.HandleFunc("/videos/{start}/{end}", getBetween).Methods("GET")

	// Live
	router.HandleFunc("/stream", getVideoShared).Methods("GET", "OPTIONS")

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
	)

	log.Fatal(http.ListenAndServe(":8000", handlers.CompressHandler(cors(router))))
}
