package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	JWT         JWTConfig
	StoragePath string
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	URL string
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port int
	Host string
}

// JWTConfig holds JWT authentication settings
type JWTConfig struct {
	Secret     string
	ExpireHours int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://lakehouse:lakehouse_dev@localhost:5432/lakehouse_finance?sslmode=disable"),
		},
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
		StoragePath: getEnv("STORAGE_PATH", "./data"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(cfg.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return cfg, nil
}

// Validate checks required configuration
func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

// ServerAddr returns the full server address
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
