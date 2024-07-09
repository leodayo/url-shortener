package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MainRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/", ShortenURL)
	r.Get("/{id}", GetOriginalURL)

	return r
}
