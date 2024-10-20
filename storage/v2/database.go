package main

import (
	"context"
	fmt "fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func configDB(ctx context.Context) (*mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s", "localhost")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to mongo: %v", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("mongo client couldn't connect with background context: %v", err)
	}
	todoDB := client.Database("setting")
	return todoDB, nil
}
