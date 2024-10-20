package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//Message is the json format
type Message struct {
	Image string
	Time  string
	Code  string
	Count int
	Name  string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//Socket handler
func getVideo(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	logger.Printf("Get video, socket upgraded for %s to watch %s", r.RemoteAddr, params["camera"])
	go sendVideo(cam, ws)
}

//Socket handler
func getMotionWatch(w http.ResponseWriter, r *http.Request) {
	//ws, err := upgrader.Upgrade(w, r, nil)
	//failOnError(err, "Couldn't upgrade")
	logger.Printf("Motion watch for %s", r.RemoteAddr)
	//go motionWatch("", ws)
}

//Socket handler
func getDoor(w http.ResponseWriter, r *http.Request) {
	//ws, err := upgrader.Upgrade(w, r, nil)
	//failOnError(err, "Couldn't upgrade")
	logger.Printf("Door watch for %s", r.RemoteAddr)
	// register client
	//go doorWatch("", ws)
}

//For the connection, get the stream and send it to the socket
func sendVideo(cam string, ws *websocket.Conn) {
	msgs, ch := listenToExchange("videoStream", strings.Replace(cam, " ", ".", -1))
	forever := make(chan bool)
	var m Message

	defer closeThis(ch)
	const duration = 3 * time.Second
	timer := time.NewTimer(duration)
	for !connect.IsClosed() {
		select {
		case d := <-msgs:
			timer.Reset(duration)
			err := json.Unmarshal(d.Body, &m)
			failOnError(err, "Json decode error")

			err = ws.WriteMessage(websocket.TextMessage, []byte(m.Image))

			if err != nil {
				log.Printf("Websocket error: %s", err)
				ws.Close()
				closeThis(ch)
				msgs = nil
				forever = nil
				d = amqp.Delivery{}
				err = nil
				break
			}

		case <-timer.C:
			fmt.Println("Timeout !")
			ws.Close()
			break
		}
		print("For loop!")
	}
	print("Ended?")
	connect = nil
	m = Message{}
	ws = nil
	return
	<-forever
	print("Ended forever")
}

func closeThis(ch *amqp.Channel) {
	log.Print("Closing channel")
	ch.Close()
	ch = nil
}

//For the connection, get motion and send it
func motionWatch(cam string, ws *websocket.Conn) {
	msgs, ch := listenToQueue("motionAlert")
	defer ch.Close()
	prev := ""
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			m := decodeMessage(d.Body)
			if prev != m.Code {
				ws.WriteMessage(websocket.TextMessage, []byte(m.Code))
				prev = m.Code
			}
		}

	}()
	<-forever
}

//For the connection, get motion and send it
func doorWatch(cam string, ws *websocket.Conn) {
	msgs, ch := listenToQueue("doorService")
	defer ch.Close()
	prev := ""
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			m := decodeMessage(d.Body)
			if prev != m.Code {
				ws.WriteMessage(websocket.TextMessage, []byte(m.Code))
				prev = m.Code
			}
		}

	}()
	<-forever
}
func decodeMessage(d []byte) Message {
	var m Message
	err := json.Unmarshal(d, &m)
	failOnError(err, "Json decode error")
	return m

}
