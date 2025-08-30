package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"exchange-rate-service/configs"
	"exchange-rate-service/internal/models"

	"github.com/go-kit/log"
	"github.com/redis/go-redis/v9"
)

// RateRepository defines the interface for rate data operations
type RateRepository interface {
	GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*models.ExchangeRate, error)
	GetHistoricalRate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*models.HistoricalRate, error)
	GetSupportedCurrencies(ctx context.Context) ([]*models.Currency, error)
	HealthCheck(ctx context.Context) (map[string]string, error)
}

// rateRepository implements RateRepository
type rateRepository struct {
	config *configs.Config
	logger log.Logger
	cache  Cache
	client *OpenERAPIClient
}

// Cache defines the cache interface
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	Ping(ctx context.Context) error
}

// NewRateRepository creates a new rate repository
func NewRateRepository(config *configs.Config, logger log.Logger) RateRepository {
	// Initialize cache (Redis)
	var cache Cache
	redisCache, err := NewRedisCache(config.Redis.Addr, config.Redis.Password, config.Redis.DB)
	if err != nil {
		logger.Log("error", err, "msg", "failed to initialize Redis cache")
		// Fallback to in-memory cache
		cache = NewInMemoryCache()
	} else {
		cache = redisCache
	}

	// Initialize single provider client (open-er-api.com)
	providerCfg := config.Providers
	if strings.ToLower(providerCfg.Name) != "open.er-api.com" && strings.ToLower(providerCfg.Name) != "openerapi" {
		if providerCfg.BaseURL == "" {
			providerCfg.BaseURL = "https://open.er-api.com/v6"
		}
		if providerCfg.Timeout == 0 {
			providerCfg.Timeout = 10 * time.Second
		}
		providerCfg.Name = "open.er-api.com"
		logger.Log("warn", "provider overridden to open.er-api.com")
	}
	client := NewOpenERAPIClient(providerCfg, logger)

	return &rateRepository{
		config: config,
		logger: logger,
		cache:  cache,
		client: client,
	}
}

// GetLatestRate retrieves the latest exchange rate
func (r *rateRepository) GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*models.ExchangeRate, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("rate:%s:%s:latest", baseCurrency, targetCurrency)
	var rate models.ExchangeRate
	if err := r.cache.Get(ctx, cacheKey, &rate); err == nil {
		r.logger.Log("msg", "rate found in cache", "base", baseCurrency, "target", targetCurrency)
		return &rate, nil
	}

	// Fetch from provider
	ratePtr, err := r.client.GetLatestRate(ctx, baseCurrency, targetCurrency)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, ratePtr, 5*time.Minute); err != nil {
		r.logger.Log("error", err, "msg", "failed to cache rate")
	}

	return ratePtr, nil
}

// GetHistoricalRate retrieves a historical exchange rate
func (r *rateRepository) GetHistoricalRate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*models.HistoricalRate, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("rate:%s:%s:%s", baseCurrency, targetCurrency, date.Format("2006-01-02"))
	var rate models.HistoricalRate
	if err := r.cache.Get(ctx, cacheKey, &rate); err == nil {
		r.logger.Log("msg", "historical rate found in cache", "base", baseCurrency, "target", targetCurrency, "date", date.Format("2006-01-02"))
		return &rate, nil
	}

	ratePtr, err := r.client.GetHistoricalRate(ctx, baseCurrency, targetCurrency, date)
	if err != nil {
		return nil, err
	}

	// Cache the result (historical rates can be cached longer)
	if err := r.cache.Set(ctx, cacheKey, ratePtr, 24*time.Hour); err != nil {
		r.logger.Log("error", err, "msg", "failed to cache historical rate")
	}

	return ratePtr, nil
}

// GetSupportedCurrencies retrieves list of supported currencies
func (r *rateRepository) GetSupportedCurrencies(ctx context.Context) ([]*models.Currency, error) {
	// Try cache first
	cacheKey := "currencies:supported"
	var currencies []*models.Currency
	if err := r.cache.Get(ctx, cacheKey, &currencies); err == nil {
		r.logger.Log("msg", "supported currencies found in cache")
		return currencies, nil
	}

	currencies, err := r.client.GetSupportedCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result (currencies list changes rarely)
	if err := r.cache.Set(ctx, cacheKey, currencies, 24*time.Hour); err != nil {
		r.logger.Log("error", err, "msg", "failed to cache currencies")
	}

	return currencies, nil
}

