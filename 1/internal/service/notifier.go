package service

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	wbfredis "github.com/wb-go/wbf/redis"

	"gitlab.com/arkine/l3/1/internal/model"
	"gitlab.com/arkine/l3/1/internal/repo"
)

type NotificationMessage struct {
	ID int64 `json:"id"`
}

type Notifier struct {
	DB      *dbpg.DB
	Repo    *repo.Repo
	Cache   *wbfredis.Client
	Client  *rabbitmq.RabbitClient
	Senders map[string]Sender
}

func NewNotifier(
	db *dbpg.DB,
	r *repo.Repo,
	c *wbfredis.Client,
	client *rabbitmq.RabbitClient,
	senders map[string]Sender,
) *Notifier {
	return &Notifier{
		DB:      db,
		Repo:    r,
		Cache:   c,
		Client:  client,
		Senders: senders,
	}
}

func (n *Notifier) senderFor(channel string) Sender {
	if s, ok := n.Senders[channel]; ok {
		return s
	}
	return NewSimulatedSender()
}

func (n *Notifier) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var payload NotificationMessage
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return nil
	}

	notification, err := n.Repo.GetNotification(ctx, payload.ID)
	if err != nil || notification == nil {
		return nil
	}

	if notification.Status != model.StatusScheduled {
		return nil
	}

	sender := n.senderFor(notification.Channel)
	if err := sender.Send(ctx, notification.Recipient, notification.Message); err != nil {
		return err
	}

	return n.Repo.UpdateStatus(ctx, notification.ID, model.StatusSent)
}
