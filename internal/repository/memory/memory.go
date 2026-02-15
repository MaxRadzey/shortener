// Package memory реализует URLRepository с хранением данных в памяти (map).
package memory

import (
	"context"
	"sync"

	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/repository"
)

// MemoryRepository хранит URL в памяти (map).
type MemoryRepository struct {
	mu   sync.RWMutex
	data map[string]models.URLEntry
}

// NewMemoryRepository возвращает пустой in-memory репозиторий.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]models.URLEntry),
	}
}

// Get возвращает оригинальный URL по short path; ErrNotFound или ErrGone при отсутствии или удалении.
func (m *MemoryRepository) Get(short string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.data[short]
	if !ok {
		return "", &repository.ErrNotFound{ShortPath: short}
	}
	if r.IsDeleted {
		return "", &repository.ErrGone{ShortPath: short}
	}
	return r.FullURL, nil
}

// Create сохраняет запись в памяти.
func (m *MemoryRepository) Create(item models.URLEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[item.ShortPath] = item
	return nil
}

// CreateBatch сохраняет пачку записей в памяти.
func (m *MemoryRepository) CreateBatch(ctx context.Context, items []models.URLEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range items {
		m.data[item.ShortPath] = item
	}
	return nil
}

// GetByUserID возвращает все записи пользователя.
func (m *MemoryRepository) GetByUserID(ctx context.Context, userID string) ([]models.UserURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []models.UserURL
	for short, r := range m.data {
		if r.UserID == userID {
			out = append(out, models.UserURL{ShortPath: short, OriginalURL: r.FullURL})
		}
	}
	return out, nil
}

// DeleteBatch проставляет is_deleted у записей пользователя по списку short path.
func (m *MemoryRepository) DeleteBatch(ctx context.Context, userID string, shortPaths []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Проставляем флаг удаления только у записей, принадлежащих пользователю
	for _, shortPath := range shortPaths {
		if entry, exists := m.data[shortPath]; exists && entry.UserID == userID {
			entry.IsDeleted = true
			m.data[shortPath] = entry
		}
	}

	return nil
}

// Ping для in-memory всегда возвращает nil (хранилище доступно).
func (m *MemoryRepository) Ping(ctx context.Context) error {
	// In-memory хранилище всегда доступно
	return nil
}
