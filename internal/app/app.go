// Package app собирает логгер, хранилище, сервис и запускает HTTP- и gRPC-серверы.
package app

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/MaxRadzey/shortener/docs"

	"github.com/MaxRadzey/shortener/internal/audit"
	"github.com/MaxRadzey/shortener/internal/config"
	grpcpkg "github.com/MaxRadzey/shortener/internal/grpc"
	"github.com/MaxRadzey/shortener/internal/httpserver"
	"github.com/MaxRadzey/shortener/internal/httpserver/handler"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/service"
	"github.com/MaxRadzey/shortener/internal/storage"
	"go.uber.org/zap"
	libgrpc "google.golang.org/grpc"
)

const shutdownTimeout = 30 * time.Second

// App — приложение: конфиг, HTTP- и gRPC-серверы и хранилище.
type App struct {
	config       *config.Config
	server       *http.Server
	grpcServer   *libgrpc.Server
	grpcListener net.Listener
	storage      *storage.StorageResult
}

// New создаёт приложение: инициализирует логгер, хранилище, сервис и оба сервера (HTTP и gRPC).
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

	httpHandler := &handler.Handler{
		Service:       urlService,
		Audit:         auditNotifier,
		TrustedSubnet: cfg.TrustedSubnet,
	}
	server := httpserver.New(cfg, httpHandler)

	grpcServer, grpcListener, err := grpcpkg.Setup(cfg, urlService, auditNotifier)
	if err != nil {
		return nil, err
	}

	return &App{
		config:       cfg,
		server:       server,
		grpcServer:   grpcServer,
		grpcListener: grpcListener,
		storage:      storageResult,
	}, nil
}

// Run запускает сервер, ждёт сигнал завершения (SIGTERM, SIGINT, SIGQUIT) и корректно останавливает приложение.
func (a *App) Run() error {
	a.startServer()
	<-a.shutdownSignal()
	return a.gracefulShutdown()
}

func (a *App) startServer() {
	go httpserver.Run(a.server, a.config)
	go grpcpkg.Run(a.grpcServer, a.grpcListener)
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

	if err := a.grpcListener.Close(); err != nil {
		logger.Log.Error("gRPC listener close error", zap.Error(err))
	}
	a.grpcServer.GracefulStop()

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
