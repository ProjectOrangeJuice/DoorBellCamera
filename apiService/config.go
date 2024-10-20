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
	Name               string //`bson:"_id"`
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
	// Zones              []zone
}

type newSettings struct {
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
	Area        int
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	camName := params["cam"]
	log.Print("get config")
	collection := conn.Collection("settings")
	filter := bson.M{"_id": camName}
	doc := collection.FindOne(context.TODO(), filter)
	var settings newSettings
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
	var settings newSettings
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
	log.Printf("Can get collection")
	cur, err := collection.Find(context.TODO(), bson.D{})
	failOnError(err, "Failed to get setting records")
	var cams []info
	for cur.Next(context.TODO()) {
		var setting newSettings
		err := cur.Decode(&setting)
		failOnError(err, "Failed to decode setting")
		log.Printf("going to get alerts for %s", setting.Name)
		l, a := getAlertsInfo(setting.Name)
		i := info{setting.Name, l, a}
		cams = append(cams, i)
	}

	json.NewEncoder(w).Encode(cams)

}

func getAlertsInfo(name string) (last string, alerts int) {
	collection := conn.Collection("video")
	log.Printf("Can get collection for video")
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

//Create camera profile
func createProfile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	camName := params["cam"]
	collection := conn.Collection("settings")

	profile := newSettings{
		Name: camName,
		// "",
		FPS: 13,
		// [[]],
		// [],
		// [],
		// [],
		// false,
		// 21,
		// 5,
		// true,
		// 10,
		// 10,
		// 3,
		// 5,
		// []
	}

	filter := bson.M{"_id": camName}
	collection.FindOneAndReplace(context.TODO(), filter, profile, options.FindOneAndReplace().SetUpsert(true))
}

//delete camera profile
//Doesn't remove videos from this profile!!
func deleteProfile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	camName := params["cam"]
	collection := conn.Collection("settings")

	filter := bson.M{"_id": camName}
	collection.DeleteMany(context.TODO(), filter)
}
