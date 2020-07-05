package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Baking struct {
	UserID                  primitive.ObjectID             `json:"user_id" bson:"user_id"`
}
