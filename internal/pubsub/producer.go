package pubsub

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func Publish(channel *amqp.Channel, exchangeName, queueName, body string, reliable bool) error {
	// Reliable publisher confirms require confirm.select support from the
	// connection.
	if reliable {
		if err := channel.Confirm(false); err != nil {
			return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
		}

		confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

		defer confirmOne(confirms)
	}
	if err := channel.Publish(
		exchangeName, // publish to an exchange
		queueName,    // routing to 0 or more queues
		true,         // mandatory
		false,        // immediate
		amqp.Publishing{
			//Headers:         amqp.Table{},
			ContentType: "text/plain",
			//ContentEncoding: "",
			Body: []byte(body),
			//DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			//Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return err
	}

	return nil
}

// One would typically keep a channel of publishings, a sequence number, and a
// set of unacknowledged sequence numbers and loop until the publishing channel
// is closed.
func confirmOne(confirms <-chan amqp.Confirmation) {
	if confirmed := <-confirms; confirmed.Ack {
		log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}
