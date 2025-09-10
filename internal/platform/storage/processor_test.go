package storage

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewImageProcessor(t *testing.T) {
	tests := []struct {
		name         string
		maxWidth     int
		maxHeight    int
		quality      int
		expectWidth  int
		expectHeight int
		expectQual   int
	}{
		{
			name:         "valid parameters",
			maxWidth:     1920,
			maxHeight:    1080,
			quality:      90,
			expectWidth:  1920,
			expectHeight: 1080,
			expectQual:   90,
		},
		{
			name:         "zero width defaults to 2000",
			maxWidth:     0,
			maxHeight:    1080,
			quality:      90,
			expectWidth:  2000,
			expectHeight: 1080,
			expectQual:   90,
		},
		{
			name:         "zero height defaults to 2000",
			maxWidth:     1920,
			maxHeight:    0,
			quality:      90,
			expectWidth:  1920,
			expectHeight: 2000,
			expectQual:   90,
		},
		{
			name:         "zero quality defaults to 85",
			maxWidth:     1920,
			maxHeight:    1080,
			quality:      0,
			expectWidth:  1920,
			expectHeight: 1080,
			expectQual:   85,
		},
		{
			name:         "quality over 100 defaults to 85",
			maxWidth:     1920,
			maxHeight:    1080,
			quality:      150,
			expectWidth:  1920,
			expectHeight: 1080,
			expectQual:   85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewImageProcessor(tt.maxWidth, tt.maxHeight, tt.quality)
			
			assert.NotNil(t, processor)
			assert.Equal(t, tt.expectWidth, processor.maxWidth)
			assert.Equal(t, tt.expectHeight, processor.maxHeight)
			assert.Equal(t, tt.expectQual, processor.quality)
		})
	}
}

func TestImageProcessor_GenerateThumbnail(t *testing.T) {
	processor := NewImageProcessor(2000, 2000, 85)

	// Create a simple test image
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, testImg, nil)
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        io.Reader
		maxWidth    int
		maxHeight   int
		expectError bool
	}{
		{
			name:        "nil data",
			data:        nil,
			maxWidth:    50,
			maxHeight:   50,
			expectError: true,
		},
		{
			name:        "zero width",
			data:        bytes.NewReader(buf.Bytes()),
			maxWidth:    0,
			maxHeight:   50,
			expectError: true,
		},
		{
			name:        "zero height",
			data:        bytes.NewReader(buf.Bytes()),
			maxWidth:    50,
			maxHeight:   0,
			expectError: true,
		},
		{
			name:        "invalid image data",
			data:        strings.NewReader("not an image"),
			maxWidth:    50,
			maxHeight:   50,
			expectError: true,
		},
		{
			name:        "valid image data",
			data:        bytes.NewReader(buf.Bytes()),
			maxWidth:    50,
			maxHeight:   50,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.GenerateThumbnail(context.Background(), tt.data, tt.maxWidth, tt.maxHeight)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify the result is valid image data
				if result != nil {
					_, _, err := image.DecodeConfig(result)
					assert.NoError(t, err, "Generated thumbnail should be valid image")
				}
			}
		})
	}
}

func TestImageProcessor_GetImageInfo(t *testing.T) {
	processor := NewImageProcessor(2000, 2000, 85)

	// Create test images of different formats
	jpegImg := image.NewRGBA(image.Rect(0, 0, 100, 200))
	var jpegBuf bytes.Buffer
	err := jpeg.Encode(&jpegBuf, jpegImg, nil)
	require.NoError(t, err)

	pngImg := image.NewRGBA(image.Rect(0, 0, 150, 100))
	var pngBuf bytes.Buffer
	err = png.Encode(&pngBuf, pngImg)
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          io.Reader
		expectError   bool
		expectedWidth int
		expectedHeight int
		expectedFormat string
	}{
		{
			name:        "nil data",
			data:        nil,
			expectError: true,
		},
		{
			name:        "invalid image data",
			data:        strings.NewReader("not an image"),
			expectError: true,
		},
		{
			name:           "valid JPEG",
			data:           bytes.NewReader(jpegBuf.Bytes()),
			expectError:    false,
			expectedWidth:  100,
			expectedHeight: 200,
			expectedFormat: "jpeg",
		},
		{
			name:           "valid PNG", 
			data:           bytes.NewReader(pngBuf.Bytes()),
			expectError:    false,
			expectedWidth:  150,
			expectedHeight: 100,
			expectedFormat: "png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := processor.GetImageInfo(context.Background(), tt.data)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				
				if info != nil {
					assert.Equal(t, tt.expectedWidth, info.Width)
					assert.Equal(t, tt.expectedHeight, info.Height)
					assert.Equal(t, tt.expectedFormat, info.Format)
					assert.NotEmpty(t, info.ColorSpace)
				}
			}
		})
	}
}

