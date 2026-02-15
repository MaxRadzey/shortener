package repository

import "fmt"

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

// ErrGone — запись найдена, но помечена удалённой (is_deleted).
type ErrGone struct {
	ShortPath string
}

func (e *ErrGone) Error() string {
	return fmt.Sprintf("url is deleted: %s", e.ShortPath)
}
