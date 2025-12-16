package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/arkine/l3/1/internal/handlers"
)

func NewRouter(h *handlers.NotificationHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("pong"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui/index.html")
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./ui"))))

	r.Post("/notify", h.Create)
	r.Get("/notify", h.List)
	r.Delete("/notify/{id}", h.Cancel)

	r.Get("/channels", handlers.Channels)

	return r
}
