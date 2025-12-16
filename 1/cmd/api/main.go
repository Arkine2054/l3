package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.com/arkine/l3/1/internal/handlers"
	"gitlab.com/arkine/l3/1/internal/queue"
	"gitlab.com/arkine/l3/1/internal/router"

	"gitlab.com/arkine/l3/1/internal/repo"
)

func main() {

	db, err := repo.Connect(os.Getenv("POSTGRES_URL"))
	if err != nil {
		log.Fatal(err)
	}
	repos := repo.New(db)

	conn, ch, err := queue.Connect(os.Getenv("RABBIT_URL"))
	if err != nil {
		log.Fatal(err)
	}

	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %s\n", err)
		}
	}(conn)

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Printf("Error closing channel: %s\n", err)
		}
	}(ch)

	h := &handlers.NotificationHandler{
		Repo: repos,
		AMQP: ch,
	}

	r := router.NewRouter(h)

	log.Println("API listening on :8083")
	err = http.ListenAndServe(":8083", r)
	if err != nil {
		log.Printf("Error listening on port 8083: %v", err)
	}
}
