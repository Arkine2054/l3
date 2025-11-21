package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/arkine/l3/3/internal/db"
	"gitlab.com/arkine/l3/3/internal/handlers"
	"gitlab.com/arkine/l3/3/internal/repository"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatal("migration failed:", err)
	}

	repo := &repository.CommentRepo{DB: database}
	handler := &handlers.CommentHandler{Repo: repo}

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/comments", handler.Create).Methods("POST")
	api.HandleFunc("/comments", handler.List).Methods("GET")
	api.HandleFunc("/comments/{id:[0-9]+}", handler.Delete).Methods("DELETE")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
