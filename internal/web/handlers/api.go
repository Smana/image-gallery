package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"image-gallery/internal/domain/image"
	"image-gallery/internal/domain/settings"
	"image-gallery/internal/platform/database"
	"image-gallery/internal/services/implementations"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	scenarioError  = "error"
	statusSuccess  = "success"
	statusError    = "error"
	queryParamTrue = "true"
	defaultUserID  = "default"
	// contentTypeJPEGAlt is the nonstandard "image/jpg" variant some clients
	// send instead of contentTypeJPEG; kept as a constant for the same
	// goconst reason as contentTypeJPEG/contentTypePNG in handlers.go.
	contentTypeJPEGAlt = "image/jpg"
	contentTypeWebP    = "image/webp"
)

func (h *Handler) listImagesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := r.URL.Query()
	tagFilters := q["tags"]
	matchAll := q.Get("match_all") == queryParamTrue
	page, errPage := positiveQueryInt(q, "page", 1)
	pageSize, errSize := positiveQueryInt(q, "page_size", image.MaxPageSize)
	if errPage != nil || errSize != nil {
		http.Error(w, "page and page_size must be positive integers", http.StatusBadRequest)
		return
	}
	pageSize = min(pageSize, image.MaxPageSize)

	ctx, span := h.startSpan(ctx, "listImagesHandler",
		attribute.String("handler", "list_images"),
		attribute.Int("tag_filters.count", len(tagFilters)),
		attribute.Bool("tag_filters.match_all", matchAll),
		attribute.Int("page", page),
		attribute.Int("page_size", pageSize),
	)
	defer h.endSpan(span)

	if len(tagFilters) > 0 {
		h.setSpanAttributes(span, attribute.StringSlice("tag_filters.tags", tagFilters))
	}

	listing, err := h.getStorageImages(ctx, &image.ListImagesRequest{
		Page:     page,
		PageSize: pageSize,
		Tags:     tagFilters,
		MatchAll: matchAll,
	})
	if err != nil {
		h.handleError(ctx, span, err, "Failed to list images", "failed to get images", "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.setSpanAttributes(span, attribute.Int("images.count", len(listing.Images)))
	if span != nil {
		span.AddEvent("images_retrieved", trace.WithAttributes(
			attribute.Int("count", len(listing.Images)),
		))
	}

	listing.ActiveFilters = tagFilters
	listing.MatchAll = matchAll
	h.renderJSONResponse(w, listing)
	h.setSpanStatus(span, codes.Ok, "")
}

func positiveQueryInt(q url.Values, key string, fallback int) (int, error) {
	raw := q.Get(key)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return n, nil
}

// ImageListResponse is the body of GET /api/images.
type ImageListResponse struct {
	Images        []ImageResponse   `json:"images"`
	TotalCount    int               `json:"total_count"`
	Page          int               `json:"page"`
	PageSize      int               `json:"page_size"`
	TotalPages    int               `json:"total_pages"`
	TagColors     map[string]string `json:"tag_colors,omitempty"`
	ActiveFilters []string          `json:"active_filters,omitempty"`
	MatchAll      bool              `json:"match_all"`
}

// getStorageImages fetches one page of images from the database with metadata
func (h *Handler) getStorageImages(ctx context.Context, req *image.ListImagesRequest) (*ImageListResponse, error) {
	ctx, span := h.startSpan(ctx, "getStorageImages",
		attribute.Int("tag_filters.count", len(req.Tags)),
		attribute.Bool("tag_filters.match_all", req.MatchAll),
	)
	defer h.endSpan(span)

	// Try ImageService first (preferred path with full metadata)
	if h.imageService != nil {
		return h.getImagesFromService(ctx, span, req)
	}

	// Fallback to storage-only approach
	images, err := h.getImagesFromStorage(ctx, span)
	if err != nil {
		return nil, err
	}
	return &ImageListResponse{Images: images, TotalCount: len(images), Page: 1, PageSize: len(images), TotalPages: 1}, nil
}

// getImagesFromService fetches images using the ImageService
func (h *Handler) getImagesFromService(ctx context.Context, span trace.Span, req *image.ListImagesRequest) (*ImageListResponse, error) {
	listResp, err := h.imageService.ListImages(ctx, req)
	if err != nil {
		h.setSpanStatus(span, codes.Error, "failed to list images from service")
		if span != nil {
			span.RecordError(err)
		}
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	h.setSpanAttributes(span, attribute.Int("images.count", len(listResp.Images)))

	images := h.convertDomainImagesToResponse(listResp.Images)
	tagColors := map[string]string{}
	for _, tag := range req.Tags {
		tagColors[tag] = settings.GetTagColor(tag)
	}
	for i := range images {
		for _, tag := range images[i].Tags {
			tagColors[tag] = settings.GetTagColor(tag)
		}
	}

	h.setSpanStatus(span, codes.Ok, "")
	return &ImageListResponse{
		Images:     images,
		TotalCount: listResp.TotalCount,
		Page:       listResp.Page,
		PageSize:   listResp.PageSize,
		TotalPages: listResp.TotalPages,
		TagColors:  tagColors,
	}, nil
}

// convertDomainImagesToResponse converts domain images to API response format
func (h *Handler) convertDomainImagesToResponse(domainImages []image.Image) []ImageResponse {
	images := make([]ImageResponse, 0, len(domainImages))
	for i := range domainImages {
		img := &domainImages[i]
		// Use proxy endpoint for reliable access from browser
		viewURL := fmt.Sprintf("/api/images/%d/view", img.ID)

		// Extract tag names
		tagNames := make([]string, 0, len(img.Tags))
		for _, tag := range img.Tags {
			tagNames = append(tagNames, tag.Name)
		}

		images = append(images, ImageResponse{
			ID:              fmt.Sprintf("%d", img.ID),
			Name:            img.OriginalFilename,
			URL:             viewURL,
			ThumbnailURL:    fmt.Sprintf("/api/images/%d/thumbnail", img.ID),
			Status:          img.Status,
			ProcessingError: img.ProcessingError,
			Size:            img.FileSize,
			UploadTime:      img.UploadedAt.Format("2006-01-02 15:04:05"),
			UploadedAt:      img.UploadedAt,
			ContentType:     img.ContentType,
			Width:           img.Width,
			Height:          img.Height,
			Tags:            tagNames,
		})
	}
	return images
}

// getImagesFromStorage fetches images from storage service (fallback)
func (h *Handler) getImagesFromStorage(ctx context.Context, span trace.Span) ([]ImageResponse, error) {
	storageServiceImpl, ok := h.storageService.(*implementations.StorageServiceImpl)
	if !ok {
		err := fmt.Errorf("storage service not available")
		h.setSpanStatus(span, codes.Error, "storage service unavailable")
		if span != nil {
			span.RecordError(err)
		}
		return nil, err
	}

	objects, err := storageServiceImpl.ListObjects(ctx, "", 100)
	if err != nil {
		h.setSpanStatus(span, codes.Error, "failed to list objects")
		if span != nil {
			span.RecordError(err)
		}
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	h.setSpanAttributes(span, attribute.Int("objects.total", len(objects)))

	images := h.filterImageObjects(objects)

	h.setSpanAttributes(span, attribute.Int("images.filtered", len(images)))
	h.setSpanStatus(span, codes.Ok, "")

	return images, nil
}

// filterImageObjects converts storage objects to image responses
func (h *Handler) filterImageObjects(objects []implementations.ObjectInfo) []ImageResponse {
	images := make([]ImageResponse, 0)
	for _, obj := range objects {
		if isImageFile(obj.Key, obj.ContentType) {
			viewURL := fmt.Sprintf("/api/images/%s/view", obj.Key)
			images = append(images, ImageResponse{
				ID:         obj.Key,
				Name:       extractOriginalFilename(obj.UserMetadata, obj.Key),
				URL:        viewURL,
				Size:       obj.Size,
				UploadTime: obj.LastModified.Format("2006-01-02 15:04:05"),
				UploadedAt: obj.LastModified,
			})
		}
	}
	return images

}

// renderJSONResponse renders images as JSON for API requests
func (h *Handler) renderJSONResponse(w http.ResponseWriter, listing *ImageListResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(listing); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Helper methods to reduce cyclomatic complexity in handlers

func (h *Handler) startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	if h.tracer != nil {
		return h.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
	}
	return ctx, nil
}

func (h *Handler) endSpan(span trace.Span) {
	if span != nil {
		span.End()
	}
}

func (h *Handler) setSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span != nil {
		span.SetAttributes(attrs...)
	}
}

func (h *Handler) addSpanEvent(span trace.Span, name string) {
	if span != nil {
		span.AddEvent(name)
	}
}

func (h *Handler) setSpanStatus(span trace.Span, code codes.Code, description string, attrs ...attribute.KeyValue) {
	if span != nil {
		if len(attrs) > 0 {
			span.SetAttributes(attrs...)
		}
		span.SetStatus(code, description)
	}
}

func (h *Handler) handleError(ctx context.Context, span trace.Span, err error, logMsg, spanMsg, imagePath string) {
	if span != nil {
		span.RecordError(err)
		if spanMsg != "" {
			span.SetStatus(codes.Error, spanMsg)
		}
	}
	if h.logger != nil && logMsg != "" {
		logger := h.logger.Error(ctx).Err(err)
		if imagePath != "" {
			logger = logger.Str("image_path", imagePath)
		}
		logger.Msg(logMsg)
	}
}

type ImageResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name,omitempty"`
	URL             string    `json:"url"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	Status          string    `json:"status"`
	ProcessingError *string   `json:"processing_error,omitempty"`
	Size            int64     `json:"size"`
	UploadTime      string    `json:"upload_time"`
	UploadedAt      time.Time `json:"uploaded_at"`
	ContentType     string    `json:"content_type,omitempty"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
}

func isImageContentType(contentType string) bool {
	supportedTypes := map[string]bool{
		contentTypeJPEG:    true,
		contentTypeJPEGAlt: true,
		contentTypePNG:     true,
		contentTypeGIF:     true,
		contentTypeWebP:    true,
	}
	return supportedTypes[contentType]
}

func isImageFile(filename, contentType string) bool {
	// First check content type if it exists
	if contentType != "" && isImageContentType(contentType) {
		return true
	}

	// If content type is missing or not recognized, check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExtensions := map[string]string{
		extJPG:  contentTypeJPEG,
		extJPEG: contentTypeJPEG,
		extPNG:  contentTypePNG,
		extGIF:  contentTypeGIF,
		extWebP: contentTypeWebP,
	}
	_, isSupported := supportedExtensions[ext]
	return isSupported
}

