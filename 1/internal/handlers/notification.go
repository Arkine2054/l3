package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	amqp "github.com/rabbitmq/amqp091-go"

	"gitlab.com/arkine/l3/1/internal/model"
	"gitlab.com/arkine/l3/1/internal/repo"
)

type NotificationHandler struct {
	Repo *repo.Repo
	AMQP *amqp.Channel
}

type NotificationInput struct {
	Recipient string `json:"recipient"`
	Channel   string `json:"channel"`
	Message   string `json:"message"`
	SendAt    string `json:"send_at"`
}

func (h *NotificationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input NotificationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	sendAt, err := time.Parse(time.RFC3339, input.SendAt)
	if err != nil {
		http.Error(w, "send_at must be RFC3339", http.StatusBadRequest)
		return
	}

	n := model.Notification{
		Recipient: input.Recipient,
		Channel:   input.Channel,
		Message:   input.Message,
		SendAt:    sendAt,
		Status:    model.StatusScheduled,
	}

	if err := n.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.Repo.CreateNotification(r.Context(), &n)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	body := []byte(fmt.Sprintf(`{"id":%d}`, id))
	if err := h.AMQP.Publish(
		"",
		"notifications",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		log.Printf("rabbit publish error: %v", err)
		http.Error(w, "failed to publish", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]int64{"id": id})
	if err != nil {
		log.Printf("create json encode error: %v", err)
	}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		v, err := strconv.Atoi(q)
		if err != nil || v <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = v
	}

	list, err := h.Repo.ListRecent(r.Context(), limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		log.Printf("list json encode error: %v", err)
	}
}

func (h *NotificationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Repo.CancelNotification(r.Context(), id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
