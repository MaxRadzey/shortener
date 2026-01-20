package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var ErrNotFound = errors.New("url not found")

type BatchItem struct {
	ShortPath string
	FullURL   string
}

type URLStorage interface {
	Get(short string) (string, error)
	Create(short, full string) error
	CreateBatch(ctx context.Context, items []BatchItem) error
}

type Storage struct {
	mu       sync.RWMutex
	data     map[string]string
	filePath string
}

func NewStorage(filePath string) (*Storage, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	data, err := readLines(filePath)
	if err != nil {
		return nil, fmt.Errorf("read urls from file error: %w", err)
	}

	return &Storage{
		data:     data,
		filePath: filePath,
	}, nil
}

func readLines(filePath string) (map[string]string, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	res := make(map[string]string)

	if ok := scanner.Scan(); !ok {
		return res, nil
	}

	err = json.Unmarshal([]byte(scanner.Text()), &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Storage) Create(id, url string) error {
	s.mu.Lock()
	s.data[id] = url
	dataCopy := make(map[string]string)
	for k, v := range s.data {
		dataCopy[k] = v
	}
	s.mu.Unlock()

	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	data, err := json.Marshal(dataCopy)
	if err != nil {
		return fmt.Errorf("serialize url error: %w", err)
	}

	_, err = file.WriteString(string(data))
	if err != nil {
		return fmt.Errorf("write url to file error: %w", err)
	}

	return nil
}

func (s *Storage) Get(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.data[id]
	if !ok {
		return "", ErrNotFound
	}

	return url, nil
}

func (s *Storage) CreateBatch(ctx context.Context, items []BatchItem) error {
	s.mu.Lock()
	// Обновляем data map атомарно
	for _, item := range items {
		s.data[item.ShortPath] = item.FullURL
	}
	dataCopy := make(map[string]string)
	for k, v := range s.data {
		dataCopy[k] = v
	}
	s.mu.Unlock()

	// Записываем весь файл за одну операцию
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	data, err := json.Marshal(dataCopy)
	if err != nil {
		return fmt.Errorf("serialize url error: %w", err)
	}

	_, err = file.WriteString(string(data))
	if err != nil {
		return fmt.Errorf("write url to file error: %w", err)
	}

	return nil
}
