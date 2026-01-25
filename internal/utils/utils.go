package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

// GetShortPath возвращает короткое уникальное строковое представление пути (URL),
// который был передан. Использует алгоритм шифрования sha1 и кодирование base64.
// Результат обрезается до 6 символов.
func GetShortPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("empty string cannot be shortened")
	}

	hasher := sha1.New()
	hasher.Write([]byte(path))
	hash := hasher.Sum(nil)

	short := base64.URLEncoding.EncodeToString(hash)[:6]
	return short, nil
}

// IsValidURL валидирует переданную строку и возвращает булево значение True,
// если строка - валидный URL, иначе False.
func IsValidURL(urlToCheck string) bool {
	_, err := url.ParseRequestURI(urlToCheck)
	return err == nil
}

// MaskDSN скрывает пароль в DSN для безопасного логирования.
// Простая маскировка - скрываем пароль после @
// postgres://user:password@host:port/db -> postgres://user:***@host:port/db
func MaskDSN(dsn string) string {
	if idx := strings.Index(dsn, "@"); idx > 0 {
		if passIdx := strings.LastIndex(dsn[:idx], ":"); passIdx > 0 {
			return dsn[:passIdx+1] + "***" + dsn[idx:]
		}
	}
	return dsn
}
