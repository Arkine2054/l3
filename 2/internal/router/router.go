package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.com/arkine/l3/2/internal/handlers"
)

func New(shorten *handlers.ShortenHandler, redirect *handlers.RedirectHandler, analytics *handlers.AnalyticsHandler) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/shorten", shorten.Create).Methods("POST")
	r.HandleFunc("/s/{alias}", redirect.Redirect).Methods("GET")
	r.HandleFunc("/analytics/{alias}", analytics.GetAnalytics).Methods("GET")

	fs := http.FileServer(http.Dir("./ui"))
	r.PathPrefix("/").Handler(fs)

	return r
}
