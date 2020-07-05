package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func GetLastUpdateTime(collection mongo.Collection, filter bson.D)(*time.Time, error){
	var data map[string]interface{}
	err := collection.FindOne(
		context.Background(), filter, options.FindOne().SetSort(map[string]int{"data.date": -1}),
	).Decode(&data)
	if err != nil {
		if err.Error() == "mongo: no documents in result"{
			return nil, nil
		}
		return nil, err
	}
	_entry, ok := data["data"]
	if !ok {
		return nil, fmt.Errorf("invalid date")
	}
	pEntry, ok := _entry.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid date")
	}
	_date, ok := pEntry["date"]
	if !ok {
		return nil, fmt.Errorf("invalid date")
	}
	pDate, ok := _date.(primitive.DateTime)
	if !ok {
		return nil, fmt.Errorf("invalid date")
	}
	date := pDate.Time()

	return &date, nil
}