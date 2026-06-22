package panel

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr  string
	AuditPath string
}

func ConfigFromEnv() Config {
	return Config{
		HTTPAddr:  getenv("AMT_HTTP_ADDR", ":8080"),
		AuditPath: os.Getenv("AMT_AUDIT_PATH"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
