package transport

import (
	"context"
	"encoding/json"
	"net/http"

	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
)

func NewGetLatestRateHTTPHandler(ep kitendpoint.Endpoint, logger log.Logger) http.Handler {
	return kithttp.NewServer(ep, decodeGetLatestRateRequest, encodeResponse)
}

func NewConvertCurrencyHTTPHandler(ep kitendpoint.Endpoint, logger log.Logger) http.Handler {
	return kithttp.NewServer(ep, decodeConvertCurrencyRequest, encodeResponse)
}

func NewGetHistoricalRatesHTTPHandler(ep kitendpoint.Endpoint, logger log.Logger) http.Handler {
	return kithttp.NewServer(ep, decodeGetHistoricalRatesRequest, encodeResponse)
}

func NewGetSupportedCurrenciesHTTPHandler(ep kitendpoint.Endpoint, logger log.Logger) http.Handler {
	return kithttp.NewServer(ep, decodeGetSupportedCurrenciesRequest, encodeResponse)
}

// decoders
func decodeGetLatestRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	return GetLatestRateRequest{From: vars["base"], To: vars["target"]}, nil
}

func decodeConvertCurrencyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req ConvertCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeGetHistoricalRatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	q := r.URL.Query()
	return GetHistoricalRatesRequest{
		From:      vars["base"],
		To:        vars["target"],
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
	}, nil
}

func decodeGetSupportedCurrenciesRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return GetSupportedCurrenciesRequest{}, nil
}

// encoder
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}
