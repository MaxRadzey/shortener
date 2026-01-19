package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MaxRadzey/shortener/internal/app"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestPingIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("DATABASE_DSN is not set; skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	gin.SetMode(gin.TestMode)

	handler := &httphandlers.Handler{
		Storage:   &FakeStorage{data: map[string]string{}},
		AppConfig: *AppConfig,
		DBPool:    pool,
	}

	router := app.SetupRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
