package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.com/arkine/l3/3/internal/models"
	"gitlab.com/arkine/l3/3/internal/repository"
)

type CommentHandler struct {
	Repo *repository.CommentRepo
}

// Создать комментарий
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c models.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Repo.Create(&c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// Удалить комментарий
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Получить список комментариев или выполнить поиск
func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	search := query.Get("search")

	limit := 50
	offset := 0

	if lStr := query.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := query.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var (
		comments []*models.Comment
		err      error
	)

	if search != "" {
		comments, err = h.Repo.Search(search, limit, offset)
	} else {
		comments, err = h.Repo.List()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
