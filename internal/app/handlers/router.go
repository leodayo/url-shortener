package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/logger"
)

func MainRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(logger.ResponseLogger, logger.RequestLogger)

	r.Post("/", ShortenURL)
	r.Get(config.ExpandPath.Path+"/{id}", GetOriginalURL)

	return r
}
