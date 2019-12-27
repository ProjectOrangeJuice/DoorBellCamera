package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"gocv.io/x/gocv"
)

func makeVideo(code string, name string) {
	var startTime string
	var endTime string
	video, err := gocv.VideoWriterFile(fmt.Sprintf("%s/%s.mp4", videoFolder, code), "avc1", 5.0, 1280, 720, true)

	//aw, err := mjpeg.New(fmt.Sprintf("%s/%s", videoFolder, m.Code), 1280, 720, 10)
	failOnError(err, "Setting up video")

	collection := conn.Collection("capture")
	filter := bson.M{"code": code}
	cur, _ := collection.Find(context.TODO(), filter)

	var fr []string
	var counter = 0
	var reasons string
	for cur.Next(context.TODO()) {
		counter++
		var record dbRecord
		err := cur.Decode(&record)
		failOnError(err, "decoding record")
		if reasons == "" {
			reasons = record.Reason
		}
		sp := strings.Split(record.Reason, ",")
		if len(sp) > 1 {
			for _, v := range sp {
				spr := strings.Split(reasons, ",")
				found := false
				for _, v2 := range spr {
					if v2 == v {
						found = true
					}
				}
				if !found {
					if reasons == "" {
						reasons = v
					} else {
						reasons = reasons + "," + v
					}
				}
			}
		}
		if startTime == "" {
			startTime = record.Time
		} else {
			endTime = record.Time
		}
		video.Write(gocv.IMRead(fmt.Sprintf("%s", record.Location), gocv.IMReadAnyColor))

	}

	//err = aw.Close()
	//failOnError(err, "Error closing")
	video.Close()
	for _, elem := range fr {
		err = os.Remove(elem)
		failOnError(err, "Failed to remove image")
	}

	log.Printf("Start time %s and end time %s", startTime, endTime)
	addToDatabase(code, name, startTime, endTime, reasons)
	squashVideo(code)
}

type videoRecord struct {
	Code   string
	Name   string
	Start  string
	End    string
	Reason string
}

func addToDatabase(code string, name string, start string, end string, reason string) {
	r := videoRecord{code, name, start, end, reason}
	collection := conn.Collection("video")
	collection.InsertOne(context.TODO(), r)

	log.Printf("Saved to db")

	//	_, err = db.Exec("DELETE FROM motion WHERE motionCode=?", code)
	//	failOnError(err, "Couldn't delete motion records")

}

func squashVideo(code string) {
	output, err := exec.Command("ffmpeg", "-i", fmt.Sprintf("%s/%s.mp4", videoFolder, code), "-crf", "50", fmt.Sprintf("%s/mobile/%s.mp4", videoFolder, code)).Output()
	//failOnError(err, "FAiled to compress video")
	log.Printf("I failed to make compressed video. %s", err)
	log.Printf("Output was %s", output)
	log.Print("Finished making mobile")
}
