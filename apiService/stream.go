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
	go sendVideo(cam, ws)
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
func sendVideo(cam string, ws *websocket.Conn) {
	msgs, ch := listenToExchange("videoStream", cam)
	var m Message
	forever := make(chan bool)
	const duration = 13 * time.Second
	timer := time.NewTimer(duration)
	alive := true
	for alive {
		select {
		case d := <-msgs:
			timer.Reset(duration)
			err := json.Unmarshal(d.Body, &m)
			failOnError(err, "Json decode error")
			sDec, _ := b64.StdEncoding.DecodeString(m.Image)
			image, _, err := image.Decode(bytes.NewReader(sDec))

			buf := new(bytes.Buffer)
			err = jpeg.Encode(buf, image, &jpeg.Options{15})
			send_s3 := buf.Bytes()

			sEnc := b64.StdEncoding.EncodeToString([]byte(send_s3))
			err = ws.WriteMessage(websocket.TextMessage, []byte(sEnc))
			if err != nil {
				ch.Close()
				ws.Close()
				alive = false
				break
			}
		case <-timer.C:
			print("Timer!")
			ch.Close()
			ws.Close()
			alive = false
			break
		}

	}
	<-forever

}
