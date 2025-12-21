package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/arkine/l3/6/internal/models"
	"gitlab.com/arkine/l3/6/internal/repository"
)

type Handlers struct {
	repo *repository.Repo
}

func New(repo *repository.Repo) *Handlers {
	return &Handlers{repo: repo}
}

func parseTimePtr(q string) *time.Time {
	if q == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, q)
	if err != nil {
		return nil
	}
	return &t
}

func (h *Handlers) CreateSale(w http.ResponseWriter, r *http.Request) {
	var s models.Sale
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	if s.Amount < 0 {
		http.Error(w, "amount must be >= 0", http.StatusBadRequest)
		return
	}

	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}

	if s.Kind != models.KindIncome && s.Kind != models.KindExpense {
		http.Error(w, "invalid kind", http.StatusBadRequest)
		return
	}

	if err := h.repo.CreateSale(r.Context(), &s); err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	writeJSON(w, http.StatusCreated, s)
}

func (h *Handlers) ListSales(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.ListFilter{Limit: 100}

	if v, _ := strconv.Atoi(q.Get("limit")); v > 0 {
		f.Limit = v
	}
	if v, _ := strconv.Atoi(q.Get("offset")); v >= 0 {
		f.Offset = v
	}
	if q.Get("desc") == "1" {
		f.Desc = true
	}
	if c := q.Get("category"); c != "" {
		f.Category = &c
	}
	if k := q.Get("kind"); k != "" {
		kk := models.Kind(k)
		f.Kind = &kk
	}

	arr, err := h.repo.ListSales(r.Context(), f)
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, arr)
}

func (h *Handlers) GetSale(w http.ResponseWriter, r *http.Request, id int64) {
	s, err := h.repo.GetSale(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed", http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, s)
}

func (h *Handlers) UpdateSale(w http.ResponseWriter, r *http.Request, id int64) {
	var s models.Sale
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	s.ID = id

	if err := h.repo.UpdateSale(r.Context(), &s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, s)
}

func (h *Handlers) DeleteSale(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.repo.DeleteSale(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from := parseTimePtr(q.Get("from"))
	to := parseTimePtr(q.Get("to"))

	var kind *models.Kind
	if k := q.Get("kind"); k != "" {
		kk := models.Kind(k)
		kind = &kk
	}

	var category *string
	if c := q.Get("category"); c != "" {
		category = &c
	}

	result, err := h.repo.GetAnalytics(
		r.Context(),
		from,
		to,
		kind,
		category,
	)
	if err != nil {
		http.Error(w, "analytics failed", http.StatusInternalServerError)
		log.Println("analytics:", err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	arr, err := h.repo.ListSales(r.Context(), repository.ListFilter{Limit: 10000})
	if err != nil {
		http.Error(w, "export failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"sales.csv\"")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	err = cw.Write([]string{"id", "kind", "amount", "category", "note", "created_at"})
	if err != nil {
		log.Printf("error writing csv: %v", err)
	}
	for _, s := range arr {
		err := cw.Write([]string{
			strconv.FormatInt(s.ID, 10),
			string(s.Kind),
			fmt.Sprintf("%.2f", s.Amount),
			s.Category,
			s.Note,
			s.CreatedAt.Format(time.RFC3339),
		})
		if err != nil {
			log.Printf("error writing csv: %v", err)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("error encoding json: %v", err)
	}
}
