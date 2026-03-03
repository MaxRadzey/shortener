// Package middleware содержит HTTP-only middleware: аутентификация по куке, gzip.
package middleware

import (
	"github.com/MaxRadzey/shortener/internal/auth"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Auth выставляет в контексте user_id из куки или создаёт нового пользователя и ставит куку.
func Auth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, hadInvalidCookie := auth.GetOrCreateUser(c.Request, c.Writer, secretKey)
		if hadInvalidCookie {
			userID = auth.CreateNewUser(c.Writer, secretKey)
			logger.Log.Debug("Invalid cookie, created new user ID", zap.String("user_id", userID))
		} else if userID != "" {
			logger.Log.Debug("Authenticated user ID", zap.String("user_id", userID))
		}
		c.Set(contextkeys.UserIDKey, userID)
		c.Next()
	}
}
