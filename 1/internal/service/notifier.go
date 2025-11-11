package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.com/arkine/l3/1/internal/cache"
	"gitlab.com/arkine/l3/1/internal/model"
	"gitlab.com/arkine/l3/1/internal/repo"
)

type NotificationMessage struct {
	ID int64 `json:"id"`
}

type Notifier struct {
	DB        *sql.DB
	Repo      *repo.Repo
	Cache     *cache.Client
	PublishCh *amqp.Channel
	Senders   map[string]Sender
}

func NewNotifier(db *sql.DB, r *repo.Repo, c *cache.Client, pubCh *amqp.Channel, senders map[string]Sender) *Notifier {
	return &Notifier{DB: db, Repo: r, Cache: c, PublishCh: pubCh, Senders: senders}
}

func (n *Notifier) senderFor(channel string) Sender {
	if s, ok := n.Senders[channel]; ok {
		return s
	}
	return NewSimulatedSender()
}

func (n *Notifier) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered while handling message: %v", r)
		}
	}()

	var m NotificationMessage
	if err := json.Unmarshal(msg.Body, &m); err != nil {
		log.Printf("bad message format: %v", err)
		if ackErr := msg.Ack(false); ackErr != nil {
			log.Printf("ack failed after bad format: %v", ackErr)
		}
		return nil
	}

	not, err := n.Repo.GetNotification(ctx, m.ID)
	if err != nil {
		log.Printf("repo get error for id=%d: %v", m.ID, err)
		if ackErr := msg.Nack(false, true); ackErr != nil {
			log.Printf("nack (requeue) failed for id=%d: %v", m.ID, ackErr)
		}
		return nil
	}

	log.Printf("Notification loaded: id=%d, recipient=%s, channel='%s', send_at=%s, status=%s",
		not.ID, not.Recipient, not.Channel, not.SendAt.Format(time.RFC3339), not.Status)

	if not.Status == model.StatusCancelled {
		log.Printf("skipping cancelled notification id=%d", not.ID)
		if err := msg.Ack(false); err != nil {
			log.Printf("failed to ack cancelled notification id=%d: %v", not.ID, err)
		}
		return nil
	}

	if err := n.Cache.SetFull(ctx, not.ID, string(not.Status), not.Attempts, nullable(not.LastError)); err != nil {
		log.Printf("redis cache write error: %v", err)
	}

	now := time.Now().UTC()
	if now.Before(not.SendAt) {
		delay := not.SendAt.Sub(now)
		log.Printf("notification id=%d scheduled in %v (current UTC: %s)", not.ID, delay, now.Format(time.RFC3339))
		time.Sleep(delay)
	}

	sender := n.senderFor(not.Channel)
	log.Printf("Selected sender for notification id=%d: %s", not.ID, sender.Name())

	log.Printf("Sending notification id=%d to %s via %s", not.ID, not.Recipient, sender.Name())
	if err := sender.Send(ctx, not.Recipient, not.Message); err != nil {
		log.Printf("send failed id=%d: %v", not.ID, err)

		not.Attempts++
		lastErr := err.Error()
		if err := n.Repo.UpdateAttemptsAndError(ctx, not.ID, not.Attempts, lastErr); err != nil {
			log.Printf("db update attempt error: %v", err)
		}
		if err := n.Cache.SetFull(ctx, not.ID, string(model.StatusFailed), not.Attempts, lastErr); err != nil {
			log.Printf("redis update failed: %v", err)
		}

		backoff := time.Duration(math.Pow(2, float64(not.Attempts))) * time.Second
		log.Printf("retrying in %v", backoff)

		go func(id int64, delay time.Duration) {
			time.Sleep(delay)
			b, err := json.Marshal(NotificationMessage{ID: id})
			if err != nil {
				log.Printf("republish marshal error (id=%d): %v", id, err)
				return
			}
			if err := n.PublishCh.Publish("", "notifications", false, false, amqp.Publishing{
				ContentType: "application/json",
				Body:        b,
			}); err != nil {
				log.Printf("republish error: %v", err)
			}
		}(not.ID, backoff)

		if err := msg.Ack(false); err != nil {
			log.Printf("ack failed after scheduling retry for id=%d: %v", not.ID, err)
		}
		return nil
	}

	if err := n.Repo.UpdateStatus(ctx, not.ID, model.StatusSent); err != nil {
		log.Printf("db update status error: %v", err)
	}
	if err := n.Cache.SetFull(ctx, not.ID, string(model.StatusSent), not.Attempts, ""); err != nil {
		log.Printf("redis final update error: %v", err)
	}

	log.Printf("sent id=%d successfully via %s", not.ID, sender.Name())
	if err := msg.Ack(false); err != nil {
		log.Printf("ack error: %v", err)
	}
	return nil
}

func nullable(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
