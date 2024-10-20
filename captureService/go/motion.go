package main

import (
	"image"
	"time"

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
		gocv.CvtColor(f.frame, &grayMap, gocv.ColorBGRToGray)
		gocv.GaussianBlur(grayMap, &blurMap, image.Pt(setting.Blur, setting.Blur), 0, 0, gocv.BorderDefault)

		//We require a prev frame to work
		if preMap.Empty() {
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
			// Go through contours
			for index, contour := range points {
				area := gocv.ContourArea(contour)

				//If the area is too small, skip
				if area < float64(zone.Area) {
					continue
				}

				rect := gocv.BoundingRect(contour)
				newBox = append(newBox, rect)

			}
		}

	}

}
