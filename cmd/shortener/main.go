// Package main — точка входа: парсит конфиг и запускает HTTP-сервер коротких ссылок.
package main

import (
	"os"

	"github.com/MaxRadzey/shortener/internal/app"
	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

// @title        Shortener API
// @version      1.0
// @description  API для сокращения URL и управления короткими ссылками.
// @host         localhost:8080
// @BasePath     /
func main() {
	AppConfig := config.New()

	config.ParseEnv(AppConfig)
	config.ParseFlags(AppConfig)

	if err := app.Run(AppConfig); err != nil {
		logger.Log.Error("failed to run app", zap.Error(err))
		os.Exit(1)
	}
	logger.Log.Info("running server", zap.String("address", AppConfig.Address))
}
