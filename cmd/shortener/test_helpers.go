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
	data map[string]string
}

func (f *FakeStorage) Get(short string) (string, error) {
	val, ok := f.data[short]
	if !ok {
		return "", dbstorage.ErrNotFound
	}
	return val, nil
}

func (f *FakeStorage) Create(short, full string) error {
	f.data[short] = full
	return nil
}

func (f *FakeStorage) CreateBatch(ctx context.Context, items []dbstorage.BatchItem) error {
	for _, item := range items {
		f.data[item.ShortPath] = item.FullURL
	}
	return nil
}

// newFakeStorage создает новый экземпляр FakeStorage.
func newFakeStorage(data map[string]string) *FakeStorage {
	if data == nil {
		data = make(map[string]string)
	}
	return &FakeStorage{data: data}
}

// setupTestHandler создает handler для тестов с указанным хранилищем.
func setupTestHandler(storage dbstorage.URLStorage) *httphandlers.Handler {
	urlService := service.NewService(storage, *AppConfig, nil)
	return &httphandlers.Handler{Service: urlService}
}

// setupTestRouter создает роутер для тестов с указанным handler.
func setupTestRouter(handler *httphandlers.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return router.SetupRouter(handler)
}
