package middleware

import (
	"github.com/MaxRadzey/shortener/internal/auth"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/gin-gonic/gin"
)

// Auth — middleware для аутентификации пользователя: валидирует куку или создаёт нового пользователя и записывает куку.
// Всегда устанавливает user_id в контекст запроса.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, hadInvalid := auth.GetOrCreateUser(c.Request, c.Writer, secret)
		if hadInvalid {
			userID = auth.CreateNewUser(c.Writer, secret)
		}
		c.Set(contextkeys.UserIDKey, userID)
		c.Next()
	}
}
