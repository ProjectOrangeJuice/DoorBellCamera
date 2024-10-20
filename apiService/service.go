package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
	router.Use(cors)
	router.HandleFunc("/config", getConfig).Methods("GET")
	router.HandleFunc("/config", setConfig).Methods("POST")
	router.HandleFunc("/stream/{camera}", getVideo).Methods("GET", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8000", router))
	log.Print("ended")
}

//To open the API to other sources (Browser ui) this will allow CORS
func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			h.ServeHTTP(w, r)
		})
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
