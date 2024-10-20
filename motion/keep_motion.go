package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/streadway/amqp"
)

//DBName is the database file name
const DBName string = "./motions.db"

//CaptureLocation is the location of the capture folder
const CaptureLocation string = "capture"

//Message is the JSON message format
type Message struct {
	Image string
	Time  string
	Code  string
	Count int
	Name  string
}

type OutMessage struct {
	Code string
	name string
}

type cameraStructure struct {
	prev        string
	notified    string
	ignoreTimer bool
}

var camera map[string]cameraStructure
var timer time.Timer

func main() {

	if dbExists(DBName) {
		readyAndListen()
	} else {
		createDatabase()
	}

}

func readyAndListen() {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.99.100:31693/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"motionAlert", // name
		false,         // durable
		false,         // delete when usused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")

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

	createTimer()
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			decodeMessage(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func createTimer() {
	timer := time.NewTimer(5 * time.Second)
	go func() {
		<-timer.C
		log.Printf("Timer is over")
		for k, v := range camera {

			if v.prev != "" && v.prev != v.notified && !v.ignoreTimer {
				v.notified = v.prev
				notifyQueue(v.prev, k)
				v.prev = ""
			}
			v.ignoreTimer = false
		}

		createTimer()
	}()
}

func decodeMessage(d []byte) {
	var m Message
	err := json.Unmarshal(d, &m)
	failOnError(err, "Json decode error")
	if _, ok := camera[m.Name]; !ok {
		camera[m.Name] = cameraStructure{"", "Nothing", true}
	}

	storeImage(m)

}

func recordDb(msg Message, loc string) {

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	defer db.Close()
	tx, err := db.Begin()
	failOnError(err, "Failed to begin on record")
	stmt, err := tx.Prepare("insert into motion(motionCode, location,time) values(?,?,?)")
	failOnError(err, "Record sql prep failed")
	defer stmt.Close()
	_, err = stmt.Exec(msg.Code, loc, msg.Time)
	failOnError(err, "Record could not insert")
	tx.Commit()
	tc := camera[msg.Name]
	if tc.prev != "" && tc.prev != msg.Code {
		log.Printf("End of prev code")
		notifyQueue(tc.prev, msg.Name)
		tc.ignoreTimer = true
		tc.prev = msg.Code
	} else if tc.prev == "" {
		tc.prev = msg.Code
		tc.ignoreTimer = true

	}
}

func notifyQueue(code string, name string) {
	log.Printf("NOTIFY")
	body := OutMessage{code, name}
	b, err := json.Marshal(body)
	conn, err := amqp.Dial("amqp://guest:guest@192.168.99.100:31693/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"imageToVideo", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})
	failOnError(err, "Failed to publish a message")

}

func storeImage(msg Message) {
	//convert base64
	bImage, err := base64.StdEncoding.DecodeString(msg.Image)
	failOnError(err, "Base64 error")
	location := fmt.Sprintf("%s/%s-%v.jpg", CaptureLocation, msg.Code, msg.Count)
	err2 := ioutil.WriteFile(location, bImage, 0644)
	failOnError(err2, "Error writing image")
	log.Printf("Stored image %s", location)
	recordDb(msg, location)
}

func createDatabase() {
	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Error on database creation")
	defer db.Close()

	sqlStmt := `CREATE TABLE 'motion' (
		'motionId'	INTEGER PRIMARY KEY AUTOINCREMENT,
		'motionCode'	TEXT,
		'location'	TEXT,
		'time'	TEXT
	);`

	_, err = db.Exec(sqlStmt)
	failOnError(err, "Error creating table")

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func dbExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
