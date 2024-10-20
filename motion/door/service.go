package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type config struct {
	ServerAddress string
	ServerPort    string
	maxThreshold  []int
	minThreshold  []int
	minCount      int
	quiet         int
	delayFailed   int
	Cameras       []pConfig
}

type pConfig struct {
	Name         string
	MaxThreshold []int
	MinThreshold []int
	MinCount     int
	Quiet        int
	DelayFailed  int
}

func main() {
	readConfig()
}

func readConfig() {
	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	failOnError(err, "Failed to read config")
	defer jsonFile.Close()
	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var configVals config
	json.Unmarshal(byteValue, &configVals)

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
