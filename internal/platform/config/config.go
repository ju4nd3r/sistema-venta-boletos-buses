package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration settings
type Config struct {
	Port        string
	GinMode     string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	DatabaseURL string
}

// Load loads configuration from environment variables and optional .env file
func Load() (*Config, error) {
	_ = godotenv.Load() // Continue with system environment variables if .env is missing

	port := getEnv("PORT", "8080")
	ginMode := getEnv("GIN_MODE", "debug")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "bus_admin")
	dbPassword := getEnv("DB_PASSWORD", "bus_secure_pass_2026")
	dbName := getEnv("DB_NAME", "bus_booking_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	return &Config{
		Port:        port,
		GinMode:     ginMode,
		DBHost:      dbHost,
		DBPort:      dbPort,
		DBUser:      dbUser,
		DBPassword:  dbPassword,
		DBName:      dbName,
		DBSSLMode:   dbSSLMode,
		DatabaseURL: dbURL,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
