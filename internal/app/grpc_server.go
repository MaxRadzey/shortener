package app

import (
	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

// startGRPCServer запускает gRPC-сервер.
func (a *App) startGRPCServer() {
	logger.Log.Info("Starting gRPC server", zap.String("address", a.config.GRPCAddress))
	if err := a.grpcServer.Serve(a.grpcListener); err != nil {
		logger.Log.Error("gRPC server error", zap.Error(err))
	}
}
