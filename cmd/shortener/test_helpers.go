package main

import (
	"context"

	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/router"
	"github.com/MaxRadzey/shortener/internal/service"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

var AppConfig = config.New()

// FakeStorage - мок хранилища для тестов.
type FakeStorage struct {
	data map[string]dbstorage.URLEntry
}

func (f *FakeStorage) Get(short string) (string, error) {
	val, ok := f.data[short]
	if !ok {
		return "", &dbstorage.ErrNotFound{ShortPath: short}
	}
	return val.FullURL, nil
}

func (f *FakeStorage) Create(item dbstorage.URLEntry) error {
	f.data[item.ShortPath] = item
	return nil
}

func (f *FakeStorage) CreateBatch(ctx context.Context, items []dbstorage.URLEntry) error {
	for _, item := range items {
		f.data[item.ShortPath] = item
	}
	return nil
}

func (f *FakeStorage) GetByUserID(ctx context.Context, userID string) ([]dbstorage.UserURL, error) {
	var out []dbstorage.UserURL
	for short, entry := range f.data {
		if entry.UserID == userID {
			out = append(out, dbstorage.UserURL{ShortPath: short, OriginalURL: entry.FullURL})
		}
	}
	return out, nil
}

func (f *FakeStorage) Ping(ctx context.Context) error {
	return nil
}

// newFakeStorage создает новый экземпляр FakeStorage.
func newFakeStorage(data map[string]string) *FakeStorage {
	entries := make(map[string]dbstorage.URLEntry)
	for short, fullURL := range data {
		entries[short] = dbstorage.URLEntry{ShortPath: short, FullURL: fullURL, UserID: ""}
	}
	return &FakeStorage{data: entries}
}

// setupTestHandler создает handler для тестов с указанным хранилищем.
func setupTestHandler(storage dbstorage.URLStorage) *httphandlers.Handler {
	urlService := service.NewService(storage, *AppConfig)
	return &httphandlers.Handler{Service: urlService}
}

// setupTestRouter создает роутер для тестов с указанным handler.
func setupTestRouter(handler *httphandlers.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return router.SetupRouter(handler, AppConfig)
}
