package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/middleware"
)

func MainRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.ResponseLogger, middleware.RequestLogger)

	r.Post("/", ShortenURL)
	r.Get(config.ExpandPath.Path+"/{id}", GetOriginalURL)

	return r
}
