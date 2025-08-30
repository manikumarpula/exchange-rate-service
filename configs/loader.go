package configs

import (
	"os"
	"strconv"
	"time"
)

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	shutdownTimeout := getEnvAsDuration("SHUTDOWN_TIMEOUT", 30*time.Second)

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvAsInt("REDIS_DB", 0)

	// Provider configurations - only using open.er-api.com
	provider := &ProviderConfig{
			Name:     "open.er-api.com",
			BaseURL:  getEnv("OPEN_ER_API_URL", "https://open.er-api.com/v6"),
			APIKey:   getEnv("OPEN_ER_API_KEY", ""),
			Timeout:  getEnvAsDuration("OPEN_ER_API_TIMEOUT", 10*time.Second),
			Priority: 1,
	}

	return &Config{
		Server: ServerConfig{
			Port:            port,
			ShutdownTimeout: shutdownTimeout,
		},
		Redis: RedisConfig{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		},
		Providers: *provider,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
