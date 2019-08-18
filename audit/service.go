package main

import (
	"log"
	"os"

	"github.com/streadway/amqp"
)

var logger *log.Logger
var server = "amqp://guest:guest@192.168.1.126:30188/"
var connect *amqp.Connection

func main() {
	setupAudit()
	var err error
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	//listen to config
	msgs, ch := listenToExchange("config", "#")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			logger.Printf("Message %s", d.Body)
		}
	}()

	<-forever

}

func setupAudit() {
	f, err := os.OpenFile("audit.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	logger = log.New(f, "CONFIG-RABBIT ", log.LstdFlags)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
