package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
	NATSURL     string
	LogLevel    string
	Environment string
}

func Load() *Config {
	return &Config{
		Port:        getEnvInt("PORT", 8008),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://gondor:gondor_dev@localhost:5432/gondor_event_bus?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		NATSURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
