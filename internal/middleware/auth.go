package middleware

import (
	"net/http"

	"github.com/MaxRadzey/shortener/internal/auth"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/gin-gonic/gin"
)

// Auth — middleware для аутентификации пользователя: валидирует куку или создаёт нового пользователя и записывает куку.
// Всегда устанавливает user_id в контекст запроса.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID string

		cookie, err := c.Request.Cookie(auth.CookieName)
		if err != nil {
			// Куки нет, создаем нового пользователя
			userID = auth.CreateNewUser(c.Writer, secret)
		} else {
			// Проверяем подпись куки
			userID, err = auth.ValidateCookie(cookie, secret)
			if err != nil {
				// Кука невалидна, создаем нового пользователя
				userID = auth.CreateNewUser(c.Writer, secret)
			}
		}

		// Добавляем userID в контекст
		c.Set(contextkeys.UserIDKey, userID)
		c.Next()
	}
}

// RequireAuth — middleware, который требует валидную куку.
// Если куки нет или она невалидна, возвращает 401 и прерывает выполнение.
// Если кука валидна, устанавливает user_id в контекст запроса.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(auth.CookieName)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateCookie(cookie, secret)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст
		c.Set(contextkeys.UserIDKey, userID)
		c.Next()
	}
}
