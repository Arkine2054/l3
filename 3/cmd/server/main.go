package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"gitlab.com/arkine/l3/3/internal/db"
	"gitlab.com/arkine/l3/3/internal/handlers"
	"gitlab.com/arkine/l3/3/internal/repository"
	"gitlab.com/arkine/l3/3/internal/router"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer func(database *sql.DB) {
		err := database.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}(database)

	if err := db.Migrate(database); err != nil {
		log.Fatal("migration failed:", err)
	}

	repo := &repository.CommentRepo{DB: database}
	handler := &handlers.CommentHandler{Repo: repo}

	r := router.NewRouter(handler)

	srvAddr := ":" + os.Getenv("PORT")
	if srvAddr == ":" {
		srvAddr = ":8080"
	}
	log.Println("Server running on", srvAddr)
	err = http.ListenAndServe(srvAddr, r)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
}
