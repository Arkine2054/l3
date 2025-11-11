package main

import (
	"log"
	"net/http"
	"os"

	"gitlab.com/arkine/l3/2/internal/database"
	"gitlab.com/arkine/l3/2/internal/handlers"
	"gitlab.com/arkine/l3/2/internal/repo"
	"gitlab.com/arkine/l3/2/internal/router"
	"gitlab.com/arkine/l3/2/internal/services"
)

func main() {
	db, err := database.New()
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}

	// Repositories
	shortRepo := repo.NewShortURLRepo(db.SQL)
	clickRepo := repo.NewClickRepo(db.SQL)

	// Services
	shortService := services.NewShortenerService(shortRepo)
	analyticsService := services.NewAnalyticsService(clickRepo)

	// Handlers
	shortenHandler := handlers.NewShortenHandler(shortService)
	redirectHandler := handlers.NewRedirectHandler(shortRepo, clickRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	// Router
	r := router.New(shortenHandler, redirectHandler, analyticsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("Server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
