package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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
	Time   string
}

//All motion handler
func allMotion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var from, to string
	params := r.URL.Query()
	ton, ok1 := params["to"]

	fromn, ok2 := params["from"]

	t := time.Now()
	if !ok1 {
		nt := t.AddDate(0, 0, 1)
		to = nt.Format("2006-01-02 15:04:05.000000")
	} else {
		to = ton[0]
	}
	if !ok2 {
		nt := t.AddDate(0, 0, -1)
		from = nt.Format("2006-01-02 15:04:05.000000")
	} else {
		from = fromn[0]
	}
	cmd := fmt.Sprintf("select id,code,reason,name,startTime from video WHERE startTime BETWEEN '%s' AND '%s'", from, to)
	log.Printf("Using %s", cmd)
	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	rows, err := db.Query(cmd)
	failOnError(err, "prep failed")
	defer rows.Close()
	var full []MotionJSON
	for rows.Next() {
		var id int
		var code, reason, name, st string
		err = rows.Scan(&id, &code, &reason, &name, &st)
		failOnError(err, "Failed to get")
		body := MotionJSON{id, code, reason, name, st}
		full = append(full, body)

	}
	b, err := json.Marshal(full)
	fmt.Printf("Json bytes are %s\n", b)
	logger.Printf("Get motion for %s. With the range of %s and %s", r.RemoteAddr, to, from)
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
	logger.Printf("Deleted motion for %s which matched %s", r.RemoteAddr, params["code"])
}

//Get the single data
func getMotion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		params := mux.Vars(r)
		logger.Printf("Get video for %s, requested %s", r.RemoteAddr, params["code"])
		http.ServeFile(w, r, fmt.Sprintf("/mnt/shared/motion/videos/%s.mp4", params["code"]))
	} else if r.Method == "DELETE" {
		delMotion(w, r)
	}
}

type keySt struct {
	Code string
}

func addDoorKey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var m keySt
	err := decoder.Decode(&m)
	failOnError(err, "Couldn't decode doorkey message")
	username := getUser(r)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, passdb, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	failOnError(err, "Database opening error")
	defer db.Close()
	sqlStatement := `INSERT INTO keys(user,key) VALUES($1,$2)`
	_, err = db.Exec(sqlStatement, username, m.Code)
	failOnError(err, "Query error")

}
