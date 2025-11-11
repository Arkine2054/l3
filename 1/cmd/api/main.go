package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/go-chi/chi/v5"

	"gitlab.com/arkine/l3/1/internal/model"
	"gitlab.com/arkine/l3/1/internal/queue"
	"gitlab.com/arkine/l3/1/internal/repo"
)

func main() {
	r := chi.NewRouter()

	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		log.Fatal("RABBIT_URL is not set")
	}

	conn, ch, err := queue.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("RabbitMQ connect failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	_, err = ch.QueueDeclare(
		"notifications",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("queue declare failed: %v", err)
	}

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Printf("%s %s", req.Method, req.URL.Path)
			next.ServeHTTP(w, req)
		})
	})

	pgURL := os.Getenv("POSTGRES_URL")
	if pgURL == "" {
		log.Fatal("POSTGRES_URL is not set")
	}
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	repos := repo.New(db)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write([]byte("pong"))
		if err != nil {
			log.Printf("failed to write pong: %v", err)
		}
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui/index.html")
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./ui"))))

	r.Post("/notify", func(w http.ResponseWriter, req *http.Request) {
		type NotificationInput struct {
			Recipient string `json:"recipient"`
			Channel   string `json:"channel"`
			Message   string `json:"message"`
			SendAt    string `json:"send_at"`
		}

		var input NotificationInput
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		missing := []string{}
		if input.Recipient == "" {
			missing = append(missing, "recipient")
		}
		if input.Channel == "" {
			missing = append(missing, "channel")
		}
		if input.Message == "" {
			missing = append(missing, "message")
		}
		if input.SendAt == "" {
			missing = append(missing, "send_at")
		}
		if len(missing) > 0 {
			http.Error(w, "missing required fields: "+strings.Join(missing, ", "), http.StatusBadRequest)
			return
		}

		sendAt, err := time.Parse(time.RFC3339, input.SendAt)
		if err != nil {
			http.Error(w, "invalid send_at format, must be ISO8601", http.StatusBadRequest)
			return
		}
		if sendAt.Before(time.Now()) {
			http.Error(w, "send_at cannot be in the past", http.StatusBadRequest)
			return
		}

		n := model.Notification{
			Recipient: input.Recipient,
			Channel:   input.Channel,
			Message:   input.Message,
			SendAt:    sendAt,
			Status:    model.StatusScheduled,
		}
		id, err := repos.CreateNotification(req.Context(), &n)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		log.Printf("notification created with id: %v", id)

		body := []byte(fmt.Sprintf(`{"id":%d}`, id))
		log.Printf("Publishing directly to RabbitMQ: %s", body)

		err = ch.Publish(
			"",
			"notifications",
			false, false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		if err != nil {
			log.Printf("Failed to publish to RabbitMQ: %v", err)
			http.Error(w, "failed to publish message", http.StatusInternalServerError)
			return
		}
		log.Printf("Message published successfully to queue 'notifications'")
		time.Sleep(1 * time.Second)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	})

	r.Get("/notify", func(w http.ResponseWriter, req *http.Request) {
		limit := 50
		if q := req.URL.Query().Get("limit"); q != "" {
			v, err := strconv.Atoi(q)
			if err != nil || v <= 0 {
				http.Error(w, "invalid limit parameter", http.StatusBadRequest)
				return
			}
			limit = v
		}

		list, err := repos.ListRecent(req.Context(), limit)
		if err != nil {
			log.Printf("GET /notify db error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(list); err != nil {
			log.Printf("GET /notify encode error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})

	r.Get("/notify/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id parameter", http.StatusBadRequest)
			return
		}

		not, err := repos.GetNotification(req.Context(), id)
		if err != nil {
			log.Printf("GET /notify/%d db error: %v", id, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if not == nil {
			http.Error(w, "notification not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(not); err != nil {
			log.Printf("GET /notify/%d encode error: %v", id, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})

	r.Post("/notify/{id}/reschedule", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id parameter", http.StatusBadRequest)
			return
		}

		type RescheduleInput struct {
			SendAt string `json:"send_at"`
		}
		var input RescheduleInput
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		newSendAt, err := time.Parse(time.RFC3339, input.SendAt)
		if err != nil {
			http.Error(w, "invalid send_at format, must be ISO8601", http.StatusBadRequest)
			return
		}
		if newSendAt.Before(time.Now()) {
			http.Error(w, "send_at cannot be in the past", http.StatusBadRequest)
			return
		}

		not, err := repos.GetNotification(req.Context(), id)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if not == nil {
			http.Error(w, "notification not found", http.StatusNotFound)
			return
		}

		not.SendAt = newSendAt
		not.Status = model.StatusScheduled
		not.Attempts = 0
		not.LastError = nil

		if err := repos.UpdateNotification(req.Context(), not.ID, not.Message, not.SendAt); err != nil {
			http.Error(w, "failed to update notification", http.StatusInternalServerError)
			return
		}

		body := []byte(fmt.Sprintf(`{"id":%d}`, not.ID))
		if err := ch.Publish("", "notifications", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}); err != nil {
			log.Printf("Failed to publish to RabbitMQ: %v", err)
			http.Error(w, "failed to publish message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(not)
	})

	r.Get("/channels", func(w http.ResponseWriter, r *http.Request) {
		channels := []string{"email", "telegram", "simulated"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(channels)
	})

	r.Delete("/notify/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id parameter", http.StatusBadRequest)
			return
		}

		err = repos.CancelNotification(req.Context(), id)
		if err != nil {
			log.Printf("DELETE /notify/%d db error: %v", id, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	addr := ":8083"
	log.Printf("API server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("API server failed: %v", err)
	}
}
