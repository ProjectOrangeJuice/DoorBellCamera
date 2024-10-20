package main

import (
	"log"

	"github.com/streadway/amqp"
)

var streamEngineChan chan interface{}

type addListener struct {
	stream chan []byte
}

type removeListener struct {
	stream chan []byte
}

func streamEngine() {
	clients := make([]chan []byte, 0)
	var frames <-chan amqp.Delivery
	var framesCh *amqp.Channel
	var alerts <-chan amqp.Delivery
	var alertsCh *amqp.Channel
	streaming := false
	for {
		select {
		case cc := <-streamEngineChan:
			switch cc.(type) {
			case addListener:
				clients = append(clients, cc.(addListener).stream)
			case removeListener:
				for index, val := range clients {
					if val == cc.(removeListener).stream {
						clients = remove(clients, index)
						break
					}
				}
			}

			// Check to see if we need to start / stop streamer
			if len(clients) > 0 && !streaming {
				log.Printf("Start streaming")
				// Start stream
				frames, framesCh = listenToExchange("camera", "camera")
				alerts, alertsCh = listenToExchange("motion", "camera")
			}
			if len(clients) == 0 && streaming {
				log.Printf("Stop streaming")
				// stop streaming
				framesCh.Close()
				alertsCh.Close()
			}
			log.Printf("Clients: %v", len(clients))

		case frame := <-frames:
			for _, client := range clients {
				client <- frame.Body
			}
		case alert := <-alerts:
			for _, client := range clients {
				client <- alert.Body
			}
		}

	}
}

func remove(s []chan []byte, i int) []chan []byte {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}
