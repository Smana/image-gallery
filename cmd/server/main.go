package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"image-gallery/internal/config"
	"image-gallery/internal/observability"
	"image-gallery/internal/platform/database"
	"image-gallery/internal/platform/server"
	"image-gallery/internal/platform/storage"
	"image-gallery/internal/services"
	"image-gallery/internal/web/handlers"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize OpenTelemetry provider first (for tracing and metrics)
	otelConfig := observability.Config{
		ServiceName:     cfg.Observability.ServiceName,
		ServiceVersion:  cfg.Observability.ServiceVersion,
		Environment:     cfg.Environment,
		TracesEndpoint:  cfg.Observability.TracesEndpoint,
		TracesEnabled:   cfg.Observability.TracesEnabled,
		MetricsEndpoint: cfg.Observability.MetricsEndpoint,
		MetricsEnabled:  cfg.Observability.MetricsEnabled,
		LogLevel:        cfg.Logging.Level,
		LogFormat:       cfg.Logging.Format,
	}

	otelProvider, err := observability.NewProvider(context.Background(), otelConfig)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry provider: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelProvider.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down OpenTelemetry provider: %v", err)
		}
	}()
	log.Println("OpenTelemetry provider initialized")

	// Initialize structured logger with trace correlation
	logger := observability.NewLogger(otelConfig)
	logger.GetZerolog().Info().
		Str("service", cfg.Observability.ServiceName).
		Str("version", cfg.Observability.ServiceVersion).
		Str("environment", cfg.Environment).
		Msg("Starting image-gallery service")

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logger.GetZerolog().Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.GetZerolog().Error().Err(err).Msg("Error closing database connection")
		}
	}()

	// Migrations are handled by Atlas - application never runs migrations:
	// - Local development: Use `make migrate` (Atlas CLI)
	// - Kubernetes: Atlas Operator handles migrations automatically
	logger.GetZerolog().Info().Msg("Database connected - migrations are handled by Atlas")

	storageClient, err := storage.NewMinIOClient(cfg.Storage)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			logger.GetZerolog().Error().Err(closeErr).Msg("Error closing database connection")
		}
		logger.GetZerolog().Fatal().Err(err).Msg("Failed to connect to storage")
	}
	logger.GetZerolog().Info().Msg("Storage client initialized")

	// Initialize dependency injection container with observability
	container, err := services.NewContainerWithObservability(cfg, db, storageClient, logger)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			logger.GetZerolog().Error().Err(closeErr).Msg("Error closing database connection")
		}
		logger.GetZerolog().Fatal().Err(err).Msg("Failed to initialize services container")
	}
	defer func() {
		if err := container.Close(); err != nil {
			logger.GetZerolog().Error().Err(err).Msg("Error closing services container")
		}
	}()
	logger.GetZerolog().Info().Msg("Services container initialized")

	handler := handlers.NewWithContainer(container)

	srv := server.New(cfg.Port, handler.Routes())

	go func() {
		logger.GetZerolog().Info().Str("port", cfg.Port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.GetZerolog().Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.GetZerolog().Info().Msg("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Force flush telemetry before shutdown
	if flushErr := otelProvider.ForceFlush(ctx); flushErr != nil {
		logger.GetZerolog().Error().Err(flushErr).Msg("Failed to flush telemetry")
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.GetZerolog().Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.GetZerolog().Info().Msg("Server exited gracefully")
	fmt.Println("Server exited")
}
