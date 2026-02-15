// Package app собирает логгер, хранилище, сервис, роутер и запускает HTTP-сервер.
package app

import (
	_ "github.com/MaxRadzey/shortener/docs"

	"github.com/MaxRadzey/shortener/internal/audit"
	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/router"
	"github.com/MaxRadzey/shortener/internal/service"
	"github.com/MaxRadzey/shortener/internal/storage"
	"go.uber.org/zap"
)

// Run инициализирует логгер, хранилище, сервис, хендлеры и запускает HTTP-сервер.
func Run(AppConfig *config.Config) error {
	if err := logger.Initialize(AppConfig.LogLevel); err != nil {
		return err
	}

	storageResult, err := storage.InitializeStorage(AppConfig.DatabaseDSN, AppConfig.FilePath)
	if err != nil {
		return err
	}

	urlService := service.NewService(storageResult.Repository, *AppConfig)
	auditNotifier := audit.NewNotifier(AppConfig.AuditFile, AppConfig.AuditURL)
	h := &httphandlers.Handler{Service: urlService, Audit: auditNotifier}

	r := router.SetupRouter(h, AppConfig)

	logger.Log.Info("Starting HTTP server", zap.String("address", AppConfig.Address))
	return r.Run(AppConfig.Address)
}
