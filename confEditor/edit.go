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

func main() {
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

//Listen for the return of the configs
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
			decodeMsg(d.Body, arg)
			rch <- "done"
		}
	}()

	<-forever
	fmt.Print("over 2")
}

func decodeMsg(msg []byte, arg string) {
	arg = strings.Replace(arg, ".", "-", -1)
	var m outMessage
	err := json.Unmarshal(msg, &m)
	failOnError(err, "Json decode error")
	err = ioutil.WriteFile(fmt.Sprintf("configs/%s.json", arg), []byte(m.Inner), 0644)
	failOnError(err, "Failed to write")

}

//Send the command to get the config file
func getCommand(arg string) {
	returnCh := make(chan string)
	go goListen(returnCh, arg)
	if m := <-returnCh; m != "ready" {
		log.Panicf("Something went wrong when waiting for ready")
	}
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ (get)")
	defer conn.Close()

	body := outMessage{"read", "test"}
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
		fmt.Printf("Config saved to %s. Edit this\n", m)
	}

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//Push the config file
func setCommand(arg string) {
	conn, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ (get)")
	defer conn.Close()
	arg2 := strings.Replace(arg, ".", "-", -1)
	dat, err := ioutil.ReadFile(fmt.Sprintf("configs/%s.json", arg2))
	failOnError(err, "Couldn't read file")
	body := outMessage{"update", string(dat)}
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
