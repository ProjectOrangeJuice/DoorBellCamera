package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var connect amqp.Connection

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", index).Methods("GET")

	log.Fatal(http.ListenAndServe(":8001", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("web/index.html", "web/other.html")
	t.Execute(w, nil)
}
