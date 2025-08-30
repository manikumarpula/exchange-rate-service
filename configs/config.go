package configs

import (
	"time"
)

type Config struct {
	Server    ServerConfig
	Redis     RedisConfig
	Providers ProviderConfig
}

type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ProviderConfig struct {
	Name     string
	BaseURL  string
	APIKey   string
	Timeout  time.Duration
	Priority int
}