// Package main — точка входа: парсит конфиг и запускает HTTP-сервер коротких ссылок.
package main

import (
	"fmt"
	"os"

	"github.com/MaxRadzey/shortener/internal/app"
	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/logger"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	version := buildVersion
	date := buildDate
	commit := buildCommit

	if version == "" {
		version = "N/A"
	}
	if date == "" {
		date = "N/A"
	}
	if commit == "" {
		commit = "N/A"
	}

	fmt.Println("Build version:", version)
	fmt.Println("Build date:", date)
	fmt.Println("Build commit:", commit)
}

// @title        Shortener API
// @version      1.0
// @description  API для сокращения URL и управления короткими ссылками.
// @host         localhost:8080
// @BasePath     /
func main() {
	printBuildInfo()

	AppConfig := config.New()

	application, err := app.New(AppConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init app: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		logger.Log.Error("failed to run app", zap.Error(err))
		os.Exit(1)
	}
}
