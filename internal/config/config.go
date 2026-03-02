// Package config загружает настройки приложения из флагов и переменных окружения.
package config

import "github.com/ilyakaznacheev/cleanenv"

// Config — настройки приложения (адрес, БД, файл, логи, аудит, ключ подписи).
type Config struct {
	Address          string `env:"SERVER_ADDRESS"`      // адрес HTTP-сервера (например :8080)
	GRPCAddress      string `env:"GRPC_SERVER_ADDRESS"` // адрес gRPC-сервера (например :9090)
	ReturningAddress string `env:"BASE_URL"`            // базовый URL для коротких ссылок в ответах
	LogLevel         string `env:"LOG_LEVEL"`           // уровень логов (info, debug и т.д.)
	FilePath         string `env:"FILE_STORAGE_PATH"`   // путь к файлу хранилища, если не используется БД
	DatabaseDSN      string `env:"DATABASE_DSN"`        // строка подключения к PostgreSQL; пусто — БД не используется

	SigningKey    string `env:"SECRET_KEY"`     // секрет для подписи куки (в проде задавать через SECRET_KEY)
	AuditFile     string `env:"AUDIT_FILE"`     // путь к файлу аудита; пусто — выключено
	AuditURL      string `env:"AUDIT_URL"`      // URL приёмника аудита; пусто — выключено
	DevMode       bool   `env:"DEV_MODE"`       // режим разработки (pprof и т.п.)
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`   // включает запуск сервера по HTTPS (TLS)
	TLSCertFile   string `env:"TLS_CERT_FILE"`  // путь к сертификату для TLS
	TLSKeyFile    string `env:"TLS_KEY_FILE"`   // путь к ключу для TLS
	TrustedSubnet string `env:"TRUSTED_SUBNET"` // CIDR доверенной подсети
}

// New возвращает готовый конфиг: дефолты + файл + переменные окружения + флаги.
func New() *Config {
	cfg := &Config{
		Address:          "localhost:8080",
		GRPCAddress:      "localhost:9090",
		ReturningAddress: "http://localhost:8080",
		LogLevel:         "info",
		FilePath:         "data.json",
		DatabaseDSN:      "postgres://shortener:shortener@localhost:5432/shortener",
		SigningKey:       "dev-signing-key-change-in-production",
		TLSCertFile:      "server.crt",
		TLSKeyFile:       "server.key",
	}
	ParseFile(cfg)
	_ = cleanenv.ReadEnv(cfg)
	ParseFlags(cfg)
	return cfg
}
