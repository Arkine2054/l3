package queue

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(url string) (*amqp.Connection, *amqp.Channel, error) {
	if url == "" {
		return nil, nil, fmt.Errorf("RABBIT_URL is empty")
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("amqp.Dial failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		err := conn.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("conn.Channel failed: %w", err)
		}
		return nil, nil, fmt.Errorf("channel creation failed: %w", err)
	}

	_, err = ch.QueueDeclare(
		"notifications",
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		err := ch.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("ch.Close failed: %w", err)
		}
		err = conn.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("conn.Close failed: %w", err)
		}

		return nil, nil, fmt.Errorf("queue declare failed: %w", err)
	}

	log.Println("RabbitMQ connected and queue 'notifications' declared")
	return conn, ch, nil
}

func Consume(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		"notifications", // queue
		"",              // consumer
		false,           // autoAck
		false,           // exclusive
		false,           // noLocal
		false,           // noWait
		nil,             // args
	)
	if err != nil {
		return nil, fmt.Errorf("consume failed: %w", err)
	}

	log.Println("Worker listening on queue 'notifications'")
	return msgs, nil
}
