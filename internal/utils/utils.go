// Package utils — хелперы: короткий хеш URL, валидация URL, маскировка DSN для логов.
package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"hash"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/MaxRadzey/shortener/internal/pool"
)

var (
	hasherPool    = pool.New(func() hash.Hash { return sha1.New() })
	base64BufPool = sync.Pool{
		New: func() interface{} { return make([]byte, base64.URLEncoding.EncodedLen(sha1.Size)) },
	}
)

// GetShortPath по строке (URL) возвращает короткий идентификатор (sha1+base64, 6 символов).
func GetShortPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("empty string cannot be shortened")
	}

	h := hasherPool.Get()
	defer hasherPool.Put(h)
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

// IsIPInTrustedSubnet проверяет, входит ли clientIP в подсеть cidr (CIDR).
// При пустом cidr возвращает false (доступ запрещён). При невалидном IP или CIDR — false.
func IsIPInTrustedSubnet(clientIP, cidr string) bool {
	if cidr == "" {
		return false
	}
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
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
