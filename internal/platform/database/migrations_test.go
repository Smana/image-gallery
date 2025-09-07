package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations(t *testing.T) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up any existing tables
	cleanupTables(t, db)

	// Run migrations
	err = RunMigrations(db)
	assert.NoError(t, err, "RunMigrations should succeed")

	// Verify tables were created
	verifyImagesTable(t, db)
	verifyTagsTable(t, db)
	verifyImageTagsTable(t, db)
	verifyIndexes(t, db)
}

func TestRunMigrationsIdempotent(t *testing.T) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up any existing tables
	cleanupTables(t, db)

	// Run migrations twice - should not fail
	err = RunMigrations(db)
	require.NoError(t, err, "First migration should succeed")

	err = RunMigrations(db)
	assert.NoError(t, err, "Second migration should also succeed (idempotent)")
}

func TestRunMigrationsWithInvalidDatabase(t *testing.T) {
	// Create a closed database connection to simulate failure
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err)
	db.Close() // Close immediately to make it invalid

	// Should fail when trying to run migrations on closed connection
	err = RunMigrations(db)
	assert.Error(t, err, "RunMigrations should fail with closed database")
}

func TestMigrationQueries(t *testing.T) {
	// Test that migration SQL is valid
	migrations := []string{
		createImagesTable,
		createTagsTable,
	}

	for i, migration := range migrations {
		t.Run(fmt.Sprintf("migration_%d", i+1), func(t *testing.T) {
			assert.NotEmpty(t, migration, "Migration should not be empty")
			assert.Contains(t, migration, "CREATE TABLE", "Migration should contain CREATE TABLE")
			assert.Contains(t, migration, "IF NOT EXISTS", "Migration should be idempotent")
		})
	}
}

func TestImagesTableSchema(t *testing.T) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err)
	defer db.Close()

	cleanupTables(t, db)
	err = RunMigrations(db)
	require.NoError(t, err)

	// Test inserting valid data
	insertSQL := `
		INSERT INTO images (
			filename, original_filename, content_type, file_size, 
			storage_path, width, height, metadata
		) VALUES (
			'test.jpg', 'original_test.jpg', 'image/jpeg', 1024,
			'/storage/test.jpg', 800, 600, '{"description": "test image"}'
		) RETURNING id
	`

	var imageID int
	err = db.QueryRow(insertSQL).Scan(&imageID)
	assert.NoError(t, err, "Should be able to insert valid image data")
	assert.Greater(t, imageID, 0, "Should return valid image ID")

	// Test that timestamps are set automatically
	var createdAt, updatedAt, uploadedAt sql.NullTime
	err = db.QueryRow("SELECT created_at, updated_at, uploaded_at FROM images WHERE id = $1", imageID).
		Scan(&createdAt, &updatedAt, &uploadedAt)
	assert.NoError(t, err)
	assert.True(t, createdAt.Valid, "created_at should be set")
	assert.True(t, updatedAt.Valid, "updated_at should be set")
	assert.True(t, uploadedAt.Valid, "uploaded_at should be set")
}

func TestTagsAndImageTagsSchema(t *testing.T) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err)
	defer db.Close()

	cleanupTables(t, db)
	err = RunMigrations(db)
	require.NoError(t, err)

	// Insert a test image
	var imageID int
	err = db.QueryRow(`
		INSERT INTO images (filename, original_filename, content_type, file_size, storage_path)
		VALUES ('test.jpg', 'test.jpg', 'image/jpeg', 1024, '/storage/test.jpg')
		RETURNING id
	`).Scan(&imageID)
	require.NoError(t, err)

	// Insert tags
	var tagID1, tagID2 int
	err = db.QueryRow("INSERT INTO tags (name) VALUES ($1) RETURNING id", "nature").Scan(&tagID1)
	require.NoError(t, err)

	err = db.QueryRow("INSERT INTO tags (name) VALUES ($1) RETURNING id", "landscape").Scan(&tagID2)
	require.NoError(t, err)

	// Test unique constraint on tag names
	_, err = db.Exec("INSERT INTO tags (name) VALUES ($1)", "nature")
	assert.Error(t, err, "Should not allow duplicate tag names")

	// Link image with tags
	_, err = db.Exec("INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2)", imageID, tagID1)
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2)", imageID, tagID2)
	assert.NoError(t, err)

	// Test cascade delete - deleting image should remove image_tags
	_, err = db.Exec("DELETE FROM images WHERE id = $1", imageID)
	assert.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM image_tags WHERE image_id = $1", imageID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "image_tags should be deleted when image is deleted")
}

// Helper functions

func cleanupTables(t *testing.T, db *sql.DB) {
	tables := []string{"image_tags", "images", "tags"}
	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		require.NoError(t, err, "Failed to drop table %s", table)
	}
}

func verifyImagesTable(t *testing.T, db *sql.DB) {
	// Check if table exists and has correct columns
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'images' AND table_schema = 'public'
		ORDER BY ordinal_position
	`

	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	expectedColumns := []string{
		"id", "filename", "original_filename", "content_type", "file_size",
		"storage_path", "thumbnail_path", "width", "height", "uploaded_at",
		"metadata", "created_at", "updated_at",
	}

	var columns []string
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString
		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		require.NoError(t, err)
		columns = append(columns, columnName)
	}

	for _, expected := range expectedColumns {
		assert.Contains(t, columns, expected, "Table should have column %s", expected)
	}
}

func verifyTagsTable(t *testing.T, db *sql.DB) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'tags' AND table_schema = 'public'
		)
	`).Scan(&exists)

	require.NoError(t, err)
	assert.True(t, exists, "tags table should exist")
}

func verifyImageTagsTable(t *testing.T, db *sql.DB) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'image_tags' AND table_schema = 'public'
		)
	`).Scan(&exists)

	require.NoError(t, err)
	assert.True(t, exists, "image_tags table should exist")
}

func verifyIndexes(t *testing.T, db *sql.DB) {
	expectedIndexes := []string{
		"idx_images_filename",
		"idx_images_content_type",
		"idx_images_uploaded_at",
	}

	query := `
		SELECT indexname FROM pg_indexes 
		WHERE tablename = 'images' AND schemaname = 'public'
	`

	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	var indexes []string
	for rows.Next() {
		var indexName string
		err := rows.Scan(&indexName)
		require.NoError(t, err)
		indexes = append(indexes, indexName)
	}

	for _, expected := range expectedIndexes {
		assert.Contains(t, indexes, expected, "Should have index %s", expected)
	}
}
