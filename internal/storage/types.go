package storage

import (
	"context"
	"fmt"
)

// ErrNotFound — ошибка, когда URL не найден в хранилище.
type ErrNotFound struct {
	ShortPath string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("url not found with short_path: %s", e.ShortPath)
}

// ErrURLAlreadyExists — ошибка, когда URL уже существует в хранилище.
type ErrURLAlreadyExists struct {
	ShortPath string
}

func (e *ErrURLAlreadyExists) Error() string {
	return fmt.Sprintf("url already exists with short_path: %s", e.ShortPath)
}

// URLEntry — запись для создания URL.
type URLEntry struct {
	ShortPath string
	FullURL   string
	UserID    string
}

// UserURL — short_path + original_url, используется в GetByUserID.
type UserURL struct {
	ShortPath   string
	OriginalURL string
}

// URLStorage — интерфейс хранилища URL.
type URLStorage interface {
	Get(short string) (string, error)
	Create(item URLEntry) error
	CreateBatch(ctx context.Context, items []URLEntry) error
	GetByUserID(ctx context.Context, userID string) ([]UserURL, error)
	Ping(ctx context.Context) error
}
