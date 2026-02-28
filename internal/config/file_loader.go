// загрузка конфигурации из JSON (CONFIG, -c, -config).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// fileConfig описывает поддерживаемые поля JSON-конфигурации.
type fileConfig struct {
	Address          *string `json:"server_address"`
	GRPCAddress      *string `json:"grpc_server_address"`
	ReturningAddress *string `json:"base_url"`
	FilePath         *string `json:"file_storage_path"`
	DatabaseDSN      *string `json:"database_dsn"`
	EnableHTTPS      *bool   `json:"enable_https"`

	LogLevel      *string `json:"log_level"`
	SigningKey    *string `json:"secret_key"`
	AuditFile     *string `json:"audit_file"`
	AuditURL      *string `json:"audit_url"`
	DevMode       *bool   `json:"dev_mode"`
	TrustedSubnet *string `json:"trusted_subnet"`
}

// ParseFile заполняет config значениями из JSON-файла.
func ParseFile(config *Config) {
	path := getConfigPath()
	if path == "" {
		return
	}

	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open config file %q: %v\n", path, err)
		return
	}
	defer f.Close()

	var fc fileConfig
	if err := json.NewDecoder(f).Decode(&fc); err != nil {
		fmt.Fprintf(os.Stderr, "cannot decode config file %q: %v\n", path, err)
		return
	}

	if fc.Address != nil {
		config.Address = *fc.Address
	}
	if fc.GRPCAddress != nil {
		config.GRPCAddress = *fc.GRPCAddress
	}
	if fc.ReturningAddress != nil {
		config.ReturningAddress = *fc.ReturningAddress
	}
	if fc.FilePath != nil {
		config.FilePath = *fc.FilePath
	}
	if fc.DatabaseDSN != nil {
		config.DatabaseDSN = *fc.DatabaseDSN
	}
	if fc.EnableHTTPS != nil {
		config.EnableHTTPS = *fc.EnableHTTPS
	}
	if fc.LogLevel != nil {
		config.LogLevel = *fc.LogLevel
	}
	if fc.SigningKey != nil {
		config.SigningKey = *fc.SigningKey
	}
	if fc.AuditFile != nil {
		config.AuditFile = *fc.AuditFile
	}
	if fc.AuditURL != nil {
		config.AuditURL = *fc.AuditURL
	}
	if fc.DevMode != nil {
		config.DevMode = *fc.DevMode
	}
	if fc.TrustedSubnet != nil {
		config.TrustedSubnet = *fc.TrustedSubnet
	}
}

// getConfigPath возвращает путь к JSON-конфигу из CONFIG или флагов -c/-config.
func getConfigPath() string {
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-c" || arg == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
			continue
		}

		if strings.HasPrefix(arg, "-c=") {
			return strings.TrimPrefix(arg, "-c=")
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config=")
		}
	}

	return ""
}
