package image

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Image validation

func TestImage_Validate(t *testing.T) {
	tests := []struct {
		name          string
		image         *Image
		expectError   bool
		expectedError error
	}{
		{
			name: "valid image",
			image: &Image{
				ID:               1,
				Filename:         "test.jpg",
				OriginalFilename: "original_test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
				StoragePath:      "/storage/test.jpg",
				Width:            intPtr(800),
				Height:           intPtr(600),
				UploadedAt:       time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty filename",
			image: &Image{
				Filename:         "",
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
			},
			expectError:   true,
			expectedError: ErrInvalidFilename,
		},
		{
			name: "invalid content type",
			image: &Image{
				Filename:         "test.jpg",
				OriginalFilename: "test.jpg",
				ContentType:      "application/pdf",
				FileSize:         1024,
			},
			expectError:   true,
			expectedError: ErrInvalidContentType,
		},
		{
			name: "file size too large",
			image: &Image{
				Filename:         "test.jpg",
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         MaxFileSize + 1,
			},
			expectError:   true,
			expectedError: ErrInvalidFileSize,
		},
		{
			name: "invalid dimensions - width too large",
			image: &Image{
				Filename:         "test.jpg",
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
				Width:            intPtr(MaxImageWidth + 1),
				Height:           intPtr(600),
			},
			expectError:   true,
			expectedError: ErrInvalidDimensions,
		},
		{
			name: "too many tags",
			image: &Image{
				Filename:         "test.jpg",
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
				Tags:             make([]Tag, MaxTagsPerImage+1),
			},
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
		{
			name: "long filename",
			image: &Image{
				Filename:         strings.Repeat("a", MaxFilenameLen+1),
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
			},
			expectError:   true,
			expectedError: ErrInvalidFilename,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.image.Validate()

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImage_GetFileExtension(t *testing.T) {
	tests := []struct {
		name             string
		originalFilename string
		expectedExt      string
	}{
		{"jpg file", "test.jpg", ".jpg"},
		{"jpeg file", "test.JPEG", ".jpeg"},
		{"png file", "test.png", ".png"},
		{"no extension", "test", ""},
		{"multiple dots", "test.backup.jpg", ".jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := &Image{OriginalFilename: tt.originalFilename}
			ext := img.GetFileExtension()
			assert.Equal(t, tt.expectedExt, ext)
		})
	}
}

func TestImage_IsImage(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{"jpeg", "image/jpeg", true},
		{"png", "image/png", true},
		{"pdf", "application/pdf", false},
		{"text", "text/plain", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := &Image{ContentType: tt.contentType}
			result := img.IsImage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImage_GetAspectRatio(t *testing.T) {
	tests := []struct {
		name     string
		width    *int
		height   *int
		expected *float64
	}{
		{"16:9 ratio", intPtr(1920), intPtr(1080), float64Ptr(16.0 / 9.0)},
		{"square", intPtr(100), intPtr(100), float64Ptr(1.0)},
		{"portrait", intPtr(600), intPtr(800), float64Ptr(0.75)},
		{"nil width", nil, intPtr(800), nil},
		{"nil height", intPtr(600), nil, nil},
		{"zero height", intPtr(600), intPtr(0), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := &Image{Width: tt.width, Height: tt.height}
			result := img.GetAspectRatio()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.InDelta(t, *tt.expected, *result, 0.001)
			}
		})
	}
}

func TestImage_GetSizeCategory(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		expected string
	}{
		{"small file", 50*1024, "small"},          // 50KB
		{"medium file", 500*1024, "medium"},       // 500KB
		{"large file", 5*1024*1024, "large"},      // 5MB
		{"xlarge file", 20*1024*1024, "xlarge"},   // 20MB
		{"tiny file", 1, "small"},                 // 1 byte
		{"boundary small", 100*1024-1, "small"},   // just under 100KB
		{"boundary medium", 100*1024, "medium"},   // exactly 100KB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := &Image{FileSize: tt.fileSize}
			result := img.GetSizeCategory()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImage_HasTag(t *testing.T) {
	tags := []Tag{
		{Name: "nature"},
		{Name: "landscape"},
		{Name: "sunset"},
	}
	img := &Image{Tags: tags}

	assert.True(t, img.HasTag("nature"))
	assert.True(t, img.HasTag("landscape"))
	assert.False(t, img.HasTag("portrait"))
	assert.False(t, img.HasTag(""))
	assert.False(t, img.HasTag("Nature")) // case sensitive
}

// Test Tag validation

func TestTag_Validate(t *testing.T) {
	tests := []struct {
		name          string
		tag           *Tag
		expectError   bool
		expectedError error
	}{
		{
			name:        "valid tag",
			tag:         &Tag{Name: "nature"},
			expectError: false,
		},
		{
			name:          "empty name",
			tag:           &Tag{Name: ""},
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
		{
			name:          "name too long",
			tag:           &Tag{Name: strings.Repeat("a", MaxTagNameLen+1)},
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
		{
			name:          "invalid characters - uppercase",
			tag:           &Tag{Name: "Nature"},
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
		{
			name:          "invalid characters - space",
			tag:           &Tag{Name: "nature photography"},
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
		{
			name:          "invalid characters - special chars",
			tag:           &Tag{Name: "nature@home"},
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
		{
			name:        "valid with numbers",
			tag:         &Tag{Name: "photo2024"},
			expectError: false,
		},
		{
			name:        "valid with hyphens",
			tag:         &Tag{Name: "black-and-white"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tag.Validate()

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTag_NormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "nature", "nature"},
		{"uppercase", "NATURE", "nature"},
		{"mixed case", "NaTuRe", "nature"},
		{"with spaces", " nature ", "nature"},
		{"with tabs", "\tnature\t", "nature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := &Tag{Name: tt.input}
			tag.NormalizeName()
			assert.Equal(t, tt.expected, tag.Name)
		})
	}
}

func TestNewTag(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedName  string
		expectedError error
	}{
		{
			name:         "valid tag",
			input:        "Nature",
			expectError:  false,
			expectedName: "nature",
		},
		{
			name:         "with spaces",
			input:        " landscape ",
			expectError:  false,
			expectedName: "landscape",
		},
		{
			name:          "empty name",
			input:         "",
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
		{
			name:          "invalid characters",
			input:         "nature@home",
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := NewTag(tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, tag)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, tag)
				assert.Equal(t, tt.expectedName, tag.Name)
				assert.False(t, tag.CreatedAt.IsZero())
			}
		})
	}
}

// Test CreateImageRequest validation

func TestCreateImageRequest_Validate(t *testing.T) {
	tests := []struct {
		name          string
		request       *CreateImageRequest
		expectError   bool
		expectedError error
	}{
		{
			name: "valid request",
			request: &CreateImageRequest{
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
				Width:            intPtr(800),
				Height:           intPtr(600),
				Metadata:         json.RawMessage(`{"camera": "Canon"}`),
				Tags:             []string{"nature", "landscape"},
			},
			expectError: false,
		},
		{
			name: "empty filename",
			request: &CreateImageRequest{
				OriginalFilename: "",
				ContentType:      "image/jpeg",
				FileSize:         1024,
			},
			expectError:   true,
			expectedError: ErrInvalidFilename,
		},
		{
			name: "invalid content type",
			request: &CreateImageRequest{
				OriginalFilename: "test.pdf",
				ContentType:      "application/pdf",
				FileSize:         1024,
			},
			expectError:   true,
			expectedError: ErrInvalidContentType,
		},
		{
			name: "invalid metadata JSON",
			request: &CreateImageRequest{
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
				Metadata:         json.RawMessage(`{"invalid": json}`),
			},
			expectError:   true,
			expectedError: ErrInvalidImageData,
		},
		{
			name: "duplicate tags",
			request: &CreateImageRequest{
				OriginalFilename: "test.jpg",
				ContentType:      "image/jpeg",
				FileSize:         1024,
				Tags:             []string{"nature", "Nature", "NATURE"},
			},
			expectError:   true,
			expectedError: ErrDuplicateTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateImageRequest_GenerateFilename(t *testing.T) {
	request := &CreateImageRequest{
		OriginalFilename: "my_photo.jpg",
	}

	filename := request.GenerateFilename()
	
	assert.Contains(t, filename, "my_photo_")
	assert.True(t, strings.HasSuffix(filename, ".jpg"))
	assert.Greater(t, len(filename), len("my_photo_.jpg"))
}

func TestCreateImageRequest_GetMimeType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expectError bool
		expected    string
	}{
		{"valid jpeg", "image/jpeg", false, "image/jpeg"},
		{"valid png", "image/png", false, "image/png"},
		{"invalid type", "application/pdf", true, ""},
		{"empty type", "", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &CreateImageRequest{ContentType: tt.contentType}
			mimeType, err := request.GetMimeType()

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, mimeType)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, mimeType)
			}
		})
	}
}

// Test ListImagesRequest validation

func TestListImagesRequest_Validate(t *testing.T) {
	tests := []struct {
		name          string
		request       *ListImagesRequest
		expectError   bool
		expectedError error
	}{
		{
			name: "valid request",
			request: &ListImagesRequest{
				Page:     1,
				PageSize: 20,
				Tag:      "nature",
			},
			expectError: false,
		},
		{
			name: "invalid page",
			request: &ListImagesRequest{
				Page:     0,
				PageSize: 20,
			},
			expectError:   true,
			expectedError: ErrInvalidPagination,
		},
		{
			name: "page size too large",
			request: &ListImagesRequest{
				Page:     1,
				PageSize: MaxPageSize + 1,
			},
			expectError:   true,
			expectedError: ErrInvalidPagination,
		},
		{
			name: "tag name too long",
			request: &ListImagesRequest{
				Page:     1,
				PageSize: 20,
				Tag:      strings.Repeat("a", MaxTagNameLen+1),
			},
			expectError:   true,
			expectedError: ErrInvalidTagName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListImagesRequest_SetDefaults(t *testing.T) {
	tests := []struct {
		name         string
		request      *ListImagesRequest
		expectedPage int
		expectedSize int
	}{
		{
			name:         "zero values",
			request:      &ListImagesRequest{},
			expectedPage: 1,
			expectedSize: DefaultPageSize,
		},
		{
			name:         "existing values",
			request:      &ListImagesRequest{Page: 2, PageSize: 10},
			expectedPage: 2,
			expectedSize: 10,
		},
		{
			name:         "zero page only",
			request:      &ListImagesRequest{PageSize: 15},
			expectedPage: 1,
			expectedSize: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.request.SetDefaults()
			assert.Equal(t, tt.expectedPage, tt.request.Page)
			assert.Equal(t, tt.expectedSize, tt.request.PageSize)
		})
	}
}

func TestListImagesRequest_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int
	}{
		{"first page", 1, 20, 0},
		{"second page", 2, 20, 20},
		{"third page", 3, 10, 20},
		{"large page", 10, 25, 225},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &ListImagesRequest{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}
			offset := request.GetOffset()
			assert.Equal(t, tt.expected, offset)
		})
	}
}

// Test ListImagesResponse methods

func TestListImagesResponse_CalculateTotalPages(t *testing.T) {
	tests := []struct {
		name         string
		totalCount   int
		pageSize     int
		expectedPages int
	}{
		{"exact division", 100, 20, 5},
		{"with remainder", 105, 20, 6},
		{"less than page size", 15, 20, 1},
		{"zero count", 0, 20, 0},
		{"zero page size", 100, 0, 0},
		{"single item", 1, 20, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &ListImagesResponse{
				TotalCount: tt.totalCount,
				PageSize:   tt.pageSize,
			}
			response.CalculateTotalPages()
			assert.Equal(t, tt.expectedPages, response.TotalPages)
		})
	}
}

func TestListImagesResponse_HasNextPage(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		totalPages int
		expected   bool
	}{
		{"has next page", 2, 5, true},
		{"last page", 5, 5, false},
		{"beyond last page", 6, 5, false},
		{"first page with more", 1, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &ListImagesResponse{
				Page:       tt.page,
				TotalPages: tt.totalPages,
			}
			result := response.HasNextPage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListImagesResponse_HasPrevPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected bool
	}{
		{"first page", 1, false},
		{"second page", 2, true},
		{"later page", 5, true},
		{"zero page", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &ListImagesResponse{Page: tt.page}
			result := response.HasPrevPage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test helper functions

func TestValidateContentTypeFromExtension(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		filename    string
		expectError bool
	}{
		{"jpg with jpeg", "image/jpeg", "test.jpg", false},
		{"jpeg with jpeg", "image/jpeg", "test.jpeg", false},
		{"png with png", "image/png", "test.png", false},
		{"jpg with png - mismatch", "image/png", "test.jpg", true},
		{"unsupported extension", "image/jpeg", "test.bmp", true},
		{"case insensitive extension", "image/jpeg", "test.JPG", false},
		{"no extension", "image/jpeg", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContentTypeFromExtension(tt.contentType, tt.filename)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidContentType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test constants and globals

func TestConstants(t *testing.T) {
	assert.Greater(t, MaxFileSize, MinFileSize)
	assert.Greater(t, MaxFilenameLen, 0)
	assert.Greater(t, MaxTagNameLen, MinTagNameLen)
	assert.Greater(t, MaxTagsPerImage, 0)
	assert.Greater(t, MaxPageSize, MinPageSize)
	assert.Greater(t, DefaultPageSize, 0)
	assert.LessOrEqual(t, DefaultPageSize, MaxPageSize)
}

func TestSupportedContentTypes(t *testing.T) {
	expectedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, contentType := range expectedTypes {
		assert.True(t, SupportedContentTypes[contentType], "Content type %s should be supported", contentType)
	}

	unsupportedTypes := []string{
		"application/pdf",
		"text/plain",
		"image/bmp",
		"video/mp4",
	}

	for _, contentType := range unsupportedTypes {
		assert.False(t, SupportedContentTypes[contentType], "Content type %s should not be supported", contentType)
	}
}

func TestDomainErrors(t *testing.T) {
	errors := []error{
		ErrInvalidImageData,
		ErrInvalidContentType,
		ErrInvalidFileSize,
		ErrInvalidFilename,
		ErrInvalidDimensions,
		ErrInvalidTagName,
		ErrInvalidPagination,
		ErrImageNotFound,
		ErrTagNotFound,
		ErrDuplicateTag,
	}

	for _, err := range errors {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

// Helper functions for tests

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// Benchmarks

func BenchmarkImage_Validate(b *testing.B) {
	img := &Image{
		Filename:         "test.jpg",
		OriginalFilename: "test.jpg",
		ContentType:      "image/jpeg",
		FileSize:         1024,
		Width:            intPtr(800),
		Height:           intPtr(600),
		Tags:             []Tag{{Name: "nature"}, {Name: "landscape"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = img.Validate()
	}
}

func BenchmarkCreateImageRequest_Validate(b *testing.B) {
	request := &CreateImageRequest{
		OriginalFilename: "test.jpg",
		ContentType:      "image/jpeg",
		FileSize:         1024,
		Width:            intPtr(800),
		Height:           intPtr(600),
		Tags:             []string{"nature", "landscape", "sunset"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = request.Validate()
	}
}

func BenchmarkTag_Validate(b *testing.B) {
	tag := &Tag{Name: "nature-photography-2024"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tag.Validate()
	}
}