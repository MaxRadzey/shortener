package service

import "fmt"

// ErrValidation — невалидный URL.
type ErrValidation struct {
	URL string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error: invalid URL %q", e.URL)
}

// ErrURLConflict — URL уже сокращён, в ShortURL лежит существующая короткая ссылка.
type ErrURLConflict struct {
	ShortURL string
}

func (e *ErrURLConflict) Error() string {
	return fmt.Sprintf("url already exists: %s", e.ShortURL)
}

// ErrNotFound — запись по short_path не найдена.
type ErrNotFound struct {
	ShortPath string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("url not found with short_path: %s", e.ShortPath)
}

// ErrGone — запись найдена, но помечена удалённой (410).
type ErrGone struct {
	ShortPath string
}

func (e *ErrGone) Error() string {
	return fmt.Sprintf("url is deleted: %s", e.ShortPath)
}
