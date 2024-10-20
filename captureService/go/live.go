package main

import (
	"github.com/streadway/amqp"
	"gocv.io/x/gocv"
)

var liveStream chan gocv.Mat

func liveStreamPush() {
	liveStream = make(chan gocv.Mat, 5)
	rabbitChannel := setupRabbitStream()
	defer rabbitChannel.Close()
	for img := range liveStream {
		//Convert the image to jpg
		buf, _ := gocv.IMEncodeWithParams(".jpg", img, []int{gocv.IMWriteJpegQuality, 50})
		rabbitChannel.Publish("", "camera", false, false, amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "bytes",
			Body:         buf,
		})
	}
}

func setupRabbitStream() *amqp.Channel {
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
