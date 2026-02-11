package config

import (
	"flag"
	"os"
)

type Config struct {
	Address          string
	ReturningAddress string
	LogLevel         string
	FilePath         string
	DatabaseDSN      string
	// SigningKey — секрет для HMAC-подписи куки (не шифрование). Кука = base64(userID).base64(hmac).
	// В проде обязательно задавать через SECRET_KEY; дефолт — только для локальной разработки.
	SigningKey string
}

func New() *Config {
	return &Config{
		Address:          "localhost:8080",
		ReturningAddress: "http://localhost:8080",
		LogLevel:         "info",
		FilePath:         "data.json",
		DatabaseDSN:      "postgres://shortener:shortener@localhost:5432/shortener",
		SigningKey:       "dev-signing-key-change-in-production",
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
	if v := os.Getenv("SECRET_KEY"); v != "" {
		config.SigningKey = v
	}
}

// ParseFlags парсит флаги командной строки и обновляет конфигурацию.
// Флаги имеют приоритет над переменными окружения.
func ParseFlags(config *Config) {
	flag.StringVar(&config.Address, "a", config.Address, "address and port to run server")
	flag.StringVar(&config.ReturningAddress, "b", config.ReturningAddress, "address to return URL")
	flag.StringVar(&config.LogLevel, "l", config.LogLevel, "log level")
	flag.StringVar(&config.FilePath, "f", config.FilePath, "file path")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database connection string")

	flag.Parse()
}
