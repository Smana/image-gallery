package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"image-gallery/internal/domain/image"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	maxUploadSize = 50 << 20 // 50MB total per request
	maxFileSize   = 10 << 20 // 10MB per file
)

// UploadResponse represents the response for a successful upload
type UploadResponse struct {
	Images []UploadedImageInfo `json:"images"`
	Count  int                 `json:"count"`
	Errors []UploadError       `json:"errors,omitempty"`
}

// UploadedImageInfo represents information about an uploaded image
type UploadedImageInfo struct {
	ID               int      `json:"id"`
	Filename         string   `json:"filename"`
	OriginalFilename string   `json:"original_filename"`
	Size             int64    `json:"size"`
	ContentType      string   `json:"content_type"`
	Width            *int     `json:"width,omitempty"`
	Height           *int     `json:"height,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	URL              string   `json:"url,omitempty"`
}

// UploadError represents an error that occurred during upload
type UploadError struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

// processedFileResult holds the result of processing a single uploaded file
type processedFileResult struct {
	uploadedImage *UploadedImageInfo
	uploadError   *UploadError
	bytesUploaded int64
}

// uploadImagesHandler handles POST requests to upload one or more images
func (h *Handler) uploadImagesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Start tracing span
	ctx, span := h.tracer.Start(ctx, "UploadImages", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Log upload initiation
	h.logger.Info(ctx).
		Str("user_agent", r.UserAgent()).
		Str("content_type", r.Header.Get("Content-Type")).
		Msg("Starting image upload request")

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse multipart form")
		h.logger.Error(ctx).Err(err).Msg("Failed to parse multipart form")
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get uploaded files
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		err := fmt.Errorf("no files provided")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no files in request")
		h.logger.Warn(ctx).Msg("Upload request with no files")
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}

	// Get tags (comma-separated)
	tagsStr := r.FormValue("tags")
	tags := parseTags(tagsStr)

	// Add span attributes
	span.SetAttributes(
		attribute.Int("upload.file_count", len(files)),
		attribute.Int("upload.tag_count", len(tags)),
	)

	h.logger.Info(ctx).
		Int("file_count", len(files)).
		Int("tag_count", len(tags)).
		Interface("tags", tags).
		Msg("Processing uploaded files")

	// Process each file
	uploadedImages := make([]UploadedImageInfo, 0, len(files))
	var uploadErrors []UploadError
	var totalBytes int64

	for _, fileHeader := range files {
		result := h.processUploadedFile(ctx, fileHeader, tags)

		if result.uploadedImage != nil {
			uploadedImages = append(uploadedImages, *result.uploadedImage)
			totalBytes += result.bytesUploaded
		}

		if result.uploadError != nil {
			uploadErrors = append(uploadErrors, *result.uploadError)
		}
	}

	// Set overall span attributes
	span.SetAttributes(
		attribute.Int("upload.success_count", len(uploadedImages)),
		attribute.Int("upload.error_count", len(uploadErrors)),
		attribute.Int64("upload.total_bytes", totalBytes),
	)

	// Determine response status
	statusCode := http.StatusCreated
	if len(uploadedImages) == 0 {
		statusCode = http.StatusBadRequest
		span.SetStatus(codes.Error, "all uploads failed")
	} else if len(uploadErrors) > 0 {
		statusCode = http.StatusMultiStatus // 207 - some succeeded, some failed
		span.SetStatus(codes.Ok, "partial success")
	} else {
		span.SetStatus(codes.Ok, "all uploads successful")
	}

	h.logger.Info(ctx).
		Int("success_count", len(uploadedImages)).
		Int("error_count", len(uploadErrors)).
		Int64("total_bytes", totalBytes).
		Int("status_code", statusCode).
		Msg("Upload request completed")

	// Build response
	response := UploadResponse{
		Images: uploadedImages,
		Count:  len(uploadedImages),
		Errors: uploadErrors,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error(ctx).Err(err).Msg("Failed to encode response")
	}
}

// processUploadedFile processes a single uploaded file and returns the result
func (h *Handler) processUploadedFile(ctx context.Context, fileHeader *multipart.FileHeader, tags []string) processedFileResult {
	// Create child span for each file
	_, fileSpan := h.tracer.Start(ctx, "ProcessUploadedFile",
		trace.WithAttributes(
			attribute.String("file.name", fileHeader.Filename),
			attribute.Int64("file.size", fileHeader.Size),
		),
	)
	defer fileSpan.End()

	h.logger.Debug(ctx).
		Str("filename", fileHeader.Filename).
		Int64("size", fileHeader.Size).
		Msg("Processing file")

	// Validate file size
	if fileHeader.Size > maxFileSize {
		err := fmt.Errorf("file too large: %d bytes (max %d bytes)", fileHeader.Size, maxFileSize)
		fileSpan.RecordError(err)
		fileSpan.SetStatus(codes.Error, "file too large")
		h.logger.Warn(ctx).Str("filename", fileHeader.Filename).Int64("size", fileHeader.Size).Msg("File too large")
		return processedFileResult{
			uploadError: &UploadError{
				Filename: fileHeader.Filename,
				Error:    err.Error(),
			},
		}
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		fileSpan.RecordError(err)
		fileSpan.SetStatus(codes.Error, "failed to open file")
		h.logger.Error(ctx).Err(err).Str("filename", fileHeader.Filename).Msg("Failed to open uploaded file")
		return processedFileResult{
			uploadError: &UploadError{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("Failed to open file: %v", err),
			},
		}
	}

	// Read file content into buffer (allows multiple reads for dimension extraction + upload)
	fileData, err := io.ReadAll(file)
	_ = file.Close() //nolint:errcheck // Explicitly ignore close error (already read data)
	if err != nil {
		fileSpan.RecordError(err)
		fileSpan.SetStatus(codes.Error, "failed to read file")
		h.logger.Error(ctx).Err(err).Str("filename", fileHeader.Filename).Msg("Failed to read file content")
		return processedFileResult{
			uploadError: &UploadError{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("Failed to read file: %v", err),
			},
		}
	}

	// Detect and validate content type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileData)
	}

	if !isSupportedImageType(contentType) {
		err := fmt.Errorf("unsupported image type: %s", contentType)
		fileSpan.RecordError(err)
		fileSpan.SetStatus(codes.Error, "unsupported content type")
		h.logger.Warn(ctx).Str("filename", fileHeader.Filename).Str("content_type", contentType).Msg("Unsupported content type")
		return processedFileResult{
			uploadError: &UploadError{
				Filename: fileHeader.Filename,
				Error:    err.Error(),
			},
		}
	}

	// Extract image dimensions
	width, height := h.extractImageDimensions(ctx, fileData, fileHeader.Filename, fileSpan)

	// Create upload request
	createReq := &image.CreateImageRequest{
		OriginalFilename: fileHeader.Filename,
		ContentType:      contentType,
		FileSize:         fileHeader.Size,
		Width:            width,
		Height:           height,
		Tags:             tags,
	}

	// Upload image via ImageService
	img, err := h.imageService.CreateImage(ctx, createReq, bytes.NewReader(fileData))
	if err != nil {
		fileSpan.RecordError(err)
		fileSpan.SetStatus(codes.Error, "image creation failed")
		h.logger.Error(ctx).Err(err).Str("filename", fileHeader.Filename).Msg("Failed to create image")
		return processedFileResult{
			uploadError: &UploadError{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("Upload failed: %v", err),
			},
		}
	}

	// Generate presigned URL for immediate access
	var imageURL string
	if h.storageService != nil {
		url, err := h.storageService.GenerateURL(ctx, img.StoragePath, 3600) // 1 hour expiry
		if err != nil {
			h.logger.Warn(ctx).Err(err).Int("image_id", img.ID).Msg("Failed to generate image URL")
		} else {
			imageURL = url
		}
	}

	// Extract tag names
	tagNames := make([]string, 0, len(img.Tags))
	for _, tag := range img.Tags {
		tagNames = append(tagNames, tag.Name)
	}

	fileSpan.SetStatus(codes.Ok, "file processed successfully")

	h.logger.Info(ctx).
		Int("image_id", img.ID).
		Str("filename", fileHeader.Filename).
		Int64("size", fileHeader.Size).
		Msg("Successfully uploaded image")

	return processedFileResult{
		uploadedImage: &UploadedImageInfo{
			ID:               img.ID,
			Filename:         img.Filename,
			OriginalFilename: img.OriginalFilename,
			Size:             img.FileSize,
			ContentType:      img.ContentType,
			Width:            img.Width,
			Height:           img.Height,
			Tags:             tagNames,
			URL:              imageURL,
		},
		bytesUploaded: fileHeader.Size,
	}
}

// extractImageDimensions extracts width and height from image data
func (h *Handler) extractImageDimensions(ctx context.Context, fileData []byte, filename string, span trace.Span) (width *int, height *int) {
	if h.container == nil {
		return nil, nil
	}

	imageInfo, err := h.container.ImageProcessor().GetImageInfo(ctx, bytes.NewReader(fileData))
	if err != nil {
		// Log but don't fail - dimensions are optional
		h.logger.Warn(ctx).Err(err).Str("filename", filename).Msg("Failed to extract image dimensions")
		return nil, nil
	}

	span.SetAttributes(
		attribute.Int("image.width", imageInfo.Width),
		attribute.Int("image.height", imageInfo.Height),
		attribute.String("image.format", imageInfo.Format),
	)

	h.logger.Debug(ctx).
		Str("filename", filename).
		Int("width", imageInfo.Width).
		Int("height", imageInfo.Height).
		Str("format", imageInfo.Format).
		Msg("Extracted image dimensions")

	return &imageInfo.Width, &imageInfo.Height
}

// parseTags parses comma-separated tags and returns cleaned tag names
func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}

	var tags []string
	parts := strings.Split(tagsStr, ",")
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// isSupportedImageType checks if the content type is a supported image format
func isSupportedImageType(contentType string) bool {
	supportedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return supportedTypes[contentType]
}
