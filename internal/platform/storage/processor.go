package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // Import for webp decoding support
)

// ImageInfo represents metadata extracted from an image
type ImageInfo struct {
	Width       int
	Height      int
	Format      string
	ColorSpace  string
	HasAlpha    bool
	Orientation int
}

// ImageProcessor implements the domain ImageProcessor interface
type ImageProcessor struct {
	maxWidth  int
	maxHeight int
	quality   int
}

// NewImageProcessor creates a new image processor
func NewImageProcessor(maxWidth, maxHeight, quality int) *ImageProcessor {
	if maxWidth <= 0 {
		maxWidth = 2000
	}
	if maxHeight <= 0 {
		maxHeight = 2000
	}
	if quality <= 0 || quality > 100 {
		quality = 85
	}

	return &ImageProcessor{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		quality:   quality,
	}
}

// GenerateThumbnail implements ImageProcessor.GenerateThumbnail
func (p *ImageProcessor) GenerateThumbnail(ctx context.Context, data io.Reader, maxWidth, maxHeight int) (io.Reader, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	if maxWidth <= 0 || maxHeight <= 0 {
		return nil, errors.New("width and height must be positive")
	}

	// Decode the image
	src, format, err := image.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate thumbnail dimensions maintaining aspect ratio
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// Calculate scaling factor
	scaleX := float64(maxWidth) / float64(srcWidth)
	scaleY := float64(maxHeight) / float64(srcHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Don't upscale images
	if scale > 1.0 {
		scale = 1.0
	}

	dstWidth := int(float64(srcWidth) * scale)
	dstHeight := int(float64(srcHeight) * scale)

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))

	// Use high-quality scaling
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	// Encode thumbnail
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "png":
		err = png.Encode(&buf, dst)
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	default:
		// Default to JPEG for all other formats including webp
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: p.quality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// GetImageInfo implements ImageProcessor.GetImageInfo
func (p *ImageProcessor) GetImageInfo(ctx context.Context, data io.Reader) (*ImageInfo, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	// Decode image to get configuration
	config, format, err := image.DecodeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	// Determine color space and alpha channel
	colorSpace := "unknown"
	hasAlpha := false

	// Try to decode the actual image to get more detailed info
	if seeker, ok := data.(io.Seeker); ok {
		// Reset reader to beginning
		seeker.Seek(0, io.SeekStart)
		
		img, _, err := image.Decode(data)
		if err == nil {
			switch img.ColorModel() {
			case color.RGBAModel, color.RGBA64Model:
				colorSpace = "rgba"
				hasAlpha = true
			case color.NRGBAModel, color.NRGBA64Model:
				colorSpace = "nrgba"
				hasAlpha = true
			case color.GrayModel, color.Gray16Model:
				colorSpace = "gray"
			case color.YCbCrModel:
				colorSpace = "ycbcr"
			default:
				colorSpace = "rgb"
			}
		}
	}

	return &ImageInfo{
		Width:       config.Width,
		Height:      config.Height,
		Format:      format,
		ColorSpace:  colorSpace,
		HasAlpha:    hasAlpha,
		Orientation: 1, // Default orientation
	}, nil
}

// Resize implements ImageProcessor.Resize
func (p *ImageProcessor) Resize(ctx context.Context, data io.Reader, width, height int) (io.Reader, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	if width <= 0 || height <= 0 {
		return nil, errors.New("width and height must be positive")
	}

	// Check maximum dimensions
	if width > p.maxWidth || height > p.maxHeight {
		return nil, fmt.Errorf("requested dimensions %dx%d exceed maximum allowed %dx%d", 
			width, height, p.maxWidth, p.maxHeight)
	}

	// Decode the image
	src, format, err := image.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Use high-quality scaling
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	// Encode resized image
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "png":
		err = png.Encode(&buf, dst)
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	default:
		// Default to JPEG
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: p.quality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// ValidateImage implements ImageProcessor.ValidateImage
func (p *ImageProcessor) ValidateImage(ctx context.Context, data io.Reader, contentType string) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}

	if contentType == "" {
		return errors.New("content type cannot be empty")
	}

	// Check if content type is supported
	supportedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !supportedTypes[contentType] {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Try to decode image to validate it's actually an image
	config, format, err := image.DecodeConfig(data)
	if err != nil {
		return fmt.Errorf("invalid image data: %w", err)
	}

	// Validate dimensions
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", config.Width, config.Height)
	}

	if config.Width > p.maxWidth || config.Height > p.maxHeight {
		return fmt.Errorf("image dimensions %dx%d exceed maximum allowed %dx%d", 
			config.Width, config.Height, p.maxWidth, p.maxHeight)
	}

	// Validate format matches content type
	expectedFormats := map[string][]string{
		"image/jpeg": {"jpeg"},
		"image/jpg":  {"jpeg"},
		"image/png":  {"png"},
		"image/gif":  {"gif"},
		"image/webp": {"webp"},
	}

	if expectedList, ok := expectedFormats[contentType]; ok {
		formatValid := false
		for _, expectedFormat := range expectedList {
			if format == expectedFormat {
				formatValid = true
				break
			}
		}
		if !formatValid {
			return fmt.Errorf("image format %s doesn't match content type %s", format, contentType)
		}
	}

	return nil
}

// OptimizeImage implements ImageProcessor.OptimizeImage
func (p *ImageProcessor) OptimizeImage(ctx context.Context, data io.Reader, quality int) (io.Reader, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	if quality <= 0 || quality > 100 {
		quality = p.quality
	}

	// Decode the image
	src, format, err := image.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "png":
		// For PNG, we can only optimize by re-encoding
		// PNG is lossless, so quality doesn't apply
		encoder := &png.Encoder{
			CompressionLevel: png.BestCompression,
		}
		err = encoder.Encode(&buf, src)
		
	case "gif":
		// For GIF, optimize by re-encoding with better compression
		err = gif.Encode(&buf, src, &gif.Options{
			NumColors: 256, // Standard GIF palette
		})
		
	default:
		// For JPEG and others, apply quality compression
		err = jpeg.Encode(&buf, src, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to optimize image: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// GetSupportedFormats returns the list of supported image formats
func (p *ImageProcessor) GetSupportedFormats() []string {
	return []string{"jpeg", "jpg", "png", "gif", "webp"}
}

// GetMaxDimensions returns the maximum allowed dimensions
func (p *ImageProcessor) GetMaxDimensions() (int, int) {
	return p.maxWidth, p.maxHeight
}

// CalculateOptimalThumbnailSize calculates optimal thumbnail dimensions
func (p *ImageProcessor) CalculateOptimalThumbnailSize(srcWidth, srcHeight, maxWidth, maxHeight int) (int, int) {
	if srcWidth <= 0 || srcHeight <= 0 {
		return maxWidth, maxHeight
	}

	// Calculate scaling factor
	scaleX := float64(maxWidth) / float64(srcWidth)
	scaleY := float64(maxHeight) / float64(srcHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Don't upscale images
	if scale > 1.0 {
		scale = 1.0
	}

	return int(float64(srcWidth) * scale), int(float64(srcHeight) * scale)
}