func TestImageProcessor_Resize(t *testing.T) {
	processor := NewImageProcessor(2000, 2000, 85)

	// Create a test image
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, testImg, nil)
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        io.Reader
		width       int
		height      int
		expectError bool
	}{
		{
			name:        "nil data",
			data:        nil,
			width:       50,
			height:      50,
			expectError: true,
		},
		{
			name:        "zero width",
			data:        bytes.NewReader(buf.Bytes()),
			width:       0,
			height:      50,
			expectError: true,
		},
		{
			name:        "zero height",
			data:        bytes.NewReader(buf.Bytes()),
			width:       50,
			height:      0,
			expectError: true,
		},
		{
			name:        "exceeds max width",
			data:        bytes.NewReader(buf.Bytes()),
			width:       3000,
			height:      50,
			expectError: true,
		},
		{
			name:        "exceeds max height",
			data:        bytes.NewReader(buf.Bytes()),
			width:       50,
			height:      3000,
			expectError: true,
		},
		{
			name:        "invalid image data",
			data:        strings.NewReader("not an image"),
			width:       50,
			height:      50,
			expectError: true,
		},
		{
			name:        "valid resize",
			data:        bytes.NewReader(buf.Bytes()),
			width:       200,
			height:      150,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.Resize(context.Background(), tt.data, tt.width, tt.height)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify the resized image has correct dimensions
				if result != nil {
					config, _, err := image.DecodeConfig(result)
					assert.NoError(t, err)
					assert.Equal(t, tt.width, config.Width)
					assert.Equal(t, tt.height, config.Height)
				}
			}
		})
	}
}

func TestImageProcessor_ValidateImage(t *testing.T) {
	processor := NewImageProcessor(2000, 2000, 85)

	// Create test images
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var jpegBuf bytes.Buffer
	err := jpeg.Encode(&jpegBuf, testImg, nil)
	require.NoError(t, err)

	// Create oversized image for testing dimension limits
	oversizedImg := image.NewRGBA(image.Rect(0, 0, 3000, 3000))
	var oversizedBuf bytes.Buffer  
	err = jpeg.Encode(&oversizedBuf, oversizedImg, nil)
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        io.Reader
		contentType string
		expectError bool
	}{
		{
			name:        "nil data",
			data:        nil,
			contentType: "image/jpeg",
			expectError: true,
		},
		{
			name:        "empty content type",
			data:        bytes.NewReader(jpegBuf.Bytes()),
			contentType: "",
			expectError: true,
		},
		{
			name:        "unsupported content type",
			data:        bytes.NewReader(jpegBuf.Bytes()),
			contentType: "application/pdf",
			expectError: true,
		},
		{
			name:        "invalid image data",
			data:        strings.NewReader("not an image"),
			contentType: "image/jpeg",
			expectError: true,
		},
		{
			name:        "oversized image",
			data:        bytes.NewReader(oversizedBuf.Bytes()),
			contentType: "image/jpeg",
			expectError: true,
		},
		{
			name:        "valid JPEG",
			data:        bytes.NewReader(jpegBuf.Bytes()),
			contentType: "image/jpeg",
			expectError: false,
		},
		{
			name:        "valid JPG content type",
			data:        bytes.NewReader(jpegBuf.Bytes()),
			contentType: "image/jpg",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateImage(context.Background(), tt.data, tt.contentType)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImageProcessor_OptimizeImage(t *testing.T) {
	processor := NewImageProcessor(2000, 2000, 85)

	// Create test images of different formats
	jpegImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var jpegBuf bytes.Buffer
	err := jpeg.Encode(&jpegBuf, jpegImg, &jpeg.Options{Quality: 100})
	require.NoError(t, err)

	pngImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var pngBuf bytes.Buffer
	err = png.Encode(&pngBuf, pngImg)
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        io.Reader
		quality     int
		expectError bool
	}{
		{
			name:        "nil data",
			data:        nil,
			quality:     75,
			expectError: true,
		},
		{
			name:        "invalid image data",
			data:        strings.NewReader("not an image"),
			quality:     75,
			expectError: true,
		},
		{
			name:        "optimize JPEG with custom quality",
			data:        bytes.NewReader(jpegBuf.Bytes()),
			quality:     50,
			expectError: false,
		},
		{
			name:        "optimize JPEG with invalid quality (uses default)",
			data:        bytes.NewReader(jpegBuf.Bytes()),
			quality:     0,
			expectError: false,
		},
		{
			name:        "optimize PNG",
			data:        bytes.NewReader(pngBuf.Bytes()),
			quality:     75, // Quality ignored for PNG
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.OptimizeImage(context.Background(), tt.data, tt.quality)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify the result is valid image data
				if result != nil {
					_, _, err := image.DecodeConfig(result)
					assert.NoError(t, err, "Optimized image should be valid")
				}
			}
		})
	}
}

