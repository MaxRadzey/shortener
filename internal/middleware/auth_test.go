package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/service"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	teststorage "github.com/MaxRadzey/shortener/internal/testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	secretKey := "test-secret-key"
	cfg := &config.Config{
		SigningKey:       secretKey,
		ReturningAddress: "http://localhost:8080",
	}

	gin.SetMode(gin.TestMode)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	storage := teststorage.NewFakeStorageWithEntries(map[string]dbstorage.URLEntry{
		"XxLlqM": {ShortPath: "XxLlqM", FullURL: "https://vk.com", UserID: userID},
	})

	// Создаем handler
	urlService := service.NewService(storage, *cfg)
	handler := &httphandlers.Handler{Service: urlService}

	tests := []struct {
		name           string
		cookieValue    string
		cookiePresent  bool
		expectedStatus int
		expectCookie   bool
		description    string
	}{
		{
			name:           "Test #1 no cookie - should create new user and return 204",
			cookiePresent:  false,
			expectedStatus: http.StatusNoContent,
			expectCookie:   true,
			description:    "При отсутствии куки миделварь должен создать нового пользователя и установить куку",
		},
		{
			name:           "Test #2 valid cookie - should authenticate and return 200",
			cookiePresent:  true,
			cookieValue:    createValidCookie(userID, secretKey),
			expectedStatus: http.StatusOK,
			expectCookie:   false,
			description:    "При валидной куке миделварь должен аутентифицировать пользователя",
		},
		{
			name:           "Test #3 invalid cookie format - should create new user",
			cookiePresent:  true,
			cookieValue:    "invalid-format",
			expectedStatus: http.StatusNoContent,
			expectCookie:   true,
			description:    "При невалидном формате куки миделварь должен создать нового пользователя",
		},
		{
			name:           "Test #4 invalid signature - should create new user",
			cookiePresent:  true,
			cookieValue:    userID + ".invalid-signature",
			expectedStatus: http.StatusNoContent,
			expectCookie:   true,
			description:    "При невалидной подписи куки миделварь должен создать нового пользователя",
		},
		{
			name:           "Test #5 empty cookie value - should create new user",
			cookiePresent:  true,
			cookieValue:    "",
			expectedStatus: http.StatusNoContent,
			expectCookie:   true,
			description:    "При пустом значении куки миделварь должен создать нового пользователя",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Создаем роутер с миделварью Auth
			r := gin.New()
			r.Use(Auth(cfg.SigningKey))
			r.GET("/api/user/urls", handler.GetUserURLs)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			if test.cookiePresent {
				req.AddCookie(&http.Cookie{
					Name:  cookieName,
					Value: test.cookieValue,
				})
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Проверяем код ответа - это главная проверка работы миделваря
			require.Equal(t, test.expectedStatus, w.Code, test.description+": Код ответа не совпадает с ожидаемым")

			// Проверяем наличие куки в ответе
			res := w.Result()
			defer res.Body.Close()
			cookies := res.Cookies()
			if test.expectCookie {
				assert.Greater(t, len(cookies), 0, "Ожидалась установка куки")
				// Проверяем, что кука валидна
				if len(cookies) > 0 {
					cookie := cookies[0]
					assert.Equal(t, cookieName, cookie.Name, "Имя куки не совпадает")
					assert.NotEmpty(t, cookie.Value, "Значение куки не должно быть пустым")
					// Проверяем, что кука валидна
					_, err := validateCookie(cookie, secretKey)
					assert.NoError(t, err, "Кука должна быть валидной")
				}
			}
		})
	}
}

// createValidCookie создает валидную куку для тестов
func createValidCookie(userID, secretKey string) string {
	signature := signData(userID, secretKey)
	return userID + "." + signature
}
