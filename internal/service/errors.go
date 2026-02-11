package service

import "fmt"

// ErrValidation представляет ошибку валидации URL
type ErrValidation struct {
	URL string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error: invalid URL %q", e.URL)
}

// ErrURLConflict представляет ошибку конфликта URL с уже существующим сокращённым URL
type ErrURLConflict struct {
	ShortURL string
}

func (e *ErrURLConflict) Error() string {
	return fmt.Sprintf("url already exists: %s", e.ShortURL)
}

// ErrNotFound представляет ошибку, когда URL не найден
type ErrNotFound struct {
	ShortPath string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("url not found with short_path: %s", e.ShortPath)
}

// ErrGone представляет ошибку, когда URL найден, но помечен как удалённый
type ErrGone struct {
	ShortPath string
}

func (e *ErrGone) Error() string {
	return fmt.Sprintf("url is deleted: %s", e.ShortPath)
}
