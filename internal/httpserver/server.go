// Package httpserver содержит создание и запуск HTTP-сервера приложения.
package httpserver

import (
	"net/http"
	"time"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/httpserver/handler"
	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

// New создаёт HTTP-сервер: роутер с хендлером и middleware, затем *http.Server.
func New(cfg *config.Config, h *handler.Handler) *http.Server {
	r := SetupRouter(h, cfg)
	return &http.Server{
		Addr:              cfg.Address,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// Run запускает HTTP-сервер (блокирующий вызов).
func Run(server *http.Server, cfg *config.Config) {
	logger.Log.Info("Starting HTTP server", zap.String("address", cfg.Address))
	var err error
	if cfg.EnableHTTPS {
		logger.Log.Info("HTTPS mode enabled")
		err = server.ListenAndServeTLS("server.crt", "server.key")
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		logger.Log.Error("HTTP server error", zap.Error(err))
	}
}
