package main

import (
	"context"
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
	Triggered     bool
}

const (
	host   = "192.168.1.135"
	port   = 5432
	user   = "door"
	passdb = "door"
	dbname = "doorservice"
)

func readConfig() {

}

var connect *amqp.Connection

const server = "amqp://guest:guest@192.168.1.135:5672/"

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

	// if it's the same motion. extend the timeout
	if held.Code == msg.Code && held.Triggered {
		held.PreviousAlert = time.Now()
	}

	if !held.PreviousAlert.IsZero() {

		diff := held.PreviousAlert.Sub(time.Now())
		//log.Printf("Time %f", diff.Minutes())
		if diff.Minutes() < -2.0 {
			if held.Code != msg.Code {
				log.Print("Reset counters")
				checkFrame = true
				held.Count = 0
				held.FalseCount = 0
				held.Point = 0
				held.Triggered = false
			} else if held.Triggered {
				held.PreviousAlert = time.Now()
				log.Print("Delayed as it's the same motion")
			} else {
				checkFrame = true
			}
		}
	} else {
		checkFrame = true
	}

	held.Code = msg.Code
	//checkFrame = true //ignore the time out
	if checkFrame {

		//Decode the points
		msg.Locations = strings.Replace(msg.Locations, "'", "\"", -1)
		var locPoints []map[string]interface{}
		err := json.Unmarshal([]byte(msg.Locations), &locPoints)
		failOnError(err, "Json failed on locpoints")
		down := false
		var smallest float64
		for _, loc := range locPoints {
			v1, err := strconv.ParseFloat(fmt.Sprintf("%v", loc["m10"]), 64)
			v2, err := strconv.ParseFloat(fmt.Sprintf("%v", loc["m00"]), 64)
			failOnError(err, "failed to convert v2")
			v3, _ := strconv.ParseFloat(fmt.Sprintf("%v", loc["m01"]), 64)
			mY := v3 / v2
			mX := v1 / v2
			log.Printf("Comparing %f with mx %f mY %f ", held.Point, mX, mY)

			if held.Point < 5 {
				held.Point = mX
			}
			if mX < held.Point {
				down = true
			}
			if smallest < 5 {
				smallest = mX
			}
			if smallest > mX {
				smallest = mX
			}

		}
		//reset smallest for previous
		held.Point = smallest

		if len(locPoints) > 0 {
			if down {
				log.Print("For this frame i would agree that it's likely to come from the gate.")
				held.Count++
			} else {
				log.Print("Added to false count")
				held.FalseCount++
			}

			if held.FalseCount > 5 {
				log.Print("False count went over. Decided to delay alerts")
				held.Triggered = true
				held.PreviousAlert = time.Now()
			}

			if held.Count > 5 {
				log.Print("I would send a notification now!")
				held.Triggered = true
				go sendNotification()
				held.PreviousAlert = time.Now()
			}
		}

	}

}

func sendNotification() {
	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	// 	"password=%s dbname=%s sslmode=disable",
	// 	host, port, user, passdb, dbname)

	// db, err := sql.Open("postgres", psqlInfo)
	// failOnError(err, "Database opening error")
	// defer db.Close()
	// sqlStatement := `SELECT key FROM keys`
	// row, err := db.Query(sqlStatement)
	// failOnError(err, "Query error")
	// defer row.Close()
	// var keys []string
	// for row.Next() {
	// 	var key string
	// 	err = row.Scan(&key)
	// 	failOnError(err, "Failed to scan")
	// 	keys = append(keys, key)
	// }
	// log.Printf("Keys! %s", keys)
	keys := []string{"erlnh0uEks8:APA91bGs_p5Na7oEpZHm44POkc5jq4c6XAJ7kd7WxbflEiGo4JtuH3CRRdOVhChVjbzXWkjeWOaV6GICSXMicDC7sRiRyIA6dkpwD262Hy-Juq0qZJg2JCqHz0O2hQ_718EtVSALz-xh"}
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
