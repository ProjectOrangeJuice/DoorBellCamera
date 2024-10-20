package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var connect *amqp.Connection
const server =  "amqp://guest:guest@192.168.1.126:30188/"
func main() {
	var err error
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")

	router := mux.NewRouter()
	router.HandleFunc("/login", signin).Methods("POST", "OPTIONS")
	//Everything with /s/.. requires you to login
	sec := router.PathPrefix("/s").Subrouter()
	sec.Use(auth)
	sec.HandleFunc("/refresh", refresh).Methods("GET", "OPTIONS")
	sec.HandleFunc("/motion", allMotion).Methods("GET", "OPTIONS")
	sec.HandleFunc("/motion/{code}", getMotion).Methods("DELETE", "GET", "OPTIONS")
	sec.HandleFunc("/stream/{code}", getVideo).Methods("GET", "OPTIONS")
	sec.HandleFunc("/service/motion", getMotionWatch).Methods("GET", "OPTIONS")
	sec.HandleFunc("/service/door", getDoor).Methods("GET", "OPTIONS")
	sec.HandleFunc("/config/{service}", setConfig).Methods("POST", "OPTIONS")
	sec.HandleFunc("/config/{service}", getConfig).Methods("GET", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8000", router))
}
