package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const videoFolder string = "videos"

//CaptureLocation is the location of the capture folder
const CaptureLocation string = "images"

var server = ""
var connect *amqp.Connection

//Message is the JSON message format
type Message struct {
	Image  string
	Time   string
	Code   string
	Count  int
	Name   string
	Blocks string
	End    bool
}

type outMessage struct {
	Code string
	Name string
}

type cameraStructure struct {
	prev        string
	notified    string
	ignoreTimer bool
}

var camera = make(map[string]*cameraStructure)
var timer time.Timer

var conn *mongo.Database

func main() {
	go http.ListenAndServe("localhost:8080", nil)
	var err error
	conn, err = configDB(context.Background())
	if err != nil {
		//log.Fatal(err)
	}
	server = "amqp://guest:guest@localhost:5672/"
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connect.Close()

	readyAndListen()

}

func readyAndListen() {
	msgs, ch := listenToExchange("motion", "#")

	createTimer()
	forever := make(chan bool)

	go func() {
		defer ch.Close()
		for d := range msgs {
			decodeMessage(d.Body)
			d.Ack(true)
		}
	}()

	//log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func createTimer() {
	timer := time.NewTimer(30 * time.Second)
	go func() {
		<-timer.C
		//log.Printf("Timer is over")
		for k, v := range camera {
			if v.prev != "" && v.prev != v.notified && !v.ignoreTimer {
				log.Println("I've ended due to timeout")
				v.notified = v.prev
				notifyQueue(v.prev, k)
				v.prev = ""
			}
			v.ignoreTimer = false
		}

		createTimer()
	}()
}

var m Message

func decodeMessage(d []byte) {

	err := json.Unmarshal(d, &m)
	failOnError(err, "Json decode error")
	if _, ok := camera[m.Name]; !ok {
		c := cameraStructure{"", "Nothing", true}
		camera[m.Name] = &c
	}
	if m.End {
		//end of motion, create video
		log.Println("I've ended as I was sent END")
		tc := camera[m.Name]
		go notifyQueue(tc.prev, m.Name)
		tc.ignoreTimer = true
		tc.notified = m.Code
		m.End = false
		tc.prev = ""

	} else {
		storeImage(m)
	}

}

func configDB(ctx context.Context) (*mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s", "localhost")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to mongo: %v", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("mongo client couldn't connect with background context: %v", err)
	}
	todoDB := client.Database("camera")
	return todoDB, nil
}

type dbRecord struct {
	Code     string
	Location string
	Time     string
	Reason   string
}

func recordDb(msg Message, loc string) {
	var record dbRecord
	record.Code = msg.Code
	record.Location = loc
	record.Reason = msg.Blocks
	record.Time = msg.Time

	collection := conn.Collection("capture")
	collection.InsertOne(context.TODO(), record)

	tc := camera[msg.Name]
	if tc.prev != "" && tc.prev != msg.Code {
		//log.Printf("End of prev code")
		notifyQueue(tc.prev, msg.Name)
		tc.ignoreTimer = true
		tc.prev = msg.Code
	} else if tc.prev == "" {
		tc.prev = msg.Code
		tc.ignoreTimer = true

	}

	//log.Println("End record")
}

func notifyQueue(code string, name string) {
	//log.Printf("making video.. code is %s and name is %s", code, name)
	makeVideo(code, name)
	//log.Println("Finished making video?")
}

func storeImage(msg Message) {
	//convert base64
	bImage, err := base64.StdEncoding.DecodeString(msg.Image)
	failOnError(err, "Base64 error")
	location := fmt.Sprintf("%s/%s-%d.jpg", CaptureLocation, msg.Code, msg.Count)
	err = ioutil.WriteFile(location, bImage, 0644)
	failOnError(err, "Error writing image")
	//log.Printf("Stored image %s", location)
	//log.Println(bImage[0])
	recordDb(msg, location)
	//log.Println("End store image")
}

func failOnError(err error, msg string) {
	if err != nil {
		//log.Fatalf("%s: %s", msg, err)
	}
}
