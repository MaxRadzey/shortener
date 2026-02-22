// Package app собирает логгер, хранилище, сервис, роутер и запускает HTTP-сервер.
package app

import (
	"errors"
	"fmt"

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

	if AppConfig.EnableHTTPS {
		if AppConfig.TLSCertFile == "" || AppConfig.TLSKeyFile == "" {
			return errors.New("HTTPS enabled: -cert and -key (or TLS_CERT_FILE and TLS_KEY_FILE) are required")
		}
		logger.Log.Info("Starting HTTPS server", zap.String("address", AppConfig.Address))
		return r.RunTLS(AppConfig.Address, AppConfig.TLSCertFile, AppConfig.TLSKeyFile)
	}
	logger.Log.Info("Starting HTTP server", zap.String("address", AppConfig.Address))
	if err := r.Run(AppConfig.Address); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
