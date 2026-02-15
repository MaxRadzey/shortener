// Package config загружает настройки приложения из флагов и переменных окружения.
package config

import (
	"flag"
	"os"
)

// Config — настройки приложения (адрес, БД, файл, логи, аудит, ключ подписи).
type Config struct {
	Address          string // адрес сервера (например :8080)
	ReturningAddress string // базовый URL для коротких ссылок в ответах
	LogLevel         string // уровень логов (info, debug и т.д.)
	FilePath         string // путь к файлу хранилища, если не используется БД
	DatabaseDSN      string // строка подключения к PostgreSQL; пусто — БД не используется
	// SigningKey — секрет для подписи куки (в проде задавать через SECRET_KEY).
	SigningKey string
	AuditFile  string // путь к файлу аудита; пусто — выключено
	AuditURL   string // URL приёмника аудита; пусто — выключено
	DevMode    bool   // режим разработки (pprof и т.п.)
}

// New возвращает конфиг с дефолтными значениями.
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

// ParseEnv подставляет в config значения из переменных окружения.
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
	if v := os.Getenv("AUDIT_FILE"); v != "" {
		config.AuditFile = v
	}
	if v := os.Getenv("AUDIT_URL"); v != "" {
		config.AuditURL = v
	}
	if v := os.Getenv("APP_ENV"); v == "dev" || v == "development" {
		config.DevMode = true
	}
}

// ParseFlags парсит флаги (-a, -b, -d, -f, -dev и др.); приоритет над env.
func ParseFlags(config *Config) {
	flag.StringVar(&config.Address, "a", config.Address, "address and port to run server")
	flag.StringVar(&config.ReturningAddress, "b", config.ReturningAddress, "address to return URL")
	flag.StringVar(&config.LogLevel, "l", config.LogLevel, "log level")
	flag.StringVar(&config.FilePath, "f", config.FilePath, "file path")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database connection string")
	flag.StringVar(&config.AuditFile, "audit-file", config.AuditFile, "path to audit log file (empty = disabled)")
	flag.StringVar(&config.AuditURL, "audit-url", config.AuditURL, "URL of remote audit receiver (empty = disabled)")
	flag.BoolVar(&config.DevMode, "dev", config.DevMode, "enable dev mode")

	flag.Parse()
}
