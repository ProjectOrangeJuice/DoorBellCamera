package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	failOnError(err, "Json error")
	w.Write([]byte(jsonout))

}

//POST for setconfig
func setConfig(w http.ResponseWriter, r *http.Request) {
	log.Print("On set")
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

	//setCommand(params["service"], string(body))
}
