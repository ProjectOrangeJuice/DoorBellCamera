package main

import (
	"log"

	"github.com/streadway/amqp"
)

func listenToQueue(q string) (<-chan amqp.Delivery, *amqp.Channel) {

	ch, err := connect.Channel()
	failOnError(err, "Failed to open a channel")

	qu, err := ch.QueueDeclare(
		q,     // name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")
	msgs, err := ch.Consume(
		qu.Name, // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	failOnError(err, "Failed to register a consumer")

	return msgs, ch
}

func listenToExchange(name string, routing string) (<-chan amqp.Delivery, *amqp.Channel) {
	log.Printf("I'm going to listen to %s with %s", name, routing)
	ch, err := connect.Channel()
	failOnError(err, "Failed to open a channel")
	err = ch.ExchangeDeclare(
		name,    // name
		"topic", // type
		false,   // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")
	err = ch.Qos(1, 0, true)
	failOnError(err, "Setting QOS")

	err = ch.QueueBind(
		q.Name,  // queue name
		routing, // routing key
		name,    // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

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
	return msgs, ch
}

func listenToFanout(name string) (<-chan amqp.Delivery, *amqp.Channel) {

	ch, err := connect.Channel()
	failOnError(err, "Failed to open a channel")
	err = ch.ExchangeDeclare(
		name,     // name
		"fanout", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		name,   // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

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
	return msgs, ch
}
