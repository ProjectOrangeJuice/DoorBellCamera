package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var connect *amqp.Connection

//For the connection, get the stream and send it to the socket
func DoStream(cam string, ws *websocket.Conn) {
	log.Printf("Setting up connection for %s", cam)
	msgs, ch := listenToExchange("videoStream", strings.Replace(cam, " ", ".", -1))
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		const duration = 3 * time.Second
		timer := time.NewTimer(duration)
		for {
			select {
			case d := <-msgs:
				timer.Reset(duration)
				var m Message
				err := json.Unmarshal(d.Body, &m)
				failOnError(err, "Json decode error")

				err = ws.WriteMessage(websocket.TextMessage, []byte(m.Image))

				if err != nil {
					log.Printf("Websocket error: %s", err)
					ws.Close()
					return
				}

			case <-timer.C:
				fmt.Println("Timeout !")
				ws.Close()
			}
		}

	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	log.Printf("Finished..")
}

//For the connection, get motion and send it
func doMotionCheck(cam string, ws *websocket.Conn) {
	log.Printf("Setting up connection for %s", cam)

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
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

//For the connection, get motion and send it
func doDoorWatch(cam string, ws *websocket.Conn) {
	log.Printf("Setting up connection for %s", cam)

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
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func decodeMessage(d []byte) Message {
	var m Message
	err := json.Unmarshal(d, &m)
	failOnError(err, "Json decode error")
	return m

}
