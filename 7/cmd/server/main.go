package main

import (
	"context"
	"database/sql"
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

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/warehouse?sslmode=disable"
	}

	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil && db.Ping() == nil {
			break
		}
		log.Println("[Startup] DB ping failed, retrying in 2s...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("[Startup] Cannot connect to DB:", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(60 * time.Minute)

	defer db.Close()
	log.Println("[Startup] Connected to PostgreSQL")

	if err := utils.RunMigrations(db, "migrate"); err != nil {
		log.Fatal("[Startup] Migration failed:", err)
	}
	log.Println("[Startup] Migrations applied successfully!")

	repo := repository.NewRepo(db)
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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
