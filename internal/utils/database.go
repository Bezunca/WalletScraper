package utils

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func ToDoc(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

func InsertOrUpdate(collection *mongo.Collection, models []mongo.WriteModel) (*mongo.BulkWriteResult, error){
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(len(models))*10*time.Second)
	defer cancel()
	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(ctx, models, opts)
	if err != nil {
		return nil, err
	}

	return result, nil
}