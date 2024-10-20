package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/icza/mjpeg"
	_ "github.com/mattn/go-sqlite3"
	"github.com/streadway/amqp"
)

type OutMessage struct {
	Code string
}

//DBName is the database file name
const DBName string = "../motions.db"

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@192.168.99.100:31693/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"imageToVideo", // name
		false,          // durable
		false,          // delete when usused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
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
			log.Printf("Received a message: %s", d.Body)
			convert(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func convert(msg []byte) {
	var m OutMessage
	var startTime string
	var endTime string
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")

	aw, err := mjpeg.New(fmt.Sprintf("videos/%s", m.Code), 1280, 720, 10)
	failOnError(err, "Setting up video")

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")

	rows, err := db.Query("select location,time from motion where motionCode = ?", m.Code)
	failOnError(err, "prep failed")
	defer rows.Close()
	root := "/home/oharris/Documents/cameraProject/motion"
	for rows.Next() {
		var location string
		var time string
		err = rows.Scan(&location, &time)
		failOnError(err, "Failed to get")

		if startTime == "" {
			startTime = time
		} else {
			endTime = time
		}

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", root, location))
		failOnError(err, "Failed reading image")
		err = aw.AddFrame(data)
		failOnError(err, "failed to add frame")
		err = os.Remove(fmt.Sprintf("%s/%s", root, location))
		failOnError(err, "Failed to remove image")

	}
	err = aw.Close()
	failOnError(err, "Error closing")
	log.Printf("Start time %s and end time %s", startTime, endTime)
	addToDatabase(m.Code, startTime, endTime)
}

func addToDatabase(code string, start string, end string) {
	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	defer db.Close()
	tx, err := db.Begin()
	failOnError(err, "Failed to begin on record")
	stmt, err := tx.Prepare("insert into video(code, startTime,endTime) values(?,?,?)")
	failOnError(err, "Record sql prep failed")
	defer stmt.Close()
	_, err = stmt.Exec(code, start, end)
	failOnError(err, "Record could not insert")
	tx.Commit()
	log.Printf("Saved to db")

	_, err = db.Exec("DELETE FROM motion WHERE motionCode=?", code)
	failOnError(err, "Couldn't delete motion records")

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
