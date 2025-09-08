package services

import (
	"database/sql"
	"log"

	"image-gallery/internal/config"
	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/storage"
	"image-gallery/internal/services/implementations"
)

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

// initializeServices initializes all services in the correct dependency order
func (c *Container) initializeServices() error {
	// Initialize repositories first
	c.imageRepository = implementations.NewImageRepository(c.db)
	c.tagRepository = implementations.NewTagRepository(c.db)
	
	// Initialize infrastructure services
	c.storageService = implementations.NewStorageService(c.storageClient)
	c.imageProcessor = implementations.NewImageProcessor()
	c.validationService = implementations.NewValidationService()
	
	// Initialize optional services (can be nil for now)
	c.eventPublisher = nil      // Will implement later
	c.cacheService = nil        // Will implement later
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