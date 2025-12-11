package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"gitlab.com/arkine/l3/7/internal/handlers"
	"gitlab.com/arkine/l3/7/internal/middlewares"
)

func NewRouter(h *handlers.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf("write health error: %v", err)
		}
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	r.Handle("/web/*", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	r.Post("/login", h.Login)
	r.Post("/register", h.Register)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWTAuth(h.Repo))

		r.With(middlewares.RequireRole("viewer", "manager", "admin")).Get("/items", h.GetItems)
		r.With(middlewares.RequireRole("manager", "admin")).Post("/items", h.CreateItem)
		r.With(middlewares.RequireRole("manager", "admin")).Put("/items/{id}", h.UpdateItem)
		r.With(middlewares.RequireRole("admin")).Delete("/items/{id}", h.DeleteItem)

		r.With(middlewares.RequireRole("viewer", "manager", "admin")).Get("/history/{id}", h.GetItemHistory)
	})

	return r
}
