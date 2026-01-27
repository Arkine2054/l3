package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func Channels(w http.ResponseWriter, _ *http.Request) {
	channels := []string{"email", "telegram", "simulated"}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(channels)
	if err != nil {
		log.Printf("json encode error: %s", err)
	}
}
