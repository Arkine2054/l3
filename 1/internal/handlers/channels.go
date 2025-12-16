package handlers

import (
	"encoding/json"
	"net/http"
)

func Channels(w http.ResponseWriter, _ *http.Request) {
	channels := []string{"email", "telegram", "simulated"}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(channels)
}
