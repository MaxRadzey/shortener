package testing

import (
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
)

// NewFakeStorage создает новый экземпляр MemoryStorage с пустыми данными.
// Используется в тестах как легковесное in-memory хранилище.
func NewFakeStorage() dbstorage.URLStorage {
	return dbstorage.NewMemoryStorage()
}

// NewFakeStorageWithData создает MemoryStorage с предзаполненными данными.
// data - map[shortPath]fullURL
func NewFakeStorageWithData(data map[string]string) dbstorage.URLStorage {
	storage := dbstorage.NewMemoryStorage()
	for short, fullURL := range data {
		_ = storage.Create(dbstorage.URLEntry{ShortPath: short, FullURL: fullURL, UserID: ""})
	}
	return storage
}

// NewFakeStorageWithEntries создает MemoryStorage с указанными записями.
func NewFakeStorageWithEntries(entries map[string]dbstorage.URLEntry) dbstorage.URLStorage {
	storage := dbstorage.NewMemoryStorage()
	for _, entry := range entries {
		_ = storage.Create(entry)
	}
	return storage
}
