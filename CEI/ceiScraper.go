package main

import (
	"WalletScraper/internal/config"
	"WalletScraper/internal/processing"
	"WalletScraper/internal/pubsub"
	"fmt"
	"log"
)

func main() {
	config.New()
	body := "teste show pa"
	configs := config.Get()

	channel, err := pubsub.GetChannel(
		configs.QueueAddress("amqps"),
		configs.ExchangeName,
		configs.PubQueueName,
		configs.SubQueueName,
		"direct",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connection established")

	err = pubsub.Listen(channel, configs.SubQueueName, processing.ScrapeCEI)
	select {}
}
