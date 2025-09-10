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

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	storageClient, err := storage.NewMinIOClient(cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to connect to storage: %v", err)
	}

	// Initialize dependency injection container
	container, err := services.NewContainer(cfg, db, storageClient)
	if err != nil {
		log.Fatalf("Failed to initialize services container: %v", err)
	}
	defer container.Close()

	handler := handlers.NewWithContainer(container)

	srv := server.New(cfg.Port, handler.Routes())

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}
