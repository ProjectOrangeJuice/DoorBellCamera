package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var connect amqp.Connection

var templates *template.Template

func main() {
	//Initiate templates
	var err error
	templates, err = template.ParseFiles("web/index.html", "web/other.html")
	failOnError(err, "Failed to read templates")

	router := mux.NewRouter()
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/o", other).Methods("GET")

	log.Fatal(http.ListenAndServe(":8001", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index", nil)
}
func other(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "other", nil)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
