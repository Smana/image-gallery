package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"image-gallery/internal/domain/image"
)

func TestImageRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given: A mock database and repository
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()

		testImage := &Image{
			Filename:         "test-image.jpg",
			OriginalFilename: "original.jpg",
			ContentType:      "image/jpeg",
			FileSize:         1024,
			StoragePath:      "images/test-image.jpg",
			Width:            intPtr(800),
			Height:           intPtr(600),
			Metadata:         Metadata{},
		}

		// Expect the INSERT query
		mock.ExpectQuery(`INSERT INTO images`).
			WithArgs(
				testImage.Filename,
				testImage.OriginalFilename,
				testImage.ContentType,
				testImage.FileSize,
				testImage.StoragePath,
				testImage.Width,
				testImage.Height,
				sqlmock.AnyArg(), // metadata JSON
				sqlmock.AnyArg(), // uploaded_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(1, time.Now(), time.Now()))

		// When: Creating an image
		err = repo.Create(ctx, testImage)

		// Then: Image should be created successfully
		assert.NoError(t, err)
		assert.NotZero(t, testImage.ID)
		assert.NotZero(t, testImage.CreatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		// Given: A mock database that returns an error
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()

		testImage := &Image{
			Filename:         "test-image.jpg",
			OriginalFilename: "original.jpg",
			ContentType:      "image/jpeg",
			FileSize:         1024,
			StoragePath:      "images/test-image.jpg",
		}

		// Expect the INSERT query to fail
		mock.ExpectQuery(`INSERT INTO images`).
			WillReturnError(sql.ErrConnDone)

		// When: Creating an image
		err = repo.Create(ctx, testImage)

		// Then: Should return error
		assert.Error(t, err)
		assert.Zero(t, testImage.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestImageRepository_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given: A mock database and repository
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()
		imageID := 1

		now := time.Now()
		
		// Expect the SELECT query
		mock.ExpectQuery(`SELECT (.+) FROM images WHERE id = \$1`).
			WithArgs(imageID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "filename", "original_filename", "content_type", "file_size",
				"storage_path", "thumbnail_path", "width", "height", "uploaded_at",
				"metadata", "created_at", "updated_at",
			}).AddRow(
				1, "test-image.jpg", "original.jpg", "image/jpeg", 1024,
				"images/test-image.jpg", sql.NullString{}, 800, 600, now,
				`{}`, now, now,
			))

		// When: Getting an image by ID
		result, err := repo.GetByID(ctx, imageID)

		// Then: Should return the image
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, imageID, result.ID)
		assert.Equal(t, "test-image.jpg", result.Filename)
		assert.Equal(t, "original.jpg", result.OriginalFilename)
		assert.Equal(t, "image/jpeg", result.ContentType)
		assert.Equal(t, int64(1024), result.FileSize)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		// Given: A mock database that returns no rows
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()
		imageID := 999

		// Expect the SELECT query to return no rows
		mock.ExpectQuery(`SELECT (.+) FROM images WHERE id = \$1`).
			WithArgs(imageID).
			WillReturnError(sql.ErrNoRows)

		// When: Getting a non-existent image
		result, err := repo.GetByID(ctx, imageID)

		// Then: Should return not found error
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestImageRepository_List(t *testing.T) {
	t.Run("Success_WithoutTag", func(t *testing.T) {
		// Given: A mock database and repository
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()

		req := &image.ListImagesRequest{
			Page:     1,
			PageSize: 10,
		}

		now := time.Now()

		// Expect the count query
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM images`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Expect the SELECT query
		mock.ExpectQuery(`SELECT (.+) FROM images ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
			WithArgs(req.PageSize, (req.Page-1)*req.PageSize).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "filename", "original_filename", "content_type", "file_size",
				"storage_path", "thumbnail_path", "width", "height", "uploaded_at",
				"metadata", "created_at", "updated_at",
			}).
				AddRow(1, "image1.jpg", "orig1.jpg", "image/jpeg", 1024, "path1", sql.NullString{}, 800, 600, now, `{}`, now, now).
				AddRow(2, "image2.jpg", "orig2.jpg", "image/png", 2048, "path2", sql.NullString{}, 1024, 768, now, `{}`, now, now))

		// When: Listing images
		response, err := repo.List(ctx, req)

		// Then: Should return paginated results
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Images, 2)
		assert.Equal(t, 2, response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		assert.Equal(t, 1, response.TotalPages)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success_WithTag", func(t *testing.T) {
		// Given: A mock database and repository with tag filter
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()

		req := &image.ListImagesRequest{
			Page:     1,
			PageSize: 10,
			Tag:      "nature",
		}

		now := time.Now()

		// Expect the count query with tag filter
		mock.ExpectQuery(`SELECT COUNT\(DISTINCT i\.id\) FROM images i JOIN image_tags it ON i\.id = it\.image_id JOIN tags t ON it\.tag_id = t\.id WHERE t\.name = \$1`).
			WithArgs("nature").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Expect the SELECT query with tag filter
		mock.ExpectQuery(`SELECT DISTINCT (.+) FROM images i JOIN image_tags it ON i\.id = it\.image_id JOIN tags t ON it\.tag_id = t\.id WHERE t\.name = \$1 ORDER BY i\.created_at DESC LIMIT \$2 OFFSET \$3`).
			WithArgs("nature", req.PageSize, (req.Page-1)*req.PageSize).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "filename", "original_filename", "content_type", "file_size",
				"storage_path", "thumbnail_path", "width", "height", "uploaded_at",
				"metadata", "created_at", "updated_at",
			}).AddRow(1, "nature1.jpg", "nature.jpg", "image/jpeg", 1024, "path1", sql.NullString{}, 800, 600, now, `{}`, now, now))

		// When: Listing images with tag filter
		response, err := repo.List(ctx, req)

		// Then: Should return filtered results
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Images, 1)
		assert.Equal(t, 1, response.TotalCount)
		assert.Equal(t, 1, response.TotalPages)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestImageRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given: A mock database and repository
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()
		imageID := 1

		// Expect transaction
		mock.ExpectBegin()
		// Expect delete from image_tags first (foreign key constraint)
		mock.ExpectExec(`DELETE FROM image_tags WHERE image_id = \$1`).
			WithArgs(imageID).
			WillReturnResult(sqlmock.NewResult(0, 2)) // 2 tags deleted
		// Expect delete from images
		mock.ExpectExec(`DELETE FROM images WHERE id = \$1`).
			WithArgs(imageID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 image deleted
		mock.ExpectCommit()

		// When: Deleting an image
		err = repo.Delete(ctx, imageID)

		// Then: Should delete successfully
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		// Given: A mock database that returns no affected rows
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewImageRepository(db)
		ctx := context.Background()
		imageID := 999

		// Expect transaction
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM image_tags WHERE image_id = \$1`).
			WithArgs(imageID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`DELETE FROM images WHERE id = \$1`).
			WithArgs(imageID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected
		mock.ExpectRollback()

		// When: Deleting a non-existent image
		err = repo.Delete(ctx, imageID)

		// Then: Should return not found error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Helper function
func intPtr(i int) *int {
	return &i
}

// Test helper to verify the AnyValue matcher works correctly
type anyTime struct{}

func (a anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}