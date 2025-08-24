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
