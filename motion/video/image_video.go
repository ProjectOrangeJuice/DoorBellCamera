package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/icza/mjpeg"
	"github.com/streadway/amqp"
)

type OutMessage struct {
	Code string
}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@192.168.99.100:31693/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"imageToVideo", // name
		false,          // durable
		false,          // delete when usused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

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
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			convert(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func convert(msg []byte) {
	var m OutMessage
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")

	aw, err := mjpeg.New(m.Code, 1280, 720, 5)
	failOnError(err, "Setting up video")

	var files []string
	root := "/home/oharris/Documents/cameraProject/motion/capture"
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	failOnError(err, "Failed to read images")
	// Create a movie from images: 1.jpg, 2.jpg, ..., 10.jpg

	for _, file := range files {
		if strings.Contains(file, m.Code) {
			data, err := ioutil.ReadFile(fmt.Sprintf("%s", file))
			failOnError(err, "Failed reading image")
			err = aw.AddFrame(data)
			failOnError(err, "failed to add frame")
			err = os.Remove(file)
			failOnError(err, "Failed to remove image")
			log.Printf("Added.. %s", file)
		}
		log.Printf("File we looked at.. %s", file)
	}

	err = aw.Close()
	failOnError(err, "Error closing")

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
