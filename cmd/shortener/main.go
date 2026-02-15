package main

import (
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
		panic(err)
	}
	logger.Log.Info("running server", zap.String("address", AppConfig.Address))
}
