package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"github.com/streadway/amqp"
	"gocv.io/x/gocv"
)

const layout = "Jan 2, 2006 3:04pm"

var rabbit *amqp.Connection

func main() {
	//Connect to the rabbit
	connectRabbit()
	go liveStreamPush()
	in := make(chan InputImage)
	s := settings{}
	go checkMotion(in, liveStream, &s)
	//Open video
	video, err := gocv.OpenVideoCapture("/home/oharris/t.mp4")
	defer video.Close()
	if err != nil {
		log.Printf("Failed to open video: %s", err)
	}

	img := gocv.NewMat()
	defer img.Close()

	streamImg := gocv.NewMat()
	defer img.Close()

	statusColor := color.RGBA{0, 255, 0, 0}

	fmt.Println("Reading the video now")
	preTime := time.Now()
	fps := 1. / 99
	for {
		//Skip some frames
		timeSince := time.Since(preTime)
		if timeSince.Seconds() < fps {
			video.Grab(1)
			continue
		}
		preTime = time.Now()

		ok := video.Read(&img)
		if !ok {
			fmt.Println("Video closed")
			return
		}

		if img.Empty() {
			//Didn't read anything useful, skipping
			continue
		}

		// Attach timestamp to stream frame
		streamImg = img //Copy the image
		t := time.Now()
		stamp := t.Format(layout)
		gocv.PutText(&streamImg, stamp, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		// Push to live stream
		//liveStream <- streamImg
		// Push to motion check
		in <- InputImage{img, streamImg}
	}
}

func connectRabbit() {
	server := "amqp://guest:guest@localhost:5672/"
	var err error
	rabbit, err = amqp.Dial(server)
	if err != nil {
		log.Printf("Failed to connect to rabbit: %s", err)
		time.Sleep(10)
		connectRabbit()
	}
	fmt.Println("Connected to rabbit")
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
