package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, group string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: group,
		}),
	}
}

func (c *Consumer) Read(ctx context.Context) (string, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return "", err
	}
	return string(m.Value), nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
