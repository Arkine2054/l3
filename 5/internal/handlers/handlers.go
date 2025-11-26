package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"gitlab.com/arkine/l3/5/internal/repository"
	"gitlab.com/arkine/l3/5/internal/service"
)

type Handlers struct {
	svc *service.Service
}

func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/events", h.handleCreateEvent).Methods("POST")
	r.HandleFunc("/events", h.handleListEvents).Methods("GET")
	r.HandleFunc("/events/{id:[0-9]+}", h.handleGetEvent).Methods("GET")
	r.HandleFunc("/events/{id:[0-9]+}/book", h.handleBookSeat).Methods("POST")
	r.HandleFunc("/events/{id:[0-9]+}/confirm", h.handleConfirmBooking).Methods("POST")

	fs := http.FileServer(http.Dir("./web"))
	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))
}

func (h *Handlers) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Title      string `json:"title"`
		Date       string `json:"date"`
		TotalSeats int    `json:"total_seats"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.TotalSeats <= 0 || req.Date == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	eventDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	event, err := h.svc.CreateEvent(ctx, req.Title, eventDate, req.TotalSeats)
	if err != nil {
		log.Printf("CreateEvent error: %v", err)
		http.Error(w, "failed to create event", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, event)
}

func (h *Handlers) handleListEvents(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	events, err := h.svc.ListEvents(ctx)
	if err != nil {
		log.Printf("ListEvents error: %v", err)
		http.Error(w, "failed to get events", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, events)
}

func (h *Handlers) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	event, err := h.svc.GetEvent(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("event not found: %v", err), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func (h *Handlers) handleBookSeat(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid event id", http.StatusBadRequest)
		return
	}

	type request struct {
		UserName string `json:"user_name"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.UserName == "" {
		http.Error(w, "user_name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	booking, err := h.svc.BookSeat(ctx, id, req.UserName)
	if err != nil {
		if errors.Is(err, repository.ErrNoSeatsAvailable) {
			http.Error(w, "no seats available", http.StatusConflict)
			return
		}
		log.Printf("BookSeat error: %v", err)
		http.Error(w, "failed to create booking", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, booking)
}

func (h *Handlers) handleConfirmBooking(w http.ResponseWriter, r *http.Request) {
	type request struct {
		BookingID int `json:"booking_id"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.BookingID <= 0 {
		http.Error(w, "booking_id required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.ConfirmBooking(ctx, req.BookingID); err != nil {
		log.Printf("ConfirmBooking error: %v", err)
		http.Error(w, "failed to confirm booking", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("writeJSON error: %v", err)
		}
	}
}
