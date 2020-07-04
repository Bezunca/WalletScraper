package process

import (
	"WalletScraper/internal/models"
	"WalletScraper/internal/rabbitmq"
	"encoding/json"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

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
			log.Println("Received Message!")

			ceiRequest := models.Scraping{}
			if err := json.Unmarshal(message.Body, &ceiRequest); err != nil {
				errorLog.Println(err)
			}

			if err := message.Ack(true); err != nil {
				errorLog.Println(err)
			}
		}

		errorLog.Println("An error occured, restarting processing")
	}
}
