package main

import (
	"WalletScraper/internal/config"
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
	if err := pubsub.Publish(channel, configs.ExchangeName, configs.PubQueueName, body, true); err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("published %dB OK", len(body))
}
