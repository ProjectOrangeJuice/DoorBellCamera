package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson"
)

type cameraSettings struct {
	Name       string
	Connection string
	FPS        int
}

func makeVideo(code string, name string) {
	log.Println("Make video")
	collection := conn.Collection("settings")
	filter := bson.M{"_id": name}
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
	st := CaptureLocation + "/" + code + "-*.jpg"
	output, err := exec.Command("ffmpeg", "-framerate", fmt.Sprintf("%d", (settings.FPS)), "-pattern_type", "glob", "-i", st, fmt.Sprintf("%s/%s.mp4", videoFolder, code)).Output()
	log.Println(output)
	failOnError(err, "c")

	thumbnail := getThumb(fr[counter/2])

	for _, elem := range fr {
		err := os.Remove(elem)
		failOnError(err, "Failed to remove image")
	}
	addToDatabase(code, name, startTime, endTime, reasons, thumbnail)
}

func getThumb(loc string) string {

	log.Printf("The addr value is %s", loc)
	existingImageFile, err := os.Open(loc)
	defer existingImageFile.Close()
	failOnError(err, "Failed to get image")

	image, _, err := image.Decode(existingImageFile)
	failOnError(err, "Failed to read image to compress")

	imageres := resize.Resize(160, 0, image, resize.Lanczos3)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, imageres, &jpeg.Options{25})
	failOnError(err, "Failed to encode")
	sends3 := buf.Bytes()

	sEnc := b64.StdEncoding.EncodeToString([]byte(sends3))

	return sEnc

}

type videoRecord struct {
	Code      string
	Name      string
	Start     string
	End       string
	Reason    string
	Thumbnail string
}

func addToDatabase(code string, name string, start string, end string, reason string, th string) {
	r := videoRecord{code, name, start, end, reason, fmt.Sprintf("data:image/jpg;base64, %s", th)}
	collection := conn.Collection("video")
	collection.InsertOne(context.TODO(), r)

	collection = conn.Collection("capture")
	filter := bson.M{"code": code}
	collection.DeleteMany(context.TODO(), filter)

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
