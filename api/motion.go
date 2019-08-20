package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

//DBName string = "/shared/motions.db"
const DBName string = "/mnt/shared/motion/motions.db"

//MotionJSON is the output for a motion
type MotionJSON struct {
	ID     int
	Code   string
	Reason string
	Name   string
}

//All motion handler
func allMotion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")

	rows, err := db.Query("select id,code,reason,name from video")
	failOnError(err, "prep failed")
	defer rows.Close()
	var full []MotionJSON
	for rows.Next() {
		var id int
		var code, reason, name string
		err = rows.Scan(&id, &code, &reason, &name)
		failOnError(err, "Failed to get")
		body := MotionJSON{id, code, reason, name}
		full = append(full, body)

	}
	b, err := json.Marshal(full)
	fmt.Printf("Json bytes are %s\n", b)
	w.Write(b)

}

//del motion handler
func delMotion(w http.ResponseWriter, r *http.Request) {
	print("Deleting motion")
	params := mux.Vars(r)

	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	tx, err := db.Begin()
	failOnError(err, "Failed to begin on record")
	stmt, err := tx.Prepare("DELETE FROM video WHERE code=?")
	failOnError(err, "Record sql prep failed")
	defer stmt.Close()
	_, err = stmt.Exec(params["code"])
	failOnError(err, "Record could not insert")
	tx.Commit()
	w.Write([]byte("Okay"))
}

//Get the single data
func getMotion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		params := mux.Vars(r)
		http.ServeFile(w, r, fmt.Sprintf("/mnt/shared/motion/videos/%s.mp4", params["code"]))
	} else if r.Method == "DELETE" {
		delMotion(w, r)
	}
}
