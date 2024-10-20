package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gocv.io/x/gocv"
)

const layout = "Jan 2, 2006 3:04pm"

var rabbit *amqp.Connection
var databaseClient *mongo.Client

func main() {
	//Connect to database
	var err error
	databaseClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Printf("Failed to connect to database: %s", err)
		return
	}
	//Connect to the rabbit
	connectRabbit()
	go liveStreamPush()
	go motionStreamPush()
	in := make(chan inputImage)
	setting := getSetting()
	settingUpdate := time.Now().Add(time.Second * 30)
	go checkMotion(in, liveStream, &setting)
	//Open video
	video, err := gocv.OpenVideoCapture(setting.Connection)
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
	fps := 1. / setting.FPS
	for {
		// Check to see if we should update our settings
		if time.Now().After(settingUpdate) {
			//Update the settings
			settingUpdate = time.Now().Add(time.Second * 30)
			setting = getSetting()
		}

		//Skip some frames
		timeSince := time.Since(preTime)
		if timeSince.Seconds() < float64(fps) {
			//video.Grab(1)
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
		in <- inputImage{img, streamImg}
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