func extractOriginalFilename(metadata map[string]string, objectKey string) string {
	if filename, exists := metadata["original-filename"]; exists {
		return filename
	}
	// Fallback to the object key (filename) if no metadata
	return filepath.Base(objectKey)
}

// testDatabaseHandler is a test endpoint to validate observability:
// - Creates trace spans for HTTP request and database queries
// - Logs at different levels including errors
// - Demonstrates full stack tracing from HTTP → Database
func (h *Handler) testDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create span for this handler
	ctx, span := h.startSpan(ctx, "testDatabaseHandler",
		attribute.String("handler", "test_database"),
		attribute.String("test.type", "observability_validation"),
	)
	defer h.endSpan(span)

	// Get scenario parameter (normal, error, slow)
	scenario := r.URL.Query().Get("scenario")
	if scenario == "" {
		scenario = "normal"
	}

	h.setSpanAttributes(span, attribute.String("test.scenario", scenario))

	response := make(map[string]any)
	response["scenario"] = scenario

	switch scenario {
	case scenarioError:
		// Simulate a database error for testing error logs and traces
		h.handleErrorScenario(ctx, span, w, response)
		return
	case "slow":
		// Simulate a slow database query
		h.handleSlowScenario(ctx, span, w, response)
		return
	default:
		// Normal scenario - successful database query
		h.handleNormalScenario(ctx, span, w, response)
		return
	}
}

