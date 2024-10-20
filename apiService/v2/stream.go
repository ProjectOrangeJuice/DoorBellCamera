package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var streamEngineChan chan interface{}
var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type addListener struct {
	stream chan []byte
}

type removeListener struct {
	stream chan []byte
}

type outMessage struct {
	Image []byte
	Alert bool
}

//Socket handler
func getVideoShared(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	stream := make(chan []byte, 10)
	add := addListener{stream}
	remove := removeListener{stream}
	streamEngineChan <- add
	t := time.NewTimer(5 * time.Second)
	var m interface{}
	alert := false
	ready := true

	// Start the socket listner
	errchan := make(chan bool)
	readychan := make(chan bool)
	go readChan(ws, errchan, readychan)

	for {
		select {
		case <-errchan:
			ws.Close()
			streamEngineChan <- remove
			return
		case <-readychan:
			ready = true

		case <-t.C:
			// Time out
			streamEngineChan <- remove
			ws.Close()
			return
		case msg := <-stream:

			err := json.Unmarshal(msg, &m)
			if err != nil {
				log.Printf("error with unmarsh stream message. Must be image?")
				if !ready {
					// Skip this message, client not ready
					continue
				}
				o := outMessage{msg, alert}
				err := ws.WriteJSON(o)
				alert = false
				if err != nil {
					log.Printf("Failed to write json %v", err)
					streamEngineChan <- remove
					return
				}
			} else {
				alert = true
			}
		}
	}

}

func readChan(ws *websocket.Conn, errc chan bool, ready chan bool) {
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading socket %v", err)
			errc <- true
			return
		}
		ready <- true
	}
}

func streamEngine() {
	streamEngineChan = make(chan interface{}, 5)
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
