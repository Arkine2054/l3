package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/arkine/l3/2/internal/services"
)

type AnalyticsHandler struct {
	service *services.AnalyticsService
}

func NewAnalyticsHandler(service *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]

	clicks, err := h.service.GetAnalytics(alias)
	if err != nil {
		http.Error(w, "failed to get analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(clicks)
	if err != nil {
		log.Println("failed encode analytics response", err)
	}
}
