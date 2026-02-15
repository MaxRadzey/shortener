package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/repository"
	"github.com/MaxRadzey/shortener/internal/utils"
)

// Service реализует бизнес-логику сокращения URL и работы с репозиторием.
type Service struct {
	repo      repository.URLRepository
	appConfig config.Config
}

// NewService возвращает сервис с заданным репозиторием и конфигом.
func NewService(repo repository.URLRepository, appConfig config.Config) *Service {
	return &Service{
		repo:      repo,
		appConfig: appConfig,
	}
}

// buildShortURL собирает полный short URL без fmt.Sprintf для меньших аллокаций.
func (s *Service) buildShortURL(shortPath string) string {
	var b strings.Builder
	b.Grow(len(s.appConfig.ReturningAddress) + 1 + len(shortPath))
	b.WriteString(s.appConfig.ReturningAddress)
	b.WriteByte('/')
	b.WriteString(shortPath)
	return b.String()
}

// CreateShortURL создаёт короткий путь для longURL и сохраняет в репозитории; при дубликате возвращает ErrURLConflict.
func (s *Service) CreateShortURL(longURL, userID string) (string, error) {
	shortPath, err := utils.GetShortPath(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate short path: %w", err)
	}

	item := models.URLEntry{ShortPath: shortPath, FullURL: longURL, UserID: userID}
	err = s.repo.Create(item)
	if err != nil {
		var urlExistsErr *repository.ErrURLAlreadyExists
		if errors.As(err, &urlExistsErr) {
			existingURL := s.buildShortURL(urlExistsErr.ShortPath)
			return existingURL, &ErrURLConflict{ShortURL: existingURL}
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	return s.buildShortURL(shortPath), nil
}

// GetLongURL возвращает оригинальный URL по short_path; ErrNotFound / ErrGone при отсутствии или удалении.
func (s *Service) GetLongURL(shortPath string) (string, error) {
	longURL, err := s.repo.Get(shortPath)
	if err != nil {
		// Преобразуем ошибки repository в ошибки service для изоляции слоёв
		var notFoundErr *repository.ErrNotFound
		var goneErr *repository.ErrGone
		if errors.As(err, &notFoundErr) {
			return "", &ErrNotFound{ShortPath: notFoundErr.ShortPath}
		}
		if errors.As(err, &goneErr) {
			return "", &ErrGone{ShortPath: goneErr.ShortPath}
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return longURL, nil
}

// Ping проверяет доступность репозитория.
func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// CreateShortURLBatch создаёт короткие ссылки для списка URL, сохраняет пачкой.
func (s *Service) CreateShortURLBatch(ctx context.Context, items []models.BatchRequestItem, userID string) ([]models.BatchResponseItem, error) {
	entries := make([]models.URLEntry, 0, len(items))
	responseItems := make([]models.BatchResponseItem, 0, len(items))

	for _, item := range items {
		shortPath, err := utils.GetShortPath(item.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short path: %w", err)
		}

		entries = append(entries, models.URLEntry{
			ShortPath: shortPath,
			FullURL:   item.OriginalURL,
			UserID:    userID,
		})

		responseItems = append(responseItems, models.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      s.buildShortURL(shortPath),
		})
	}

	err := s.repo.CreateBatch(ctx, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch URLs: %w", err)
	}

	return responseItems, nil
}

// GetUserURLs возвращает список коротких ссылок пользователя.
func (s *Service) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLItem, error) {
	rows, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]models.UserURLItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.UserURLItem{
			ShortURL:    s.buildShortURL(r.ShortPath),
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

// DeleteURLs помечает URL пользователя как удалённые пачками (асинхронно вызывается из хендлера).
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
		if err := s.repo.DeleteBatch(ctx, userID, buffer); err != nil {
			lastErr = fmt.Errorf("failed to delete batch: %w", err)
			// Продолжаем обработку остальных буферов
		}
	}

	return lastErr
}
