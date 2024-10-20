package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"google.golang.org/api/option"
)

type config struct {
	ServerAddress string
	ServerPort    int
	Camereas      []cameras
}
type cameras struct {
	Name string
}

//Message is the json format
type Message struct {
	Image     string
	Time      string
	Code      string
	Count     int
	Name      string
	Reason    string
	Locations string
}

type hold struct {
	Code          string
	Point         float64
	Count         int
	PreviousAlert time.Time
	FalseCount    int
}

const (
	host   = "localhost"
	port   = 5432
	user   = "door"
	passdb = "door"
	dbname = "doorservice"
)

func readConfig() {

}

var connect *amqp.Connection

const server = "amqp://guest:guest@192.168.1.126:30188/"

func main() {

	var err error
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	msgs, ch := listenToFanout("motion")
	var phold hold
	forever := make(chan bool)
	go func() {
		defer ch.Close()
		for d := range msgs {
			decodeMessage(d.Body, &phold)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func decodeMessage(d []byte, held *hold) {
	var m Message
	err := json.Unmarshal(d, &m)
	failOnError(err, "Json decode error")
	decideFate(m, held)
}

func decideFate(msg Message, held *hold) {
	checkFrame := false
	if !held.PreviousAlert.IsZero() {

		diff := held.PreviousAlert.Sub(time.Now())
		//log.Printf("Time %f", diff.Minutes())
		if diff.Minutes() < -5.0 {
			if held.Code != msg.Code {
				//Don't check
				checkFrame = true
			} else {
				held.PreviousAlert = time.Now()
			}
		}
	} else {
		checkFrame = true
	}
	//checkFrame = true //ignore the time out
	if checkFrame {
		//Decode the points
		msg.Locations = strings.Replace(msg.Locations, "'", "\"", -1)
		var locPoints []map[string]interface{}
		err := json.Unmarshal([]byte(msg.Locations), &locPoints)
		failOnError(err, "Json failed on locpoints")
		down := false
		for _, loc := range locPoints {
			v1, err := strconv.ParseFloat(fmt.Sprintf("%v", loc["m10"]), 64)
			v2, err := strconv.ParseFloat(fmt.Sprintf("%v", loc["m00"]), 64)
			failOnError(err, "failed to convert v2")
			//v3, _ := strconv.ParseFloat(fmt.Sprintf("%v", loc["m01"]), 64)
			//mY := v3 / v2
			mX := v1 / v2
			log.Printf("I worked the value as %f", mX)
			if held.Code != msg.Code {
				log.Print("Reset.")
				held.Count = 0
				held.FalseCount = 0
				held.Point = 0
				if held.Point < 5 {
					held.Point = mX
				} else {

					if mX < held.Point {
						held.Point = mX
					}
				}
			} else {
				if held.Point < 5 {
					held.Point = mX
				}
				if mX < held.Point {
					down = true
					held.Point = mX
				}
			}
		}
		if len(locPoints) > 0 {
			log.Printf("Smallest point is %f", held.Point)
			held.Code = msg.Code
			if down {
				log.Print("For this frame i would agree that it's likely to come from the gate.")
				held.Count++
			} else {
				log.Print("Added to false count")
				held.FalseCount++
			}

			if held.FalseCount > 5 {
				log.Print("False count went over. Decided to delay alerts")
				held.PreviousAlert = time.Now()
			}

			if held.Count > 5 {
				log.Print("I would send a notification now!")
				go sendNotification()
				held.PreviousAlert = time.Now()
			}
		}

	}

}

func sendNotification() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, passdb, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	failOnError(err, "Database opening error")
	defer db.Close()
	sqlStatement := `SELECT key FROM keys`
	row, err := db.Query(sqlStatement)
	failOnError(err, "Query error")
	defer row.Close()
	var keys []string
	for row.Next() {
		var key string
		err = row.Scan(&key)
		failOnError(err, "Failed to scan")
		keys = append(keys, key)
	}
	log.Printf("Keys! %s", keys)

	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("There was an error on the first step. %s", err)
	}

	// Obtain a messaging.Client from the App.
	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	// See documentation on defining a message payload.
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: "Door service",
			Body:  "I think someone is there!",
		},

		Tokens: keys,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.SendMulticast(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)

}
