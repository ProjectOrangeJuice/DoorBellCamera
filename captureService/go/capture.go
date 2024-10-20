package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"gocv.io/x/gocv"
)

type out struct {
	CameraName string
	Time       string
	Image      string
}

const server = "amqp://guest:guest@localhost:5672/"
const waitTime = 1000000000 / 1

func main() {
	connect, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	ch, err := connect.Channel()
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
	failOnError(err, "Exchange failed")

	stream, err := gocv.OpenVideoCapture("rtsp://192.168.1.120")
	defer stream.Close()
	failOnError(err, "Failed to open stream")

	img := gocv.NewMat()
	defer img.Close()
	// wait := time.Nanosecond * 200
	start := time.Now().UnixNano()
	for {

		now := time.Now().UnixNano()
		diff := now - start
		if diff > 200000000 {
			start = now
			//one second
			if ok := stream.Read(&img); !ok {
				fmt.Printf("Device closed: %v\n", "Streamer..")
				return
			}
			if img.Empty() {
				continue
			}

			//convert it to a thing we can read
			buf, _ := gocv.IMEncode(".jpg", img)
			encoded := base64.StdEncoding.EncodeToString([]byte(buf))
			output := out{"Hello", "Now", encoded}
			b, err := json.Marshal(output)

			err = ch.Publish(
				"videoStream", // exchange
				"Hello",       // routing key
				false,         // mandatory
				false,         // immediate
				amqp.Publishing{
					Body: []byte(b),
				})
			failOnError(err, "Failed to publish a message")
		} else {
			stream.Grab(1)
		}
	}

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
