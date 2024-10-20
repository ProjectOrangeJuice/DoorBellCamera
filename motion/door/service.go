package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

//Message is the JSON message format
type Message struct {
	Image  string
	Time   string
	Code   string
	Count  int
	Name   string
	Blocks []int
}

type config struct {
	ServerAddress string
	ServerPort    string
	MaxThreshold  []int
	MinThreshold  []int
	MinCount      int
	Quiet         int64
	DelayFailed   int64
	Cameras       []pConfig
}

type pConfig struct {
	Name         string
	MaxThreshold []int
	MinThreshold []int
	MinCount     int
	Quiet        int64
	DelayFailed  int64
}

type pre struct {
	Code   string
	Alert  int64
	Failed int64
}

var configVals config
var camera = make(map[string]*pre)

func main() {
	readConfig()
	listenForMotion()
}

func readConfig() {
	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	failOnError(err, "Failed to read config")
	defer jsonFile.Close()
	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &configVals)
	log.Printf("Config values set to.. %v", configVals)
}

func listenForMotion() {
	conn, err := amqp.Dial(configVals.ServerAddress)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"motionAlert", // name
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
			decodeMessage(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func decodeMessage(d []byte) {
	var m Message
	err := json.Unmarshal(d, &m)
	failOnError(err, "Json decode error")
	if _, ok := camera[m.Name]; !ok {
		c := pre{"", 0, 0}
		camera[m.Name] = &c
		log.Print("Made new camera")
	}

	decideFate(m)

}

func decideFate(m Message) {
	current := camera[m.Name]
	curConfig := getCam(m.Name)
	alert := false
	if current.Alert != 0 && current.Failed != 0 {
		if time.Since(time.Unix(current.Alert, 0)).Seconds() > float64(curConfig.Quiet) &&
			time.Since(time.Unix(current.Failed, 0)).Seconds() > float64(curConfig.DelayFailed) {
			alert = true
		}
	} else {
		alert = true
	}

	if alert {
		if len(curConfig.MaxThreshold) == len(m.Blocks) {
			alerted := false
			for index, key := range m.Blocks {
				if key >= curConfig.MinThreshold[index] && key <= curConfig.MaxThreshold[index] {
					sendAlert(m)
					current.Alert = time.Now().Unix()
					alerted = true
				}
			}
			if !alerted {
				log.Printf("Failed for alert testing")
				current.Failed = time.Now().Unix()
			}
		} else {
			log.Fatalf("Incorrect number of threshold values.. ")
		}
	}
	current.Code = m.Code

}

func sendAlert(m Message) {
	log.Printf("Sending alert!")
}

func getCam(name string) pConfig {

	for _, value := range configVals.Cameras {
		if value.Name == name {
			return value
		}
	}
	//Generate the default one

	v := pConfig{name, configVals.MaxThreshold, configVals.MinThreshold, configVals.MinCount, configVals.Quiet, configVals.DelayFailed}
	return v
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
