package audit

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

// FileReceiver пишет события в файл по одной строке (append).
type FileReceiver struct {
	path string
	mu   sync.Mutex
}

// NewFileReceiver возвращает приёмник для записи в указанный файл.
func NewFileReceiver(path string) *FileReceiver {
	return &FileReceiver{path: path}
}

// Notify дописывает JSON события в конец файла.
func (f *FileReceiver) Notify(event Event) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("audit: failed to marshal event", zap.Error(err))
		return
	}

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Log.Error("audit: failed to open audit file", zap.String("path", f.path), zap.Error(err))
		return
	}
	defer file.Close()

	if _, err := file.Write(append(data, '\n')); err != nil {
		logger.Log.Error("audit: failed to write audit event", zap.Error(err))
	}
}
