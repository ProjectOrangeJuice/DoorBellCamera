package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cameraSettings struct {
	Name               string `bson:"_id"`
	Connection         string
	FPS                int
	Area               [][]int
	Amount             []int
	Threshold          []int
	MinCount           []int
	Motion             bool
	Blur               int
	BoxJump            int
	Debug              bool
	BufferBefore       int
	BufferAfter        int
	NoMoveRefreshCount int
	SmallMove          int
	Zones              []zone
}

type zone struct {
	X1          int
	Y1          int
	X2          int
	Y2          int
	Threshold   int
	MinCount    int
	BoxJump     int
	SmallIgnore int
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	camName := params["cam"]
	log.Print("get config")
	collection := conn.Collection("settings")
	filter := bson.M{"_id": camName}
	doc := collection.FindOne(context.TODO(), filter)
	var settings cameraSettings
	err := doc.Decode(&settings)
	if err != nil {
		//Content not found
	}
	log.Printf("%s", settings)
	json.NewEncoder(w).Encode(settings)

}

func setConfig(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	camName := params["cam"]
	decoder := json.NewDecoder(r.Body)
	var settings cameraSettings
	err := decoder.Decode(&settings)
	failOnError(err, "decode new settings")
	settings.Name = camName
	log.Printf("Set config %v", settings)
	collection := conn.Collection("settings")
	filter := bson.M{"_id": camName}
	collection.FindOneAndReplace(context.TODO(), filter, settings, options.FindOneAndReplace().SetUpsert(true))
}

type info struct {
	Name      string
	LastAlert string
	Alerts24  int
}

//Get information about cameras on this system
func getInformation(w http.ResponseWriter, r *http.Request) {
	log.Print("get camera information")
	collection := conn.Collection("settings")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.M{}, findOptions)
	failOnError(err, "Failed to get setting records")
	var cams []info
	for cur.Next(context.TODO()) {
		var setting cameraSettings
		err := cur.Decode(&setting)
		failOnError(err, "Failed to decode setting")
		l, a := getAlertsInfo(setting.Name)
		i := info{setting.Name, l, a}
		cams = append(cams, i)
	}

	json.NewEncoder(w).Encode(cams)

}

func getAlertsInfo(name string) (last string, alerts int) {
	collection := conn.Collection("video")
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"start", -1}})
	t := time.Now()
	t = t.Add(time.Duration(-24) * time.Hour)
	stamp := t.Unix()
	filter := bson.M{"name": name,
		"start": bson.M{"$gt": strconv.FormatInt(stamp, 10)}}
	cur, err := collection.Find(context.TODO(), filter, findOptions)

	failOnError(err, "Failed to get video records")

	total := 0
	lastStamp := ""
	for cur.Next(context.TODO()) {
		if total == 0 {
			var record videoRecord
			cur.Decode(&record)
			lastStamp = record.Start
		}
		total++
	}
	return lastStamp, total
}
