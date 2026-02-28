package grpc

import (
	"net"

	"github.com/MaxRadzey/shortener/internal/audit"
	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/grpc/proto"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/service"
	"go.uber.org/zap"
	libgrpc "google.golang.org/grpc"
)

// NewServer создаёт gRPC-сервер и listener: interceptor auth, регистрация ShortenerService.
func NewServer(cfg *config.Config, svc *service.Service, auditNotifier *audit.Notifier) (*libgrpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return nil, nil, err
	}
	srv := libgrpc.NewServer(libgrpc.UnaryInterceptor(AuthUnaryInterceptor(cfg.SigningKey)))
	proto.RegisterShortenerServiceServer(srv, &Server{
		Service: svc,
		Audit:   auditNotifier,
		SignKey: cfg.SigningKey,
	})
	return srv, listener, nil
}

// Run запускает gRPC-сервер (блокирующий вызов).
func Run(server *libgrpc.Server, listener net.Listener) {
	logger.Log.Info("Starting gRPC server", zap.String("address", listener.Addr().String()))
	if err := server.Serve(listener); err != nil {
		logger.Log.Error("gRPC server error", zap.Error(err))
	}
}
