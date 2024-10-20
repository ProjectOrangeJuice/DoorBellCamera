package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Get motions ( pagination )

// Delete motion

// Get motion
const videoLoc = "../storage/videos"

func getHQVideo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("Get video for %s, requested %s", r.RemoteAddr, params["code"])
	http.ServeFile(w, r, fmt.Sprintf("%s/%s.mp4", videoLoc, params["code"]))

}

func getLQVideo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Printf("Get video for %s, requested %s", r.RemoteAddr, params["code"])
	http.ServeFile(w, r, fmt.Sprintf("%s/s/%s.mp4", videoLoc, params["code"]))

}
