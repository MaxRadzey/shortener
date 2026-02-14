package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

const remoteTimeout = 10 * time.Second

// RemoteReceiver отправляет события аудита на удалённый сервер методом POST.
type RemoteReceiver struct {
	url    string
	client *http.Client
}

// NewRemoteReceiver создаёт приёмник отправки на удалённый URL.
func NewRemoteReceiver(url string) *RemoteReceiver {
	return &RemoteReceiver{
		url: url,
		client: &http.Client{
			Timeout: remoteTimeout,
		},
	}
}

// Notify отправляет событие POST на сконфигурированный URL.
func (r *RemoteReceiver) Notify(event Event) {
	go func() {
		data, err := json.Marshal(event)
		if err != nil {
			logger.Log.Error("audit: failed to marshal event for remote", zap.Error(err))
			return
		}

		req, err := http.NewRequest(http.MethodPost, r.url, bytes.NewReader(data))
		if err != nil {
			logger.Log.Error("audit: failed to create request", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := r.client.Do(req)
		if err != nil {
			logger.Log.Error("audit: failed to send event to remote", zap.String("url", r.url), zap.Error(err))
			return
		}
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			logger.Log.Warn("audit: remote returned non-success status", zap.Int("status", resp.StatusCode), zap.String("url", r.url))
		}
	}()
}
