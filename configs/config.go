package configs

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	AppPort string
	CORS    CORSConfig
	APIKey  string
	Cache   CacheConfig
	Logger  *slog.Logger
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	DefaultExpiration time.Duration `json:"default_expiration"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cacheDefaultExpiration, err := parseDuration("CACHE_DEFAULT_EXPIRATION", 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cache default expiration: %w", err)
	}

	cacheCleanupInterval, err := parseDuration("CACHE_CLEANUP_INTERVAL", 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cache cleanup interval: %w", err)
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		return nil, fmt.Errorf("ALLOWED_ORIGINS environment variable is required")
	}

	allowedMethods := os.Getenv("ALLOWED_METHODS")
	if allowedMethods == "" {
		return nil, fmt.Errorf("ALLOWED_METHODS environment variable is required")
	}

	allowedHeaders := os.Getenv("ALLOWED_HEADERS")
	if allowedHeaders == "" {
		return nil, fmt.Errorf("ALLOWED_HEADERS environment variable is required")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is required")
	}

	appPort := os.Getenv("PORT")
	if appPort == "" {
		return nil, fmt.Errorf("PORT environment variable is required")
	}

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return &Config{
		AppPort: appPort,
		CORS: CORSConfig{
			AllowedOrigins:   strings.Split(allowedOrigins, ","),
			AllowedMethods:   strings.Split(allowedMethods, ","),
			AllowedHeaders:   strings.Split(allowedHeaders, ","),
			AllowCredentials: true,
		},
		APIKey: apiKey,
		Cache: CacheConfig{
			DefaultExpiration: cacheDefaultExpiration,
			CleanupInterval:   cacheCleanupInterval,
		},
		Logger: logger,
	}, nil
}

func parseDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value for %s: %s", key, value)
	}
	return parsed, nil
}
