// Package app собирает логгер, хранилище, сервис, роутер и запускает HTTP-сервер.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

const shutdownTimeout = 30 * time.Second

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

	srv := &http.Server{
		Addr:    AppConfig.Address,
		Handler: r,
	}

	if AppConfig.EnableHTTPS {
		if AppConfig.TLSCertFile == "" || AppConfig.TLSKeyFile == "" {
			return errors.New("HTTPS enabled: -cert and -key (or TLS_CERT_FILE and TLS_KEY_FILE) are required")
		}
		logger.Log.Info("Starting HTTPS server", zap.String("address", AppConfig.Address))
		go func() {
			if err := srv.ListenAndServeTLS(AppConfig.TLSCertFile, AppConfig.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				logger.Log.Error("HTTPS server error", zap.Error(err))
			}
		}()
	} else {
		logger.Log.Info("Starting HTTP server", zap.String("address", AppConfig.Address))
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log.Error("HTTP server error", zap.Error(err))
			}
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigChan

	logger.Log.Info("Shutdown signal received, draining connections")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	auditNotifier.Shutdown()
	if storageResult.DB != nil {
		storageResult.DB.Close()
	}
	logger.Log.Info("Server stopped")
	return nil
}
