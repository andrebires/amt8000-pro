package panel

import (
	"os"
	"strconv"
)

type Config struct {
	Host     string
	Port     int
	Password string
	HTTPAddr string
}

func ConfigFromEnv() Config {
	return Config{
		Host:     getenv("AMT_HOST", "192.168.4.1"),
		Port:     getenvInt("AMT_PORT", 9009),
		Password: getenv("AMT_PASSWORD", ""),
		HTTPAddr: getenv("AMT_HTTP_ADDR", ":8080"),
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
