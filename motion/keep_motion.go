package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/streadway/amqp"
)

//DBName is the database file name
const DBName string = "./motions.db"

//Message is the JSON message format
type Message struct {
	Image string
	Time  string
	Code  string
}

func main() {

	if dbExists(DBName) {
		readyAndListen()
	} else {
		createDatabase()
	}

}

func readyAndListen() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			decodeMessage(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func decodeMessage(d []byte) {

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
