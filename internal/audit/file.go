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
	file *os.File
	mu   sync.Mutex
}

// NewFileReceiver возвращает приёмник для записи в указанный файл.
// Файл открывается сразу и остаётся открытым на всё время жизни программы.
func NewFileReceiver(path string) *FileReceiver {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Log.Error("audit: failed to open audit file", zap.String("path", path), zap.Error(err))
		return &FileReceiver{}
	}
	return &FileReceiver{file: file}
}

// Notify дописывает JSON события в конец файла.
func (f *FileReceiver) Notify(event Event) {
	if f.file == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("audit: failed to marshal event", zap.Error(err))
		return
	}

	f.mu.Lock()
	_, err = f.file.Write(append(data, '\n'))
	f.mu.Unlock()

	if err != nil {
		logger.Log.Error("audit: failed to write audit event", zap.Error(err))
	}
}
