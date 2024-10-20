package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"gocv.io/x/gocv"
)

var liveStream chan gocv.Mat
var motionStream chan []byte

func liveStreamPush() {
	liveStream = make(chan gocv.Mat, 5)
	rabbitChannel := setupLiveStream()
	defer rabbitChannel.Close()
	for img := range liveStream {
		//Convert the image to jpg
		buf, err := gocv.IMEncodeWithParams(".jpg", img, []int{gocv.IMWriteJpegQuality, 50})
		if err != nil {
			log.Printf("Error encoding %s", err)
		}
		rabbitChannel.Publish("camera", "", false, false, amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "bytes",
			Body:         buf,
		})
	}
}

func setupLiveStream() *amqp.Channel {
	ch, err := rabbit.Channel()
	failOnError(err, "Failed to open a channel")
	err = ch.ExchangeDeclare(
		"camera", // name
		"fanout", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare exchange")
	return ch
}

func setupMotionStream() *amqp.Channel {
	ch, err := rabbit.Channel()
	failOnError(err, "Failed to open a channel")
	err = ch.ExchangeDeclare(
		"motion", // name
		"fanout", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare exchange")
	return ch
}

func motionStreamPush() {
	motionStream = make(chan []byte, 5)
	rabbitChannel := setupLiveStream()
	defer rabbitChannel.Close()
	for data := range motionStream {
		fmt.Println("Push frame")
		//Convert the image to jpg
		rabbitChannel.Publish("", "camera", false, false, amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "bytes",
			Body:         data,
		})
	}

}
