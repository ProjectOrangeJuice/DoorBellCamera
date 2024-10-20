package main

import (
	fmt "fmt"
	"time"

	"github.com/golang/protobuf/proto"
)

type cameraStructure struct {
	prev        string
	notified    string
	ignoreTimer bool
}

var timer time.Timer

const imageLocation string = "../images"
const fullVideoLocation string = "../videos"
const smallVideoLocation string = "../videos/s"

func main() {
	//go http.ListenAndServe("localhost:8080", nil)
	mstream := make(chan *Buffer)
	go recvMotionImg(mstream)
	buf := &Buffer{}
	c := setupRabbit()
	for input := range c {
		fmt.Println("Got message")
		input.Ack(true)
		// Pass the message to our video creator
		err := proto.Unmarshal(input.Body, buf)
		if err != nil {
			fmt.Printf("Failed to unmash %v\n", err)
			continue
		}
		mstream <- buf // Must block before getting next image
		// Otherwise we will change the buffer while its working
	}
}
