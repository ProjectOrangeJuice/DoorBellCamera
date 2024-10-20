package main

import (
	"log"
	"time"
)

func main() {
	const secToNano = time.Nanosecond * time.Second
	se := int64(5)
	wait := secToNano.Nanoseconds() / se
	log.Printf("Time %d", wait)
}
