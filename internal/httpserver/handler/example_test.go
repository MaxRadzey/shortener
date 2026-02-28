package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/MaxRadzey/shortener/internal/httpserver"
	"github.com/MaxRadzey/shortener/internal/httpserver/handler"
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/service"
	teststorage "github.com/MaxRadzey/shortener/internal/testing"
	"github.com/gin-gonic/gin"
)

var exampleConfig = &config.Config{
	ReturningAddress: "http://localhost:8080",
	SigningKey:       "example-key",
}

func ExampleHandler_CreateURL() {
	repo := teststorage.NewFakeRepositoryWithData(nil)
	svc := service.NewService(repo, *exampleConfig)
	h := &handler.Handler{Service: svc}
	gin.SetMode(gin.TestMode)
	r := httpserver.SetupRouter(h, exampleConfig)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/page"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	println("Status:", w.Code)
	println("Body:", strings.TrimSpace(w.Body.String()))
}

func ExampleHandler_GetURL() {
	repo := teststorage.NewFakeRepositoryWithData(map[string]string{"abc123": "https://ya.ru"})
	svc := service.NewService(repo, *exampleConfig)
	h := &handler.Handler{Service: svc}
	gin.SetMode(gin.TestMode)
	r := httpserver.SetupRouter(h, exampleConfig)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	println("Status:", w.Code)
	println("Location:", w.Header().Get("Location"))
}

func ExampleHandler_GetURLJSON() {
	repo := teststorage.NewFakeRepositoryWithData(nil)
	svc := service.NewService(repo, *exampleConfig)
	h := &handler.Handler{Service: svc}
	gin.SetMode(gin.TestMode)
	r := httpserver.SetupRouter(h, exampleConfig)

	body, _ := json.Marshal(models.Request{URL: "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp models.Response
	_ = json.NewDecoder(w.Body).Decode(&resp)
	println("Status:", w.Code)
	println("Result:", resp.Result)
}

func ExampleHandler_CreateURLBatch() {
	repo := teststorage.NewFakeRepositoryWithData(nil)
	svc := service.NewService(repo, *exampleConfig)
	h := &handler.Handler{Service: svc}
	gin.SetMode(gin.TestMode)
	r := httpserver.SetupRouter(h, exampleConfig)

	body, _ := json.Marshal([]models.BatchRequestItem{
		{CorrelationID: "1", OriginalURL: "https://a.com"},
		{CorrelationID: "2", OriginalURL: "https://b.com"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	println("Status:", w.Code)
}

func ExampleHandler_Ping() {
	repo := teststorage.NewFakeRepositoryWithData(nil)
	svc := service.NewService(repo, *exampleConfig)
	h := &handler.Handler{Service: svc}
	gin.SetMode(gin.TestMode)
	r := httpserver.SetupRouter(h, exampleConfig)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	println("Status:", w.Code)
	println("Body:", strings.TrimSpace(w.Body.String()))
}

func ExampleHandler_GetUserURLs() {
	userID := "user-1"
	repo := teststorage.NewFakeRepositoryWithEntries(map[string]models.URLEntry{
		"abc": {ShortPath: "abc", FullURL: "https://first.com", UserID: userID},
		"def": {ShortPath: "def", FullURL: "https://second.com", UserID: userID},
	})
	svc := service.NewService(repo, *exampleConfig)
	h := &handler.Handler{Service: svc}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(contextkeys.UserIDKey, userID)
		c.Next()
	})
	r.GET("/api/user/urls", h.GetUserURLs)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	println("Status:", w.Code)
}

func ExampleHandler_DeleteURLs() {
	userID := "user-1"
	repo := teststorage.NewFakeRepositoryWithEntries(map[string]models.URLEntry{
		"abc": {ShortPath: "abc", FullURL: "https://first.com", UserID: userID},
	})
	svc := service.NewService(repo, *exampleConfig)
	h := &handler.Handler{Service: svc}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(contextkeys.UserIDKey, userID)
		c.Next()
	})
	r.DELETE("/api/user/urls", h.DeleteURLs)

	body, _ := json.Marshal([]string{"http://localhost:8080/abc"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	println("Status:", w.Code)
}
