package api

import (
	"encoding/json"
	"net/http"
	"time"

	"exchange-rate-service/internal/models"
	"exchange-rate-service/internal/service"
	"exchange-rate-service/internal/errors"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
)

// Handlers handles HTTP requests
type Handlers struct {
	exchangeService service.ExchangeService
	logger          log.Logger
}

// NewHandlers creates new HTTP handlers
func NewHandlers(exchangeService service.ExchangeService, logger log.Logger) *Handlers {
	return &Handlers{
		exchangeService: exchangeService,
		logger:          logger,
	}
}

// HealthCheck handles health check requests
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.Log("method", "HealthCheck", "remote_addr", r.RemoteAddr)

	ctx := r.Context()
	health, err := h.exchangeService.HealthCheck(ctx)
	if err != nil {
		h.logger.Log("error", err, "method", "HealthCheck")
		models.WriteInternalError(w, "Health check failed")
		return
	}

	models.WriteSuccess(w, health, "Service is healthy")
}

// GetLatestRate handles latest rate requests
func (h *Handlers) GetLatestRate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseCurrency := vars["base"]
	targetCurrency := vars["target"]

	h.logger.Log("method", "GetLatestRate", "base", baseCurrency, "target", targetCurrency, "remote_addr", r.RemoteAddr)

	ctx := r.Context()
	rate, err := h.exchangeService.GetLatestRate(ctx, baseCurrency, targetCurrency)
	if err != nil {
		h.logger.Log("error", err, "method", "GetLatestRate")

		if utils.IsValidationError(err) {
			models.WriteBadRequest(w, err.Error())
			return
		}

		models.WriteInternalError(w, "Failed to get latest rate")
		return
	}

	models.WriteSuccess(w, rate, "Latest rate retrieved successfully")
}

// ConvertCurrency handles currency conversion requests
func (h *Handlers) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	h.logger.Log("method", "ConvertCurrency", "remote_addr", r.RemoteAddr)

	var req models.ConversionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Log("error", err, "method", "ConvertCurrency")
		models.WriteBadRequest(w, "Invalid request body")
		return
	}

	ctx := r.Context()
	response, err := h.exchangeService.ConvertCurrency(ctx, &req)
	if err != nil {
		h.logger.Log("error", err, "method", "ConvertCurrency")

		if utils.IsValidationError(err) {
			models.WriteBadRequest(w, err.Error())
			return
		}

		models.WriteInternalError(w, "Failed to convert currency")
		return
	}

	models.WriteSuccess(w, response, "Currency converted successfully")
}

// GetHistoricalRate handles historical rate requests
func (h *Handlers) GetHistoricalRate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseCurrency := vars["base"]
	targetCurrency := vars["target"]
	dateStr := vars["date"]

	h.logger.Log("method", "GetHistoricalRate", "base", baseCurrency, "target", targetCurrency, "date", dateStr, "remote_addr", r.RemoteAddr)

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.Log("error", err, "method", "GetHistoricalRate")
		models.WriteBadRequest(w, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	ctx := r.Context()
	rate, err := h.exchangeService.GetHistoricalRate(ctx, baseCurrency, targetCurrency, date)
	if err != nil {
		h.logger.Log("error", err, "method", "GetHistoricalRate")

		if utils.IsValidationError(err) {
			models.WriteBadRequest(w, err.Error())
			return
		}

		models.WriteInternalError(w, "Failed to get historical rate")
		return
	}

	models.WriteSuccess(w, rate, "Historical rate retrieved successfully")
}

// GetSupportedCurrencies handles supported currencies requests
func (h *Handlers) GetSupportedCurrencies(w http.ResponseWriter, r *http.Request) {
	h.logger.Log("method", "GetSupportedCurrencies", "remote_addr", r.RemoteAddr)

	ctx := r.Context()
	currencies, err := h.exchangeService.GetSupportedCurrencies(ctx)
	if err != nil {
		h.logger.Log("error", err, "method", "GetSupportedCurrencies")
		models.WriteInternalError(w, "Failed to get supported currencies")
		return
	}

	models.WriteSuccess(w, currencies, "Supported currencies retrieved successfully")
}

// GetRates handles bulk rates requests
func (h *Handlers) GetRates(w http.ResponseWriter, r *http.Request) {
	h.logger.Log("method", "GetRates", "remote_addr", r.RemoteAddr)

	// Parse query parameters
	baseCurrency := r.URL.Query().Get("base")
	if baseCurrency == "" {
		baseCurrency = "USD" // Default base currency
	}

	// Get supported currencies first
	ctx := r.Context()
	currencies, err := h.exchangeService.GetSupportedCurrencies(ctx)
	if err != nil {
		h.logger.Log("error", err, "method", "GetRates")
		models.WriteInternalError(w, "Failed to get supported currencies")
		return
	}

	// Get rates for all currencies
	var rates []*models.ExchangeRate
	for _, currency := range currencies {
		if currency.Code == baseCurrency {
			continue
		}

		rate, err := h.exchangeService.GetLatestRate(ctx, baseCurrency, currency.Code)
		if err != nil {
			h.logger.Log("error", err, "method", "GetRates", "base", baseCurrency, "target", currency.Code)
			continue // Skip failed rates, continue with others
		}

		rates = append(rates, rate)
	}

	response := map[string]interface{}{
		"base_currency": baseCurrency,
		"rates":         rates,
		"count":         len(rates),
	}

	models.WriteSuccess(w, response, "Rates retrieved successfully")
}

// GetTimeSeries handles time series requests
func (h *Handlers) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseCurrency := vars["base"]
	targetCurrency := vars["target"]

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		models.WriteBadRequest(w, "start_date and end_date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		models.WriteBadRequest(w, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		models.WriteBadRequest(w, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	if startDate.After(endDate) {
		models.WriteBadRequest(w, "start_date must be before end_date")
		return
	}

	h.logger.Log("method", "GetTimeSeries", "base", baseCurrency, "target", targetCurrency, "start", startDateStr, "end", endDateStr, "remote_addr", r.RemoteAddr)

	// Get rates for each date in the range
	ctx := r.Context()
	var rates []*models.HistoricalRate
	currentDate := startDate
	for !currentDate.After(endDate) {
		rate, err := h.exchangeService.GetHistoricalRate(ctx, baseCurrency, targetCurrency, currentDate)
		if err != nil {
			h.logger.Log("error", err, "method", "GetTimeSeries", "date", currentDate.Format("2006-01-02"))
			// Continue with other dates
		} else {
			rates = append(rates, rate)
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	response := map[string]interface{}{
		"base_currency":   baseCurrency,
		"target_currency": targetCurrency,
		"start_date":      startDateStr,
		"end_date":        endDateStr,
		"rates":           rates,
		"count":           len(rates),
	}

	models.WriteSuccess(w, response, "Time series retrieved successfully")
}
