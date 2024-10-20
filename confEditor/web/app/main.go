package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/streadway/amqp"
)

var server = "amqp://guest:guest@192.168.1.126:30188/"

//const DBName string = "/shared/motions.db"
const DBName string = "/mnt/shared/motion/motions.db"

type outMessage struct {
	Task  string
	Inner string
}
type microJson struct {
	Motions []MotionJSON
}

//Message is the json format
type Message struct {
	Image string
	Time  string
	Code  string
	Count int
	Name  string
}

type MotionJSON struct {
	Id     int
	Code   string
	Reason string
	Name   string
}

//This is for the websockets
var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	var err error
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	router := mux.NewRouter()
	router.HandleFunc("/config/{service}", getConfig).Methods("GET", "OPTIONS")
	sec := router.PathPrefix("/s").Subrouter()
	sec.Use(auth)
	sec.HandleFunc("/motion", getMotions).Methods("GET", "OPTIONS")
	router.HandleFunc("/motion/{code}", getMotion).Methods("GET", "OPTIONS")
	router.HandleFunc("/delete/{code}", delMotion).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/config/{service}", setConfig).Methods("POST")
	router.HandleFunc("/stream/{camera}", wsHandler)
	router.HandleFunc("/streamMotion", wsHandlerMotion)
	router.HandleFunc("/streamDoor", wsHandlerDoor)
	router.HandleFunc("/login", Signin).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}

//All motion handler
func getMotions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")

	rows, err := db.Query("select id,code,reason,name from video")
	failOnError(err, "prep failed")
	defer rows.Close()
	var full []MotionJSON
	for rows.Next() {
		var id int
		var code, reason, name string
		err = rows.Scan(&id, &code, &reason, &name)
		failOnError(err, "Failed to get")
		body := MotionJSON{id, code, reason, name}
		full = append(full, body)

	}

	//f1 := microJson{full}
	b, err := json.Marshal(full)
	fmt.Printf("Json bytes are %s\n", b)
	w.Write(b)

}

//del motion handler
func delMotion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Access-Control-Allow-Origin, access-control-allow-headers")

	print("Deleting motion")
	params := mux.Vars(r)

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	tx, err := db.Begin()
	failOnError(err, "Failed to begin on record")
	stmt, err := tx.Prepare("DELETE FROM video WHERE code=?")
	failOnError(err, "Record sql prep failed")
	defer stmt.Close()
	_, err = stmt.Exec(params["code"])
	failOnError(err, "Record could not insert")
	tx.Commit()
	w.Write([]byte("Okay"))
}

//Get the single data
func getMotion(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	http.ServeFile(w, r, fmt.Sprintf("/mnt/shared/motion/videos/%s.mp4", params["code"]))
}

//Socket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	go DoStream(cam, ws)
}

//Socket handler
func wsHandlerMotion(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	go doMotionCheck(cam, ws)
}

//Socket handler
func wsHandlerDoor(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	go doDoorWatch(cam, ws)
}

//GET for getconfig
func getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	params := mux.Vars(r)
	msg := getCommand(params["service"])
	body := outMessage{params["service"], msg}
	b, err := json.Marshal(body)
	failOnError(err, "failed to create json to send")
	w.Write(b)

}

//POST for setconfig
func setConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	params := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	failOnError(err, "failed to read body")
	setCommand(params["service"], string(body))
}

//Listens for the return of the config file
func goListen(rch chan string, arg string) {
	msgs, ch := listenToExchange("config", "config.test")
	defer ch.Close()
	rch <- "ready"
	forever := make(chan bool)

	go func() {
		for d := range msgs {

			rch <- decodeMsg(d.Body, arg)
		}
	}()

	<-forever
	fmt.Print("over 2")
}

func decodeMsg(msg []byte, arg string) string {
	arg = strings.Replace(arg, ".", "-", -1)
	var m outMessage
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")
	return m.Inner
}

//Get the config file
func getCommand(arg string) string {
	returnCh := make(chan string)
	go goListen(returnCh, arg)
	if m := <-returnCh; m != "ready" {
		log.Panicf("Something went wrong when waiting for ready")
	}

	_, ch := listenToExchange("config", arg)
	body := outMessage{"read", "test"}
	b, err := json.Marshal(body)
	failOnError(err, "Failed to make json")
	defer ch.Close()
	go func() {
		err := ch.Publish(
			"config", // exchange
			arg,      // routing key
			false,    // mandatory
			false,    // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(b),
			})

		failOnError(err, "Failed to publish a message")
	}()

	if m := <-returnCh; m == "error" {
		log.Print("Something went wrong when returning")
	} else {
		return m
	}
	return "error"
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//Set the config file
func setCommand(arg string, config string) {

	body := outMessage{"update", config}
	b, err := json.Marshal(body)
	_, ch := listenToExchange("config", arg)
	defer ch.Close()

	err = ch.Publish(
		"config", // exchange
		arg,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(b),
		})

	failOnError(err, "Failed to publish a message")
	log.Print("Sent new config")

}
