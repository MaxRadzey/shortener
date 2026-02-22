// Package testing содержит хелперы для тестов (фейковые репозитории).
package testing

import (
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/repository"
	"github.com/MaxRadzey/shortener/internal/repository/memory"
)

// NewFakeRepository создает новый экземпляр MemoryRepository с пустыми данными.
// Используется в тестах как легковесное in-memory хранилище.
func NewFakeRepository() repository.URLRepository {
	return memory.NewMemoryRepository()
}

// NewFakeRepositoryWithData создает MemoryRepository с предзаполненными данными.
// data - map[shortPath]fullURL
func NewFakeRepositoryWithData(data map[string]string) repository.URLRepository {
	repo := memory.NewMemoryRepository()
	for short, fullURL := range data {
		_ = repo.Create(models.URLEntry{ShortPath: short, FullURL: fullURL, UserID: ""})
	}
	return repo
}

// NewFakeRepositoryWithEntries создает MemoryRepository с указанными записями.
func NewFakeRepositoryWithEntries(entries map[string]models.URLEntry) repository.URLRepository {
	repo := memory.NewMemoryRepository()
	for _, entry := range entries {
		_ = repo.Create(entry)
	}
	return repo
}
