package config

import "os"

type Config struct {
	Address          string
	ReturningAddress string
	LogLevel         string
	FilePath         string
	DBConfig         DBConfig
}

type DBConfig struct {
	DSN string
}

func New() *Config {
	return &Config{
		Address:          "localhost:8080",
		ReturningAddress: "http://localhost:8080",
		LogLevel:         "info",
		FilePath:         "data.json",
		DBConfig: DBConfig{
			DSN: "postgres://username:password@localhost:5432/database_name",
		},
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
	if DSN := os.Getenv("DATABASE_DSN"); DSN != "" {
		config.DBConfig.DSN = DSN
	}
}
