package queue

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	_, err = ch.QueueDeclare("notifications", true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}
	return conn, ch, nil
}

func Consume(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume("notifications", "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	log.Println("Worker listening on queue 'notifications'")
	return msgs, nil
}