func (h *Handler) handleErrorScenario(ctx context.Context, span trace.Span, w http.ResponseWriter, response map[string]any) {
	// Try to query a non-existent table to trigger a database error
	var count int
	query := "SELECT COUNT(*) FROM non_existent_table_for_testing"

	h.addSpanEvent(span, "executing_error_query")

	err := h.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		// This is expected - log the error and record in span
		h.handleError(ctx, span, err, "Database error (expected for testing)", "database_error_scenario", "")

		response["status"] = statusError
		response["error"] = err.Error()
		response["message"] = "Successfully generated database error for testing"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
			h.logger.Error(ctx).Err(encodeErr).Msg("Failed to encode error response")
		}
		return
	}
}

func (h *Handler) handleSlowScenario(ctx context.Context, span trace.Span, w http.ResponseWriter, response map[string]any) {
	// Execute a slow query using pg_sleep
	// IMPORTANT: pg_sleep() is a void function but still needs to be scanned
	// to properly consume the result and close the connection
	query := "SELECT pg_sleep(2), NOW() as current_time"

	h.addSpanEvent(span, "executing_slow_query")

	var sleepResult any // pg_sleep returns void but must be scanned
	var currentTime string
	if err := h.db.QueryRowContext(ctx, query).Scan(&sleepResult, &currentTime); err != nil {
		h.handleError(ctx, span, err, "Slow query failed", "slow_query_error", "")
		http.Error(w, "Slow query failed", http.StatusInternalServerError)
		return
	}

	h.setSpanAttributes(span,
		attribute.String("query.type", "slow"),
		attribute.Int("query.duration_seconds", 2),
	)

	response["status"] = statusSuccess
	response["message"] = "Completed slow query (2 seconds)"
	response["database_time"] = currentTime

	h.setSpanStatus(span, codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error(ctx).Err(err).Msg("Failed to encode response")
	}
}

