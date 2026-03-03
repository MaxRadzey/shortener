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
	"google.golang.org/grpc/credentials"
)

// Setup создаёт gRPC-сервер и listener.
func Setup(cfg *config.Config, svc *service.Service, auditNotifier *audit.Notifier) (*libgrpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return nil, nil, err
	}
	opts := []libgrpc.ServerOption{libgrpc.UnaryInterceptor(AuthUnaryInterceptor(cfg.SigningKey))}
	if cfg.EnableHTTPS {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			listener.Close()
			return nil, nil, err
		}
		opts = append(opts, libgrpc.Creds(creds))
	}
	srv := libgrpc.NewServer(opts...)
	proto.RegisterShortenerServiceServer(srv, NewGRPCServer(svc, auditNotifier, cfg.SigningKey))
	return srv, listener, nil
}

// Run запускает gRPC-сервер.
func Run(server *libgrpc.Server, listener net.Listener) {
	logger.Log.Info("Starting gRPC server", zap.String("address", listener.Addr().String()))
	if err := server.Serve(listener); err != nil {
		logger.Log.Error("gRPC server error", zap.Error(err))
	}
}
