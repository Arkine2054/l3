package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"

	"gitlab.com/arkine/l3/1/internal/cache"
	"gitlab.com/arkine/l3/1/internal/queue"
	"gitlab.com/arkine/l3/1/internal/repo"
	"gitlab.com/arkine/l3/1/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgURL := os.Getenv("POSTGRES_URL")
	rabbitURL := os.Getenv("RABBIT_URL")
	redisURL := os.Getenv("REDIS_URL")

	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	repos := repo.New(db)

	cacheClient := cache.New(redisURL)
	if err := cacheClient.Ping(ctx); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer cacheClient.Close()

	conn, ch, err := queue.Connect(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	defer ch.Close()

	senders := make(map[string]service.Sender)

	smtpAddr := os.Getenv("SMTP_ADDR")
	smtpFrom := os.Getenv("SMTP_FROM")
	if smtpAddr != "" && smtpFrom != "" {
		senders["email"] = service.NewEmailSender(smtpAddr, smtpFrom)
		log.Println("Email sender initialized")
	}

	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgToken != "" {
		senders["telegram"] = service.NewTelegramSender(tgToken)
	}

	senders["simulated"] = service.NewSimulatedSender()

	for k := range senders {
		log.Printf("Sender registered: %s", k)
	}

	notifier := service.NewNotifier(db, repos, cacheClient, ch, senders)

	msgs, err := queue.Consume(ch)
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	workerCount := 5
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case d, ok := <-msgs:
					if !ok {
						return
					}
					handleMessage(notifier, d, id)
				case <-ctx.Done():
					return
				}
			}
		}(i + 1)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	log.Println("shutting down gracefully...")
	cancel()
	wg.Wait()
	log.Println("all workers stopped")
}

func handleMessage(notifier *service.Notifier, d amqp.Delivery, workerID int) {
	log.Printf("Worker %d received message: %s", workerID, d.Body)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered in worker %d: %v", workerID, r)
			err := d.Nack(false, true)
			if err != nil {
				log.Printf("Nack recover error: %v", err)
			}
			return
		}
	}()

	if err := notifier.HandleMessage(context.Background(), d); err != nil {
		log.Printf("worker %d: error handling message: %v", workerID, err)
		err := d.Nack(false, true)
		if err != nil {
			log.Printf("Nack handle message error: %v", err)
		}
		return
	}

	if err := d.Ack(false); err != nil {
		log.Printf("worker %d: failed to ack message: %v", workerID, err)
	} else {
		log.Printf("Worker %d: message acked successfully", workerID)
	}
}
