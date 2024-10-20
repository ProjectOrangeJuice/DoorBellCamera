package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type outMessage struct {
	Task  string
	Inner string
}

//GET for getconfig
func getConfig(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	output := client.HGetAll(params["service"])
	o2, _ := output.Result()
	jsonout, err := json.Marshal(o2)
	logger.Printf("Getting config for %s. Asked by %s and returned %s", params["service"], r.RemoteAddr, jsonout)
	failOnError(err, "Json error")
	w.Write([]byte(jsonout))

}

//POST for setconfig
func setConfig(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	failOnError(err, "failed to read body")
	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(body, &objmap)
	failOnError(err, "Json error")
	for index, key := range objmap {
		var new string

		new = strings.Replace(string(*key), string('"'), "", -1)

		client.HSet(params["service"], index, new)
	}

	logger.Printf("Set config for %s. Set by %s with %v", params["service"], r.RemoteAddr, objmap)
}

type userStruct struct {
	user     string
	password string
}

func makeUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var up userStruct
	err := decoder.Decode(&up)
	failOnError(err, "failed to decode userstruct")
	username := getUser(r)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, passdb, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	failOnError(err, "Database opening error")
	defer db.Close()
	sqlStatement := `INSERT INTO public.accounts ("user", password)
	SELECT $1,$1
	WHERE
	NOT EXISTS ( SELECT user FROM public.accounts WHERE "user"=$1);
	`
	results, err := db.Exec(sqlStatement, up.user, up.password)
	failOnError(err, "Query error")
	ra, err := results.RowsAffected()
	failOnError(err, "Failed to get rows")
	if ra > 0 {
		logger.Printf("created %s by %s", up.user, username)
		w.WriteHeader(200)
	} else {
		w.WriteHeader(409)
	}

}
