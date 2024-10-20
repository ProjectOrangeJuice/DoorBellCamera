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

//For the connection, get the stream and send it to the socket
func DoStream(cam string, ws *websocket.Conn) {
	log.Printf("Setting up connection for %s", cam)
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"videoStream", // name
		"topic",       // type
		false,         // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		strings.Replace(cam, " ", ".", -1), // routing key
		"videoStream",                      // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

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
