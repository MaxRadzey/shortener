package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const CookieName = "user_id"

// GetOrCreateUser: если куки нет — создаёт пользователя, ставит куку, возвращает (userID, false).
// Если кука есть, но невалидна — возвращает ("", true). Если валидна — (userID, false).
func GetOrCreateUser(req *http.Request, w http.ResponseWriter, secret string) (userID string, hadInvalidCookie bool) {
	cookie, err := req.Cookie(CookieName)
	if err != nil {
		userID = generateUserID()
		setAuthCookie(w, userID, secret)
		return userID, false
	}
	userID, err = validateCookie(cookie, secret)
	if err != nil {
		return "", true
	}
	return userID, false
}

// CreateNewUser создаёт нового пользователя, выставляет куку, возвращает userID.
func CreateNewUser(w http.ResponseWriter, secret string) string {
	userID := generateUserID()
	setAuthCookie(w, userID, secret)
	return userID
}

// NewCookie возвращает подписанную куку для userID. Нужна в тестах.
// userID должен быть валидным UUID.
func NewCookie(userID, secret string) *http.Cookie {
	signature := signData(userID, secret)
	value := userID + "." + signature
	return &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func generateUserID() string {
	return uuid.New().String()
}

func signData(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func setAuthCookie(w http.ResponseWriter, userID, secret string) {
	signature := signData(userID, secret)
	value := userID + "." + signature
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// ValidateCookie проверяет валидность куки и возвращает userID, если кука валидна.
// Возвращает ошибку, если кука невалидна или отсутствует.
func ValidateCookie(cookie *http.Cookie, secret string) (string, error) {
	if cookie == nil || cookie.Value == "" {
		return "", errors.New("empty cookie value")
	}
	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid cookie format")
	}
	userID := parts[0]
	signature := parts[1]
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}
	expected := signData(userID, secret)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return "", errors.New("invalid signature")
	}
	return userID, nil
}

func validateCookie(cookie *http.Cookie, secret string) (string, error) {
	return ValidateCookie(cookie, secret)
}
