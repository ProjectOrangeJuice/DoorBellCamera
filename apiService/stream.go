package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

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
	log.Printf("Get video, socket upgraded for %s to watch %s", r.RemoteAddr, params["camera"])
	go sendVideo(cam, ws, false)
}

//Socket handler
func getCompressedVideo(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	log.Printf("Get video, socket upgraded for %s to watch %s", r.RemoteAddr, params["camera"])
	go sendVideo(cam, ws, true)
}

//Message is the json format
type Message struct {
	Image string
	Time  string
	Code  string
	Count int
	Name  string
}

//For the connection, get the stream and send it to the socket
func sendVideo(cam string, ws *websocket.Conn, compressed bool) {
	msgs, ch := listenToExchange("videoStream", cam)
	var m Message
	//forever := make(chan bool)
	p := make(chan bool)
	go pingponger(ws, p)
	const duration = 13 * time.Second
	timer := time.NewTimer(duration)
	alive := true
	for alive {
		select {
		case d := <-msgs:
			timer.Reset(duration)
			err := json.Unmarshal(d.Body, &m)
			failOnError(err, "Json decode error")
			if compressed {
				sDec, _ := b64.StdEncoding.DecodeString(m.Image)
				image, _, err := image.Decode(bytes.NewReader(sDec))
				failOnError(err, "Failed to read image to compress")

				buf := new(bytes.Buffer)
				err = jpeg.Encode(buf, image, &jpeg.Options{15})
				sends3 := buf.Bytes()

				sEnc := b64.StdEncoding.EncodeToString([]byte(sends3))
				err = ws.WriteMessage(websocket.TextMessage, []byte(sEnc))
			} else {
				err = ws.WriteMessage(websocket.TextMessage, []byte(m.Image))
			}
			if err != nil {

				alive = false
				break
			}
		case <-p:
			log.Println("Ending connection due to ping pong1")
			alive = false
			break

		case <-timer.C:
			//Connection to camera failed
			print("Timer!")
			alive = false
			break

		}
	}
	ch.Close()
	ws.Close()
}

//Socket handler
func getMotionAlert(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client

	log.Printf("Get motion, socket upgraded for %s to watch", r.RemoteAddr)
	go getMotionAlerts(ws)
}

type alert struct {
	Name string
	Time string
}

//For the connection, get the stream and send it to the socket
func getMotionAlerts(ws *websocket.Conn) {
	msgs, ch := listenToExchange("motion", "#")
	var m alert
	p := make(chan bool)
	go pingponger(ws, p)
	alive := true
	for alive {
		select {
		case d := <-msgs:
			err := json.Unmarshal(d.Body, &m)
			failOnError(err, "Json decode error")
			b, _ := json.Marshal(m)
			err = ws.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				alive = false
				break
			}
		case <-p:
			log.Println("Ending connection due to ping pong")
			alive = false
			break
		}

	}
	ch.Close()
	ws.Close()

}

func pingponger(ws *websocket.Conn, c chan bool) {
	ws.WriteMessage(websocket.TextMessage, []byte("PING"))
	for {
		_, bytes, err := ws.ReadMessage()
		if err != nil || string(bytes) != "PONG" {
			break
		}
		time.Sleep(5 * time.Second)
		err = ws.WriteMessage(websocket.TextMessage, []byte("PING"))
		if err != nil {
			log.Println("Write Error: ", err)
			break
		}
	}
	c <- false

}
