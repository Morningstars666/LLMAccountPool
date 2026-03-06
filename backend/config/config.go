package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type DatabaseType string

const (
	SQLiteType   DatabaseType = "sqlite"
	PostgresType DatabaseType = "postgres"
	MySQLType    DatabaseType = "mysql"
)

type Config struct {
	ServerPort       string
	ServerHost       string
	DatabaseURL      string
	DatabaseType     DatabaseType
	JWTSecret        string
	AllowedOrigins   []string
	MaxLoginAttempts int
	LockoutDuration  int

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int

	EnableWALMode bool
	BusyTimeout   int
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

	dbType := DatabaseType(strings.ToLower(GetEnv("DB_TYPE", "sqlite")))
	if dbType != SQLiteType && dbType != PostgresType && dbType != MySQLType {
		dbType = SQLiteType
	}

	var databaseURL string
	if dbType == SQLiteType {
		databaseURL = GetEnv("DATABASE_URL", "../data/llmaccountpool.db")
	} else {
		databaseURL = buildDSN(dbType)
	}

	maxOpenConns := GetEnvAsInt("DB_MAX_OPEN_CONNS", 100)
	maxIdleConns := GetEnvAsInt("DB_MAX_IDLE_CONNS", 20)
	connMaxLifetime := GetEnvAsInt("DB_CONN_MAX_LIFETIME", 300)
	connMaxIdleTime := GetEnvAsInt("DB_CONN_MAX_IDLE_TIME", 60)

	enableWALMode := GetEnvAsBool("DB_ENABLE_WAL_MODE", false)
	busyTimeout := GetEnvAsInt("DB_BUSY_TIMEOUT", 5000)

	return &Config{
		ServerPort:       GetEnv("SERVER_PORT", "8080"),
		ServerHost:       GetEnv("SERVER_HOST", "http://localhost:8080"),
		DatabaseURL:      databaseURL,
		DatabaseType:     dbType,
		JWTSecret:        jwtSecret,
		AllowedOrigins:   allowedOrigins,
		MaxLoginAttempts: GetEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:  GetEnvAsInt("LOCKOUT_DURATION", 15),

		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,

		EnableWALMode: enableWALMode,
		BusyTimeout:   busyTimeout,
	}
}

func buildDSN(dbType DatabaseType) string {
	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "")
	user := GetEnv("DB_USER", "postgres")
	password := GetEnv("DB_PASSWORD", "")
	dbname := GetEnv("DB_NAME", "llmaccountpool")
	sslmode := GetEnv("DB_SSLMODE", "disable")

	if dbType == PostgresType {
		if port == "" {
			port = "5432"
		}
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
			host, port, user, password, dbname, sslmode,
		)
	}

	if dbType == MySQLType {
		if port == "" {
			port = "3306"
		}
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
			user, password, host, port, dbname,
		)
	}

	return ""
}

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		lower := strings.ToLower(value)
		if lower == "true" || lower == "1" || lower == "yes" {
			return true
		}
		if lower == "false" || lower == "0" || lower == "no" {
			return false
		}
	}
	return defaultValue
}
