package handlers

import (
	"encoding/json"
	"net/http"

	"gitlab.com/arkine/l3/2/internal/services"
)

type ShortenHandler struct {
	service *services.ShortenerService
}

func NewShortenHandler(s *services.ShortenerService) *ShortenHandler {
	return &ShortenHandler{service: s}
}

func (h *ShortenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target string `json:"target"`
		Alias  string `json:"alias,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateShortURL(req.Target, req.Alias, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
