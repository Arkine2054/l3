package router

import (
	"net/http"

	"gitlab.com/arkine/l3/3/internal/handlers"

	"github.com/gorilla/mux"
)

func NewRouter(h *handlers.CommentHandler) *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/comments", h.Create).Methods("POST")
	api.HandleFunc("/comments", h.List).Methods("GET")
	api.HandleFunc("/comments/{id:[0-9]+}", h.Delete).Methods("DELETE")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	return r
}
