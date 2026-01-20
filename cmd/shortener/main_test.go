package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaxRadzey/shortener/internal/models"

	"github.com/MaxRadzey/shortener/internal/config"

	"github.com/MaxRadzey/shortener/internal/app"
	"github.com/gin-gonic/gin"

	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/service"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
)

var AppConfig = config.New()

type FakeStorage struct {
	data map[string]string
}

func (f *FakeStorage) Get(short string) (string, error) {
	val, ok := f.data[short]
	if !ok {
		return "", dbstorage.ErrNotFound
	}
	return val, nil
}

func (f *FakeStorage) Create(short, full string) error {
	f.data[short] = full
	return nil
}

func (f *FakeStorage) CreateBatch(ctx context.Context, items []dbstorage.BatchItem) error {
	for _, item := range items {
		f.data[item.ShortPath] = item.FullURL
	}
	return nil
}

func TestGetShortPath(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{
			name:    "Test get short value of URL",
			value:   "https://vk.com",
			want:    "XxLlqM",
			wantErr: false,
		},
		{
			name:    "Test get short value of URL",
			value:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := utils.GetShortPath(test.value)
			if !test.wantErr {
				require.NoError(t, err)
				assert.Equal(t, test.want, value)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		want     bool
	}{
		{
			name:     "Valid URL",
			inputURL: "https://vk.ccom",
			want:     true,
		},
		{
			name:     "Invalid URL",
			inputURL: "123",
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := utils.IsValidURL(test.inputURL)
			assert.Equal(t, result, test.want)
		})
	}
}

func TestGetURL(t *testing.T) {
	storage := &FakeStorage{
		data: map[string]string{
			"XxLlqM": "https://vk.com",
		},
	}

	urlService := service.NewService(storage, *AppConfig, nil)
	handler := &httphandlers.Handler{Service: urlService}

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
			gin.SetMode(gin.TestMode)

			router := app.SetupRouter(handler)

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
	storage := &FakeStorage{
		data: map[string]string{},
	}

	urlService := service.NewService(storage, *AppConfig, nil)
	handler := &httphandlers.Handler{Service: urlService}

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
			gin.SetMode(gin.TestMode)

			router := app.SetupRouter(handler)

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
	storage := &FakeStorage{
		data: map[string]string{},
	}

	type want struct {
		code     int
		response string
	}

	urlService := service.NewService(storage, *AppConfig, nil)
	handler := &httphandlers.Handler{Service: urlService}

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
			gin.SetMode(gin.TestMode)

			router := app.SetupRouter(handler)

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
