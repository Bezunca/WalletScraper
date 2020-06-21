package pubsub

import (
	"WalletScraper/internal/config"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func Listen(channel *amqp.Channel, queueName, body string) error {
	fmt.Println("setup listener")
	fmt.Println(queueName)

	deliveries, err := channel.Consume(
		queueName,         // name
		"simple-consumer", // consumerTag,
		false,             // noAck
		false,             // exclusive
		false,             // noLocal
		false,             // noWait
		nil,               // arguments
	)
	if err != nil {
		return err
	}

	done := make(chan error)

	go handle(channel, deliveries, done)
	return nil
}

func handle(channel *amqp.Channel, deliveries <-chan amqp.Delivery, done chan error) {
	configs := config.Get()
	fmt.Println("New message")
	for d := range deliveries {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		d.Ack(false)

		if err := Publish(channel, configs.ExchangeName, configs.PubQueueName, string(d.Body[:]), true); err != nil {
			log.Fatalf("%s", err)
		}
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
