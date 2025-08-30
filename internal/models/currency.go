package models

import (
	"time"
)

// Currency represents a currency code
type Currency struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol,omitempty"`
	IsBase      bool   `json:"is_base,omitempty"`
	IsSupported bool   `json:"is_supported"`
}

// ExchangeRate represents an exchange rate between two currencies
type ExchangeRate struct {
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	Provider       string    `json:"provider"`
	FetchedAt      time.Time `json:"fetched_at"`
	IsStale        bool      `json:"is_stale,omitempty"`
	TTL            int64     `json:"ttl,omitempty"`
}

// ConversionRequest represents a currency conversion request
type ConversionRequest struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date,omitempty"` // Optional historical date
}

// ConversionResponse represents a currency conversion response
type ConversionResponse struct {
	FromCurrency    string    `json:"from_currency"`
	ToCurrency      string    `json:"to_currency"`
	Amount          float64   `json:"amount"`
	ConvertedAmount float64   `json:"converted_amount"`
	Rate            float64   `json:"rate"`
	Provider        string    `json:"provider"`
	FetchedAt       time.Time `json:"fetched_at"`
}

// HistoricalRate represents a historical exchange rate
type HistoricalRate struct {
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	Date           time.Time `json:"date"`
	Provider       string    `json:"provider"`
	FetchedAt      time.Time `json:"fetched_at"`
}

// ProviderResponse represents a response from an exchange rate provider
type ProviderResponse struct {
	Success bool                   `json:"success"`
	Base    string                 `json:"base,omitempty"`
	Date    string                 `json:"date,omitempty"`
	Rates   map[string]float64     `json:"rates,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Raw     map[string]interface{} `json:"raw,omitempty"`
}

// OpenERAPIResponse represents the response from open.er-api.com
type OpenERAPIResponse struct {
	Result             string             `json:"result"`
	Provider           string             `json:"provider"`
	Documentation      string             `json:"documentation"`
	TermsOfUse         string             `json:"terms_of_use"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string             `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64              `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string             `json:"time_next_update_utc"`
	TimeEOLUnix        int64              `json:"time_eol_unix"`
	BaseCode           string             `json:"base_code"`
	Rates              map[string]float64 `json:"rates"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Providers map[string]string `json:"providers"`
	Cache     string            `json:"cache"`
}
