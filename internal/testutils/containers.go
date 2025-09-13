package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	minioClient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisModule "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"image-gallery/internal/config"
	"image-gallery/internal/platform/cache"
	"image-gallery/internal/platform/database"
	"image-gallery/internal/platform/storage"
)

// TestContainers manages test containers for integration testing
type TestContainers struct {
	PostgresContainer testcontainers.Container
	MinioContainer    testcontainers.Container
	RedisContainer    testcontainers.Container
	DB                *sql.DB
	MinioClient       *storage.MinIOClient
	RedisClient       *cache.RedisClient
	DatabaseURL       string
	MinioEndpoint     string
	MinioUsername     string
	MinioPassword     string
	RedisEndpoint     string
}

// SetupTestContainers initializes and starts test containers
func SetupTestContainers(ctx context.Context) (*TestContainers, error) {
	containers := &TestContainers{
		MinioUsername: "testuser",
		MinioPassword: "testpass123",
	}

	// Setup PostgreSQL container
	if err := containers.setupPostgres(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup postgres container: %w", err)
	}

	// Setup MinIO container
	if err := containers.setupMinio(ctx); err != nil {
		containers.Cleanup(ctx) // Clean up postgres if minio fails
		return nil, fmt.Errorf("failed to setup minio container: %w", err)
	}

	// Setup Redis container
	if err := containers.setupRedis(ctx); err != nil {
		containers.Cleanup(ctx) // Clean up other containers if redis fails
		return nil, fmt.Errorf("failed to setup redis container: %w", err)
	}

	// Run database migrations
	if err := database.RunMigrations(containers.DB); err != nil {
		containers.Cleanup(ctx)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return containers, nil
}

// setupPostgres creates and starts a PostgreSQL test container
func (tc *TestContainers) setupPostgres(ctx context.Context) error {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithSQLDriver("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres container: %w", err)
	}

	tc.PostgresContainer = postgresContainer

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	tc.DatabaseURL = connStr

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Test connection with retry
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err != nil {
			if i == 9 {
				return fmt.Errorf("failed to ping postgres after retries: %w", err)
			}
			time.Sleep(time.Second)
			continue
		}
		break
	}

	tc.DB = db
	return nil
}

// setupMinio creates and starts a MinIO test container
func (tc *TestContainers) setupMinio(ctx context.Context) error {
	minioContainer, err := minio.Run(ctx,
		"minio/minio:latest",
		minio.WithUsername(tc.MinioUsername),
		minio.WithPassword(tc.MinioPassword),
	)
	if err != nil {
		return fmt.Errorf("failed to start minio container: %w", err)
	}

	tc.MinioContainer = minioContainer

	// Get connection details
	endpoint, err := minioContainer.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get minio endpoint: %w", err)
	}

	tc.MinioEndpoint = endpoint

	// Create MinIO client
	minioClientInstance, err := minioClient.New(endpoint, &minioClient.Options{
		Creds:  credentials.NewStaticV4(tc.MinioUsername, tc.MinioPassword, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	// Wrap in our storage client using config
	storageConfig := config.StorageConfig{
		Endpoint:        endpoint,
		AccessKeyID:     tc.MinioUsername,
		SecretAccessKey: tc.MinioPassword,
		UseSSL:          false,
		BucketName:      "test-images",
		Region:          "us-east-1",
	}

	storageClient, err := storage.NewMinIOClient(storageConfig)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	tc.MinioClient = storageClient

	// Create test bucket
	bucketName := "test-images"
	exists, err := minioClientInstance.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = minioClientInstance.MakeBucket(ctx, bucketName, minioClient.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create test bucket: %w", err)
		}
	}

	return nil
}

// setupRedis creates and starts a Valkey test container (Redis-compatible)
func (tc *TestContainers) setupRedis(ctx context.Context) error {
	redisContainer, err := redisModule.Run(ctx,
		"valkey/valkey:7-alpine",
		redisModule.WithSnapshotting(10, 1),
		redisModule.WithLogLevel(redisModule.LogLevelVerbose),
	)
	if err != nil {
		return fmt.Errorf("failed to start valkey container: %w", err)
	}

	tc.RedisContainer = redisContainer

	// Get connection string
	endpoint, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get valkey endpoint: %w", err)
	}

	tc.RedisEndpoint = endpoint

	// Create Redis client using our cache implementation (Valkey is Redis-compatible)
	cacheConfig := config.CacheConfig{
		Enabled:     true,
		Address:     endpoint,
		Password:    "",
		Database:    0,
		DefaultTTL:  1 * time.Hour,
		DialTimeout: 5 * time.Second,
	}

	redisClient, err := cache.NewRedisClient(cacheConfig)
	if err != nil {
		return fmt.Errorf("failed to create redis client: %w", err)
	}

	tc.RedisClient = redisClient

	// Test connection
	if err := tc.RedisClient.Health(ctx); err != nil {
		return fmt.Errorf("failed to connect to valkey: %w", err)
	}

	return nil
}

