package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//Message is the json format
type Message struct {
	Image string
	Time  string
	Code  string
	Count int
	Name  string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//Socket handler
func getVideo(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	go sendVideo(cam, ws)
}

//For the connection, get the stream and send it to the socket
func sendVideo(cam string, ws *websocket.Conn) {
	log.Printf("Setting up connection for %s", cam)
	msgs, ch := listenToExchange("videoStream", strings.Replace(cam, " ", ".", -1))
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		const duration = 3 * time.Second
		timer := time.NewTimer(duration)
		for {
			select {
			case d := <-msgs:
				timer.Reset(duration)
				var m Message
				err := json.Unmarshal(d.Body, &m)
				failOnError(err, "Json decode error")

				err = ws.WriteMessage(websocket.TextMessage, []byte(m.Image))

				if err != nil {
					log.Printf("Websocket error: %s", err)
					ws.Close()
					return
				}

			case <-timer.C:
				fmt.Println("Timeout !")
				ws.Close()
			}
		}

	}()
	<-forever
}
