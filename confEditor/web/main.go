package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var server = "amqp://guest:guest@192.168.99.100:31693/"

type OutMessage struct {
	Task  string
	Inner string
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/config/{service}", getConfig).Methods("GET")
	router.HandleFunc("/config/{service}", setConfig).Methods("POST")
	//router.HandleFunc("/stream",wsHandler)
	http.ListenAndServe(":8000", router)
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	msg := getCommand(params["service"])
	body := OutMessage{params["service"], msg}
	b, err := json.Marshal(body)
	failOnError(err, "failed to create json to send")
	w.Write(b)

}

func setConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text")
	params := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	failOnError(err, "failed to read body")
	setCommand(params["service"], string(body))
}

func goListen(rch chan string, arg string) {
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,        // queue name
		"config.test", // routing key
		"config",      // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	rch <- "ready"
	forever := make(chan bool)

	go func() {
		for d := range msgs {

			rch <- decodeMsg(d.Body, arg)
		}
	}()

	<-forever
	fmt.Print("over 2")
}

func decodeMsg(msg []byte, arg string) string {
	arg = strings.Replace(arg, ".", "-", -1)
	var m OutMessage
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")
	return m.Inner
}

func getCommand(arg string) string {
	returnCh := make(chan string)
	go goListen(returnCh, arg)
	if m := <-returnCh; m != "ready" {
		log.Panicf("Something went wrong when waiting for ready")
	}
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ (get)")
	defer conn.Close()

	body := OutMessage{"read", "test"}
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
	go func() {
		err = ch.Publish(
			"config", // exchange
			arg,      // routing key
			false,    // mandatory
			false,    // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(b),
			})

		failOnError(err, "Failed to publish a message")
	}()

	if m := <-returnCh; m == "error" {
		log.Print("Something went wrong when returning")
	} else {
		return m
	}
	return "error"
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func setCommand(arg string, config string) {
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ (get)")
	defer conn.Close()
	body := OutMessage{"update", config}
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
		"config", // exchange
		arg,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(b),
		})

	failOnError(err, "Failed to publish a message")
	log.Print("Sent new config")

}