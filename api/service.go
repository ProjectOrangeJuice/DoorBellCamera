package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var connect amqp.Connection

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/login", signin).Methods("POST")
	//Everything with /s/.. requires you to login
	sec := router.PathPrefix("/s").Subrouter()
	sec.Use(auth)
	sec.HandleFunc("/refresh", refresh).Methods("GET")
	sec.HandleFunc("/motion", allMotion).Methods("GET")
	sec.HandleFunc("/motion/{code}", getMotion).Methods("DELETE", "GET", "OPTIONS")
	log.Fatal(http.ListenAndServe(":8000", router))
}
