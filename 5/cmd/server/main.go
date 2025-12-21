package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"gitlab.com/arkine/l3/5/internal/handlers"
	"gitlab.com/arkine/l3/5/internal/repository"
	"gitlab.com/arkine/l3/5/internal/service"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	expirationMinutes := 2
	if val := os.Getenv("BOOKING_EXPIRATION_MINUTES"); val != "" {
		if m, err := strconv.Atoi(val); err == nil {
			expirationMinutes = m
		}
	}

	cleanInterval := 1 * time.Minute
	if val := os.Getenv("CLEAN_INTERVAL_MINUTES"); val != "" {
		if m, err := strconv.Atoi(val); err == nil {
			cleanInterval = time.Duration(m) * time.Minute
		}
	}

	db, err := repository.ConnectPostgres(dsn)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("DB close failed: %v", err)
		}
	}(db)

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, time.Duration(expirationMinutes)*time.Minute)
	svc.StartBookingCleaner(cleanInterval)
	defer svc.StopBookingCleaner()

	r := mux.NewRouter()

	h := handlers.NewHandlers(svc)
	h.RegisterRoutes(r)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Println("Shutting down server...")

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Server Shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("Server is running on :8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ListenAndServe error: %v", err)
	}

	<-idleConnsClosed
	log.Println("Server stopped gracefully")
}
