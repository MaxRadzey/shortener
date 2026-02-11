package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// fileRecord — формат одной записи в файле (url, user_id, is_deleted).
type fileRecord struct {
	FullURL   string `json:"url"`
	UserID    string `json:"user_id"`
	IsDeleted bool   `json:"is_deleted"`
}

type Storage struct {
	mu       sync.RWMutex
	filePath string
}

func NewStorage(filePath string) (*Storage, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	return &Storage{
		filePath: filePath,
	}, nil
}

func (s *Storage) readFromFile() (map[string]URLEntry, error) {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	res := make(map[string]URLEntry)

	if ok := scanner.Scan(); !ok {
		return res, nil
	}

	line := scanner.Bytes()

	// Новый формат: map[string]fileRecord
	var byShortNew map[string]fileRecord
	if err := json.Unmarshal(line, &byShortNew); err == nil {
		for k, v := range byShortNew {
			res[k] = URLEntry{ShortPath: k, FullURL: v.FullURL, UserID: v.UserID, IsDeleted: v.IsDeleted}
		}
		return res, nil
	}

	// Старый формат (совместимость): map[string]string — без user_id и is_deleted
	var byShort map[string]string
	if err := json.Unmarshal(line, &byShort); err != nil {
		return nil, err
	}
	for k, v := range byShort {
		res[k] = URLEntry{ShortPath: k, FullURL: v, UserID: "", IsDeleted: false}
	}
	return res, nil
}

func (s *Storage) Create(item URLEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFromFile()
	if err != nil {
		return fmt.Errorf("read from file error: %w", err)
	}

	data[item.ShortPath] = item
	return s.writeToFile(data)
}

func (s *Storage) Get(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.readFromFile()
	if err != nil {
		return "", fmt.Errorf("read from file error: %w", err)
	}

	r, ok := data[id]
	if !ok {
		return "", &ErrNotFound{ShortPath: id}
	}
	if r.IsDeleted {
		return "", &ErrGone{ShortPath: id}
	}
	return r.FullURL, nil
}

func (s *Storage) CreateBatch(ctx context.Context, items []URLEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFromFile()
	if err != nil {
		return fmt.Errorf("read from file error: %w", err)
	}

	for _, item := range items {
		data[item.ShortPath] = item
	}

	return s.writeToFile(data)
}

func (s *Storage) writeToFile(data map[string]URLEntry) error {
	byShort := make(map[string]fileRecord)
	for k, r := range data {
		byShort[k] = fileRecord{FullURL: r.FullURL, UserID: r.UserID, IsDeleted: r.IsDeleted}
	}
	raw, err := json.Marshal(byShort)
	if err != nil {
		return fmt.Errorf("serialize url error: %w", err)
	}
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err = file.WriteString(string(raw)); err != nil {
		return fmt.Errorf("write url to file error: %w", err)
	}
	return nil
}

func (s *Storage) GetByUserID(ctx context.Context, userID string) ([]UserURL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.readFromFile()
	if err != nil {
		return nil, fmt.Errorf("read from file error: %w", err)
	}

	var out []UserURL
	for short, r := range data {
		if r.UserID == userID {
			out = append(out, UserURL{ShortPath: short, OriginalURL: r.FullURL})
		}
	}
	return out, nil
}

// DeleteBatch проставляет флаг is_deleted у записей по списку short_path для указанного пользователя.
func (s *Storage) DeleteBatch(ctx context.Context, userID string, shortPaths []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFromFile()
	if err != nil {
		return fmt.Errorf("read from file error: %w", err)
	}

	for _, shortPath := range shortPaths {
		if entry, exists := data[shortPath]; exists && entry.UserID == userID {
			entry.IsDeleted = true
			data[shortPath] = entry
		}
	}

	return s.writeToFile(data)
}

func (s *Storage) Ping(ctx context.Context) error {
	// Проверяем доступность файла для записи
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("file storage not available: %w", err)
	}
	_ = file.Close()
	return nil
}
