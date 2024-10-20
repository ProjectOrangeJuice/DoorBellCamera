package main

import (
	"context"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

const videoFolder string = "videos"
const captureLocation string = "images"

type cameraStructure struct {
	prev        string
	notified    string
	ignoreTimer bool
}

var timer time.Timer
var conn *mongo.Database

func main() {
	//go http.ListenAndServe("localhost:8080", nil)
	var err error
	conn, err = configDB(context.Background())
	if err != nil {
		//log.Fatal(err)
	}
	server = "amqp://guest:guest@localhost:5672/"
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connect.Close()

	readyAndListen()

}
