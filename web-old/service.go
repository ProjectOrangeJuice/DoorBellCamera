package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

const (
	host   = "sqldata"
	port   = 3306
	user   = "door"
	passdb = "pass"
	dbname = "doorservice"
)

var connect amqp.Connection

var templates *template.Template

type pageContent struct {
	Title string
	Side  []side
	Cam   []string
	Code  string
}

type side struct {
	Title    string
	Location string
	Current  bool
}

func main() {
	host, err := net.LookupIP(host)
	if err != nil {
		fmt.Println("Unknown host")
	} else {
		fmt.Println("IP address: ", host)
	}
	//Initiate templates

	templates, err = template.ParseFiles("web/index.html", "web/templates/header.html", "web/templates/footer.html",
		"web/templates/side.html", "web/dash.html", "web/cameras.html", "web/edit.html",
		"web/edit/cam.html", "web/inspect.html", "web/templates/alert.html", "web/edit/user.html")
	failOnError(err, "Failed to read templates")

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/live", live).Methods("GET")
	router.HandleFunc("/config", edit).Methods("GET")
	router.HandleFunc("/add/{name}", addCam).Methods("POST")
	router.HandleFunc("/inspect/{code}", inspect).Methods("GET")

	log.Fatal(http.ListenAndServe(":8001", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	data := pageContent{Title: "Dash", Side: makeSide("Dash")}
	decideTemplate(w, r, "dash", data)

}
func edit(w http.ResponseWriter, r *http.Request) {
	data := pageContent{Title: "Config edit", Side: makeSide("Config"), Cam: getCams()}
	decideTemplate(w, r, "config", data)

}

func inspect(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	data := pageContent{Title: "Inspect", Side: makeSide("none"), Code: params["code"]}
	decideTemplate(w, r, "inspect", data)

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
	data := pageContent{Title: "Dash", Side: makeSide("Live"), Cam: getCams()}
	decideTemplate(w, r, "live", data)

}

func addCam(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("token")
	if err == nil {
		params := mux.Vars(r)
		sqlInfo := fmt.Sprintf("%s:%s@tcp(%s)/%s",
			user, passdb, host, dbname)

		db, err := sql.Open("mysql", sqlInfo)
		failOnError(err, "Database opening error")
		defer db.Close()
		sqlStatement := `INSERT INTO cameras(name) VALUES (?)`
		stmt, err := db.Prepare(sqlStatement)
		failOnError(err, "making statement failed")

		_, err = stmt.Exec(params["name"])
		failOnError(err, "Failed to insert")
	} else {
		log.Printf("Cam error %s", err)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getCams() []string {
	sqlInfo := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		user, passdb, host, dbname)

	db, err := sql.Open("mysql", sqlInfo)
	failOnError(err, "Database opening error")
	failOnError(err, "Database opening error")
	defer db.Close()
	sqlStatement := `SELECT name FROM cameras`
	row, err := db.Query(sqlStatement)
	failOnError(err, "Query error")
	defer row.Close()
	var cams []string
	for row.Next() {
		var cam string
		err = row.Scan(&cam)
		failOnError(err, "Failed to scan")
		cams = append(cams, cam)
	}
	log.Printf("cams.. %s", cams)
	return cams
}

func makeSide(name string) []side {
	titles := []string{"Dash", "Live", "Config"}
	loc := []string{"/", "/live", "/config"}
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
