package main

import (
	"WalletScraper/internal"
	"WalletScraper/internal/config"
	"log"
)

func main() {
	config.New()
	body := "teste show pa"
	configs := config.Get()
	if err := internal.Publish(
		configs.QueueAddress("amqps"),
		configs.ExchangeName,
		"direct",
		configs.QueueName,
		body,
		true,
	); err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("published %dB OK", len(body))
}
