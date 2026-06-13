// Package config loads environment configuration.
package config

import "os"

// Config holds all environment-derived service configuration.
type Config struct {
	DatabaseURL    string
	InternalAPIKey string
	Port           string
	LogLevel       string
}

// Load reads configuration from environment variables.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	return Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		InternalAPIKey: os.Getenv("INTERNAL_API_KEY"),
		Port:           port,
		LogLevel:       logLevel,
	}
}
