package main

import (
	"fmt"
	"image"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
	"gocv.io/x/gocv"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type buffered struct {
	Time   string
	Name   string
	Image  []byte
	Code   string
	Count  int64
	Blocks []image.Point
}

func main() {
	go http.ListenAndServe(":8080", nil)
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"camera", // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	//enc := gob.NewDecoder(&network)
	forever := make(chan bool)
	window := gocv.NewWindow("Capture Window")
	defer window.Close()
	go func() {
		for d := range msgs {
			buf := &Buffer{}
			d.Ack(true)
			fmt.Printf("Message size: %v", (float32(len(d.Body)) / 1000000))
			err := proto.Unmarshal(d.Body, buf)
			if err != nil {
				fmt.Printf("Failed to unmash %v\n", err)
				continue
			}
			v, err := gocv.IMDecode(buf.Image, gocv.IMReadAnyColor)
			//failOnError(err, "Reading image")
			if err != nil {
				fmt.Printf("Error displaying image %s\n", err)
				continue
			}
			window.IMShow(v)
			if window.WaitKey(1) == 27 {
				break
			}

			v.Close()
			fmt.Println("Finished message")

		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
