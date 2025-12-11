package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"gitlab.com/arkine/l3/7/internal/models"
	"gitlab.com/arkine/l3/7/internal/repository"
	"gitlab.com/arkine/l3/7/internal/utils"
)

type Handler struct {
	Repo *repository.Repo
}

func NewHandler(r *repository.Repo) *Handler {
	return &Handler{Repo: r}
}
func getUserID(r *http.Request) int64 {
	claims := utils.GetClaimsFromContext(r.Context())
	return utils.GetUserIDFromClaims(claims)
}

func getRole(r *http.Request) string {
	claims := utils.GetClaimsFromContext(r.Context())
	if claims == nil {
		return ""
	}
	if role, ok := claims["role"].(string); ok {
		return role
	}
	return ""
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("[Login] Trying user: '%s'", req.Username)
	user, err := h.Repo.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		utils.JSONError(w, "Invalid credentials", http.StatusBadRequest)
		log.Printf("[Login] Failed login attempt for user '%s': user not found", req.Username)
		return
	}
	if !utils.CheckPassword(user.Password, req.Password) {
		utils.JSONError(w, "Invalid credentials", http.StatusBadRequest)
		log.Printf("[Login] Failed login attempt for user '%s': wrong password", req.Username)
		return
	}

	tokenStr, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		utils.JSONError(w, "Token creation failed", http.StatusInternalServerError)
		return
	}
	utils.JSON(w, map[string]string{"token": tokenStr})
	log.Printf("[Login] User '%s' logged in successfully", req.Username)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		utils.JSONError(w, "Username and password required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		utils.JSONError(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "viewer"
	}
	if req.Role == "admin" {
		utils.JSONError(w, "Cannot register as admin", http.StatusForbidden)
		return
	}

	user := models.User{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}

	if err := h.Repo.CreateUser(r.Context(), user); err != nil {
		utils.JSONError(w, "User already exists or DB error", http.StatusBadRequest)
		log.Printf("[Register] Failed to register user '%s': %v", req.Username, err)
		return
	}

	utils.JSONOK(w)
	log.Printf("[Register] New user registered: '%s' with role '%s'", req.Username, req.Role)
}

func (h *Handler) GetItems(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	log.Printf("[GetItems] Role=%s requested items", role)

	items, err := h.Repo.GetAllItems(r.Context())
	if err != nil {
		utils.JSONError(w, "DB error", http.StatusInternalServerError)
		log.Printf("[GetItems] DB error: %v", err)
		return
	}
	utils.JSON(w, items)
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)

	var i models.Item
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		utils.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := h.Repo.CreateItem(r.Context(), i); err != nil {
		utils.JSONError(w, "Unable to create", http.StatusInternalServerError)
		log.Printf("[CreateItem] Role=%s failed to create item %+v: %v", role, i, err)
		return
	}

	utils.JSONOK(w)
	log.Printf("[CreateItem] Role=%s created item %+v", role, i)
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	err := h.Repo.SetCurrentUser(r.Context(), getUserID(r))
	if err != nil {
		log.Printf("[UpdateItem] Failed to update user: %v", err)
	}
	role := getRole(r)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSONError(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var i models.Item
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		utils.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := h.Repo.UpdateItem(r.Context(), id, i); err != nil {
		utils.JSONError(w, "Unable to update", http.StatusInternalServerError)
		log.Printf("[UpdateItem] Role=%s failed to update item %d: %v", role, id, err)
		return
	}

	utils.JSONOK(w)
	log.Printf("[UpdateItem] Role=%s updated item %d: %+v", role, id, i)
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	err := h.Repo.SetCurrentUser(r.Context(), getUserID(r))
	if err != nil {
		log.Printf("[DeleteItem] Failed to delete user '%s': %v", getUserID(r), err)
	}

	role := getRole(r)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSONError(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	log.Printf("[DeleteItem] Role=%s deleting item id=%d", role, id)

	if err := h.Repo.DeleteItem(r.Context(), id); err != nil {
		utils.JSONError(w, "Unable to delete", http.StatusInternalServerError)
		log.Printf("[DeleteItem] Failed: %v", err)
		return
	}

	utils.JSONOK(w)
	log.Printf("[DeleteItem] Deleted item id=%d", id)
}

func (h *Handler) GetItemHistory(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSONError(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	history, err := h.Repo.GetHistoryForItem(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "Unable to get history", http.StatusInternalServerError)
		log.Printf("[GetItemHistory] Role=%s failed to get history for item %d: %v", role, id, err)
		return
	}

	utils.JSON(w, history)
	log.Printf("[GetItemHistory] Role=%s accessed history for item %d", role, id)
}
