package app

import (
	"net/http"

	"github.com/leodayo/url-shortener/internal/app/handlers"
)

func Run() error {
	return http.ListenAndServe("localhost:8080", handlers.MainRouter())
}
