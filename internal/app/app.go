package app

import (
	"net/http"

	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/app/handlers"
)

func Run() error {
	config.ParseFlags()
	err := config.ParseEnv()
	if err != nil {
		return err
	}

	return http.ListenAndServe(config.ServerAddress, handlers.MainRouter())
}
