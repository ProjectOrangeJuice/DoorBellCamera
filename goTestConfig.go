package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type OutMessage struct {
	Task  string
	Inner string
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.1.126:30188/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	jsonS :=
		`{
		"serverAddress" : "192.168.99.100",
		"serverPort" : "31693",
		"threshold":80,
		"minCount":5,
		"cameras":[
			{
				"name":"Test Camera@Dontuse",
				"threshold":10,
				"minCount":1
			}
		]
	
		
	
	}
	`
	body := OutMessage{"update", jsonS}
	b, err := json.Marshal(body)

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"config", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a exchange")
	err = ch.Publish(
		"config",       // exchange
		"motion.check", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(b),
		})

	failOnError(err, "Failed to publish a message")

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
