// Package config загружает настройки приложения из флагов и переменных окружения.
package config

// Config — настройки приложения (адрес, БД, файл, логи, аудит, ключ подписи).
type Config struct {
	Address          string // адрес сервера (например :8080)
	ReturningAddress string // базовый URL для коротких ссылок в ответах
	LogLevel         string // уровень логов (info, debug и т.д.)
	FilePath         string // путь к файлу хранилища, если не используется БД
	DatabaseDSN      string // строка подключения к PostgreSQL; пусто — БД не используется

	SigningKey  string // SigningKey — секрет для подписи куки (в проде задавать через SECRET_KEY)
	AuditFile   string // путь к файлу аудита; пусто — выключено
	AuditURL    string // URL приёмника аудита; пусто — выключено
	DevMode     bool   // режим разработки (pprof и т.п.)
	EnableHTTPS bool   // включает запуск сервера по HTTPS (TLS)
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
