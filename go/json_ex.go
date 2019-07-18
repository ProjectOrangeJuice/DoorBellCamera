// Go offers built-in support for JSON encoding and
// decoding, including to and from built-in and custom
// data types.

package main

import (
	"encoding/json"
	"log"
)

type Message struct {
	Image string
	Time  float32
}

func main() {
	b := []byte(`{"time": 1563397240.9809768, "image": "b64.decode('utf-8')"}`)
	var m Message
	err := json.Unmarshal(b, &m)
	log.Printf("error %s or json %s", err, m)
}
