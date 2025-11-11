package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

type DB struct {
	SQL   *sql.DB
	Redis *redis.Client
}

func New() (*DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user1:pass1@localhost:5432/urlshortener?sslmode=disable"
	}

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	log.Println("Connected to Postgres")

	var rdb *redis.Client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr != "" {
		rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Printf("Redis unavailable: %v", err)
			rdb = nil
		} else {
			log.Println("Connected to Redis")
		}
	}

	return &DB{SQL: sqlDB, Redis: rdb}, nil
}
