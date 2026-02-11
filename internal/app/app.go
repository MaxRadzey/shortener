package app

import (
	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/router"
	"github.com/MaxRadzey/shortener/internal/service"
	"github.com/MaxRadzey/shortener/internal/storage"
	"go.uber.org/zap"
)

// Run запускает http сервер.
func Run(AppConfig *config.Config) error {
	if err := logger.Initialize(AppConfig.LogLevel); err != nil {
		return err
	}

	storageResult, err := storage.InitializeStorage(AppConfig.DatabaseDSN, AppConfig.FilePath)
	if err != nil {
		return err
	}

	urlService := service.NewService(storageResult.Repository, *AppConfig)
	h := &httphandlers.Handler{Service: urlService}

	r := router.SetupRouter(h, AppConfig)

	logger.Log.Info("Starting HTTP server", zap.String("address", AppConfig.Address))
	return r.Run(AppConfig.Address)
}
