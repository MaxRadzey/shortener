package config

import "os"

type Config struct {
	Address          string
	ReturningAddress string
}

func New() *Config {
	return &Config{
		Address:          "localhost:8080",
		ReturningAddress: "http://localhost:8080",
	}
}

func ParseEnv(config *Config) {
	if Address := os.Getenv("SERVER_ADDRESS"); Address != "" {
		config.Address = Address
	}
	if ReturningAddress := os.Getenv("BASE_URL"); ReturningAddress != "" {
		config.ReturningAddress = ReturningAddress
	}
}
