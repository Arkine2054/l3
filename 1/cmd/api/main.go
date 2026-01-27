package main

import (
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"

	"gitlab.com/arkine/l3/1/internal/handlers"
	"gitlab.com/arkine/l3/1/internal/repo"
	"gitlab.com/arkine/l3/1/internal/router"
)

func main() {
	zlog.InitConsole()

	if err := zlog.SetLevel("debug"); err != nil {
		panic(err)
	}

	cfg := config.New()

	err := cfg.LoadEnvFiles(".env")
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	cfg.EnableEnv("")

	dsn := fmt.Sprintf(
		cfg.GetString("postgres.url"),
	)

	db, err := dbpg.New(dsn, nil, &dbpg.Options{
		MaxOpenConns:    cfg.GetInt("postgres.max_open_conns"),
		MaxIdleConns:    cfg.GetInt("postgres.max_idle_conns"),
		ConnMaxLifetime: cfg.GetDuration("postgres.conn_max_lifetime"),
	})
	if err != nil {
		log.Printf("Error opening database connection: %v\n", err)
	}

	repos := repo.New(db)

	clientConfig := rabbitmq.ClientConfig{
		URL:            cfg.GetString("rabbit.url"),
		ConnectionName: "RabbitMQ",
	}

	client, err := rabbitmq.NewClient(clientConfig)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := client.GetChannel()
	if err != nil {
		log.Fatal(err)
	}

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Printf("Error closing channel: %s\n", err)
		}
	}(ch)

	handler := &handlers.NotificationHandler{
		Repo:   repos,
		Client: client,
	}

	r := router.NewRouter(handler)

	log.Println("API listening on :8083")
	err = http.ListenAndServe(":8083", r)
	if err != nil {
		log.Printf("Error listening on port 8083: %v", err)
	}

}
