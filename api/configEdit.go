package main

import (
	"encoding/json"
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
