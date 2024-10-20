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

type outMessage struct {
	Task  string
	Inner string
}

//GET for getconfig
func getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	params := mux.Vars(r)
	msg := getCommand(params["service"])
	body := outMessage{params["service"], msg}
	b, err := json.Marshal(body)
	failOnError(err, "failed to create json to send")
	w.Write(b)

}

//POST for setconfig
func setConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	params := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	failOnError(err, "failed to read body")
	setCommand(params["service"], string(body))
}

//Listens for the return of the config file
func goListen(rch chan string, arg string) {
	msgs, ch := listenToExchange("config", "config.test")
	defer ch.Close()
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
	var m outMessage
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")
	return m.Inner
}

//Get the config file
func getCommand(arg string) string {
	returnCh := make(chan string)
	go goListen(returnCh, arg)
	if m := <-returnCh; m != "ready" {
		log.Panicf("Something went wrong when waiting for ready")
	}

	_, ch := listenToExchange("config", arg)
	body := outMessage{"read", "test"}
	b, err := json.Marshal(body)
	failOnError(err, "Failed to make json")
	defer ch.Close()
	go func() {
		err := ch.Publish(
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

//Set the config file
func setCommand(arg string, config string) {

	body := outMessage{"update", config}
	b, err := json.Marshal(body)
	_, ch := listenToExchange("config", arg)
	defer ch.Close()

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
