package service

import (
	"context"
	"time"

	"exchange-rate-service/internal/models"
	"exchange-rate-service/internal/repository"
	"exchange-rate-service/internal/errors"

	"github.com/go-kit/log"
)

// ExchangeService defines the interface for exchange rate operations
type ExchangeService interface {
	GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*models.ExchangeRate, error)
	ConvertCurrency(ctx context.Context, req *models.ConversionRequest) (*models.ConversionResponse, error)
	GetHistoricalRate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*models.HistoricalRate, error)
	GetSupportedCurrencies(ctx context.Context) ([]*models.Currency, error)
	HealthCheck(ctx context.Context) (*models.HealthResponse, error)
}

// exchangeService implements ExchangeService
type exchangeService struct {
	rateRepo repository.RateRepository
	logger   log.Logger
}

// NewExchangeService creates a new exchange service
func NewExchangeService(rateRepo repository.RateRepository, logger log.Logger) ExchangeService {
	return &exchangeService{
		rateRepo: rateRepo,
		logger:   logger,
	}
}

// GetLatestRate retrieves the latest exchange rate
func (s *exchangeService) GetLatestRate(ctx context.Context, baseCurrency, targetCurrency string) (*models.ExchangeRate, error) {
	s.logger.Log("method", "GetLatestRate", "base", baseCurrency, "target", targetCurrency)

	// Validate currencies
	if err := s.validateCurrencies(baseCurrency, targetCurrency); err != nil {
		return nil, err
	}

	// Get rate from repository
	rate, err := s.rateRepo.GetLatestRate(ctx, baseCurrency, targetCurrency)
	if err != nil {
		s.logger.Log("error", err, "method", "GetLatestRate")
		return nil, err
	}

	return rate, nil
}

// ConvertCurrency converts an amount from one currency to another
func (s *exchangeService) ConvertCurrency(ctx context.Context, req *models.ConversionRequest) (*models.ConversionResponse, error) {
	s.logger.Log("method", "ConvertCurrency", "from", req.FromCurrency, "to", req.ToCurrency, "amount", req.Amount)

	// Validate request
	if err := s.validateConversionRequest(req); err != nil {
		return nil, err
	}

	var rate interface{}
	var err error

	// If date is specified, get historical rate
	if req.Date != "" {
		date, parseErr := time.Parse("2006-01-02", req.Date)
		if parseErr != nil {
			return nil,errors.NewValidationError("invalid date format", "date must be in YYYY-MM-DD format")
		}
		rate, err = s.rateRepo.GetHistoricalRate(ctx, req.FromCurrency, req.ToCurrency, date)
	} else {
		rate, err = s.rateRepo.GetLatestRate(ctx, req.FromCurrency, req.ToCurrency)
	}

	if err != nil {
		s.logger.Log("error", err, "method", "ConvertCurrency")
		return nil, err
	}

	// Calculate converted amount and extract rate info
	var rateValue float64
	var provider string
	var fetchedAt time.Time

	if req.Date != "" {
		// Historical rate
		if histRate, ok := rate.(*models.HistoricalRate); ok {
			rateValue = histRate.Rate
			provider = histRate.Provider
			fetchedAt = histRate.FetchedAt
		}
	} else {
		// Latest rate
		if latestRate, ok := rate.(*models.ExchangeRate); ok {
			rateValue = latestRate.Rate
			provider = latestRate.Provider
			fetchedAt = latestRate.FetchedAt
		}
	}

	convertedAmount := req.Amount * rateValue

	response := &models.ConversionResponse{
		FromCurrency:    req.FromCurrency,
		ToCurrency:      req.ToCurrency,
		Amount:          req.Amount,
		ConvertedAmount: convertedAmount,
		Rate:            rateValue,
		Provider:        provider,
		FetchedAt:       fetchedAt,
	}

	return response, nil
}

// GetHistoricalRate retrieves a historical exchange rate
func (s *exchangeService) GetHistoricalRate(ctx context.Context, baseCurrency, targetCurrency string, date time.Time) (*models.HistoricalRate, error) {
	s.logger.Log("method", "GetHistoricalRate", "base", baseCurrency, "target", targetCurrency, "date", date.Format("2006-01-02"))

	// Validate currencies
	if err := s.validateCurrencies(baseCurrency, targetCurrency); err != nil {
		return nil, err
	}

	// Get historical rate from repository
	rate, err := s.rateRepo.GetHistoricalRate(ctx, baseCurrency, targetCurrency, date)
	if err != nil {
		s.logger.Log("error", err, "method", "GetHistoricalRate")
		return nil, err
	}

	return rate, nil
}

// GetSupportedCurrencies retrieves list of supported currencies
func (s *exchangeService) GetSupportedCurrencies(ctx context.Context) ([]*models.Currency, error) {
	s.logger.Log("method", "GetSupportedCurrencies")

	currencies, err := s.rateRepo.GetSupportedCurrencies(ctx)
	if err != nil {
		s.logger.Log("error", err, "method", "GetSupportedCurrencies")
		return nil, err
	}

	return currencies, nil
}

// HealthCheck performs a health check
func (s *exchangeService) HealthCheck(ctx context.Context) (*models.HealthResponse, error) {
	s.logger.Log("method", "HealthCheck")

	// Check repository health
	providers, err := s.rateRepo.HealthCheck(ctx)
	if err != nil {
		s.logger.Log("error", err, "method", "HealthCheck")
		return nil, err
	}

	response := &models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Providers: providers,
		Cache:     "connected",
	}

	return response, nil
}

// validateCurrencies validates currency codes
func (s *exchangeService) validateCurrencies(baseCurrency, targetCurrency string) error {
	if baseCurrency == "" {
		return errors.NewValidationError("base currency is required", "base_currency cannot be empty")
	}
	if targetCurrency == "" {
		return errors.NewValidationError("target currency is required", "target_currency cannot be empty")
	}
	if baseCurrency == targetCurrency {
		return errors.NewValidationError("currencies must be different", "base_currency and target_currency cannot be the same")
	}
	return nil
}

// validateConversionRequest validates conversion request
func (s *exchangeService) validateConversionRequest(req *models.ConversionRequest) error {
	if req.FromCurrency == "" {
		return errors.NewValidationError("from currency is required", "from_currency cannot be empty")
	}
	if req.ToCurrency == "" {
		return errors.NewValidationError("to currency is required", "to_currency cannot be empty")
	}
	if req.Amount <= 0 {
		return errors.NewValidationError("amount must be positive", "amount must be greater than 0")
	}
	if req.FromCurrency == req.ToCurrency {
		return errors.NewValidationError("currencies must be different", "from_currency and to_currency cannot be the same")
	}
	return nil
}
