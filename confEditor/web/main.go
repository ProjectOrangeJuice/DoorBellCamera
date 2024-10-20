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
}

//This is for the websockets
var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/config/{service}", getConfig).Methods("GET", "OPTIONS")
	router.HandleFunc("/motion", getMotions).Methods("GET", "OPTIONS")
	router.HandleFunc("/motion/{code}", getMotion).Methods("GET", "OPTIONS")
	router.HandleFunc("/config/{service}", setConfig).Methods("POST")
	router.HandleFunc("/stream/{camera}", wsHandler)

	log.Fatal(http.ListenAndServe(":8000", router))
}

//All motion handler
func getMotions(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")

	rows, err := db.Query("select id,code,reason from video")
	failOnError(err, "prep failed")
	defer rows.Close()
	var full []MotionJSON
	for rows.Next() {
		var id int
		var code, reason string
		err = rows.Scan(&id, &code, &reason)
		failOnError(err, "Failed to get")
		body := MotionJSON{id, code, reason}
		full = append(full, body)

	}
	fmt.Printf("Bodycurrently is %+v\n", full)
	fmt.Printf("Single one is %v", full[1])
	//f1 := microJson{full}
	b, err := json.Marshal(full)
	fmt.Printf("Json bytes are %s\n", b)
	w.Write(b)

}

//Get the single data
func getMotion(w http.ResponseWriter, r *http.Request) {

}

//Socket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	go doStream(cam, ws)
}

//For the connection, get the stream and send it to the socket
func doStream(cam string, ws *websocket.Conn) {
	log.Printf("Setting up connection for %s", cam)
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"videoStream", // name
		"topic",       // type
		false,         // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		strings.Replace(cam, " ", ".", -1), // routing key
		"videoStream",                      // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var m Message
			err := json.Unmarshal(d.Body, &m)
			failOnError(err, "Json decode error")

			err = ws.WriteMessage(websocket.TextMessage, []byte(m.Image))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				ws.Close()
				return
			}

		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	log.Printf("Finished..")
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
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,        // queue name
		"config.test", // routing key
		"config",      // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
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
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ (get)")
	defer conn.Close()

	body := outMessage{"read", "test"}
	b, err := json.Marshal(body)

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"config", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a exchange")
	go func() {
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
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ (get)")
	defer conn.Close()
	body := outMessage{"update", config}
	b, err := json.Marshal(body)

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"config", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a exchange")

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
