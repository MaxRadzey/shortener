package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/MaxRadzey/shortener/internal/audit"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/MaxRadzey/shortener/internal/grpc/proto"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/service"
	"github.com/MaxRadzey/shortener/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server реализует ShortenerService и является фасадом над service.Service.
type Server struct {
	proto.UnimplementedShortenerServiceServer
	service *service.Service
	audit   *audit.Notifier
	signKey string
}

// NewGRPCServer возвращает gRPC-фасад с приватными полями (в т.ч. signKey не экспортируется).
func NewGRPCServer(svc *service.Service, auditNotifier *audit.Notifier, signKey string) *Server {
	return &Server{
		service: svc,
		audit:   auditNotifier,
		signKey: signKey,
	}
}

func userIDFromContext(ctx context.Context) string {
	v := ctx.Value(contextkeys.UserIDContextKey)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func (s *Server) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	userID := userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization")
	}
	urlStr := strings.TrimSpace(req.GetUrl())
	if urlStr == "" || !utils.IsValidURL(urlStr) {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}
	result, err := s.service.CreateShortURL(urlStr, userID)
	if err != nil {
		var conflict *service.ErrURLConflict
		if errors.As(err, &conflict) {
			if s.audit != nil {
				s.audit.Notify(audit.NewEvent("shorten", userID, urlStr))
			}
			return &proto.URLShortenResponse{Result: conflict.ShortURL}, nil
		}
		logger.Log.Error("ShortenURL", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	if s.audit != nil {
		s.audit.Notify(audit.NewEvent("shorten", userID, urlStr))
	}
	return &proto.URLShortenResponse{Result: result}, nil
}

func (s *Server) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	id := strings.TrimSpace(req.GetId())
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing id")
	}
	longURL, err := s.service.GetLongURL(id)
	if err != nil {
		var notFound *service.ErrNotFound
		var gone *service.ErrGone
		if errors.As(err, &notFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		if errors.As(err, &gone) {
			return nil, status.Error(codes.FailedPrecondition, "url deleted")
		}
		logger.Log.Error("ExpandURL", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &proto.URLExpandResponse{Result: longURL}, nil
}

func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID := userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization")
	}
	items, err := s.service.GetUserURLs(ctx, userID)
	if err != nil {
		logger.Log.Error("ListUserURLs", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := make([]*proto.URLData, 0, len(items))
	for _, it := range items {
		out = append(out, &proto.URLData{ShortUrl: it.ShortURL, OriginalUrl: it.OriginalURL})
	}
	return &proto.UserURLsResponse{Url: out}, nil
}
