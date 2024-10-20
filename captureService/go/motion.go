package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"gocv.io/x/gocv"
)

type buffered struct {
	Time   string
	Name   string
	Image  []byte
	Code   string
	Count  int64
	Blocks []image.Point
}

type inputImage struct {
	frame gocv.Mat
	image gocv.Mat
}

var (
	counter       = 0
	blocks        [][]image.Rectangle
	buffer        []buffered
	bufferCounter = 0
	code          = ""
	sendFrame     = 0
	sentBuffer    = false
	noMovement    = 0
	//send buffer
	network bytes.Buffer
	enc     *gob.Encoder
	//Colours
	red    = color.RGBA{255, 0, 0, 255}
	orange = color.RGBA{255, 153, 0, 255}
	green  = color.RGBA{51, 204, 51, 255}
	purple = color.RGBA{102, 0, 255, 255}
)

//Takes in images
//Delivers to channel when motion is detected
//OUT SHOULD BE BUFFERED STRUCT
func checkMotion(in chan inputImage, out chan gocv.Mat, setting *settings) {
	enc = gob.NewEncoder(&network)
	grayMap := gocv.NewMat()
	defer grayMap.Close()
	blurMap := gocv.NewMat()
	defer blurMap.Close()
	roiMap := gocv.NewMat()
	defer roiMap.Close()
	preRoiMap := gocv.NewMat()
	defer preRoiMap.Close()
	diffMap := gocv.NewMat()
	defer diffMap.Close()
	thresMap := gocv.NewMat()
	defer thresMap.Close()
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()

	preMap := gocv.NewMat()
	defer preMap.Close()

	for f := range in {
		//Set vars for this frame
		fameNum := time.Now().Unix()
		motion := false
		var boxesLocations []image.Point
		gocv.CvtColor(f.frame, &grayMap, gocv.ColorBGRToGray)
		gocv.GaussianBlur(grayMap, &blurMap, image.Pt(setting.Blur, setting.Blur), 0, 0, gocv.BorderDefault)

		//We require a prev frame to work
		if preMap.Empty() {
			code = string(time.Now().Unix())
			blurMap.CopyTo(&preMap)
			continue
		}

		// Do work on each of the ROI
		for i, zone := range setting.Zones {

			//Crop this roi
			roiMap = blurMap.Region(image.Rectangle{image.Point{zone.X1, zone.Y1}, image.Point{zone.X2, zone.Y2}})
			preRoiMap = preMap.Region(image.Rectangle{image.Point{zone.X1, zone.Y1}, image.Point{zone.X2, zone.Y2}})

			//Calculate the difference between the two frames
			gocv.AbsDiff(roiMap, preRoiMap, &diffMap)
			gocv.Threshold(diffMap, &thresMap, float32(zone.Threshold), 255, gocv.ThresholdBinary)
			gocv.Dilate(thresMap, &thresMap, kernel)

			//Find contours
			points := gocv.FindContours(thresMap, gocv.RetrievalExternal, gocv.ChainApproxSimple)

			//Setup our difference boxes
			var newBox []image.Rectangle
			noMove := false
			// Go through contours
			for _, contour := range points {
				area := gocv.ContourArea(contour)

				//If the area is too small, skip
				if area < float64(zone.Area) {
					continue
				}

				rect := gocv.BoundingRect(contour)
				newBox = append(newBox, rect)

				midX := rect.Min.X + rect.Dx()
				midY := rect.Min.Y + rect.Dy()
				x, y := findClosestBox(midX, midY, i)
				blocks[i] = newBox

				if len(blocks) == 0 || len(blocks[i]) == 0 {
					fmt.Printf("No blocks")
					// No prev boxes
					//Draw box in ORANGE
					if setting.Debug {
						gocv.Rectangle(&f.image, rect, orange, 2)
					}
					continue
				}

				if x > zone.BoxJump || y > zone.BoxJump {
					// Box is too far (large gap)
					// RED
					if setting.Debug {
						gocv.Rectangle(&f.image, rect, red, 2)
					}
					continue
				}
				if x < zone.SmallIgnore || y < zone.SmallIgnore {
					// Box moved too little
					noMove = true
					// PURPLE
					if setting.Debug {
						gocv.Rectangle(&f.image, rect, purple, 2)
					}
					continue
				}
				// Motion box
				// Green
				if setting.Debug {
					gocv.Rectangle(&f.image, rect, green, 2)
				}
				motion = true

				// Add boundary boxes to locations
				boxesLocations = append(boxesLocations, image.Point{midX, midY})
			}
			if !motion && noMove {
				noMovement++
			} else if motion && !noMove {
				noMovement = 0
			}
		}

		// Checked the image for motion
		// Now we deal with the counter
		if motion {
			counter++
			if counter > setting.MinCount {
				sendFrame = setting.BufferAfter
				counter = setting.MinCount
				sentBuffer = true
			}
		} else {
			counter--
			if counter < 0 {
				code = string(time.Now().Unix())
				counter = 0
				if sentBuffer {
					//send END
					sentBuffer = false
				}
			}
		}

		// Add this frame to the buffer
		bufImg, _ := gocv.IMEncodeWithParams(".jpg", f.image, []int{gocv.IMWriteJpegQuality, 80})
		buf := buffered{Time: string(time.Now().Unix()), Name: setting.Name,
			Image: bufImg, Count: fameNum, Blocks: boxesLocations}

		//We can't pre make the buffer as it can dynamically change
		//But we roundrobin the buffer
		if len(buffer) < setting.BufferBefore {
			buffer = append(buffer, buf)
		} else {
			buffer[bufferCounter] = buf
			bufferCounter++
			if bufferCounter > len(buffer)-1 {
				bufferCounter = 0
			}
		}

		if sendFrame > 0 {
			//Send the buffer
		}

		//No movement should cause a refresh of background
		if noMovement > setting.NoMoveRefreshCount {
			blurMap.CopyTo(&preMap)
			noMovement = 0
		}

		sendFrame--
		if sendFrame < 0 {
			sendFrame = 0
		}
	}

}

func sendBuffer(code string) {
	for _, b := range buffer {
		b.Code = code
		err := enc.Encode(b)
		if err != nil {
			log.Printf("Failed to gob encode %s", err)
			continue
		}
		motionStream <- network.Bytes()
	}
}

func findClosestBox(x int, y int, index int) (int, int) {
	difX := -1
	difY := -1
	for _, block := range blocks[index] {
		midX := block.Min.X + block.Dx()
		midY := block.Min.Y + block.Dy()
		dy := abs(midY - y)
		dx := abs(midX - x)
		if difX == 1 || difX < dx && difY < dy {
			difX = dx
			difY = dy
		}
	}
	return difX, difY

}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
