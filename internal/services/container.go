package services

import (
	"database/sql"
	"log"

	"image-gallery/internal/config"
	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/cache"
	"image-gallery/internal/platform/database"
	"image-gallery/internal/platform/storage"
	"image-gallery/internal/services/implementations"
)

// TestConfig provides test-specific configuration for integration testing
type TestConfig struct {
	DatabaseURL string
}

// Container holds all the application dependencies
type Container struct {
	config *config.Config
	db     *sql.DB
	
	// Storage
	storageClient *storage.MinIOClient
	storageService image.StorageService
	
	// Repositories
	imageRepository image.Repository
	tagRepository   image.TagRepository
	
	// Services
	imageService      image.ImageService
	tagService        image.TagService
	imageProcessor    image.ImageProcessor
	validationService image.ValidationService
	
	// Infrastructure services (optional - can be nil for now)
	eventPublisher      image.EventPublisher
	cacheService       image.CacheService
	searchService      image.SearchService
	auditService       image.AuditService
	notificationService image.NotificationService
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config, db *sql.DB, storageClient *storage.MinIOClient) (*Container, error) {
	container := &Container{
		config:        cfg,
		db:           db,
		storageClient: storageClient,
	}
	
	if err := container.initializeServices(); err != nil {
		return nil, err
	}
	
	return container, nil
}

// NewContainerForTest creates a new dependency injection container for testing
func NewContainerForTest(testCfg *TestConfig, db *sql.DB, storageClient *storage.MinIOClient) (*Container, error) {
	// Create a minimal config for testing
	cfg := &config.Config{
		Environment: "test",
		DatabaseURL: testCfg.DatabaseURL,
		Storage: config.StorageConfig{
			BucketName: "test-images",
		},
	}
	
	return NewContainer(cfg, db, storageClient)
}

// initializeServices initializes all services in the correct dependency order
func (c *Container) initializeServices() error {
	// Initialize database repositories first
	dbImageRepo := database.NewImageRepository(c.db)
	
	// Initialize domain repository adapters
	c.imageRepository = implementations.NewImageRepositoryAdapter(dbImageRepo)
	c.tagRepository = implementations.NewTagRepository(c.db)
	
	// Initialize infrastructure services
	// Try to create full storage service, fallback to MinIOClient wrapper
	storageConfig := &c.config.Storage
	if fullStorageService, err := storage.NewService(storageConfig); err == nil {
		c.storageService = implementations.NewStorageServiceWithService(fullStorageService)
	} else {
		c.storageService = implementations.NewStorageService(c.storageClient)
	}
	c.imageProcessor = implementations.NewImageProcessor()
	c.validationService = implementations.NewValidationService()
	
	// Initialize cache service (optional)
	if c.config.Cache.Enabled {
		if redisClient, err := cache.NewRedisClient(c.config.Cache); err == nil {
			c.cacheService = implementations.NewCacheService(redisClient)
			log.Println("Cache service initialized with Valkey/Redis")
		} else {
			log.Printf("Failed to initialize cache service: %v", err)
			c.cacheService = nil
		}
	} else {
		log.Println("Cache service disabled")
		c.cacheService = nil
	}
	
	// Initialize optional services (can be nil for now)
	c.eventPublisher = nil      // Will implement later
	c.searchService = nil       // Will implement later  
	c.auditService = nil        // Will implement later
	c.notificationService = nil // Will implement later
	
	// Initialize domain services
	c.imageService = implementations.NewImageService(
		c.imageRepository,
		c.tagRepository,
		c.storageService,
		c.imageProcessor,
		c.validationService,
		c.eventPublisher,
		c.cacheService,
	)
	
	c.tagService = implementations.NewTagService(
		c.tagRepository,
		c.validationService,
		c.eventPublisher,
	)
	
	log.Println("Dependency injection container initialized successfully")
	return nil
}

// Getters for accessing services

func (c *Container) Config() *config.Config {
	return c.config
}

func (c *Container) DB() *sql.DB {
	return c.db
}

func (c *Container) StorageClient() *storage.MinIOClient {
	return c.storageClient
}

func (c *Container) StorageService() image.StorageService {
	return c.storageService
}

func (c *Container) ImageRepository() image.Repository {
	return c.imageRepository
}

func (c *Container) TagRepository() image.TagRepository {
	return c.tagRepository
}

func (c *Container) ImageService() image.ImageService {
	return c.imageService
}

func (c *Container) TagService() image.TagService {
	return c.tagService
}

func (c *Container) ImageProcessor() image.ImageProcessor {
	return c.imageProcessor
}

func (c *Container) ValidationService() image.ValidationService {
	return c.validationService
}

func (c *Container) EventPublisher() image.EventPublisher {
	return c.eventPublisher
}

func (c *Container) CacheService() image.CacheService {
	return c.cacheService
}

func (c *Container) SearchService() image.SearchService {
	return c.searchService
}

func (c *Container) AuditService() image.AuditService {
	return c.auditService
}

func (c *Container) NotificationService() image.NotificationService {
	return c.notificationService
}

// Close cleans up resources
func (c *Container) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}