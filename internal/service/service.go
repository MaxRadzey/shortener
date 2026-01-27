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

func (s *Service) CreateShortURL(longURL, userID string) (string, error) {
	shortPath, err := utils.GetShortPath(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate short path: %w", err)
	}

	item := dbstorage.URLEntry{ShortPath: shortPath, FullURL: longURL, UserID: userID}
	err = s.storage.Create(item)
	if err != nil {
		var urlExistsErr *dbstorage.ErrURLAlreadyExists
		if errors.As(err, &urlExistsErr) {
			existingURL := fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, urlExistsErr.ShortPath)
			return existingURL, &ErrURLConflict{ShortURL: existingURL}
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, shortPath), nil
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
func (s *Service) CreateShortURLBatch(ctx context.Context, items []models.BatchRequestItem, userID string) ([]models.BatchResponseItem, error) {
	entries := make([]dbstorage.URLEntry, 0, len(items))
	responseItems := make([]models.BatchResponseItem, 0, len(items))

	for _, item := range items {
		shortPath, err := utils.GetShortPath(item.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short path: %w", err)
		}

		entries = append(entries, dbstorage.URLEntry{
			ShortPath: shortPath,
			FullURL:   item.OriginalURL,
			UserID:    userID,
		})

		shortURL := fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, shortPath)
		responseItems = append(responseItems, models.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	err := s.storage.CreateBatch(ctx, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch URLs: %w", err)
	}

	return responseItems, nil
}

// GetUserURLs возвращает все сокращённые пользователем URL.
// При отсутствии записей — пустой слайс; хендлер в таком случае отдаёт 204.
func (s *Service) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLItem, error) {
	rows, err := s.storage.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]models.UserURLItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.UserURLItem{
			ShortURL:    fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, r.ShortPath),
			OriginalURL: r.OriginalURL,
		})
	}
	return out, nil
}
