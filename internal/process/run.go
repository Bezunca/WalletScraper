package process

import (
	"WalletScraper/internal/cei"
	"WalletScraper/internal/config"
	"WalletScraper/internal/models"
	"WalletScraper/internal/rabbitmq"
	internalRSA "WalletScraper/internal/rsa"
	"WalletScraper/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	ceiModels "github.com/Bezunca/ceilib/models"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"time"
)

func getRMQMessage(message amqp.Delivery)(*models.Wallets, error){
	ceiRequest := models.Wallets{}
	if err := json.Unmarshal(message.Body, &ceiRequest); err != nil {
		return nil, err
	}

	if ceiRequest.WalletsCredentials.CEI == nil{
		return nil, fmt.Errorf("cannot parse received data")
	}

	return &ceiRequest, nil
}

func reject(message amqp.Delivery, errorLog *log.Logger, err error)  {
	errorLog.Print(err)
	if err := message.Reject(false); err != nil {
		errorLog.Println(err)
	}
}

func nack(message amqp.Delivery, errorLog *log.Logger, err error)  {
	errorLog.Print(err)
	if err := message.Nack(false, true); err != nil {
		errorLog.Println(err)
	}
}

func getCEICredentials(ceiWalletsData *models.Wallets)(*ceiModels.CEI, error){
	configs := config.Get()

	password, err := internalRSA.Decrypt(
		configs.RSAPasswordEncryptionKey, ceiWalletsData.WalletsCredentials.CEI.Password, "cei_password",
	)
	if err != nil{
		return nil, err
	}

	ceiCredentials := ceiModels.CEI{
		User:     ceiWalletsData.WalletsCredentials.CEI.User,
		Password: password,
	}

	return &ceiCredentials, nil
}

func scrape(
	message amqp.Delivery, mongoClient *mongo.Client, ceiCredentials ceiModels.CEI, userID primitive.ObjectID,
	errorLog *log.Logger,
)(*[]mongo.WriteModel, []interface{}, *[]mongo.WriteModel, error){
	dividendModels, err := cei.ScrapeDividends(ceiCredentials, userID)
	if err != nil{
		reject(message, errorLog, err)
		return nil, nil, nil, err
	}
	trades, err := cei.ScrapeTrades(mongoClient, ceiCredentials, userID)
	if err != nil{
		reject(message, errorLog, err)
		return nil, nil, nil, err
	}
	portfolioModels, err := cei.ScrapePortfolio(ceiCredentials, userID)
	if err != nil{
		reject(message, errorLog, err)
		return nil, nil, nil, err
	}
	return dividendModels, trades, portfolioModels, nil
}

func saveData(
	message amqp.Delivery, mongoClient *mongo.Client, dividendModels *[]mongo.WriteModel, trades []interface{},
	portfolioModels *[]mongo.WriteModel, errorLog *log.Logger,
) error{
	configs := config.Get()

	dividendCollection := mongoClient.Database(configs.ApplicationDatabase).Collection("user_dividends")
	tradeCollection := mongoClient.Database(configs.ApplicationDatabase).Collection("user_trades")
	portfolioCollection := mongoClient.Database(configs.ApplicationDatabase).Collection("user_portfolio")

	if dividendModels != nil && len(*dividendModels) > 0 {
		_, err := utils.InsertOrUpdate(dividendCollection, *dividendModels)
		if err != nil {
			nack(message, errorLog, err)
			return err
		}
	}
	if trades != nil && len(trades) > 0{
		_, err := tradeCollection.InsertMany(context.Background(), trades)
		if err != nil {
			nack(message, errorLog, err)
			return err
		}
	}
	if portfolioModels != nil && len(*portfolioModels) > 0 {
		_, err := utils.InsertOrUpdate(portfolioCollection, *portfolioModels)
		if err != nil {
			nack(message, errorLog, err)
			return err
		}
	}

	return nil
}

func Run(rmq *rabbitmq.Session, mongoClient *mongo.Client) {
	errorLog := log.New(os.Stderr, "ERROR - ", log.LUTC|log.Ldate|log.Lmsgprefix|log.Ltime)

	for {
		stream, err := rmq.Stream()
		if err != nil {
			errorLog.Printf("Error on opening RabbitMQ consumer, retrying...\nError: %v", err)
			<-time.After(1 * time.Second)
			continue
		}

		log.Println("Listening...")
		for message := range stream {
			ceiRequest, err := getRMQMessage(message)
			if err != nil{
				reject(message, errorLog, err)
				continue
			}
			userID, err := primitive.ObjectIDFromHex(ceiRequest.ID)
			if err != nil{
				reject(message, errorLog, err)
				continue
			}

			log.Printf("Starting Data Capture For User %s", ceiRequest.WalletsCredentials.CEI.User)
			ceiCredentials, err := getCEICredentials(ceiRequest)
			if err != nil{
				reject(message, errorLog, err)
				continue
			}

			dividendModels, trades, portfolioModels, err := scrape(
				message, mongoClient, *ceiCredentials, userID, errorLog,
			)
			if err != nil{
				nack(message, errorLog, err)
				continue
			}
			err = saveData(message, mongoClient, dividendModels, trades, portfolioModels, errorLog)
			if err != nil{
				nack(message, errorLog, err)
				continue
			}

			_bakingMsg := models.Baking{UserID: userID}
			bakingMsg, err := json.Marshal(_bakingMsg)
			if err != nil{
				nack(message, errorLog, err)
				continue
			}
			err = rmq.Push(bakingMsg)
			if err != nil{
				nack(message, errorLog, err)
				continue
			}

			if err := message.Ack(false); err != nil {
				errorLog.Println(err)
				continue
			}
			log.Printf("SUCESSFULLY Scraped User %s", ceiRequest.WalletsCredentials.CEI)
		}

		errorLog.Println("An error occurred, restarting processing")
	}
}
