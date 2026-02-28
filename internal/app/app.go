// Package app собирает логгер, хранилище, сервис, роутер и запускает HTTP-сервер.
package app

import (
	"context"
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

// App — приложение: конфиг, HTTP-сервер и хранилище.
type App struct {
	config  *config.Config
	server  *http.Server
	storage *storage.StorageResult
}

// New создаёт приложение: инициализирует логгер, хранилище, сервис, хендлеры, роутер и HTTP-сервер.
func New(cfg *config.Config) (*App, error) {
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return nil, err
	}

	storageResult, err := storage.InitializeStorage(cfg.DatabaseDSN, cfg.FilePath)
	if err != nil {
		return nil, err
	}

	urlService := service.NewService(storageResult.Repository, *cfg)
	auditNotifier := audit.NewNotifier(cfg.AuditFile, cfg.AuditURL)
	h := &httphandlers.Handler{Service: urlService, Audit: auditNotifier}
	r := router.SetupRouter(h, cfg)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	return &App{
		config:  cfg,
		server:  server,
		storage: storageResult,
	}, nil
}

// Run запускает сервер, ждёт сигнал завершения (SIGTERM, SIGINT, SIGQUIT) и корректно останавливает приложение.
func (a *App) Run() error {
	a.startServer()
	<-a.shutdownSignal()
	return a.gracefulShutdown()
}

func (a *App) startServer() {
	logger.Log.Info("Starting HTTP server", zap.String("address", a.config.Address))
	go func() {
		var err error
		if a.config.EnableHTTPS {
			logger.Log.Info("HTTPS mode enabled")
			err = a.server.ListenAndServeTLS("server.crt", "server.key")
		} else {
			err = a.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Error("server error", zap.Error(err))
		}
	}()
}

func (a *App) shutdownSignal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	return ch
}

func (a *App) gracefulShutdown() error {
	logger.Log.Info("Shutdown signal received, draining connections...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		logger.Log.Error("server shutdown error", zap.Error(err))
		return err
	}

	a.closeStorage()
	logger.Log.Info("Server stopped gracefully")
	return nil
}

func (a *App) closeStorage() {
	if a.storage != nil && a.storage.DB != nil {
		a.storage.DB.Close()
		logger.Log.Info("PostgreSQL connection pool closed")
	}
}
