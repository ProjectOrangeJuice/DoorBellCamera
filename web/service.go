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
	Current  bool
}

func main() {
	//Initiate templates
	var err error
	templates, err = template.ParseFiles("web/index.html", "web/templates/header.html", "web/templates/footer.html",
		"web/templates/side.html", "web/dash.html", "web/cameras.html", "web/edit.html")
	failOnError(err, "Failed to read templates")

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/live", live).Methods("GET")
	router.HandleFunc("/config", edit).Methods("GET")

	log.Fatal(http.ListenAndServe(":8001", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	data := pageContent{Title: "Dash", Side: makeSide("Dash")}
	decideTemplate(w, r, "dash", data)

}
func edit(w http.ResponseWriter, r *http.Request) {
	data := pageContent{Title: "Config edit", Side: makeSide("Config")}
	decideTemplate(w, r, "config", data)

}

func decideTemplate(w http.ResponseWriter, r *http.Request, templateName string, data2 pageContent) {
	c, err := r.Cookie("token")
	if err != nil {
		data := pageContent{Title: "Login"}
		templates.ExecuteTemplate(w, "index", data)
	} else {
		if c.Expires.After(time.Now()) {
			data := pageContent{Title: "Login"}
			templates.ExecuteTemplate(w, "index", data)
		} else {
			templates.ExecuteTemplate(w, templateName, data2)
		}

	}
}

func live(w http.ResponseWriter, r *http.Request) {
	data := pageContent{Title: "Dash", Side: makeSide("Live")}
	decideTemplate(w, r, "live", data)

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func makeSide(name string) []side {
	titles := []string{"Dash", "Live", "Config"}
	loc := []string{"/", "live", "config"}
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
