package config

import (
	"os"
	"strconv"
)

// ParseEnv подставляет в config значения из переменных окружения.
func ParseEnv(config *Config) {
	if Address := os.Getenv("SERVER_ADDRESS"); Address != "" {
		config.Address = Address
	}
	if v := os.Getenv("GRPC_SERVER_ADDRESS"); v != "" {
		config.GRPCAddress = v
	}
	if ReturningAddress := os.Getenv("BASE_URL"); ReturningAddress != "" {
		config.ReturningAddress = ReturningAddress
	}
	if LogLevel := os.Getenv("LOG_LEVEL"); LogLevel != "" {
		config.LogLevel = LogLevel
	}
	if FilePath := os.Getenv("FILE_STORAGE_PATH"); FilePath != "" {
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
	if v := os.Getenv("ENABLE_HTTPS"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil && b {
			config.EnableHTTPS = true
		}
	}
	if v := os.Getenv("TRUSTED_SUBNET"); v != "" {
		config.TrustedSubnet = v
	}
}
