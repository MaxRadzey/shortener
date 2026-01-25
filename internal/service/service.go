package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/models"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
)

type Service struct {
	storage   dbstorage.URLStorage
	appConfig config.Config
}

func NewService(storage dbstorage.URLStorage, appConfig config.Config) *Service {
	return &Service{
		storage:   storage,
		appConfig: appConfig,
	}
}

func (s *Service) CreateShortURL(longURL string) (string, error) {
	shortPath, err := utils.GetShortPath(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate short path: %w", err)
	}

	err = s.storage.Create(shortPath, longURL)
	if err != nil {
		// Проверяем, является ли ошибка конфликтом существующего URL
		var urlExistsErr *dbstorage.ErrURLAlreadyExists
		if errors.As(err, &urlExistsErr) {
			// Формируем полный URL для существующего short_path
			existingURL := fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, urlExistsErr.ShortPath)
			return existingURL, &ErrURLConflict{ShortURL: existingURL}
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	result := fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, shortPath)
	return result, nil
}

func (s *Service) GetLongURL(shortPath string) (string, error) {
	longURL, err := s.storage.Get(shortPath)
	if err != nil {
		return "", err
	}

	return longURL, nil
}

// Ping проверяет соединение с хранилищем.
// Возвращает ошибку, если хранилище недоступно.
func (s *Service) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

// CreateShortURLBatch создает короткие URL для множества URL в одном запросе.
// Генерирует короткие пути и сохраняет их атомарно.
func (s *Service) CreateShortURLBatch(ctx context.Context, items []models.BatchRequestItem) ([]models.BatchResponseItem, error) {
	batchItems := make([]dbstorage.BatchItem, 0, len(items))
	responseItems := make([]models.BatchResponseItem, 0, len(items))

	for _, item := range items {
		shortPath, err := utils.GetShortPath(item.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short path: %w", err)
		}

		batchItems = append(batchItems, dbstorage.BatchItem{
			ShortPath: shortPath,
			FullURL:   item.OriginalURL,
		})

		shortURL := fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, shortPath)
		responseItems = append(responseItems, models.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	// Сохраняем все записи атомарно
	err := s.storage.CreateBatch(ctx, batchItems)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch URLs: %w", err)
	}

	return responseItems, nil
}
