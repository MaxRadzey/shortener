package storage

import (
	"context"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]URLEntry
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]URLEntry),
	}
}

func (m *MemoryStorage) Get(short string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.data[short]
	if !ok {
		return "", &ErrNotFound{ShortPath: short}
	}
	if r.IsDeleted {
		return "", &ErrGone{ShortPath: short}
	}
	return r.FullURL, nil
}

func (m *MemoryStorage) Create(item URLEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[item.ShortPath] = item
	return nil
}

func (m *MemoryStorage) CreateBatch(ctx context.Context, items []URLEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range items {
		m.data[item.ShortPath] = item
	}
	return nil
}

func (m *MemoryStorage) GetByUserID(ctx context.Context, userID string) ([]UserURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []UserURL
	for short, r := range m.data {
		if r.UserID == userID {
			out = append(out, UserURL{ShortPath: short, OriginalURL: r.FullURL})
		}
	}
	return out, nil
}

func (m *MemoryStorage) DeleteBatch(ctx context.Context, userID string, shortPaths []string) error {
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

func (m *MemoryStorage) Ping(ctx context.Context) error {
	// In-memory хранилище всегда доступно
	return nil
}
