package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"gitlab.com/arkine/l3/6/internal/models"
	"gitlab.com/arkine/l3/6/internal/repository"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@db:5432/salesdb?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepo(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateSale(repo, w, r)
		case http.MethodGet:
			handleListSales(repo, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/items/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetSale(repo, w, r, id)
		case http.MethodPut:
			handleUpdateSale(repo, w, r, id)
		case http.MethodDelete:
			handleDeleteSale(repo, w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/analytics", func(w http.ResponseWriter, r *http.Request) {
		handleAnalytics(repo, w, r)
	})
	mux.HandleFunc("/export.csv", func(w http.ResponseWriter, r *http.Request) {
		handleExportCSV(repo, w, r)
	})
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	addr := ":8080"
	log.Printf("listening %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
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

func handleCreateSale(repo *repository.Repo, w http.ResponseWriter, r *http.Request) {
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
	ctx := r.Context()
	if err := repo.CreateSale(ctx, &s); err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		log.Println("create:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(s)
	if err != nil {
		log.Println("Write create sale error:", err)
	}
}

func handleListSales(repo *repository.Repo, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.ListFilter{Limit: 100}
	if l := q.Get("limit"); l != "" {
		if v, _ := strconv.Atoi(l); v > 0 {
			f.Limit = v
		}
	}
	if o := q.Get("offset"); o != "" {
		if v, _ := strconv.Atoi(o); v >= 0 {
			f.Offset = v
		}
	}
	if s := q.Get("sort"); s != "" {
		f.SortBy = s
	}
	if d := q.Get("desc"); d == "1" {
		f.Desc = true
	}
	if cat := q.Get("category"); cat != "" {
		f.Category = &cat
	}
	if k := q.Get("kind"); k != "" {
		kk := models.Kind(k)
		f.Kind = &kk
	}
	if from := q.Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			f.From = &t
		}
	}
	if to := q.Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			f.To = &t
		}
	}

	ctx := r.Context()
	arr, err := repo.ListSales(ctx, f)
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		log.Println("list:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(arr)
	if err != nil {
		log.Println("Write list sales error:", err)
	}
}

func handleGetSale(repo *repository.Repo, w http.ResponseWriter, r *http.Request, id int64) {
	s, err := repo.GetSale(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed", http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(s)
	if err != nil {
		log.Println("Write get sale error:", err)
	}
}

func handleUpdateSale(repo *repository.Repo, w http.ResponseWriter, r *http.Request, id int64) {
	var s models.Sale
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	s.ID = id
	if err := repo.UpdateSale(r.Context(), &s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "update failed", http.StatusInternalServerError)
		log.Println("update:", err)
		return
	}
	err := json.NewEncoder(w).Encode(s)
	if err != nil {
		log.Println("Write update sale error:", err)
	}
}

func handleDeleteSale(repo *repository.Repo, w http.ResponseWriter, r *http.Request, id int64) {
	if err := repo.DeleteSale(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleAnalytics(repo *repository.Repo, w http.ResponseWriter, r *http.Request) {
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
	a, err := repo.GetAnalytics(r.Context(), from, to, kind, category)
	if err != nil {
		http.Error(w, "analytics failed", http.StatusInternalServerError)
		log.Println("analytics:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(a)
	if err != nil {
		log.Println("Write analytics error:", err)
	}
}

func handleExportCSV(repo *repository.Repo, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.ListFilter{Limit: 10000}
	if cat := q.Get("category"); cat != "" {
		f.Category = &cat
	}
	if k := q.Get("kind"); k != "" {
		kk := models.Kind(k)
		f.Kind = &kk
	}
	if from := q.Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			f.From = &t
		}
	}
	if to := q.Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			f.To = &t
		}
	}

	arr, err := repo.ListSales(r.Context(), f)
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
		log.Println("Write export csv error:", err)
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
			log.Println("Write export csv error:", err)
		}
	}
}
