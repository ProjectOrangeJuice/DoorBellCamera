package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var conn *mongo.Database
var connect *amqp.Connection
var logger *log.Logger

const server = "amqp://guest:guest@localhost:5672/"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	var err error
	socketList = make(map[string][]*sharedWS)

	//Create a database connection
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")

	conn, err = configDB(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()
	router.HandleFunc("/config/{cam}", getConfig).Methods("GET")
	router.HandleFunc("/config/{cam}", setConfig).Methods("POST")
	router.HandleFunc("/stream/{camera}", getVideoShared).Methods("GET", "OPTIONS")
	router.HandleFunc("/mobilestream/{camera}", getCompressedVideo).Methods("GET", "OPTIONS")
	router.HandleFunc("/motionAlert/{camera}", getMotionAlert).Methods("GET", "OPTIONS")
	router.HandleFunc("/motion", getMotions).Methods("GET")
	router.HandleFunc("/motion/{start}/{end}", getMotionBetweenDates).Methods("GET")
	router.HandleFunc("/motion/{cam}/{start}/{end}", searchMotion).Methods("GET")
	router.HandleFunc("/from24", getMotion24).Methods("GET")
	router.HandleFunc("/from24/{cam}", get24).Methods("GET")
	router.HandleFunc("/motion/{code}", getMotion).Methods("GET")
	router.HandleFunc("/motion/{code}", deleteMotion).Methods("DELETE")

	router.HandleFunc("/information", getInformation).Methods("GET")

	router.HandleFunc("/clean", clearLastMonth).Methods("DELETE")

	router.HandleFunc("/profile/{cam}", createProfile).Methods("POST")
	router.HandleFunc("/profile/{cam}", deleteProfile).Methods("DELETE")

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
	)

	log.Fatal(http.ListenAndServe(":8000", handlers.CompressHandler(cors(router))))
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
