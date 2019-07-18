package main

import (
	"log"

	"gocv.io/x/gocv"
)

func main() {
	m := gocv.IMRead("T.png", gocv.IMReadAnyColor)

	var difference gocv.Mat
	gocv.AbsDiff(m, m, &difference)
	all := difference.ToBytes()
	var noneZero = 0
	for _, element := range all {
		if element != 0 {
			noneZero++
		}
	}

	log.Printf("None zereos... %d", noneZero)
}
