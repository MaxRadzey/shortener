package main

import (
	"github.com/MaxRadzey/shortener/internal/app"
	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

func main() {
	AppConfig := config.New()

	config.ParseEnv(AppConfig)

	ParseFlag(AppConfig)

	if err := app.Run(AppConfig); err != nil {
		panic(err)
	}
	logger.Log.Info("running server", zap.String("address", AppConfig.Address))
}
