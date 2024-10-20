package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gocv.io/x/gocv"
)

func makeVideo(code string, name string) {
	var startTime string
	var endTime string
	video, err := gocv.VideoWriterFile(fmt.Sprintf("%s/%s.mp4", videoFolder, code), "avc1", 5.0, 1280, 720, true)

	//aw, err := mjpeg.New(fmt.Sprintf("%s/%s", videoFolder, m.Code), 1280, 720, 10)
	failOnError(err, "Setting up video")

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	rows, err := db.Query("select location,time,reason from motion where motionCode = ?", code)
	failOnError(err, "prep failed")
	defer rows.Close()

	var fr []string
	var counter = 0
	var reasons string
	for rows.Next() {
		counter++
		var location string
		var time string
		var reason string
		err = rows.Scan(&location, &time, &reason)
		failOnError(err, "Failed to get")
		sp := strings.Split(reason, ",")
		if len(sp) > 1 {
			for _, v := range sp {
				spr := strings.Split(reasons, ",")
				found := false
				for _, v2 := range spr {
					log.Printf("Checking if %s is %s", v2, v)
					if v2 == v {
						found = true
					}
				}
				if !found {
					if reasons == "" {
						reasons = v
					} else {
						reasons = reasons + "," + v
					}
				}
			}
		}
		if startTime == "" {
			startTime = time
		} else {
			endTime = time
		}
		video.Write(gocv.IMRead(fmt.Sprintf("%s", location), gocv.IMReadAnyColor))

	}

	//err = aw.Close()
	//failOnError(err, "Error closing")
	video.Close()
	for _, elem := range fr {
		err = os.Remove(elem)
		failOnError(err, "Failed to remove image")
	}

	log.Printf("Start time %s and end time %s", startTime, endTime)
	addToDatabase(code, name, startTime, endTime, reasons)
	squashVideo(code)
}

func addToDatabase(code string, name string, start string, end string, reason string) {

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	defer db.Close()
	tx, err := db.Begin()
	failOnError(err, "Failed to begin on record")
	stmt, err := tx.Prepare("insert into video(code,name, startTime,endTime ,reason) values(?,?,?,?,?)")
	failOnError(err, "Record sql prep failed")
	defer stmt.Close()
	_, err = stmt.Exec(code, name, start, end, reason)
	failOnError(err, "Record could not insert")
	tx.Commit()
	log.Printf("Saved to db")

	//	_, err = db.Exec("DELETE FROM motion WHERE motionCode=?", code)
	//	failOnError(err, "Couldn't delete motion records")

}

func squashVideo(code string) {
	cmds := exec.Command("ffmpeg", "-i", fmt.Sprintf("%s/%s.mp4", videoFolder, code), "-crf", "50", fmt.Sprintf("%s/mobile/%s.mp4", videoFolder, code))
	err := cmds.Run()
	failOnError(err, "FAiled to compress video")
	log.Print("Finished making mobile")
}
