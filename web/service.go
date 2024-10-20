package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var connect amqp.Connection

var templates *template.Template

type pageContent struct {
	Title string
	Side  []side
}

type side struct {
	Title    string
	Location string
	current  bool
}

func main() {
	//Initiate templates
	var err error
	templates, err = template.ParseFiles("web/index.html", "web/templates/header.html", "web/templates/footer.html",
		"web/templates/side.html", "web/dash.html")
	failOnError(err, "Failed to read templates")

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/o", other).Methods("GET")

	log.Fatal(http.ListenAndServe(":8001", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		data := pageContent{Title: "Login"}
		templates.ExecuteTemplate(w, "index", data)
	} else {
		if c.Expires.After(time.Now()) {
			data := pageContent{Title: "Login"}
			templates.ExecuteTemplate(w, "index", data)
		} else {
			data := pageContent{Title: "Dash", Side: makeSide("Dash")}
			templates.ExecuteTemplate(w, "dash", data)
		}

	}

}
func other(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "other", nil)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func makeSide(name string) []side {
	titles := []string{"Dash", "Live", "Config"}
	loc := []string{"dash", "live", "config"}
	sidebar := make([]side, len(titles))
	for index, key := range titles {
		if name == key {
			sidebar[index] = side{key, loc[index], true}
		} else {
			sidebar[index] = side{key, loc[index], false}
		}

	}
	return sidebar
}
