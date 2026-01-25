package storage

import (
	"context"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (m *MemoryStorage) Get(short string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, ok := m.data[short]
	if !ok {
		return "", ErrNotFound
	}

	return url, nil
}

func (m *MemoryStorage) Create(short, full string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[short] = full
	return nil
}

func (m *MemoryStorage) CreateBatch(ctx context.Context, items []BatchItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Атомарно добавляем все записи в map
	for _, item := range items {
		m.data[item.ShortPath] = item.FullURL
	}

	return nil
}

func (m *MemoryStorage) Ping(ctx context.Context) error {
	// In-memory хранилище всегда доступно
	return nil
}
