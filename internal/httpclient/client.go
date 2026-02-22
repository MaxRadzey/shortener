// Package httpclient предоставляет HTTP-клиент с автоматическими ретраями при ошибках и 5xx.
package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

// Параметры ретраев для RetryableClient.
const (
	MaxRetries        = 3               // максимальное число попыток (включая первую)
	InitialRetryDelay = 1 * time.Second // базовая задержка перед повтором; далее exponential backoff
)

// RetryableClient оборачивает http.Client и повторяет запрос при сетевой ошибке или ответе 5xx.
type RetryableClient struct {
	client *http.Client
}

// NewRetryableClient создаёт клиент с поддержкой ретраев; передаётся уже настроенный *http.Client.
func NewRetryableClient(client *http.Client) *RetryableClient {
	return &RetryableClient{client: client}
}

// shouldRetry возвращает true, если запрос стоит повторить (ошибка или статус 5xx).
func shouldRetry(err error, statusCode int) bool {
	if err != nil {
		return true
	}
	return statusCode >= 500 && statusCode < 600
}

// Do выполняет запрос через обёрнутый http.Client; при ошибке или 5xx повторяет до MaxRetries раз с exponential backoff.
// Тело запроса при ретраях переиспользуется (читается в память до цикла).
func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}

	var resp *http.Response
	var err error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(InitialRetryDelay * time.Duration(1<<uint(attempt-1)))
			logger.Log.Debug("httpclient: retry", zap.Int("attempt", attempt+1), zap.String("url", req.URL.String()))
		}
		if len(body) > 0 {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		resp, err = c.client.Do(req)
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		if !shouldRetry(err, code) {
			return resp, err
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if attempt < MaxRetries {
			logger.Log.Warn("httpclient: request failed, will retry", zap.Error(err), zap.Int("status", code), zap.String("url", req.URL.String()))
		}
	}
	return resp, err
}
