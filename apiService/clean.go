package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func clearLastMonth(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	monthAgo := t.AddDate(0, -1, 0)

	collection := conn.Collection("video")
	findOptions := options.Find()
	// Sort by
	findOptions.SetSort(bson.D{{"start", -1}})
	filter := bson.M{
		"start": bson.M{"$lt": strconv.FormatInt(int64(monthAgo.Unix()), 10)},
	}
	cur, err := collection.Find(context.TODO(), filter, findOptions)
	failOnError(err, "Failed to get video records")

	for cur.Next(context.TODO()) {
		var record videoRecord
		err := cur.Decode(&record)
		failOnError(err, "Failed to decode record")

		filter2 := bson.M{"code": record.Code}
		collection.DeleteOne(context.TODO(), filter2)
		err = os.Remove(fmt.Sprintf("%s/%s.mp4", videoLoc, record.Code))
		failOnError(err, "Failed to delete")
	}

}
