package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/streadway/amqp"
	"gocv.io/x/gocv"
)

type outMessage struct {
	Code string
}

//DBName is the database file name
const DBName string = "/shared/motions.db"
const captureFolder string = "/shared/capture"
const videoFolder string = "/shared/videos"
const configLocation string = "/shared/config.txt"

var server = ""

func main() {

	if dbExists(DBName) {
		readyListen()
	} else {
		makeDatabase()
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

func makeDatabase() {
	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Error on database creation")
	defer db.Close()

	sqlStmt := `CREATE TABLE 'motion' (
		'motionId'	INTEGER PRIMARY KEY AUTOINCREMENT,
		'motionCode'	TEXT,
		'location'	TEXT,
		'time'	TEXT,
		'reason' TEXT
	);`

	_, err = db.Exec(sqlStmt)

	sqlStmt = `CREATE TABLE 'video' (
		'id'	INTEGER PRIMARY KEY AUTOINCREMENT,
		'code'	TEXT,
		'startTime'	TEXT,
		'endTime'	TEXT,
		'reason' TEXT
	);`

	_, err = db.Exec(sqlStmt)
	failOnError(err, "Error creating table")
	readyListen()
}

func readyListen() {
	file, err := os.Open(configLocation)
	failOnError(err, "Couldn't open config")
	defer file.Close()
	serverb, _ := ioutil.ReadAll(file)
	server = strings.TrimSpace(string(serverb))
	failOnError(err, "Failed to read config")
	conn, err := amqp.Dial(server)
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
	var m outMessage
	var startTime string
	var endTime string
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")
	video, err := gocv.VideoWriterFile(fmt.Sprintf("%s/%s", videoFolder, m.Code), "avc1", 5.0, 1280, 720, true)

	//aw, err := mjpeg.New(fmt.Sprintf("%s/%s", videoFolder, m.Code), 1280, 720, 10)
	failOnError(err, "Setting up video")

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")

	rows, err := db.Query("select location,time,reason from motion where motionCode = ?", m.Code)
	failOnError(err, "prep failed")
	defer rows.Close()

	var fr []string
	var high = 0
	for rows.Next() {
		var location string
		var time string
		var reason string
		err = rows.Scan(&location, &time, &reason)
		failOnError(err, "Failed to get")
		s := strings.Split(reason, "-")
		total := 0
		for _, val := range s {
			t, _ := strconv.Atoi(val)
			total += t
		}
		if total > high {
			total = high
		}
		if startTime == "" {
			startTime = time
		} else {
			endTime = time
		}
		video.Write(gocv.IMRead(fmt.Sprintf("%s", location), gocv.IMReadAnyColor))
		//data, err := ioutil.ReadFile(fmt.Sprintf("%s", location))
		//failOnError(err, "Failed reading image")
		//err = aw.AddFrame(data)
		//failOnError(err, "failed to add frame")
		//fr = append(fr, fmt.Sprintf("%s", location))
		//exec.Command("ffmpeg", "-r 15", fmt.Sprintf("-i %s*", m.Code), "-vcodec libx264", "-crf 20", "-pix_fmt yuv420p", fmt.Sprintf("%s/%s.mp4", videoFolder, m.Code))

	}
	//err = aw.Close()
	//failOnError(err, "Error closing")
	video.Close()
	for _, elem := range fr {
		err = os.Remove(elem)
		failOnError(err, "Failed to remove image")
	}

	log.Printf("Start time %s and end time %s", startTime, endTime)
	addToDatabase(m.Code, startTime, endTime, high)
}

func addToDatabase(code string, start string, end string, high int) {
	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	defer db.Close()
	tx, err := db.Begin()
	failOnError(err, "Failed to begin on record")
	stmt, err := tx.Prepare("insert into video(code, startTime,endTime , reason) values(?,?,?,?)")
	failOnError(err, "Record sql prep failed")
	defer stmt.Close()
	_, err = stmt.Exec(code, start, end, strconv.Itoa(high))
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
