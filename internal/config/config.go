package config

import "os"

type Config struct {
	Address          string
	ReturningAddress string
	LogLevel         string
	FilePath         string
	DatabaseDSN      string
}

func New() *Config {
	return &Config{
		Address:          "localhost:8080",
		ReturningAddress: "http://localhost:8080",
		LogLevel:         "info",
		FilePath:         "data.json",
	}
}

func ParseEnv(config *Config) {
	if Address := os.Getenv("SERVER_ADDRESS"); Address != "" {
		config.Address = Address
	}
	if ReturningAddress := os.Getenv("BASE_URL"); ReturningAddress != "" {
		config.ReturningAddress = ReturningAddress
	}
	if LogLevel := os.Getenv("LOG_LEVEL"); LogLevel != "" {
		config.LogLevel = LogLevel
	}
	if FilePath := os.Getenv("FILE_PATH"); FilePath != "" {
		config.FilePath = FilePath
	}
	if DatabaseDSN := os.Getenv("DATABASE_DSN"); DatabaseDSN != "" {
		config.DatabaseDSN = DatabaseDSN
	}
}
