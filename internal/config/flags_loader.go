package config

import "flag"

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
	flag.BoolVar(&config.EnableHTTPS, "s", config.EnableHTTPS, "enable HTTPS (TLS) server mode")

	flag.Parse()
}
