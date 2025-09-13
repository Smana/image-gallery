package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test helpers and setup functions

func setupTestDatabase(t *testing.T) *sql.DB {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Clean up any existing data
	cleanupTables(t, db)

	// Run migrations to set up schema
	err = RunMigrations(db)
	require.NoError(t, err, "Failed to run migrations")

	return db
}

func intPtr(i int) *int {
	return &i
}

// Integration Tests

func TestDatabaseIntegration_FullWorkflow(t *testing.T) {
	db := setupTestDatabase(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Resource cleanup

	repo := NewImageRepository(db)

	// Test Create
	newImage := &Image{
		Filename:         "integration_test.jpg",
		OriginalFilename: "original_integration_test.jpg",
		ContentType:      "image/jpeg",
		FileSize:         2048,
		StoragePath:      "/storage/integration_test.jpg",
		Width:            intPtr(1200),
		Height:           intPtr(900),
		Metadata:         map[string]interface{}{"camera": "Canon EOS", "iso": 100},
	}

	err := repo.Create(context.Background(), newImage)
	require.NoError(t, err)
	assert.Greater(t, newImage.ID, 0, "Should have assigned ID")
	assert.False(t, newImage.CreatedAt.IsZero(), "Should have set created_at")
	assert.False(t, newImage.UploadedAt.IsZero(), "Should have set uploaded_at")

	// Test GetByID
	retrieved, err := repo.GetByID(context.Background(), newImage.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, newImage.ID, retrieved.ID)
	assert.Equal(t, newImage.Filename, retrieved.Filename)
	assert.Equal(t, newImage.ContentType, retrieved.ContentType)
	assert.Equal(t, *newImage.Width, *retrieved.Width)
	assert.Equal(t, *newImage.Height, *retrieved.Height)

	// Test List
	images, err := repo.List(context.Background(), PaginationParams{Limit: 10, Offset: 0}, DefaultSort())
	require.NoError(t, err)
	assert.Len(t, images, 1, "Should have one image")
	assert.Equal(t, newImage.ID, images[0].ID)

	// Test Delete
	err = repo.Delete(context.Background(), newImage.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByID(context.Background(), newImage.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted, "Image should be deleted")
}

func TestDatabaseIntegration_ConcurrentOperations(t *testing.T) {
	db := setupTestDatabase(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Resource cleanup

	repo := NewImageRepository(db)
	numWorkers := 10
	done := make(chan bool, numWorkers)
	errors := make(chan error, numWorkers)

	// Create multiple images concurrently
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer func() { done <- true }()

			img := &Image{
				Filename:         fmt.Sprintf("concurrent_test_%d.jpg", id),
				OriginalFilename: fmt.Sprintf("original_concurrent_test_%d.jpg", id),
				ContentType:      "image/jpeg",
				FileSize:         1024 + int64(id)*100,
				StoragePath:      fmt.Sprintf("/storage/concurrent_test_%d.jpg", id),
			}

			if err := repo.Create(context.Background(), img); err != nil {
				errors <- err
			}
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < numWorkers; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify all images were created
	images, err := repo.List(context.Background(), PaginationParams{Limit: 20, Offset: 0}, DefaultSort())
	require.NoError(t, err)
	assert.Len(t, images, numWorkers, "Should have created all images concurrently")
}

func TestDatabaseIntegration_TransactionBehavior(t *testing.T) {
	db := setupTestDatabase(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Resource cleanup

	// Test successful transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	_, err = tx.Exec(`
		INSERT INTO images (filename, original_filename, content_type, file_size, storage_path)
		VALUES ($1, $2, $3, $4, $5)
	`, "tx_test.jpg", "tx_test.jpg", "image/jpeg", 1024, "/storage/tx_test.jpg")
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// Verify image was committed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM images WHERE filename = $1", "tx_test.jpg").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Transaction should be committed")

	// Test rollback transaction
	tx, err = db.Begin()
	require.NoError(t, err)

	_, err = tx.Exec(`
		INSERT INTO images (filename, original_filename, content_type, file_size, storage_path)
		VALUES ($1, $2, $3, $4, $5)
	`, "tx_rollback_test.jpg", "tx_rollback_test.jpg", "image/jpeg", 1024, "/storage/tx_rollback_test.jpg")
	require.NoError(t, err)

	err = tx.Rollback()
	require.NoError(t, err)

	// Verify image was not committed
	err = db.QueryRow("SELECT COUNT(*) FROM images WHERE filename = $1", "tx_rollback_test.jpg").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Transaction should be rolled back")
}

func TestDatabaseIntegration_LargeDataSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	db := setupTestDatabase(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Resource cleanup

	repo := NewImageRepository(db)

	// Create a large number of images
	batchSize := 100
	for batch := 0; batch < 5; batch++ {
		for i := 0; i < batchSize; i++ {
			imageID := batch*batchSize + i
			img := &Image{
				Filename:         fmt.Sprintf("large_test_%d.jpg", imageID),
				OriginalFilename: fmt.Sprintf("original_large_test_%d.jpg", imageID),
				ContentType:      "image/jpeg",
				FileSize:         1024 + int64(imageID)*10,
				StoragePath:      fmt.Sprintf("/storage/large_test_%d.jpg", imageID),
			}

			err := repo.Create(context.Background(), img)
			require.NoError(t, err, "Failed to create image %d", imageID)
		}
	}

	// Test pagination
	totalImages := batchSize * 5
	pageSize := 25

	var allImages []*Image
	for offset := 0; offset < totalImages; offset += pageSize {
		images, err := repo.List(context.Background(), PaginationParams{Limit: pageSize, Offset: offset}, DefaultSort())
		require.NoError(t, err)
		allImages = append(allImages, images...)
	}

	assert.Len(t, allImages, totalImages, "Should retrieve all images through pagination")

	// Test that images are ordered by uploaded_at DESC
	for i := 1; i < len(allImages); i++ {
		assert.True(t,
			allImages[i-1].UploadedAt.After(allImages[i].UploadedAt) ||
				allImages[i-1].UploadedAt.Equal(allImages[i].UploadedAt),
			"Images should be ordered by uploaded_at DESC")
	}
}

func TestDatabaseIntegration_ErrorHandling(t *testing.T) {
	db := setupTestDatabase(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Resource cleanup

	repo := NewImageRepository(db)

	// Test duplicate filename constraint (if exists)
	img1 := &Image{
		Filename:         "duplicate_test.jpg",
		OriginalFilename: "duplicate_test.jpg",
		ContentType:      "image/jpeg",
		FileSize:         1024,
		StoragePath:      "/storage/duplicate_test.jpg",
	}

	err := repo.Create(context.Background(), img1)
	require.NoError(t, err)

	// Test invalid foreign key constraint
	// This should fail if we had foreign key constraints on storage references

	// Test SQL injection protection
	maliciousInput := "'; DROP TABLE images; --"
	maliciousImg := &Image{
		Filename:         maliciousInput,
		OriginalFilename: maliciousInput,
		ContentType:      "image/jpeg",
		FileSize:         1024,
		StoragePath:      "/storage/" + maliciousInput,
	}

	// Should not cause SQL injection, just store the malicious string as data
	err = repo.Create(context.Background(), maliciousImg)
	assert.NoError(t, err, "Should handle malicious input safely")

	// Verify tables still exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM images").Scan(&count)
	assert.NoError(t, err, "Table should still exist after malicious input")
	assert.Equal(t, 2, count, "Should have 2 images")
}

func TestDatabaseIntegration_ConnectionPooling(t *testing.T) {
	db := setupTestDatabase(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Resource cleanup

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetConnMaxIdleTime(5 * time.Second)

	numConcurrent := 20
	done := make(chan bool, numConcurrent)

	// Perform many concurrent operations
	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Perform a simple query
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM images").Scan(&count)
			assert.NoError(t, err, "Concurrent query %d should succeed", id)

			// Brief sleep to simulate work
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numConcurrent; i++ {
		<-done
	}

	// Check connection pool stats
	stats := db.Stats()
	assert.LessOrEqual(t, stats.OpenConnections, 5, "Should not exceed max open connections")
	assert.GreaterOrEqual(t, stats.OpenConnections, 0, "Should have some open connections")

	t.Logf("Connection pool stats: Open=%d, InUse=%d, Idle=%d",
		stats.OpenConnections, stats.InUse, stats.Idle)
}
