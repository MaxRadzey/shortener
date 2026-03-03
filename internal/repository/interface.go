// Package repository задаёт интерфейс хранилища URL и типы ошибок.
package repository

import (
	"context"

	"github.com/MaxRadzey/shortener/internal/models"
)

// URLRepository — контракт хранилища: Get/Create/CreateBatch, GetByUserID, DeleteBatch, Ping и тд.
type URLRepository interface {
	Get(short string) (string, error)
	Create(item models.URLEntry) error
	CreateBatch(ctx context.Context, items []models.URLEntry) error
	GetByUserID(ctx context.Context, userID string) ([]models.UserURL, error)
	DeleteBatch(ctx context.Context, userID string, shortPaths []string) error
	CountURLs(ctx context.Context) (int, error)
	CountUsers(ctx context.Context) (int, error)
	Ping(ctx context.Context) error
}
