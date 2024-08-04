package app

import (
	"net/http"

	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/app/handlers"
	"github.com/leodayo/url-shortener/internal/app/storage"
	"github.com/leodayo/url-shortener/internal/logger"
)

func Run() error {
	config.ParseFlags()
	err := config.ParseEnv()
	if err != nil {
		return err
	}

	if err := logger.Initialize("debug"); err != nil {
		return err
	}

	if err := storage.InitFileStorage(); err != nil {
		return err
	}

	return http.ListenAndServe(config.ServerAddress, handlers.MainRouter())
}
