package models

import (
	"github.com/Bezunca/ceilib/scraper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WalletsCredentials struct {
	CEI *CEI `json:"cei" bson:"cei"`
}

type Wallets struct {
	ID                 string             `json:"user_id" bson:"user_id"`
	WalletsCredentials WalletsCredentials `json:"wallets_credentials" bson:"wallets_credentials"`
}

type MongoDividends struct {
	scraper.DividendStats                                  `json:"data" bson:"data"`
	UserID                  primitive.ObjectID             `json:"user_id" bson:"user_id"`
}

type MongoTrades struct {
	scraper.Trade                                          `json:"data" bson:"data"`
	UserID                  primitive.ObjectID             `json:"user_id" bson:"user_id"`
}

type MongoPortfolio struct {
	scraper.Asset                                          `json:"data" bson:"data"`
	UserID                  primitive.ObjectID             `json:"user_id" bson:"user_id"`
}
