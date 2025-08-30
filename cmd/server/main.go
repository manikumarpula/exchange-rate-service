package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"exchange-rate-service/configs"
	"exchange-rate-service/internal/api"
	"exchange-rate-service/internal/repository"
	"exchange-rate-service/internal/service"
	"exchange-rate-service/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := utils.NewLogger()

	// Initialize repositories
	rateRepo := repository.NewRateRepository(cfg, logger)

	// Initialize service layer
	exchangeService := service.NewExchangeService(rateRepo, logger)

	// Initialize HTTP handlers
	handlers := api.NewHandlers(exchangeService, logger)

	// Setup routes
	router := api.NewRouter(handlers)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Log("msg", "Starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log("err", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log("msg", "Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log("err", "Server forced to shutdown", "error", err)
	}

	logger.Log("msg", "Server exited")
}
