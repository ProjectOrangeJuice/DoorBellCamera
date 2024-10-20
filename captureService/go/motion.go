package main

import (
	"image"
	"log"

	"gocv.io/x/gocv"
)

type Buffered struct {
	Time      string
	Name      string
	Image     []byte
	Code      string
	Count     int
	Blocks    string
	Locations string
}

type InputImage struct {
	frame gocv.Mat
	image gocv.Mat
}

//Takes in images
//Delivers to channel when motion is detected
//OUT SHOULD BE BUFFERED STRUCT
func checkMotion(in chan InputImage, out chan gocv.Mat, setting *settings) {

	grayMap := gocv.NewMat()
	defer grayMap.Close()
	blurMap := gocv.NewMat()
	defer blurMap.Close()

	for f := range in {
		log.Printf("GOT IMAGE")
		//Set vars for this frame
		// fameNum := time.Now().Unix()
		// motion := false
		gocv.CvtColor(f.frame, &grayMap, gocv.ColorBGRToGray)
		gocv.GaussianBlur(grayMap, &blurMap, image.Pt(21, 21), 0, 0, gocv.BorderDefault)
		log.Printf("ABOUT TO SEND")
		liveStream <- blurMap
		log.Printf("SENT")

	}

}
