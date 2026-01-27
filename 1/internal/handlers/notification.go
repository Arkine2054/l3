package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"

	"gitlab.com/arkine/l3/1/internal/model"
	"gitlab.com/arkine/l3/1/internal/repo"
)

type NotificationHandler struct {
	Repo   *repo.Repo
	Client *rabbitmq.RabbitClient
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

	id, err := h.Repo.CreateNotification(r.Context(), &n)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	ch, err := h.Client.GetChannel()
	if err != nil {
		http.Error(w, "rabbit error", http.StatusInternalServerError)
		return
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			http.Error(w, "rabbit error", http.StatusInternalServerError)
		}
	}(ch)

	delay := time.Until(sendAt)
	if delay < 0 {
		delay = 0
	}

	body, err := json.Marshal(map[string]int64{"id": id})
	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)
	}

	if err := ch.Publish(
		"notifications.delayed",
		"notifications",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp.Table{
				"x-delay": int64(delay / time.Millisecond),
			},
		},
	); err != nil {
		http.Error(w, "publish failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(strconv.FormatInt(id, 10)))
	if err != nil {
		log.Printf("failed to write response: %s", err)
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
		log.Printf("failed to encode response: %s", err)
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
