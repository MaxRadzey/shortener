package auth

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyCookieValue = errors.New("empty cookie value")
	ErrInvalidCookieFormat = errors.New("invalid cookie format")
	ErrInvalidSignature = errors.New("invalid signature")
)

// ErrInvalidUserID представляет ошибку невалидного user ID
type ErrInvalidUserID struct {
	UserID string
	Err    error
}

func (e *ErrInvalidUserID) Error() string {
	return fmt.Sprintf("invalid user ID %q: %v", e.UserID, e.Err)
}

func (e *ErrInvalidUserID) Unwrap() error {
	return e.Err
}

