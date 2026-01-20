package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/models"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrValidation = errors.New("validation error")
)

type Service struct {
	storage   dbstorage.URLStorage
	appConfig config.Config
	db        *pgxpool.Pool
}

func NewService(storage dbstorage.URLStorage, appConfig config.Config, db *pgxpool.Pool) *Service {
	return &Service{
		storage:   storage,
		appConfig: appConfig,
		db:        db,
	}
}

func (s *Service) CreateShortURL(longURL string) (string, error) {
	if !utils.IsValidURL(longURL) {
		return "", ErrValidation
	}

	shortPath, err := utils.GetShortPath(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate short path: %w", err)
	}

	err = s.storage.Create(shortPath, longURL)
	if err != nil {
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

// Ping проверяет соединение с базой данных.
// Возвращает ошибку, если БД недоступна или соединение не установлено.
func (s *Service) Ping(ctx context.Context) error {
	if s.db == nil {
		return errors.New("database connection not available")
	}
	return s.db.Ping(ctx)
}

// CreateShortURLBatch создает короткие URL для множества URL в одном запросе.
// Валидирует все URL перед обработкой, генерирует короткие пути и сохраняет их атомарно.
func (s *Service) CreateShortURLBatch(ctx context.Context, items []models.BatchRequestItem) ([]models.BatchResponseItem, error) {
	batchItems := make([]dbstorage.BatchItem, 0, len(items))
	responseItems := make([]models.BatchResponseItem, 0, len(items))

	for _, item := range items {
		if !utils.IsValidURL(item.OriginalURL) {
			return nil, ErrValidation
		}

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