func (h *Handler) handleNormalScenario(ctx context.Context, span trace.Span, w http.ResponseWriter, response map[string]any) {
	// Execute a normal query
	query := "SELECT version(), NOW() as current_time"

	h.addSpanEvent(span, "executing_version_query")

	var version, currentTime string
	if err := h.db.QueryRowContext(ctx, query).Scan(&version, &currentTime); err != nil {
		h.handleError(ctx, span, err, "Version query failed", "version_query_error", "")
		http.Error(w, "Version query failed", http.StatusInternalServerError)
		return
	}

	h.setSpanAttributes(span,
		attribute.String("database.version", version),
		attribute.String("query.type", "normal"),
	)

	response["status"] = statusSuccess
	response["database_version"] = version
	response["database_time"] = currentTime
	response["message"] = "Database query successful"

	h.setSpanStatus(span, codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error(ctx).Err(err).Msg("Failed to encode response")
	}
}

// deleteImageHandler deletes an image from both database and storage
func (h *Handler) deleteImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create child span for this handler
	ctx, span := h.startSpan(ctx, "DeleteImageHandler",
		attribute.String("handler", "delete_image"),
	)
	defer h.endSpan(span)

	// Parse image ID from URL
	imageIDStr := chi.URLParam(r, "id")
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		h.handleError(ctx, span, err, "Invalid image ID", "invalid_id", "")
		if h.logger != nil {
			h.logger.Warn(ctx).Str("image_id", imageIDStr).Msg("Invalid image ID format")
		}
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	h.setSpanAttributes(span, attribute.Int("image.id", imageID))
	h.addSpanEvent(span, "deleting_image")

	if h.logger != nil {
		h.logger.Info(ctx).Int("image_id", imageID).Msg("Starting image deletion")
	}

	// Get image info before deletion for logging
	img, err := h.imageService.GetImage(ctx, imageID)
	if err != nil {
		h.handleError(ctx, span, err, "Image not found", "image_not_found", "")
		if h.logger != nil {
			h.logger.Error(ctx).Err(err).Int("image_id", imageID).Msg("Failed to find image for deletion")
		}
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	h.setSpanAttributes(span,
		attribute.String("image.filename", img.Filename),
		attribute.String("image.storage_path", img.StoragePath),
	)

	// Delete the image
	if err := h.imageService.DeleteImage(ctx, imageID); err != nil {
		h.handleError(ctx, span, err, "Failed to delete image", "delete_failed", "")
		if h.logger != nil {
			h.logger.Error(ctx).Err(err).
				Int("image_id", imageID).
				Str("filename", img.Filename).
				Msg("Failed to delete image")
		}
		http.Error(w, "Failed to delete image", http.StatusInternalServerError)
		return
	}

	h.setSpanStatus(span, codes.Ok, "")
	h.addSpanEvent(span, "image_deleted_successfully")

	if h.logger != nil {
		h.logger.Info(ctx).
			Int("image_id", imageID).
			Str("filename", img.Filename).
			Str("storage_path", img.StoragePath).
			Msg("Image deleted successfully")
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"status":   "success",
		"message":  "Image deleted successfully",
		"image_id": imageID,
	}); err != nil {
		h.handleError(ctx, span, err, "Failed to encode response", "", "")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getPredefinedTagsHandler returns the list of predefined tags
func (h *Handler) getPredefinedTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create child span for this handler
	ctx, span := h.startSpan(ctx, "GetPredefinedTagsHandler",
		attribute.String("handler", "get_predefined_tags"),
	)
	defer h.endSpan(span)

	if h.logger != nil {
		h.logger.Debug(ctx).Msg("Fetching predefined tags")
	}

	// Get predefined tags directly from database repository to access description field
	dbTagRepo := database.NewTagRepository(h.db)
	tags, err := dbTagRepo.GetPredefined(ctx)
	if err != nil {
		h.handleError(ctx, span, err, "Failed to get predefined tags", "fetch_failed", "")
		if h.logger != nil {
			h.logger.Error(ctx).Err(err).Msg("Failed to fetch predefined tags")
		}
		http.Error(w, "Failed to get predefined tags", http.StatusInternalServerError)
		return
	}

	h.setSpanAttributes(span, attribute.Int("tags.count", len(tags)))
	h.setSpanStatus(span, codes.Ok, "")

	if h.logger != nil {
		h.logger.Debug(ctx).Int("count", len(tags)).Msg("Predefined tags fetched successfully")
	}

	// Convert database tags to response format with colors
	type TagResponse struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Color       string `json:"color"`
		ColorClass  string `json:"color_class"`
	}

	response := make([]TagResponse, len(tags))
	for i, tag := range tags {
		desc := ""
		if tag.Description != nil {
			desc = *tag.Description
		}
		response[i] = TagResponse{
			ID:          tag.ID,
			Name:        tag.Name,
			Description: desc,
			Color:       settings.GetTagColor(tag.Name),
			ColorClass:  settings.GetLightTagColorClass(tag.Name),
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.handleError(ctx, span, err, "Failed to encode response", "", "")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
