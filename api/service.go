package main

import (
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

}
