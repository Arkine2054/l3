package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf(err.Error())
	}
}

func JSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(map[string]string{"error": msg})
	if err != nil {
		log.Printf(err.Error())
	}
}

func JSONOK(w http.ResponseWriter) {
	JSON(w, map[string]string{"status": "ok"})
}
