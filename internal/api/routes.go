package api

import (
	"net/http"

	"exchange-rate-service/internal/transport"

	"github.com/gorilla/mux"
)

// NewRouter creates a new HTTP router with all routes
func NewRouter(handlers *Handlers) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	// Health check
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// API v1 routes
	v1 := router.PathPrefix("/api/v1").Subrouter()

	// Build go-kit endpoints
	eps := transport.MakeEndpoints(handlers.exchangeService, handlers.logger)

	// Currency routes
	v1.Handle("/currencies", transport.NewGetSupportedCurrenciesHTTPHandler(eps.GetSupportedCurrenciesEndpoint, handlers.logger)).Methods("GET")
	v1.HandleFunc("/rates", handlers.GetRates).Methods("GET")

	// Exchange rate routes
	v1.Handle("/rates/{base}/{target}", transport.NewGetLatestRateHTTPHandler(eps.GetLatestRateEndpoint, handlers.logger)).Methods("GET")
	// Historical single-date remains via handler (since free tier not supported)
	v1.HandleFunc("/rates/{base}/{target}/{date}", handlers.GetHistoricalRate).Methods("GET")

	// Conversion routes
	v1.Handle("/convert", transport.NewConvertCurrencyHTTPHandler(eps.ConvertCurrencyEndpoint, handlers.logger)).Methods("POST")

	// Time series routes (range) via go-kit endpoint
	v1.Handle("/timeseries/{base}/{target}", transport.NewGetHistoricalRatesHTTPHandler(eps.GetHistoricalRatesEndpoint, handlers.logger)).Methods("GET")

	// Documentation
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Exchange Rate Service API</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 20px 0; padding: 15px; border-left: 4px solid #007cba; background: #f8f9fa; }
        .method { font-weight: bold; color: #007cba; }
        .url { font-family: monospace; background: #e9ecef; padding: 5px; }
        .description { margin-top: 10px; color: #6c757d; }
    </style>
</head>
<body>
    <h1>Exchange Rate Service API</h1>
    <p>Welcome to the Exchange Rate Service. This service provides real-time and historical exchange rates from multiple providers.</p>
    
    <h2>Available Endpoints</h2>
    
    <div class="endpoint">
        <div class="method">GET</div>
        <div class="url">/health</div>
        <div class="description">Health check endpoint to verify service status</div>
    </div>
    
    <div class="endpoint">
        <div class="method">GET</div>
        <div class="url">/api/v1/currencies</div>
        <div class="description">Get list of supported currencies</div>
    </div>
    
    <div class="endpoint">
        <div class="method">GET</div>
        <div class="url">/api/v1/rates?base=USD</div>
        <div class="description">Get exchange rates for all currencies relative to base currency</div>
    </div>
    
    <div class="endpoint">
        <div class="method">GET</div>
        <div class="url">/api/v1/rates/{base}/{target}</div>
        <div class="description">Get latest exchange rate between two currencies</div>
    </div>
    
    <div class="endpoint">
        <div class="method">GET</div>
        <div class="url">/api/v1/rates/{base}/{target}/{date}</div>
        <div class="description">Get historical exchange rate for a specific date (YYYY-MM-DD)</div>
    </div>
    
    <div class="endpoint">
        <div class="method">POST</div>
        <div class="url">/api/v1/convert</div>
        <div class="description">Convert amount from one currency to another</div>
    </div>
    
    <div class="endpoint">
        <div class="method">GET</div>
        <div class="url">/api/v1/timeseries/{base}/{target}?start_date=2024-01-01&end_date=2024-01-31</div>
        <div class="description">Get exchange rates for a date range</div>
    </div>
    
    <h2>Example Usage</h2>
    <p><strong>Get USD to EUR rate:</strong> <code>GET /api/v1/rates/USD/EUR</code></p>
    <p><strong>Convert 100 USD to EUR:</strong> <code>POST /api/v1/convert</code> with body: <code>{"from_currency": "USD", "to_currency": "EUR", "amount": 100}</code></p>
    <p><strong>Get historical rate:</strong> <code>GET /api/v1/rates/USD/EUR/2024-01-15</code></p>
    
    <h2>Providers</h2>
    <p>This service aggregates data from multiple exchange rate providers:</p>
    <ul>
        <li>open.er-api.com (primary)</li>
    </ul>
    
    <h2>Response Format</h2>
    <p>All responses follow this format:</p>
    <pre>{
  "success": true,
  "data": {...},
  "message": "Success message",
  "timestamp": "2024-01-01T00:00:00Z"
}</pre>
</body>
</html>
        `))
	}).Methods("GET")

	return router
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request details
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
