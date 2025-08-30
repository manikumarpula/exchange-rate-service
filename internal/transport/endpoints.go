package transport

import (
	"context"
	"fmt"
	"time"

	"exchange-rate-service/internal/models"
	"exchange-rate-service/internal/service"

	kitendpoint "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
)

// Endpoints aggregates all go-kit endpoints for the service.
type Endpoints struct {
	GetLatestRateEndpoint          kitendpoint.Endpoint
	ConvertCurrencyEndpoint        kitendpoint.Endpoint
	GetHistoricalRatesEndpoint     kitendpoint.Endpoint
	GetSupportedCurrenciesEndpoint kitendpoint.Endpoint
}

// MakeEndpoints constructs all endpoints with middleware.
func MakeEndpoints(svc service.ExchangeService, logger log.Logger) Endpoints {
	var getLatestRateEndpoint kitendpoint.Endpoint
	{
		getLatestRateEndpoint = makeGetLatestRateEndpoint(svc)
		getLatestRateEndpoint = LoggingMiddleware(log.With(logger, "method", "GetLatestRate"))(getLatestRateEndpoint)
		getLatestRateEndpoint = RecoveryMiddleware(logger)(getLatestRateEndpoint)
	}

	var convertCurrencyEndpoint kitendpoint.Endpoint
	{
		convertCurrencyEndpoint = makeConvertCurrencyEndpoint(svc)
		convertCurrencyEndpoint = LoggingMiddleware(log.With(logger, "method", "ConvertCurrency"))(convertCurrencyEndpoint)
		convertCurrencyEndpoint = RecoveryMiddleware(logger)(convertCurrencyEndpoint)
	}

	var getHistoricalRatesEndpoint kitendpoint.Endpoint
	{
		getHistoricalRatesEndpoint = makeGetHistoricalRatesEndpoint(svc)
		getHistoricalRatesEndpoint = LoggingMiddleware(log.With(logger, "method", "GetHistoricalRates"))(getHistoricalRatesEndpoint)
		getHistoricalRatesEndpoint = RecoveryMiddleware(logger)(getHistoricalRatesEndpoint)
	}

	var getSupportedCurrenciesEndpoint kitendpoint.Endpoint
	{
		getSupportedCurrenciesEndpoint = makeGetSupportedCurrenciesEndpoint(svc)
		getSupportedCurrenciesEndpoint = LoggingMiddleware(log.With(logger, "method", "GetSupportedCurrencies"))(getSupportedCurrenciesEndpoint)
		getSupportedCurrenciesEndpoint = RecoveryMiddleware(logger)(getSupportedCurrenciesEndpoint)
	}

	return Endpoints{
		GetLatestRateEndpoint:          getLatestRateEndpoint,
		ConvertCurrencyEndpoint:        convertCurrencyEndpoint,
		GetHistoricalRatesEndpoint:     getHistoricalRatesEndpoint,
		GetSupportedCurrenciesEndpoint: getSupportedCurrenciesEndpoint,
	}
}

// Request/Response DTOs
type GetLatestRateRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GetLatestRateResponse struct {
	Rate  interface{} `json:"rate,omitempty"`
	Error string      `json:"error,omitempty"`
}

type ConvertCurrencyRequest struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date,omitempty"`
}

type ConvertCurrencyResponse struct {
	Conversion interface{} `json:"conversion,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type GetHistoricalRatesRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type GetHistoricalRatesResponse struct {
	Rates []interface{} `json:"rates,omitempty"`
	Error string        `json:"error,omitempty"`
}

type GetSupportedCurrenciesRequest struct{}

type GetSupportedCurrenciesResponse struct {
	Currencies interface{} `json:"currencies,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// Endpoint makers
func makeGetLatestRateEndpoint(svc service.ExchangeService) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetLatestRateRequest)
		rate, err := svc.GetLatestRate(ctx, req.From, req.To)
		if err != nil {
			return GetLatestRateResponse{Error: err.Error()}, nil
		}
		return GetLatestRateResponse{Rate: rate}, nil
	}
}

func makeConvertCurrencyEndpoint(svc service.ExchangeService) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ConvertCurrencyRequest)
		cr := &models.ConversionRequest{
			FromCurrency: req.From,
			ToCurrency:   req.To,
			Amount:       req.Amount,
			Date:         req.Date,
		}
		conversion, err := svc.ConvertCurrency(ctx, cr)
		if err != nil {
			return ConvertCurrencyResponse{Error: err.Error()}, nil
		}
		return ConvertCurrencyResponse{Conversion: conversion}, nil
	}
}

func makeGetHistoricalRatesEndpoint(svc service.ExchangeService) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetHistoricalRatesRequest)
		start, err := parseDate(req.StartDate)
		if err != nil {
			return GetHistoricalRatesResponse{Error: err.Error()}, nil
		}
		end, err := parseDate(req.EndDate)
		if err != nil {
			return GetHistoricalRatesResponse{Error: err.Error()}, nil
		}
		if end.Before(start) {
			return GetHistoricalRatesResponse{Error: "end_date must be after start_date"}, nil
		}

		var rates []interface{}
		for d := start; !d.After(end); d = d.Add(24 * time.Hour) {
			rate, rerr := svc.GetHistoricalRate(ctx, req.From, req.To, d)
			if rerr != nil {
				return GetHistoricalRatesResponse{Error: rerr.Error()}, nil
			}
			rates = append(rates, rate)
		}

		return GetHistoricalRatesResponse{Rates: rates}, nil
	}
}

func makeGetSupportedCurrenciesEndpoint(svc service.ExchangeService) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		currencies, err := svc.GetSupportedCurrencies(ctx)
		if err != nil {
			return GetSupportedCurrenciesResponse{Error: err.Error()}, nil
		}
		return GetSupportedCurrenciesResponse{Currencies: currencies}, nil
	}
}

// Helper: parse multiple formats
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date string is empty")
	}
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"02-01-2006",
		"02/01/2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// Middleware (simple versions)
type EndpointMiddleware func(kitendpoint.Endpoint) kitendpoint.Endpoint

func LoggingMiddleware(logger log.Logger) EndpointMiddleware {
	return func(next kitendpoint.Endpoint) kitendpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			start := time.Now()
			resp, err := next(ctx, request)
			_ = logger.Log("took", time.Since(start))
			return resp, err
		}
	}
}

func RecoveryMiddleware(logger log.Logger) EndpointMiddleware {
	return func(next kitendpoint.Endpoint) kitendpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			defer func() {
				if r := recover(); r != nil {
					_ = logger.Log("panic", r)
				}
			}()
			return next(ctx, request)
		}
	}
}
