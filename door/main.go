package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/streadway/amqp"
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
	Image    string
	Time     string
	Code     string
	Count    int
	Name     string
	Reason   string
	Location string
}

type hold struct {
	Code  string
	Point int
	Count int
}

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

	//Decode the points
	log.Printf("Loc points are.. %s", msg.Location)
	var locPoints []map[string]interface{}
	err := json.Unmarshal([]byte(msg.Location), &locPoints)
	failOnError(err, "Json failed on locpoints")
	down := false
	for _, loc := range locPoints {
		v2, _ := strconv.Atoi(fmt.Sprintf("%v", loc["m00"]))
		v3, _ := strconv.Atoi(fmt.Sprintf("%v", loc["m01"]))
		mY := v3 / v2
		if held.Code != msg.Code {
			log.Print("The code doesn't match. finding the largest point")
			held.Count = 0
			if mY > held.Point {
				held.Point = mY
			}
		} else {
			log.Print("The code matches. We can compare now")
			if mY > held.Point {
				down = true
				held.Point = mY
			}
		}
	}

	log.Printf("Largest point is %s or %s", held.Point)
	held.Code = msg.Code
	if down {
		log.Print("For this frame i would agree that it's likely to come from the gate.")
		held.Count++
	}

	if held.Count > 3 {
		log.Print("I would send a notification now!")
	}

}
