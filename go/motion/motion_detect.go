package main

//THIS FAILS.
import (
	"encoding/base64"
	"encoding/json"
	"log"

	"gocv.io/x/gocv"

	"github.com/streadway/amqp"
)

type videoStreamJson struct {
	Time  string
	Image string
}

var preImage = gocv.NewMat()

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@192.168.99.100:31693/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"videoStream", // name
		false,         // durable
		false,         // delete when usused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
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
			//log.Printf("Received a message: %s", d.Body)
			readData(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func readData(msg []byte) {
	var newMessage videoStreamJson
	err := json.Unmarshal(msg, &newMessage)
	failOnError(err, "Json failed")
	//convert base64
	bImage, err := base64.StdEncoding.DecodeString(newMessage.Image)
	failOnError(err, "Base64 error")
	readImage(bImage)
}

func readImage(imageData []byte) {
	image, err := gocv.IMDecode(imageData, gocv.IMReadAnyColor)
	failOnError(err, "Decoding image failed")

	if preImage.Empty() {
		preImage = image

	} else {
		var difference gocv.Mat
		gocv.AbsDiff(image, image, &difference)
		all := difference.ToBytes()
		var noneZero = 0
		for _, element := range all {
			if element != 0 {
				noneZero++
			}
		}

		log.Printf("None zereos... %d", noneZero)

	}

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
