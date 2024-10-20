package main

import (
	"context"
	fmt "fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type settings struct {
	Name               string
	Connection         string
	FPS                int
	MinCount           int
	Motion             bool
	Blur               int
	Debug              bool
	BufferBefore       int
	BufferAfter        int
	NoMoveRefreshCount int
	Zones              []zone
}
type zone struct {
	X1          int
	Y1          int
	X2          int
	Y2          int
	Threshold   int
	BoxJump     int
	SmallIgnore int
	Area        int
}

func recvMotionImg(buf chan *Buffer) {
	databaseClient, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	conn := databaseClient.Database("doorbell")
	db := conn.Collection("setting")
	filter := bson.M{"_id": 0}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	doc := db.FindOne(ctx, filter)
	cancel()
	var s settings
	doc.Decode(&s)

	vidStream := make(chan string)
	go makeVideo(vidStream)

	db = conn.Collection("motion")
	timer := time.NewTimer(5 * time.Second)
	codeUsed := ""
	for {
		select {
		case msg := <-buf:

			if codeUsed == "" {
				codeUsed = msg.Code
			} else if codeUsed != msg.Code {
				log.Println("Make video due to new code")
				vidStream <- codeUsed
				codeUsed = msg.Code
			}

			timer.Reset(5 * time.Second)
			fmt.Printf("About to write image %s\n", msg.Code)
			// Store the image
			location := fmt.Sprintf("%s/%s-%d.jpg", imageLocation, msg.Code, msg.Count)
			err := ioutil.WriteFile(location, msg.Image, 0644)
			if err != nil {
				log.Fatalf("Failed to write image %s", err)
			}
			fmt.Println("Written image")

		case <-timer.C:
			// end to last code
			if codeUsed != "" {
				vidStream <- codeUsed
				log.Println("Make video due to timeout")
				codeUsed = ""
			}
			timer.Reset(5 * time.Second)
		}
	}
}

func makeVideo(codes chan string) {
	for vid := range codes {

		saveToFull := fmt.Sprintf("%s/%s.mp4", fullVideoLocation, vid)
		saveToSmall := fmt.Sprintf("%s/%s.mp4", smallVideoLocation, vid)
		imgs := fmt.Sprintf("%s/%s-*.jpg", imageLocation, vid)
		fmt.Printf("code: %s imgs: %s\n", vid, imgs)
		output, err := exec.Command("ffmpeg", "-framerate", "5", "-pattern_type", "glob", "-i", imgs, saveToFull).Output()
		log.Println(output)
		failOnError(err, "c")

		// Squash
		output, err = exec.Command("ffmpeg", "-i", saveToFull, "-vcodec", "libx265", "-crf", "45", saveToSmall).Output()
		log.Println(output)
		failOnError(err, "c")

		// Remove images
		imgLoc := fmt.Sprintf("%s/%s-*.jpg", imageLocation, strings.Replace(vid, " ", "\\ ", -1))
		files, err := filepath.Glob(imgLoc)
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				panic(err)
			}
		}
	}
}

//ffmpeg images to video
// ffmpeg -framerate 5 -pattern_type glob -i "*.jpg" out.mp4
//fmpeg squash video
// ffmpeg -i out.mp4 -vcodec libx265 -crf 45 squash.mp4
