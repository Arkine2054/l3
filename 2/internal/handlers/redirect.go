package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/arkine/l3/2/internal/repo"
	"gitlab.com/arkine/l3/2/internal/utils"
)

type RedirectHandler struct {
	shortRepo *repo.ShortURLRepo
	clickRepo *repo.ClickRepo
}

func NewRedirectHandler(shortRepo *repo.ShortURLRepo, clickRepo *repo.ClickRepo) *RedirectHandler {
	return &RedirectHandler{shortRepo: shortRepo, clickRepo: clickRepo}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]

	shortURL, err := h.shortRepo.FindByAlias(alias)
	if err != nil || shortURL == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	ip := utils.ClientIP(r)
	userAgent := r.UserAgent()
	referer := r.Referer()

	if err := h.clickRepo.SaveClick(shortURL.ID, userAgent, ip, referer); err != nil {
		log.Printf("error saving click: %v", err)
	}

	http.Redirect(w, r, shortURL.Target, http.StatusFound)
}
