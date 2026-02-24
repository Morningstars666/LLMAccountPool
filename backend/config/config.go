package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServerPort       string
	ServerHost       string
	DatabaseURL      string
	JWTSecret        string
	AllowedOrigins   []string
	MaxLoginAttempts int
	LockoutDuration  int // minutes
}

func LoadConfig() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	allowedOrigins := []string{}
	if originsEnv := os.Getenv("ALLOWED_ORIGINS"); originsEnv != "" {
		for _, origin := range strings.Split(originsEnv, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		ServerHost:       getEnv("SERVER_HOST", "http://localhost:8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "../data/llmaccountpool.db"),
		JWTSecret:        jwtSecret,
		AllowedOrigins:   allowedOrigins,
		MaxLoginAttempts: getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:  getEnvAsInt("LOCKOUT_DURATION", 15),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		return intValue
	}
	return defaultValue
}
