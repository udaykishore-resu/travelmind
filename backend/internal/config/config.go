package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Environment string
	Version     string
	LogLevel    string

	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Claude   ClaudeConfig
	GCP      GCPConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	PostgresURL  string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret         string
	JWTExpiry         int // seconds
	OAuthClientID     string
	OAuthClientSecret string
}

// ClaudeConfig holds Claude API configuration
type ClaudeConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// GCPConfig holds Google Cloud Platform configuration
type GCPConfig struct {
	ProjectID           string
	FirestoreCollection string
	BigQueryDataset     string
	PubSubTopic         string
	StorageBucket       string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	env := getEnv("ENVIRONMENT", "development")
	version := getEnv("VERSION", "dev")

	cfg := &Config{
		Environment: env,
		Version:     version,
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},

		Database: DatabaseConfig{
			PostgresURL:  getEnv("DATABASE_URL", ""),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLife:  getEnvInt("DB_CONN_MAX_LIFE", 5),
		},

		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},

		Auth: AuthConfig{
			JWTSecret:         getEnv("JWT_SECRET", ""),
			JWTExpiry:         getEnvInt("JWT_EXPIRY", 3600),
			OAuthClientID:     getEnv("OAUTH_CLIENT_ID", ""),
			OAuthClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),
		},

		Claude: ClaudeConfig{
			APIKey:    getEnv("CLAUDE_API_KEY", ""),
			Model:     getEnv("CLAUDE_MODEL", "claude-opus-4-1"),
			MaxTokens: getEnvInt("CLAUDE_MAX_TOKENS", 4096),
		},

		GCP: GCPConfig{
			ProjectID:           getEnv("GCP_PROJECT_ID", ""),
			FirestoreCollection: getEnv("FIRESTORE_COLLECTION", "bookings"),
			BigQueryDataset:     getEnv("BIGQUERY_DATASET", "analytics"),
			PubSubTopic:         getEnv("PUBSUB_TOPIC", "booking-events"),
			StorageBucket:       getEnv("STORAGE_BUCKET", "travelmind-storage"),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required configuration is set
func (c *Config) Validate() error {
	if c.Database.PostgresURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if c.Claude.APIKey == "" {
		return fmt.Errorf("CLAUDE_API_KEY environment variable is required")
	}
	if c.GCP.ProjectID == "" {
		return fmt.Errorf("GCP_PROJECT_ID environment variable is required")
	}
	return nil
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
