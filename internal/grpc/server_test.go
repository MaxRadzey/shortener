package grpc

import (
	"context"
	"testing"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/MaxRadzey/shortener/internal/grpc/proto"
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/repository"
	"github.com/MaxRadzey/shortener/internal/service"
	teststorage "github.com/MaxRadzey/shortener/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var testGRPCConfig = config.Config{
	ReturningAddress: "http://localhost:8080",
	SigningKey:       "test-key",
}

func setupTestServer(t *testing.T, repo repository.URLRepository) *Server {
	t.Helper()
	cfg := testGRPCConfig
	svc := service.NewService(repo, cfg)
	return NewGRPCServer(svc, nil, cfg.SigningKey)
}

func ctxWithUser(userID string) context.Context {
	return context.WithValue(context.Background(), contextkeys.UserIDContextKey, userID)
}

func TestServer_ShortenURL(t *testing.T) {
	repo := teststorage.NewFakeRepositoryWithData(nil)
	srv := setupTestServer(t, repo)
	ctx := ctxWithUser("user-1")

	t.Run("success", func(t *testing.T) {
		resp, err := srv.ShortenURL(ctx, &proto.URLShortenRequest{Url: "https://example.com"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Contains(t, resp.Result, "http://localhost:8080/")
	})

	t.Run("no user in context returns Unauthenticated", func(t *testing.T) {
		_, err := srv.ShortenURL(context.Background(), &proto.URLShortenRequest{Url: "https://example.com"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("invalid url returns InvalidArgument", func(t *testing.T) {
		_, err := srv.ShortenURL(ctx, &proto.URLShortenRequest{Url: "not-a-url"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("empty url returns InvalidArgument", func(t *testing.T) {
		_, err := srv.ShortenURL(ctx, &proto.URLShortenRequest{Url: ""})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("same url twice returns same short URL", func(t *testing.T) {
		repo2 := teststorage.NewFakeRepositoryWithData(nil)
		srv2 := setupTestServer(t, repo2)
		url := "https://example.com/again"
		resp1, err := srv2.ShortenURL(ctx, &proto.URLShortenRequest{Url: url})
		require.NoError(t, err)
		require.NotEmpty(t, resp1.Result)
		resp2, err := srv2.ShortenURL(ctx, &proto.URLShortenRequest{Url: url})
		require.NoError(t, err)
		assert.Equal(t, resp1.Result, resp2.Result)
	})
}

func TestServer_ExpandURL(t *testing.T) {
	repo := teststorage.NewFakeRepositoryWithEntries(map[string]models.URLEntry{
		"abc":  {ShortPath: "abc", FullURL: "https://example.com", UserID: "u1", IsDeleted: false},
		"gone": {ShortPath: "gone", FullURL: "https://deleted.com", UserID: "u1", IsDeleted: true},
	})
	srv := setupTestServer(t, repo)

	t.Run("success", func(t *testing.T) {
		resp, err := srv.ExpandURL(context.Background(), &proto.URLExpandRequest{Id: "abc"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "https://example.com", resp.Result)
	})

	t.Run("not found returns NotFound", func(t *testing.T) {
		_, err := srv.ExpandURL(context.Background(), &proto.URLExpandRequest{Id: "nonexistent"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("deleted returns FailedPrecondition", func(t *testing.T) {
		_, err := srv.ExpandURL(context.Background(), &proto.URLExpandRequest{Id: "gone"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
	})

	t.Run("empty id returns InvalidArgument", func(t *testing.T) {
		_, err := srv.ExpandURL(context.Background(), &proto.URLExpandRequest{Id: ""})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestServer_ListUserURLs(t *testing.T) {
	userID := "user-123"
	repo := teststorage.NewFakeRepositoryWithEntries(map[string]models.URLEntry{
		"a1": {ShortPath: "a1", FullURL: "https://a.com", UserID: userID},
		"b2": {ShortPath: "b2", FullURL: "https://b.com", UserID: userID},
		"c3": {ShortPath: "c3", FullURL: "https://c.com", UserID: "other-user"},
	})
	srv := setupTestServer(t, repo)
	ctx := ctxWithUser(userID)

	t.Run("success returns only user urls", func(t *testing.T) {
		resp, err := srv.ListUserURLs(ctx, &emptypb.Empty{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Url, 2)
		shortURLs := make(map[string]string)
		for _, u := range resp.Url {
			shortURLs[u.ShortUrl] = u.OriginalUrl
		}
		assert.Equal(t, "https://a.com", shortURLs["http://localhost:8080/a1"])
		assert.Equal(t, "https://b.com", shortURLs["http://localhost:8080/b2"])
	})

	t.Run("no user in context returns Unauthenticated", func(t *testing.T) {
		_, err := srv.ListUserURLs(context.Background(), &emptypb.Empty{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("empty list for user with no urls", func(t *testing.T) {
		emptyRepo := teststorage.NewFakeRepository()
		srv2 := setupTestServer(t, emptyRepo)
		resp, err := srv2.ListUserURLs(ctx, &emptypb.Empty{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Url)
	})
}
