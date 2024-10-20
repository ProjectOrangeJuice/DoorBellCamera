package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/streadway/amqp"
	"gocv.io/x/gocv"
)

type out struct {
	CameraName string
	Time       string
	Image      string
}

const server = "amqp://guest:guest@localhost:5672/"
const secondInNano = time.Nanosecond * time.Second

var connect *amqp.Connection
var ch *amqp.Channel

func main() {
	go func() {
		http.ListenAndServe("localhost:8080", nil)
	}()
	connect, ch = setupRabbit()
	errChan := make(chan *amqp.Error)
	errChan = connect.NotifyClose(errChan)
	go watchCloses(errChan)
	defer connect.Close()
	defer ch.Close()

	//get the settings and update every minute
	getSettings()
	setupMotion()

	stream, err := gocv.OpenVideoCapture("rtsp://192.168.1.120")
	defer stream.Close()
	failOnError(err, "Failed to open stream")
	img := gocv.NewMat()
	defer img.Close()
	//Fps delay timer
	start := time.Now().UnixNano()
	for {
		for {
			//Check to see if we need this frame
			now := time.Now().UnixNano()
			diff := now - start
			if int64(diff) > (secondInNano.Nanoseconds() / int64(setting.FPS)) {
				start = time.Now().UnixNano() //Update time
				//Grab the frame
				if ok := stream.Read(&img); !ok {
					fmt.Printf("Device closed: %v\n", "Streamer..")
					break
				}
				if img.Empty() {
					log.Printf("Frame was empty")
					continue
				}

				gocv.Flip(img, &img, 1)

				//go sendFrame(img, ch)
				detect(img)

			} else {
				//Ignore this frame
				stream.Grab(1)
			}
		}
		//In the event something fails, it will try again.
		time.Sleep(time.Second * 2)
		log.Printf("Reconnecting camera")
		stream, err := gocv.OpenVideoCapture("rtsp://192.168.1.120")
		defer stream.Close()
		failOnError(err, "Failed to open stream")
	}

}

func watchCloses(err chan *amqp.Error) {
	off, ok := <-err
	if ok {
		log.Printf("Connection closed for rabbit %s", off.Reason)
		connect, ch = setupRabbit()
	}
}

func sendFrame(frame gocv.Mat, ch *amqp.Channel) {
	//Add timestamp
	pt := image.Pt(10, 30)
	stamp := time.Now().Format("01-01-2006 15:04:05")
	gocv.PutText(&frame, stamp, pt, gocv.FontHersheyComplex, 1, color.RGBA{255, 0, 0, 255}, 1)

	//convert it to a thing we can read
	buf, _ := gocv.IMEncode(".jpg", frame)
	encoded := base64.StdEncoding.EncodeToString([]byte(buf))
	output := out{setting.Name, string(time.Now().Unix()), encoded}
	b, err := json.Marshal(output)

	err = ch.Publish(
		"videoStream", // exchange
		setting.Name,  // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			Body: []byte(b),
		})
	if err != nil {
		log.Printf("Rabbit isn't connected?")
	}

}

func setupRabbit() (*amqp.Connection, *amqp.Channel) {
	connect, err := amqp.Dial(server)
	failOnError(err, "Failed to connect to RabbitMQ")
	ch, err := connect.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		"videoStream", // name
		"topic",       // type
		false,         // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Exchange failed")
	return connect, ch
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
