package main

//CLI for interacting with the config editor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

type outMessage struct {
	Task  string
	Inner string
}

//should take this variable as an argument from the terminal
var server = "amqp://guest:guest@192.168.1.126:30188/"
var connect *amqp.Connection

func main() {
	var err error
	connect, err = amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	//Get the input from the terminal
	input := ""
	read := true
	reader := bufio.NewReader(os.Stdin)
	for read {
		//Options
		fmt.Printf("%-20s%-5s%-20s\n", "get [service]", "-", "Get the config file of the service")
		fmt.Printf("%-20s%-5s%-20s\n", "set [service]", "-", "Set the config file of the service")
		input, _ = reader.ReadString('\n')
		input = input[:len(input)-1]
		args := strings.Fields(input)
		if len(args) == 2 {
			switch args[0] {
			case "get":
				getCommand(args[1])
			case "set":
				setCommand(args[1])
			default:
				read = false
			}
		}
	}
}

//Listens for the return of the config file
func goListen(rch chan string, arg string) {
	msgs, ch := listenToExchange("config", "config.test")
	defer ch.Close()
	rch <- "ready"
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			decodeMsg(d.Body, arg)
			rch <- "done"
		}
	}()

	<-forever
	fmt.Print("over 2")
}

//Get the config file
func getCommand(arg string) {
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
		fmt.Printf("Config saved to %s. Edit this\n", m)
	}

}

func decodeMsg(msg []byte, arg string) {
	arg = strings.Replace(arg, ".", "-", -1)
	var m outMessage
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")
	err = ioutil.WriteFile(fmt.Sprintf("configs/%s.json", arg), []byte(m.Inner), 0644)
	failOnError(err, "Failed to write")

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//Set the config file
func setCommand(arg string) {
	arg2 := strings.Replace(arg, ".", "-", -1)
	dat, err := ioutil.ReadFile(fmt.Sprintf("configs/%s.json", arg2))
	failOnError(err, "Couldn't read file")
	body := outMessage{"update", string(dat)}
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