func TestImageProcessor_GetSupportedFormats(t *testing.T) {
	processor := NewImageProcessor(2000, 2000, 85)
	formats := processor.GetSupportedFormats()

	expectedFormats := []string{"jpeg", "jpg", "png", "gif", "webp"}
	assert.ElementsMatch(t, expectedFormats, formats)
}

func TestImageProcessor_GetMaxDimensions(t *testing.T) {
	processor := NewImageProcessor(1920, 1080, 85)
	width, height := processor.GetMaxDimensions()

	assert.Equal(t, 1920, width)
	assert.Equal(t, 1080, height)
}

func TestImageProcessor_CalculateOptimalThumbnailSize(t *testing.T) {
	processor := NewImageProcessor(2000, 2000, 85)

	tests := []struct {
		name               string
		srcWidth           int
		srcHeight          int
		maxWidth           int
		maxHeight          int
		expectedWidth      int
		expectedHeight     int
	}{
		{
			name:           "landscape image fits within limits",
			srcWidth:       1000,
			srcHeight:      500,
			maxWidth:       800,
			maxHeight:      600,
			expectedWidth:  800,
			expectedHeight: 400,
		},
		{
			name:           "portrait image fits within limits",
			srcWidth:       500,
			srcHeight:      1000,
			maxWidth:       800,
			maxHeight:      600,
			expectedWidth:  300,
			expectedHeight: 600,
		},
		{
			name:           "small image no upscaling",
			srcWidth:       200,
			srcHeight:      100,
			maxWidth:       800,
			maxHeight:      600,
			expectedWidth:  200,
			expectedHeight: 100,
		},
		{
			name:           "square image",
			srcWidth:       400,
			srcHeight:      400,
			maxWidth:       300,
			maxHeight:      300,
			expectedWidth:  300,
			expectedHeight: 300,
		},
		{
			name:           "invalid source dimensions",
			srcWidth:       0,
			srcHeight:      100,
			maxWidth:       300,
			maxHeight:      300,
			expectedWidth:  300,
			expectedHeight: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := processor.CalculateOptimalThumbnailSize(
				tt.srcWidth, tt.srcHeight, tt.maxWidth, tt.maxHeight)
			
			assert.Equal(t, tt.expectedWidth, width)
			assert.Equal(t, tt.expectedHeight, height)
		})
	}
}

// Integration test with real image processing
func TestImageProcessor_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	processor := NewImageProcessor(1000, 1000, 85)

	// Create a larger test image
	testImg := image.NewRGBA(image.Rect(0, 0, 400, 300))
	
	// Fill with some pattern for better testing
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			testImg.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256), 
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, testImg, &jpeg.Options{Quality: 100})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("full workflow", func(t *testing.T) {
		// 1. Validate the image
		err := processor.ValidateImage(ctx, bytes.NewReader(buf.Bytes()), "image/jpeg")
		assert.NoError(t, err)

		// 2. Get image info
		info, err := processor.GetImageInfo(ctx, bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, 400, info.Width)
		assert.Equal(t, 300, info.Height)
		assert.Equal(t, "jpeg", info.Format)

		// 3. Generate thumbnail
		thumb, err := processor.GenerateThumbnail(ctx, bytes.NewReader(buf.Bytes()), 100, 75)
		require.NoError(t, err)
		
		// Verify thumbnail dimensions
		thumbConfig, _, err := image.DecodeConfig(thumb)
		require.NoError(t, err)
		assert.Equal(t, 100, thumbConfig.Width)
		assert.Equal(t, 75, thumbConfig.Height)

		// 4. Resize image
		resized, err := processor.Resize(ctx, bytes.NewReader(buf.Bytes()), 200, 150)
		require.NoError(t, err)
		
		// Verify resized dimensions
		resizedConfig, _, err := image.DecodeConfig(resized)
		require.NoError(t, err)
		assert.Equal(t, 200, resizedConfig.Width)
		assert.Equal(t, 150, resizedConfig.Height)

		// 5. Optimize image
		optimized, err := processor.OptimizeImage(ctx, bytes.NewReader(buf.Bytes()), 60)
		require.NoError(t, err)
		
		// Verify optimized image is valid
		_, _, err = image.DecodeConfig(optimized)
		assert.NoError(t, err)
	})
}