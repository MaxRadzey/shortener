package grpc

import (
	"context"
	"strings"

	"github.com/MaxRadzey/shortener/internal/auth"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const metadataKeyAuthorization = "authorization"

// AuthUnaryInterceptor возвращает unary interceptor: читает authorization из metadata,
// проверяет подпись и кладёт user_id в контекст.
func AuthUnaryInterceptor(signingKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userID := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			vals := md.Get(metadataKeyAuthorization)
			if len(vals) > 0 {
				token := strings.TrimSpace(vals[0])
				if id, err := auth.ValidateAuthValue(token, signingKey); err == nil {
					userID = id
				}
			}
		}
		if userID != "" {
			ctx = context.WithValue(ctx, contextkeys.UserIDKey, userID)
		}
		return handler(ctx, req)
	}
}
