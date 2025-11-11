package main

import (
	"context"
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
	// --- Настройка окружения ---
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

	// --- Подключение к базе ---
	db, err := repository.ConnectPostgres(dsn)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, time.Duration(expirationMinutes)*time.Minute)
	svc.StartBookingCleaner(cleanInterval)
	defer svc.StopBookingCleaner()

	// --- HTTP сервер ---
	r := mux.NewRouter()

	// API маршруты
	h := handlers.NewHandlers(svc)
	h.RegisterRoutes(r)

	// --- Отдача фронта ---
	fs := http.FileServer(http.Dir("./web"))
	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// --- Graceful shutdown ---
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
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe error: %v", err)
	}

	<-idleConnsClosed
	log.Println("Server stopped gracefully")
}
