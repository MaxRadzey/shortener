// Package main — точка входа сервиса сокращения URL: парсит конфиг и запускает HTTP-сервер.
// При старте выводит в stdout информацию о сборке (версия, дата, коммит), если она задана через -ldflags.
package main

import (
	"fmt"
	"os"

	"github.com/MaxRadzey/shortener/internal/app"
	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

// Переменные сборки, подставляемые при сборке через -ldflags (иначе остаются пустыми, при выводе — "N/A").
var (
	buildVersion string // версия приложения (например, тег или номер релиза)
	buildDate    string // дата и время сборки в UTC
	buildCommit  string // хеш коммита Git
)

// orNA возвращает s, если он не пустой, иначе "N/A" (для вывода полей сборки).
func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// printBuildInfo выводит в stdout информацию о сборке в формате:
// Build version: <buildVersion>
// Build date: <buildDate>
// Build commit: <buildCommit>
// Пустые значения выводятся как "N/A".
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", orNA(buildVersion))
	fmt.Printf("Build date: %s\n", orNA(buildDate))
	fmt.Printf("Build commit: %s\n", orNA(buildCommit))
}

// @title        Shortener API
// @version      1.0
// @description  API для сокращения URL и управления короткими ссылками.
// @host         localhost:8080
// @BasePath     /
func main() {
	printBuildInfo()

	AppConfig := config.New()

	if path := config.GetConfigFilePath(); path != "" {
		if err := config.ParseConfigFile(AppConfig, path); err != nil {
			logger.Log.Fatal("failed to load config file", zap.String("path", path), zap.Error(err))
		}
	}
	config.ParseEnv(AppConfig)
	config.ParseFlags(AppConfig)

	if err := app.Run(AppConfig); err != nil {
		logger.Log.Error("failed to run app", zap.Error(err))
		os.Exit(1)
	}
	logger.Log.Info("running server", zap.String("address", AppConfig.Address))
}
