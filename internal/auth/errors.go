package auth

import (
	"errors"
	"fmt"
)

// Ошибки валидации куки аутентификации.
var (
	// ErrEmptyCookieValue — кука пустая или отсутствует.
	ErrEmptyCookieValue = errors.New("empty cookie value")
	// ErrInvalidCookieFormat — неверный формат (ожидается userID.signature).
	ErrInvalidCookieFormat = errors.New("invalid cookie format")
	// ErrInvalidSignature — подпись куки не совпадает.
	ErrInvalidSignature = errors.New("invalid signature")
)

// ErrInvalidUserID — user_id в куке не прошёл проверку (например, не UUID).
type ErrInvalidUserID struct {
	UserID string
	Err    error
}

// Error возвращает текстовое представление ошибки.
func (e *ErrInvalidUserID) Error() string {
	return fmt.Sprintf("invalid user ID %q: %v", e.UserID, e.Err)
}

// Unwrap возвращает вложенную ошибку.
func (e *ErrInvalidUserID) Unwrap() error {
	return e.Err
}
