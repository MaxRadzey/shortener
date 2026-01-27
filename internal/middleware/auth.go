package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const cookieName = "user_id"

// Auth — middleware для аутентификации пользователя.
func Auth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID string

		cookie, err := c.Request.Cookie(cookieName)
		if err != nil {
			// Куки нет, создаем нового пользователя
			userID = generateUserID()
			setAuthCookie(c.Writer, userID, secretKey)
			logger.Log.Debug("Created new user ID", zap.String("user_id", userID))
		} else {
			// Проверяем подпись куки
			userID, err = validateCookie(cookie, secretKey)
			if err != nil {
				// Кука невалидна, создаем нового пользователя
				userID = generateUserID()
				setAuthCookie(c.Writer, userID, secretKey)
				logger.Log.Debug("Invalid cookie, created new user ID", zap.String("user_id", userID))
			} else {
				logger.Log.Debug("Authenticated user ID", zap.String("user_id", userID))
			}
		}

		// Добавляем userID в контекст Gin
		c.Set(contextkeys.UserIDKey, userID)
		c.Next()
	}
}


// generateUserID создает новый UUID для пользователя
func generateUserID() string {
	return uuid.New().String()
}

// signData создает подпись для данных
func signData(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// setAuthCookie устанавливает куку
func setAuthCookie(w http.ResponseWriter, userID, secretKey string) {
	signature := signData(userID, secretKey)
	value := userID + "." + signature

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour), // 1 год
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

// validateCookie проверяет валидность куки
func validateCookie(cookie *http.Cookie, secretKey string) (string, error) {
	if cookie.Value == "" {
		return "", errors.New("empty cookie value")
	}

	// Разделяем userID и подпись
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return "", errors.New("invalid cookie format")
	}

	userID := parts[0]
	signature := parts[1]

	// Проверяем что userID - валидный UUID
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Проверяем подпись
	expectedSignature := signData(userID, secretKey)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", errors.New("invalid signature")
	}

	return userID, nil
}
