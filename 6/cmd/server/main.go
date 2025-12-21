package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"gitlab.com/arkine/l3/6/internal/handlers"
	"gitlab.com/arkine/l3/6/internal/repository"
	"gitlab.com/arkine/l3/6/internal/router"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing db connection: %v", err)
		}
	}(db)

	repo := repository.NewRepo(db)
	h := handlers.New(repo)
	r := router.New(h)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
