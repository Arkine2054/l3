package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/arkine/l3/4/internal/handlers"
)

func NewRouter(h *handlers.UploadHandler) *mux.Router {
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	r.PathPrefix("/data/processed/").Handler(
		http.StripPrefix("/data/processed/", http.FileServer(http.Dir("./data/processed"))),
	)
	r.PathPrefix("/data/thumbs/").Handler(
		http.StripPrefix("/data/thumbs/", http.FileServer(http.Dir("./data/thumbs"))),
	)
	r.PathPrefix("/data/original/").Handler(
		http.StripPrefix("/data/original/", http.FileServer(http.Dir("./data/original"))),
	)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	r.HandleFunc("/upload", h.Upload).Methods("POST")
	r.HandleFunc("/image/{id:[0-9]+}", h.GetImage).Methods("GET")
	r.HandleFunc("/image/{id:[0-9]+}", h.DeleteImage).Methods("DELETE")

	r.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		images, err := h.Repo.List()
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(images)
		if err != nil {
			log.Printf("Error encoding json: %v", err)
		}
	}).Methods("GET")

	return r
}