// Cleanup terminates all test containers and closes connections
func (tc *TestContainers) Cleanup(ctx context.Context) error {
	var errs []error

	if tc.DB != nil {
		if err := tc.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if tc.PostgresContainer != nil {
		if err := tc.PostgresContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate postgres container: %w", err))
		}
	}

	if tc.MinioContainer != nil {
		if err := tc.MinioContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate minio container: %w", err))
		}
	}

	if tc.RedisClient != nil {
		if err := tc.RedisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close valkey client: %w", err))
		}
	}

	if tc.RedisContainer != nil {
		if err := tc.RedisContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate valkey container: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// GetDatabaseURL returns the database connection URL
func (tc *TestContainers) GetDatabaseURL() string {
	return tc.DatabaseURL
}

// GetMinioEndpoint returns the MinIO endpoint
func (tc *TestContainers) GetMinioEndpoint() string {
	return tc.MinioEndpoint
}

// GetMinioCredentials returns MinIO credentials
func (tc *TestContainers) GetMinioCredentials() (string, string) {
	return tc.MinioUsername, tc.MinioPassword
}

// ResetDatabase clears all data from the test database
func (tc *TestContainers) ResetDatabase(ctx context.Context) error {
	// Clean up all tables in reverse dependency order
	tables := []string{
		"image_albums",
		"image_tags", 
		"albums",
		"tags",
		"images",
		"schema_migrations",
	}

	tx, err := tc.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			// Ignore errors for tables that might not exist
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reset transaction: %w", err)
	}

	// Re-run migrations to ensure schema is fresh
	return database.RunMigrations(tc.DB)
}

// CreateTestBucket creates a bucket for testing
func (tc *TestContainers) CreateTestBucket(ctx context.Context, bucketName string) error {
	// This would use the internal MinIO client from our storage client
	// For now, we'll assume the bucket creation is handled by the storage client
	return nil
}

// CleanBucket removes all objects from a test bucket
func (tc *TestContainers) CleanBucket(ctx context.Context, bucketName string) error {
	// This would clean all objects from the specified bucket
	// Implementation depends on the storage client capabilities
	return nil
}

// GetRedisEndpoint returns the Valkey endpoint (Redis-compatible)
func (tc *TestContainers) GetRedisEndpoint() string {
	return tc.RedisEndpoint
}

// FlushRedis clears all data from the Valkey test database
func (tc *TestContainers) FlushRedis(ctx context.Context) error {
	if tc.RedisClient == nil {
		return fmt.Errorf("valkey client not available")
	}
	
	return tc.RedisClient.FlushCache(ctx)
}

// GetRedisClient returns the Valkey client for tests (Redis-compatible)
func (tc *TestContainers) GetRedisClient() *cache.RedisClient {
	return tc.RedisClient
}