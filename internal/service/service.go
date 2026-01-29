package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

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


const (
	// deleteBatchSize размер буфера для batch update при удалении URL
	deleteBatchSize = 100
	// deleteWorkers количество воркеров для обработки удаления
	deleteWorkers = 3
)

// DeleteURLs удаляет сокращённые URL по списку идентификаторов для указанного пользователя.
// Использует паттерн fanIn для эффективного batch update: несколько воркеров собирают
// shortPaths в буферы, которые затем обрабатываются batch update операциями.
// Вызывается асинхронно из хендлера, поэтому ошибки логируются, но не возвращаются пользователю.
func (s *Service) DeleteURLs(ctx context.Context, userID string, shortUrls []string) error {
	// Создаем входной канал для shortPaths
	inputChan := make(chan string, len(shortUrls))
	for _, shortPath := range shortUrls {
		inputChan <- shortPath
	}
	close(inputChan)

	// Создаем канал для буферов (fan-in паттерн)
	bufferChan := make(chan []string, deleteWorkers)

	// Запускаем воркеры для сбора данных в буферы
	var wg sync.WaitGroup
	for i := 0; i < deleteWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := make([]string, 0, deleteBatchSize)

			for shortPath := range inputChan {
				buffer = append(buffer, shortPath)

				// Когда буфер заполнен, отправляем его на обработку
				if len(buffer) >= deleteBatchSize {
					bufferChan <- buffer
					buffer = make([]string, 0, deleteBatchSize)
				}
			}

			// Отправляем оставшиеся элементы
			if len(buffer) > 0 {
				bufferChan <- buffer
			}
		}()
	}

	// Закрываем канал буферов после завершения всех воркеров
	go func() {
		wg.Wait()
		close(bufferChan)
	}()

	// Обрабатываем буферы batch update операциями
	var lastErr error
	for buffer := range bufferChan {
		if err := s.storage.DeleteBatch(ctx, userID, buffer); err != nil {
			lastErr = fmt.Errorf("failed to delete batch: %w", err)
			// Продолжаем обработку остальных буферов
		}
	}

	return lastErr
}