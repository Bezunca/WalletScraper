package pubsub

import (
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

	go handle(deliveries, done)
	return nil
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	fmt.Println("New message")
	for d := range deliveries {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
