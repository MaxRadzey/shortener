package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"hash"
	"net/url"
	"strings"
	"sync"
)

var (
	hasherPool = sync.Pool{
		New: func() interface{} { return sha1.New() },
	}
	base64BufPool = sync.Pool{
		New: func() interface{} { return make([]byte, base64.URLEncoding.EncodedLen(sha1.Size)) },
	}
)

// GetShortPath по строке (URL) возвращает короткий идентификатор (sha1+base64, 6 символов).
func GetShortPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("empty string cannot be shortened")
	}

	h := hasherPool.Get().(hash.Hash)
	defer hasherPool.Put(h)
	h.Reset()
	h.Write([]byte(path))
	hash := h.Sum(nil)

	buf := base64BufPool.Get().([]byte)
	defer base64BufPool.Put(buf)
	base64.URLEncoding.Encode(buf, hash)
	return string(buf[:6]), nil
}

// IsValidURL проверяет, что строка — валидный URL.
func IsValidURL(urlToCheck string) bool {
	_, err := url.ParseRequestURI(urlToCheck)
	return err == nil
}

// MaskDSN возвращает DSN с замаскированным паролем для логов.
// postgres://user:password@host:port/db -> postgres://user:***@host:port/db
func MaskDSN(dsn string) string {
	if idx := strings.Index(dsn, "@"); idx > 0 {
		if passIdx := strings.LastIndex(dsn[:idx], ":"); passIdx > 0 {
			return dsn[:passIdx+1] + "***" + dsn[idx:]
		}
	}
	return dsn
}
