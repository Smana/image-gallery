package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repository implementation for testing database operations
type ImageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

type Image struct {
	ID               int       `json:"id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	ContentType      string    `json:"content_type"`
	FileSize         int64     `json:"file_size"`
	StoragePath      string    `json:"storage_path"`
	ThumbnailPath    *string   `json:"thumbnail_path,omitempty"`
	Width            *int      `json:"width,omitempty"`
	Height           *int      `json:"height,omitempty"`
	UploadedAt       time.Time `json:"uploaded_at"`
	Metadata         []byte    `json:"metadata,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (r *ImageRepository) Create(img *Image) error {
	query := `
		INSERT INTO images (
			filename, original_filename, content_type, file_size, 
			storage_path, thumbnail_path, width, height, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, uploaded_at, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		img.Filename, img.OriginalFilename, img.ContentType, img.FileSize,
		img.StoragePath, img.ThumbnailPath, img.Width, img.Height, img.Metadata,
	).Scan(&img.ID, &img.UploadedAt, &img.CreatedAt, &img.UpdatedAt)
}

func (r *ImageRepository) GetByID(id int) (*Image, error) {
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images WHERE id = $1
	`

	img := &Image{}
	err := r.db.QueryRow(query, id).Scan(
		&img.ID, &img.Filename, &img.OriginalFilename, &img.ContentType,
		&img.FileSize, &img.StoragePath, &img.ThumbnailPath, &img.Width,
		&img.Height, &img.UploadedAt, &img.Metadata, &img.CreatedAt, &img.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return img, nil
}

func (r *ImageRepository) List(limit, offset int) ([]*Image, error) {
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images 
		ORDER BY uploaded_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*Image
	for rows.Next() {
		img := &Image{}
		err := rows.Scan(
			&img.ID, &img.Filename, &img.OriginalFilename, &img.ContentType,
			&img.FileSize, &img.StoragePath, &img.ThumbnailPath, &img.Width,
			&img.Height, &img.UploadedAt, &img.Metadata, &img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, rows.Err()
}

func (r *ImageRepository) Delete(id int) error {
	query := `DELETE FROM images WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Tests using sqlmock

func TestImageRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	img := &Image{
		Filename:         "test.jpg",
		OriginalFilename: "original_test.jpg",
		ContentType:      "image/jpeg",
		FileSize:         1024,
		StoragePath:      "/storage/test.jpg",
		Width:            &[]int{800}[0],
		Height:           &[]int{600}[0],
		Metadata:         []byte(`{"description":"test"}`),
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "uploaded_at", "created_at", "updated_at"}).
		AddRow(1, now, now, now)

	mock.ExpectQuery(`INSERT INTO images`).
		WithArgs(
			img.Filename, img.OriginalFilename, img.ContentType, img.FileSize,
			img.StoragePath, img.ThumbnailPath, img.Width, img.Height, img.Metadata,
		).
		WillReturnRows(rows)

	err = repo.Create(img)

	assert.NoError(t, err)
	assert.Equal(t, 1, img.ID)
	assert.WithinDuration(t, now, img.UploadedAt, time.Second)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	img := &Image{
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
		FileSize:    1024,
	}

	mock.ExpectQuery(`INSERT INTO images`).
		WillReturnError(errors.New("database error"))

	err = repo.Create(img)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	now := time.Now()
	width, height := 800, 600
	rows := sqlmock.NewRows([]string{
		"id", "filename", "original_filename", "content_type", "file_size",
		"storage_path", "thumbnail_path", "width", "height", "uploaded_at",
		"metadata", "created_at", "updated_at",
	}).AddRow(
		1, "test.jpg", "original.jpg", "image/jpeg", 1024,
		"/storage/test.jpg", nil, width, height, now,
		[]byte(`{}`), now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM images WHERE id`).
		WithArgs(1).
		WillReturnRows(rows)

	img, err := repo.GetByID(1)

	require.NoError(t, err)
	require.NotNil(t, img)
	assert.Equal(t, 1, img.ID)
	assert.Equal(t, "test.jpg", img.Filename)
	assert.Equal(t, "image/jpeg", img.ContentType)
	assert.Equal(t, &width, img.Width)
	assert.Equal(t, &height, img.Height)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	mock.ExpectQuery(`SELECT (.+) FROM images WHERE id`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	img, err := repo.GetByID(999)

	assert.NoError(t, err)
	assert.Nil(t, img)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "filename", "original_filename", "content_type", "file_size",
		"storage_path", "thumbnail_path", "width", "height", "uploaded_at",
		"metadata", "created_at", "updated_at",
	}).
		AddRow(1, "test1.jpg", "orig1.jpg", "image/jpeg", 1024, "/storage/test1.jpg", nil, nil, nil, now, []byte(`{}`), now, now).
		AddRow(2, "test2.png", "orig2.png", "image/png", 2048, "/storage/test2.png", nil, nil, nil, now, []byte(`{}`), now, now)

	mock.ExpectQuery(`SELECT (.+) FROM images ORDER BY uploaded_at DESC LIMIT`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	images, err := repo.List(10, 0)

	require.NoError(t, err)
	require.Len(t, images, 2)
	assert.Equal(t, "test1.jpg", images[0].Filename)
	assert.Equal(t, "test2.png", images[1].Filename)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_List_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "filename", "original_filename", "content_type", "file_size",
		"storage_path", "thumbnail_path", "width", "height", "uploaded_at",
		"metadata", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM images ORDER BY uploaded_at DESC LIMIT`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	images, err := repo.List(10, 0)

	assert.NoError(t, err)
	assert.Empty(t, images)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	mock.ExpectExec(`DELETE FROM images WHERE id`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	mock.ExpectExec(`DELETE FROM images WHERE id`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(999)

	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewImageRepository(db)

	mock.ExpectExec(`DELETE FROM images WHERE id`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	err = repo.Delete(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test transaction handling
func TestImageRepository_Transaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Test successful transaction
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO images`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	_, err = tx.Exec(`INSERT INTO images (filename, content_type, file_size, storage_path) VALUES ($1, $2, $3, $4)`,
		"test.jpg", "image/jpeg", 1024, "/storage/test.jpg")
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestImageRepository_TransactionRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Test transaction rollback on error
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO images`).
		WillReturnError(errors.New("constraint violation"))
	mock.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	_, err = tx.Exec(`INSERT INTO images (filename, content_type, file_size, storage_path) VALUES ($1, $2, $3, $4)`,
		"test.jpg", "image/jpeg", 1024, "/storage/test.jpg")
	assert.Error(t, err)

	err = tx.Rollback()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