// HealthCheck performs a health check
func (r *rateRepository) HealthCheck(ctx context.Context) (map[string]string, error) {
	providers := make(map[string]string)

	// Check cache health
	if err := r.cache.Ping(ctx); err != nil {
		providers["cache"] = "unhealthy"
	} else {
		providers["cache"] = "healthy"
	}

	// Check provider health
	if r.client != nil {
		if err := r.client.HealthCheck(ctx); err != nil {
			providers[r.client.Name()] = "unhealthy"
		} else {
			providers[r.client.Name()] = "healthy"
		}
	} else {
		providers["open.er-api.com"] = "unconfigured"
	}

	return providers, nil
}

// InMemoryCache implements a simple in-memory cache
type InMemoryCache struct {
	data map[string]interface{}
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]interface{}),
	}
}

func (c *InMemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	// Simple implementation - in real world, you'd want proper serialization
	return fmt.Errorf("in-memory cache not implemented")
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Simple implementation - in real world, you'd want proper serialization
	return nil
}

func (c *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := c.data[key]
	return exists, nil
}

func (c *InMemoryCache) Ping(ctx context.Context) error {
	return nil
}

// OpenERAPIClient implements ProviderClient for open.er-api.com API
type OpenERAPIClient struct {
	name    string
	baseURL string
	client  *http.Client
	logger  log.Logger
}

// NewOpenERAPIClient creates a new client for open.er-api.com API
func NewOpenERAPIClient(config configs.ProviderConfig, logger log.Logger) *OpenERAPIClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &OpenERAPIClient{
		name:    config.Name,
		baseURL: config.BaseURL,
		client:  httpClient,
		logger:  logger,
	}
}

// Name returns the provider name
func (c *OpenERAPIClient) Name() string {
	return c.name
}

// GetLatestRate retrieves the latest exchange rate from open.er-api.com
func (c *OpenERAPIClient) GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*models.ExchangeRate, error) {
	url := fmt.Sprintf("%s/latest/%s", c.baseURL, baseCurrency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp models.OpenERAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Result != "success" {
		return nil, fmt.Errorf("API returned error result: %s", apiResp.Result)
	}

	rate, exists := apiResp.Rates[targetCurrency]
	if !exists {
		return nil, fmt.Errorf("rate not found for %s", targetCurrency)
	}

	return &models.ExchangeRate{
		BaseCurrency:   baseCurrency,
		TargetCurrency: targetCurrency,
		Rate:           rate,
		Provider:       c.name,
		FetchedAt:      time.Now(),
	}, nil
}

// GetHistoricalRate retrieves a historical exchange rate from open.er-api.com
// Note: Historical rates endpoint may not be available in free tier
func (c *OpenERAPIClient) GetHistoricalRate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*models.HistoricalRate, error) {
	// For now, return an error indicating historical rates are not supported
	// In a production environment, you might want to implement a fallback strategy
	// or use a different provider that supports historical rates
	return nil, fmt.Errorf("historical rates not supported by %s in free tier", c.name)
}

// GetSupportedCurrencies retrieves list of supported currencies from open.er-api.com
func (c *OpenERAPIClient) GetSupportedCurrencies(ctx context.Context) ([]*models.Currency, error) {
	// For open.er-api.com, we can get currencies by making a request to get rates for USD
	// and then extract the currency codes from the response
	url := fmt.Sprintf("%s/latest/USD", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp models.OpenERAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Result != "success" {
		return nil, fmt.Errorf("API returned error result: %s", apiResp.Result)
	}

	var currencies []*models.Currency
	for code := range apiResp.Rates {
		currencies = append(currencies, &models.Currency{
			Code:        code,
			Name:        code, // We don't have names, so use code as name
			IsSupported: true,
		})
	}

	return currencies, nil
}

// HealthCheck performs a health check against the open.er-api.com API
func (c *OpenERAPIClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/latest/USD", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Use a shorter timeout for health checks
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// RedisCache implements Redis cache
type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, jsonData, expiration).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
