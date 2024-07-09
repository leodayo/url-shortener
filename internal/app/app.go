package app

import (
	"net/http"

	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/app/handlers"
)

func Run() error {
	config.ParseFlags()
	return http.ListenAndServe(config.ServerAddress, handlers.MainRouter())
}
