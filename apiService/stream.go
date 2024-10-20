package main

import (
	"bytes"
	b64 "encoding/base64"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
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

const duration = 20 * time.Second

//For the connection, get the stream and send it to the socket
func sendVideo(cam string, ws *websocket.Conn, compressed bool) {
	msgs, ch := listenToExchange("videoStream", cam)
	//ws.SetCompressionLevel(9)
	var lock sync.Mutex
	var m Message
	//forever := make(chan bool)
	p := make(chan bool)
	go pingponger(ws, p, &lock)

	timer := time.NewTimer(duration)
	alive := true
	last := time.Now().UnixNano()
	firstFrame := true
	for alive {
		select {
		case d := <-msgs:
			var err error
			timer.Reset(duration)
			currentTime := time.Now().UnixNano()
			diff := currentTime - last
			var waitTime int64
			waitTime = 1000000000 / 2
			if compressed && (diff > waitTime || firstFrame) {
				firstFrame = false
				err := json.Unmarshal(d.Body, &m)
				failOnError(err, "Json decode error")
				sDec, _ := b64.StdEncoding.DecodeString(m.Image)
				image, _, err := image.Decode(bytes.NewReader(sDec))
				failOnError(err, "Failed to read image to compress")

				buf := new(bytes.Buffer)
				err = jpeg.Encode(buf, image, &jpeg.Options{10})
				sends3 := buf.Bytes()

				sEnc := b64.StdEncoding.EncodeToString([]byte(sends3))
				lock.Lock()
				err = ws.WriteMessage(websocket.TextMessage, []byte(sEnc))
				lock.Unlock()
				//skip every other frame
				last = time.Now().UnixNano()
			} else if !compressed {
				err := json.Unmarshal(d.Body, &m)
				failOnError(err, "Json decode error")
				lock.Lock()
				err = ws.WriteMessage(websocket.TextMessage, []byte(m.Image))
				lock.Unlock()
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
	m = Message{}
	close(p)
	ch.Close()
	ws.Close()
}

//Socket handler
func getMotionAlert(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	failOnError(err, "Couldn't upgrade")
	// register client
	params := mux.Vars(r)
	cam := params["camera"]
	log.Printf("Get motion, socket upgraded for %s to watch", r.RemoteAddr)
	go getMotionAlerts(ws, cam)
}

type alert struct {
	Name string
	Time string
}

//For the connection, get the stream and send it to the socket
func getMotionAlerts(ws *websocket.Conn, camera string) {
	var lock sync.Mutex
	msgs, ch := listenToExchange("motion", camera)
	var m alert
	p := make(chan bool)
	go pingponger(ws, p, &lock)
	alive := true
	for alive {
		select {
		case d := <-msgs:
			err := json.Unmarshal(d.Body, &m)
			failOnError(err, "Json decode error")
			b, _ := json.Marshal(m)
			lock.Lock()
			err = ws.WriteMessage(websocket.TextMessage, b)
			lock.Unlock()
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

func pingponger(ws *websocket.Conn, c chan bool, lock *sync.Mutex) {
	lock.Lock()
	ws.WriteMessage(websocket.TextMessage, []byte("PING"))
	lock.Unlock()
	for {
		_, bytes, err := ws.ReadMessage()
		if err != nil || string(bytes) != "PONG" {
			break
		}
		time.Sleep(5 * time.Second)
		lock.Lock()
		err = ws.WriteMessage(websocket.TextMessage, []byte("PING"))
		lock.Unlock()
		if err != nil {
			log.Println("Write Error: ", err)
			break
		}
	}
	c <- false

}
