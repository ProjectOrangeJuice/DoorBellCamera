package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type cameraSettings struct {
	Name       string
	Connection string
	FPS        int
	Area       [][]int
	Amount     []int
	Threshold  []int
	MinCount   []int
	Motion     bool
}

func makeVideo(code string, name string) {
	log.Println("Make video")
	collection := conn.Collection("settings")
	filter := bson.M{"_id": 0}
	doc := collection.FindOne(context.TODO(), filter)
	var settings cameraSettings
	doc.Decode(&settings)

	collection = conn.Collection("capture")
	filter = bson.M{"code": code}
	cur, _ := collection.Find(context.TODO(), filter)

	var fr []string
	var counter = 0
	var reasons string
	var startTime string
	var endTime string
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
		fr = append(fr, record.Location)
	}
	st := "images/" + code + "-%03d.jpg"
	output, err := exec.Command("ffmpeg", "-i", st, "-framerate", fmt.Sprintf("%d", settings.FPS), fmt.Sprintf("%s/%s.mp4", videoFolder, code)).Output()
	log.Println(output)
	failOnError(err, "c")
	for _, elem := range fr {
		err := os.Remove(elem)
		failOnError(err, "Failed to remove image")
	}
	addToDatabase(code, name, startTime, endTime, reasons)
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

	//log.Printf("Saved to db")

	//	_, err = db.Exec("DELETE FROM motion WHERE motionCode=?", code)
	//	failOnError(err, "Couldn't delete motion records")

}

func squashVideo(code string) {
	//output, err := exec.Command("ffmpeg", "-i", fmt.Sprintf("%s/%s.mp4", videoFolder, code), "-crf", "50", fmt.Sprintf("%s/mobile/%s.mp4", videoFolder, code)).Output()
	//failOnError(err, "FAiled to compress video")
	//log.Printf("I failed to make compressed video. %s", err)
	//log.Printf("Output was %s", output)
	//log.Print("Finished making mobile")
}
