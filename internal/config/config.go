package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	SSLMode  string
	Schema   string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
	Issuer     string
}

type CORSConfig struct {
	AllowedOrigins []string
}

// Load configuration from environment variables with defaults and validation
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:         getEnvString("SERVER_HOST", "localhost"),
			Port:         getEnvInt("SERVER_PORT", 8000),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Name:     getEnvString("DB_NAME", "authz"),
			Host:     getEnvString("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Username: getEnvString("DB_USERNAME", "postgres"),
			Password: getEnvString("DB_PASSWORD", "Bruh0!0!"),
			SSLMode:  getEnvString("DB_SSL_MODE", "disable"),
			Schema:   getEnvString("DB_SCHEMA", "public"),
		},
		JWT: JWTConfig{
			Secret:     getEnvString("JWT_SECRET", "ai33yUUcmRPI64hq06ViG0404On-nMebsCtY4nTFqOg"),
			Expiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			Issuer:     getEnvString("JWT_ISSUER", "myapp"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvStringSlice("ACCESS_CONTROL_ALLOW_ORIGIN", []string{"*"}),
		},
	}

	if err := validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate checks required fields and business rules
func validate(config *Config) error {
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if config.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}

	if config.Database.Port < 1 || config.Database.Port > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535")
	}

	if !config.CORS.IsValidOrigins() {
		return fmt.Errorf("ACCESS_CONTROL_ALLOW_ORIGIN contains invalid origin")
	}

	// NOTE: add more validation if needed

	return nil
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		var result []string
		for item := range strings.SplitSeq(value, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

// DatabaseURL returns a formatted PostgreSQL connection string
func (d DatabaseConfig) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		d.Username, d.Password, d.Host, d.Port, d.Name, d.SSLMode, d.Schema,
	)
}

// ServerAddress returns the formatted server address
func (s ServerConfig) ServerAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// IsValidOrigins basic validation (allow * or valid http/https URLs)
func (c CORSConfig) IsValidOrigins() bool {
	for _, origin := range c.AllowedOrigins {
		if origin == "*" {
			return true
		}

		u, err := url.Parse(origin)
		if err != nil {
			return false
		}

		if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return false
		}
	}
	return true
}
