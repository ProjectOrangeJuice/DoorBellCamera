package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	fmt "fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

type videoDB struct {
	Code  string
	Size  int64
	Image string
}

func recvMotionImg(buf chan *Buffer) {
	vidStream := make(chan string)
	go makeVideo(vidStream)
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
		setting := getSetting()

		saveToFull := fmt.Sprintf("%s/%s.mp4", fullVideoLocation, vid)
		saveToSmall := fmt.Sprintf("%s/%s.mp4", smallVideoLocation, vid)
		imgs := fmt.Sprintf("%s/%s-*.jpg", imageLocation, vid)
		fmt.Printf("code: %s imgs: %s\n", vid, imgs)
		output, err := exec.Command("ffmpeg", "-framerate", string(setting.FPS), "-pattern_type", "glob", "-i", imgs, saveToFull).Output()
		log.Println(output)
		failOnError(err, "c")

		// Squash
		output, err = exec.Command("ffmpeg", "-i", saveToFull, "-vcodec", "libx265", "-crf", "45", saveToSmall).Output()
		log.Println(output)
		failOnError(err, "c")

		// Generate thumbnail
		imgLoc := fmt.Sprintf("%s/%s-*.jpg", imageLocation, strings.Replace(vid, " ", "\\ ", -1))
		files, err := filepath.Glob(imgLoc)
		if err != nil {
			panic(err)
		}

		// Get the middle image
		mid := len(files) / 2
		existingImageFile, err := os.Open(files[mid])
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

		// Remove images
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				panic(err)
			}
		}

		// Get the video file size
		fi, err := os.Stat(saveToFull)
		if err != nil {
			fmt.Printf("error getting size %v\n", err)
		}
		// get the size
		size := fi.Size()

		// Update the database
		dbRecord := videoDB{vid, size, sEnc}
		conn := databaseClient.Database("doorbell")
		db := conn.Collection("video")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		db.InsertOne(ctx, dbRecord)
		cancel()
	}
}

//ffmpeg images to video
// ffmpeg -framerate 5 -pattern_type glob -i "*.jpg" out.mp4
//fmpeg squash video
// ffmpeg -i out.mp4 -vcodec libx265 -crf 45 squash.mp4
