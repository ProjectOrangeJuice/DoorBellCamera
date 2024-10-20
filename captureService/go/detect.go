package main

import (
	"image"
	"log"

	"gocv.io/x/gocv"
)

type extras struct {
	Prev        gocv.BackgroundSubtractorMOG2
	PrevDefined bool
	Section     gocv.Mat
	Kernel      gocv.Mat
}

var extra *extras = &extras{gocv.NewBackgroundSubtractorMOG2(), false, gocv.NewMat(), gocv.NewMat()}

func setupMotion() {
	extra.Kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
}

func detect(frame gocv.Mat) {
	// //frameNum := time.Now().UnixNano()

	for index, zone := range setting.Area {
		log.Printf("%v %v", index, zone)
		// 	//crop our image
		extra.Section = frame.Region(image.Rect(zone[2], zone[0], zone[3], zone[1]))

		// 	//Pretend we only have one zone
		extra.Prev.Apply(extra.Section, &extra.Section)

		// 	// remaining cleanup of the image to use for finding contours.
		// 	// first use threshold
		// 	gocv.Threshold(extra.Section, &extra.Section, float32(setting.Threshold[index]), 255, gocv.ThresholdBinary)

		// 	// then dilate

		// 	gocv.Dilate(extra.Section, &extra.Section, extra.Kernel)

		// 	// now find contours
		// 	// contours := gocv.FindContours(extra.Section, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		// 	// for _, c := range contours {
		// 	// 	area := gocv.ContourArea(c)
		// 	// 	if area < float64(setting.Amount[index]) {
		// 	// 		continue
		// 	// 	}

		// 	// 	rect := gocv.BoundingRect(c)
		// 	// 	gocv.Rectangle(&frame, rect, color.RGBA{0, 0, 255, 0}, 2)
		// 	// }

		// 	//sendFrame(frame, ch)

	}

	//log.Printf("Motion? %v", motion)
}
