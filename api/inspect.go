package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

type inspectS struct {
	Reason string
	Image  string
}

func getImage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	db, err := sql.Open("sqlite3", DBName)
	failOnError(err, "Record failed because of DB error")
	cmd := fmt.Sprintf("select reason from motion WHERE location = %s", params["location"])
	rows, err := db.Query(cmd)
	failOnError(err, "prep failed")
	defer rows.Close()
	if !rows.Next() {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var reason string
	err = rows.Scan(&reason)
	failOnError(err, "Failed to get")

	dat, err := ioutil.ReadFile(params["location"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Printf("Inspect failed for %s with error %s", params["location"], err)
		return
	}
	encoded := base64.StdEncoding.EncodeToString(dat)
	msg := inspectS{reason, encoded}
	bytes, err := json.Marshal(msg)
	failOnError(err, "Failed to convert json")
	w.Write(bytes)

}
