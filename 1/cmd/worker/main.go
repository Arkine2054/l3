package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	wbfredis "github.com/wb-go/wbf/redis"

	"gitlab.com/arkine/l3/1/internal/repo"
	"gitlab.com/arkine/l3/1/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.New()
	if err := cfg.LoadEnvFiles(".env"); err != nil {
		log.Fatal(err)
	}
	cfg.EnableEnv("")

	db, err := dbpg.New(cfg.GetString("postgres.url"), nil, nil)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}

	redisClient := wbfredis.New(
		cfg.GetString("redis.addr"),
		cfg.GetString("redis.password"),
		cfg.GetInt("redis.db"),
	)
	defer redisClient.Close()

	_ = redisClient.Ping(ctx)
	log.Println("Redis connected")

	client, err := rabbitmq.NewClient(rabbitmq.ClientConfig{
		URL:            cfg.GetString("rabbit.url"),
		ConnectionName: "worker",
	})
	if err != nil {
		log.Fatalf("rabbit client error: %v", err)
	}
	defer client.Close()

	ch, err := client.GetChannel()
	if err != nil {
		log.Fatalf("failed to get channel: %v", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		"notifications.delayed",
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-delayed-type": "direct",
		},
	); err != nil {
		log.Fatalf("exchange declare failed: %v", err)
	}

	if _, err := ch.QueueDeclare(
		"notifications",
		true,
		false,
		true,
		false,
		nil,
	); err != nil {
		log.Fatalf("queue declare failed: %v", err)
	}

	if err := ch.QueueBind(
		"notifications",
		"notifications",
		"notifications.delayed",
		false,
		nil,
	); err != nil {
		log.Fatalf("queue bind failed: %v", err)
	}

	repos := repo.New(db)

	senders := map[string]service.Sender{
		"simulated": service.NewSimulatedSender(),
	}
	log.Println("sender registered: simulated")

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

	notifier := service.NewNotifier(
		db,
		repos,
		redisClient,
		client,
		senders,
	)

	cfgConsumer := rabbitmq.ConsumerConfig{
		Queue:         "notifications",
		AutoAck:       false,
		PrefetchCount: 10,
		Workers:       1,
	}

	consumer := rabbitmq.NewConsumer(client, cfgConsumer, notifier.HandleMessage)
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatalf("consumer error: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("shutdown")
	cancel()
	time.Sleep(300 * time.Millisecond)
}
