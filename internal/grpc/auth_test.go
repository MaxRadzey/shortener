package grpc

import (
	"context"
	"testing"

	"github.com/MaxRadzey/shortener/internal/auth"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthUnaryInterceptor(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New().String()
	validToken := auth.NewCookie(userID, secret).Value

	interceptor := AuthUnaryInterceptor(secret)

	t.Run("valid token in metadata sets userID in context", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", validToken))
		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			v := ctx.Value(contextkeys.UserIDKey)
			require.NotNil(t, v)
			assert.Equal(t, userID, v.(string))
			return "ok", nil
		}
		out, err := interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", out)
	})

	t.Run("invalid token leaves userID empty", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "invalid.token"))
		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			v := ctx.Value(contextkeys.UserIDKey)
			assert.Nil(t, v)
			return "ok", nil
		}
		out, err := interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", out)
	})

	t.Run("no metadata leaves userID empty", func(t *testing.T) {
		ctx := context.Background()
		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			v := ctx.Value(contextkeys.UserIDKey)
			assert.Nil(t, v)
			return "ok", nil
		}
		out, err := interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", out)
	})
}
