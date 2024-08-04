package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/app/middleware"
)

func MainRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.ResponseLogger, middleware.RequestLogger)

	r.Get(config.ExpandPath.Path+"/{id}", GetOriginalURL)
	r.Post("/", ShortenURL)
	r.Post("/api/shorten", ShortenURLJSON)

	return r
}
