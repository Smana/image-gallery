package image

import (
	"encoding/json"
	"time"
)

type Image struct {
	ID               int             `json:"id" db:"id"`
	Filename         string          `json:"filename" db:"filename"`
	OriginalFilename string          `json:"original_filename" db:"original_filename"`
	ContentType      string          `json:"content_type" db:"content_type"`
	FileSize         int64           `json:"file_size" db:"file_size"`
	StoragePath      string          `json:"storage_path" db:"storage_path"`
	ThumbnailPath    *string         `json:"thumbnail_path" db:"thumbnail_path"`
	Width            *int            `json:"width" db:"width"`
	Height           *int            `json:"height" db:"height"`
	UploadedAt       time.Time       `json:"uploaded_at" db:"uploaded_at"`
	Metadata         json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
	Tags             []Tag           `json:"tags,omitempty"`
}

type Tag struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateImageRequest struct {
	OriginalFilename string          `json:"original_filename"`
	ContentType      string          `json:"content_type"`
	FileSize         int64           `json:"file_size"`
	Width            *int            `json:"width"`
	Height           *int            `json:"height"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
}

type UpdateImageRequest struct {
	Tags []string `json:"tags"`
}

type ListImagesRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Tag      string `json:"tag" form:"tag"`
}

type ListImagesResponse struct {
	Images     []Image `json:"images"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}