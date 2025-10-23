package config

import (
	"fmt"
	"os"
)

// Config holds the database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewConfig creates a new configuration from environment variables
func NewConfig() (*Config, error) {
	cfg := &Config{
		Host:     getEnv("PGHOST", "localhost"),
		Port:     getEnv("PGPORT", "5432"),
		User:     getEnv("PGUSER", "postgres"),
		Password: getEnv("PGPASSWORD", ""),
		Database: getEnv("PGDATABASE", "postgres"),
		SSLMode:  getEnv("PGSSLMODE", "disable"),
	}

	if cfg.Password == "" {
		return nil, fmt.Errorf("PGPASSWORD environment variable is required")
	}

	return cfg, nil
}

// ConnectionString returns the PostgreSQL connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
