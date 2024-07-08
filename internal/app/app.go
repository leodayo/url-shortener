package app

import (
	"net/http"

	"github.com/leodayo/url-shortener/internal/app/handlers"
)

func Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.ShortenURL)
	mux.HandleFunc("/{id}", handlers.GetOriginalURL)

	return http.ListenAndServe("localhost:8080", mux)
}
