package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"
	"gitlab.com/arkine/l3/7/internal/handlers"
	"gitlab.com/arkine/l3/7/internal/repository"
	"gitlab.com/arkine/l3/7/internal/server"
	"gitlab.com/arkine/l3/7/internal/utils"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("[Startup] WARNING: JWT_SECRET not set, using default insecure secret")
		secret = "secret"
	}
	utils.SetJWTKey(secret)

	repo := repository.NewRepo()
	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			log.Println("[Startup] Failed to close DB:", err)
		}
	}(repo.DB)

	h := handlers.NewHandler(repo)

	r := server.NewRouter(h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("[Startup] Server running on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[Startup] ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("[Shutdown] Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("[Shutdown] Server forced to shutdown:", err)
	}
	log.Println("[Shutdown] Server exited properly")
}
