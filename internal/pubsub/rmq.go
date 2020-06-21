package pubsub

import (
	"crypto/tls"

	"github.com/streadway/amqp"
)

func GetChannel(amqpURI, exchangeName, pubQueueName, subQueueName, exchangeType string) (*amqp.Channel, error) {
	connection, err := amqp.DialTLS(
		amqpURI,
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return nil, err
	}
	//defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, err
	}

	pubQueue, err := channel.QueueDeclare(
		pubQueueName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)

	if err = channel.QueueBind(
		pubQueue.Name, // name of the queue
		pubQueueName,  // bindingKey
		exchangeName,  // sourceExchange
		false,         // noWait
		nil,           // arguments
	); err != nil {
		return nil, err
	}

	//subQueue, err := channel.QueueDeclare(
	//	subQueueName, // name
	//	true,         // durable
	//	false,        // delete when unused
	//	false,        // exclusive
	//	false,        // no-wait
	//	nil,          // arguments
	//)
	//
	//if err = channel.QueueBind(
	//	subQueue.Name, // name of the queue
	//	subQueueName,  // bindingKey
	//	exchangeName,  // sourceExchange
	//	false,         // noWait
	//	nil,           // arguments
	//); err != nil {
	//	return nil, err
	//}

	return channel, nil
}
