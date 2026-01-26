package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaxRadzey/shortener/internal/auth"
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetURL(t *testing.T) {
	storage := newFakeStorage(map[string]string{
		"XxLlqM": "https://vk.com",
	})

	handler := setupTestHandler(storage)
	router := setupTestRouter(handler)

	type want struct {
		code     int
		Location string
	}

	tests := []struct {
		name    string
		method  string
		request string
		want    want
	}{
		{
			name:    "Test #1 send valid request",
			method:  http.MethodGet,
			request: "XxLlqM",
			want: want{
				code:     http.StatusTemporaryRedirect,
				Location: "https://vk.com",
			},
		},
		{
			name:    "Test #2 invalid method",
			method:  http.MethodPost,
			request: "XxLlqM",
			want: want{
				code:     http.StatusMethodNotAllowed,
				Location: "",
			},
		},
		{
			name:    "Test #3 send not found data",
			method:  http.MethodGet,
			request: "FFF113",
			want: want{
				code:     http.StatusNotFound,
				Location: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "/"+test.request, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")

			if test.want.Location != "" {
				assert.Equal(t, test.want.Location, w.Header().Get("Location"), "Заголовок Location не совпадает с ожидаемым")
			}
		})
	}
}

func TestCreateURL(t *testing.T) {
	storage := newFakeStorage(nil)
	handler := setupTestHandler(storage)
	router := setupTestRouter(handler)

	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name        string
		method      string
		body        string
		contentType string
		want        want
	}{
		{
			name:        "Test #1 send valid request",
			method:      http.MethodPost,
			body:        "https://vk.com",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:     http.StatusCreated,
				response: AppConfig.ReturningAddress + "/XxLlqM",
			},
		},
		{
			name:        "Test #2 invalid method",
			method:      http.MethodGet,
			body:        "https://vk.com",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: "Method not allowed!",
			},
		},
		{
			name:        "Test #3 send invalid body",
			method:      http.MethodPost,
			body:        "123",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:     http.StatusBadRequest,
				response: "Invalid Body!",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestBody := strings.NewReader(test.body)

			r := httptest.NewRequest(test.method, "/", requestBody)
			r.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			require.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			require.Equal(t, test.want.response, w.Body.String(), "Body не совпадает с ожидаемым")
		})
	}
}

