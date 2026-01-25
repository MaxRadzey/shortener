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
