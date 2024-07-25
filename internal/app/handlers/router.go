package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leodayo/url-shortener/internal/app/config"
)

func MainRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/", ShortenURL)
	r.Get(config.ExpandPath.Path+"/{id}", GetOriginalURL)

	return r
}
