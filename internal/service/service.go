package service

import (
	"errors"
	"fmt"

	"github.com/MaxRadzey/shortener/internal/config"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
)

var (
	ErrValidation = errors.New("validation error")
)

type Service struct {
	storage   dbstorage.URLStorage
	appConfig config.Config
}

func NewService(storage dbstorage.URLStorage, appConfig config.Config) *Service {
	return &Service{
		storage:   storage,
		appConfig: appConfig,
	}
}

func (s *Service) CreateShortURL(longURL string) (string, error) {
	if !utils.IsValidURL(longURL) {
		return "", ErrValidation
	}

	shortPath, err := utils.GetShortPath(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate short path: %w", err)
	}

	err = s.storage.Create(shortPath, longURL)
	if err != nil {
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	result := fmt.Sprintf("%s/%s", s.appConfig.ReturningAddress, shortPath)
	return result, nil
}

func (s *Service) GetLongURL(shortPath string) (string, error) {
	longURL, err := s.storage.Get(shortPath)
	if err != nil {
		return "", err
	}

	return longURL, nil
}