func TestGetURLJSON(t *testing.T) {
	storage := newFakeStorage(nil)
	handler := setupTestHandler(storage)
	router := setupTestRouter(handler)

	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name        string
		method      string
		request     interface{}
		contentType string
		want        want
	}{
		{
			name:        "Test #1 invalid method",
			method:      http.MethodGet,
			request:     models.Request{URL: "vk.com"},
			contentType: "application/json",
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: "Method not allowed!",
			},
		},
		{
			name:        "Test #2 Ok",
			method:      http.MethodPost,
			request:     models.Request{URL: "https://vk.com"},
			contentType: "application/json",
			want: want{
				code:     http.StatusCreated,
				response: `{"result":"http://localhost:8080/XxLlqM"}`,
			},
		},
		{
			name:        "Test #3 invalid body",
			method:      http.MethodPost,
			request:     "vk.com",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:     http.StatusBadRequest,
				response: "invalid request",
			},
		},
		{
			name:        "Test #4 empty request body",
			method:      http.MethodPost,
			request:     models.Request{URL: ""},
			contentType: "application/json",
			want: want{
				code:     http.StatusBadRequest,
				response: "invalid request",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var requestBody *bytes.Reader
			switch v := test.request.(type) {
			case models.Request:
				b, _ := json.Marshal(v)
				requestBody = bytes.NewReader(b)
			case string:
				requestBody = bytes.NewReader([]byte(v))
			}

			r := httptest.NewRequest(test.method, "/api/shorten", requestBody)
			r.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			require.Equal(t, strings.TrimSpace(test.want.response), strings.TrimSpace(w.Body.String()), "Body не совпадает с ожидаемым")
			require.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestCreateURLBatch(t *testing.T) {
	storage := newFakeStorage(nil)
	handler := setupTestHandler(storage)
	router := setupTestRouter(handler)

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name        string
		method      string
		request     interface{}
		contentType string
		want        want
	}{
		{
			name:   "Test #1 valid batch request",
			method: http.MethodPost,
			request: []models.BatchRequestItem{
				{CorrelationID: "1", OriginalURL: "https://vk.com"},
				{CorrelationID: "2", OriginalURL: "https://ya.ru"},
			},
			contentType: "application/json",
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
				response:    `[{"correlation_id":"1","short_url":"http://localhost:8080/XxLlqM"},{"correlation_id":"2","short_url":"http://localhost:8080/` + getShortPathForURL("https://ya.ru") + `"}]`,
			},
		},
		{
			name:   "Test #2 invalid method",
			method: http.MethodGet,
			request: []models.BatchRequestItem{
				{CorrelationID: "1", OriginalURL: "https://vk.com"},
			},
			contentType: "application/json",
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				response:    "Method not allowed!",
			},
		},
		{
			name:        "Test #3 invalid JSON body",
			method:      http.MethodPost,
			request:     "invalid json",
			contentType: "application/json",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				response:    "invalid request",
			},
		},
		{
			name:        "Test #4 empty batch",
			method:      http.MethodPost,
			request:     []models.BatchRequestItem{},
			contentType: "application/json",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				response:    "invalid request",
			},
		},
		{
			name:   "Test #5 invalid URL in batch",
			method: http.MethodPost,
			request: []models.BatchRequestItem{
				{CorrelationID: "1", OriginalURL: "invalid-url"},
			},
			contentType: "application/json",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				response:    "invalid request",
			},
		},
		{
			name:   "Test #6 single item batch",
			method: http.MethodPost,
			request: []models.BatchRequestItem{
				{CorrelationID: "single", OriginalURL: "https://vk.com"},
			},
			contentType: "application/json",
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
				response:    `[{"correlation_id":"single","short_url":"http://localhost:8080/XxLlqM"}]`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var requestBody *bytes.Reader
			switch v := test.request.(type) {
			case []models.BatchRequestItem:
				b, _ := json.Marshal(v)
				requestBody = bytes.NewReader(b)
			case string:
				requestBody = bytes.NewReader([]byte(v))
			default:
				requestBody = bytes.NewReader([]byte{})
			}

			r := httptest.NewRequest(test.method, "/api/shorten/batch", requestBody)
			r.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			require.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")

			if test.want.contentType != "" {
				contentType := w.Header().Get("Content-Type")
				assert.Contains(t, contentType, test.want.contentType, "Content-Type не совпадает с ожидаемым")
			}

			if test.want.response != "" {
				actualResponse := strings.TrimSpace(w.Body.String())
				if test.want.code == http.StatusCreated {
					var actualItems []models.BatchResponseItem
					err := json.Unmarshal([]byte(actualResponse), &actualItems)
					require.NoError(t, err, "Ответ должен быть валидным JSON")

					var expectedItems []models.BatchResponseItem
					err = json.Unmarshal([]byte(test.want.response), &expectedItems)
					require.NoError(t, err, "Ожидаемый ответ должен быть валидным JSON")

					assert.Equal(t, len(expectedItems), len(actualItems), "Количество элементов не совпадает")
					for i, expected := range expectedItems {
						assert.Equal(t, expected.CorrelationID, actualItems[i].CorrelationID, "CorrelationID не совпадает")
						assert.Equal(t, expected.ShortURL, actualItems[i].ShortURL, "ShortURL не совпадает")
					}
				} else {
					assert.Equal(t, test.want.response, actualResponse, "Body не совпадает с ожидаемым")
				}
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	t.Run("no cookie returns 401", func(t *testing.T) {
		storage := newFakeStorage(nil)
		handler := setupTestHandler(storage)
		router := setupTestRouter(handler)

		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Empty(t, w.Header().Get("Set-Cookie"), "на /api/user/urls куку не выдаём")
	})

	t.Run("invalid cookie returns 401", func(t *testing.T) {
		storage := newFakeStorage(nil)
		handler := setupTestHandler(storage)
		router := setupTestRouter(handler)

		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "invalid.garbage"})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid cookie no URLs returns 204", func(t *testing.T) {
		storage := newFakeStorage(nil)
		handler := setupTestHandler(storage)
		router := setupTestRouter(handler)

		userID := uuid.New().String()
		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		r.AddCookie(auth.NewCookie(userID, AppConfig.SigningKey))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		require.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("valid cookie with URLs returns 200 and JSON", func(t *testing.T) {
		storage := newFakeStorage(nil)
		handler := setupTestHandler(storage)
		router := setupTestRouter(handler)

		userID := uuid.New().String()
		cookie := auth.NewCookie(userID, AppConfig.SigningKey)

		// Shorten via POST /api/shorten with same user cookie
		body := bytes.NewReader([]byte(`{"url":"https://example.com"}`))
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Get user URLs
		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		r.AddCookie(cookie)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, r)

		require.Equal(t, http.StatusOK, w2.Code)
		var items []models.UserURLItem
		err := json.NewDecoder(w2.Body).Decode(&items)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "https://example.com", items[0].OriginalURL)
		assert.Contains(t, items[0].ShortURL, AppConfig.ReturningAddress)
	})
}
