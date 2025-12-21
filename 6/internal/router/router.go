package router

import (
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/arkine/l3/6/internal/handlers"
)

func New(h *handlers.Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateSale(w, r)
		case http.MethodGet:
			h.ListSales(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/items/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.GetSale(w, r, id)
		case http.MethodPut:
			h.UpdateSale(w, r, id)
		case http.MethodDelete:
			h.DeleteSale(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/analytics", h.GetAnalytics)
	mux.HandleFunc("/export.csv", h.ExportCSV)
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	return mux
}
