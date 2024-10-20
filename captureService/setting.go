package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type settings struct {
	Name               string
	Connection         string
	FPS                int
	MinCount           int
	Motion             bool
	Blur               int
	Debug              bool
	BufferBefore       int
	BufferAfter        int
	NoMoveRefreshCount int
	Zones              []zone
}
type zone struct {
	X1          int
	Y1          int
	X2          int
	Y2          int
	Threshold   int
	BoxJump     int
	SmallIgnore int
	Area        int
}

func getSetting() settings {
	conn := databaseClient.Database("doorbell")
	db := conn.Collection("setting")
	filter := bson.M{"_id": 0}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	doc := db.FindOne(ctx, filter)
	cancel()
	var s settings
	doc.Decode(&s)
	return s
}

func genTestSetting() settings {
	z := zone{10, 10, 500, 500, 20, 400, 2, 50}
	zo := make([]zone, 1)
	zo[0] = z
	s := settings{"test", "", 5, 3, true, 21, true, 5, 5, 3, zo}
	return s
}
