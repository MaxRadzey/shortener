package app

import (
	"net/http"

	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

// startHTTPServer запускает HTTP-сервер.
func (a *App) startHTTPServer() {
	logger.Log.Info("Starting HTTP server", zap.String("address", a.config.Address))
	var err error
	if a.config.EnableHTTPS {
		logger.Log.Info("HTTPS mode enabled")
		err = a.server.ListenAndServeTLS("server.crt", "server.key")
	} else {
		err = a.server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		logger.Log.Error("HTTP server error", zap.Error(err))
	}
}
