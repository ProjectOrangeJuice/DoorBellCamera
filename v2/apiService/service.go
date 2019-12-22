package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var connect *amqp.Connection

var logger *log.Logger

const server = "amqp://guest:guest@192.168.1.126:5672/"

func main() {
	setupLogging()

	var err error
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	router := mux.NewRouter()
	router.HandleFunc("/stream/{camera}", getVideo).Methods("GET", "OPTIONS")
	logger.Fatal(http.ListenAndServe(":8000", router))
	logger.Print("ended")
}

func setupLogging() {
	f, err := os.OpenFile("log/api.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	logger = log.New(f, "api-1 ", log.LstdFlags)
	mw := io.MultiWriter(os.Stdout, f)
	logger.SetOutput(mw)
}
