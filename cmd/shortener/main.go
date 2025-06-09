package main

import (
	"github.com/MaxRadzey/shortener/internal/app"
	"github.com/MaxRadzey/shortener/internal/config"
)

func main() {
	AppConfig := config.New()

	config.ParseEnv(AppConfig)

	ParseFlag(AppConfig)

	if err := app.Run(AppConfig); err != nil {
		panic(err)
	}
}
