package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/service"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStorage - мок хранилища для тестов миделваря
type TestStorage struct {
	data map[string]dbstorage.URLEntry
}

func (t *TestStorage) Get(short string) (string, error) {
	val, ok := t.data[short]
	if !ok {
		return "", &dbstorage.ErrNotFound{ShortPath: short}
	}
	return val.FullURL, nil
}

func (t *TestStorage) Create(item dbstorage.URLEntry) error {
	t.data[item.ShortPath] = item
	return nil
}

func (t *TestStorage) CreateBatch(ctx context.Context, items []dbstorage.URLEntry) error {
	for _, item := range items {
		t.data[item.ShortPath] = item
	}
	return nil
}

func (t *TestStorage) GetByUserID(ctx context.Context, userID string) ([]dbstorage.UserURL, error) {
	var out []dbstorage.UserURL
	for short, entry := range t.data {
		if entry.UserID == userID {
			out = append(out, dbstorage.UserURL{ShortPath: short, OriginalURL: entry.FullURL})
		}
	}
	return out, nil
}

func (t *TestStorage) Ping(ctx context.Context) error {
	return nil
}

func TestAuth(t *testing.T) {
	secretKey := "test-secret-key"
	cfg := &config.Config{
		SigningKey:       secretKey,
		ReturningAddress: "http://localhost:8080",
	}

	gin.SetMode(gin.TestMode)

	// Создаем storage с данными для пользователя
	userID := "test-user-id"
	storage := &TestStorage{
		data: map[string]dbstorage.URLEntry{
			"XxLlqM": {ShortPath: "XxLlqM", FullURL: "https://vk.com", UserID: userID},
		},
	}

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
			cookies := w.Result().Cookies()
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
