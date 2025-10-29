package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"image-gallery/internal/services/implementations"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) listImagesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create child span for this handler
	var span trace.Span
	if h.tracer != nil {
		ctx, span = h.tracer.Start(ctx, "listImagesHandler",
			trace.WithAttributes(
				attribute.String("handler", "list_images"),
				attribute.Bool("htmx_request", r.Header.Get("HX-Request") != ""),
			),
		)
		defer span.End()
	}

	images, err := h.getStorageImages(ctx)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to get images")
		}
		if h.logger != nil {
			h.logger.Error(ctx).Err(err).Msg("Failed to list images")
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if span != nil {
		span.SetAttributes(attribute.Int("images.count", len(images)))
		span.AddEvent("images_retrieved", trace.WithAttributes(
			attribute.Int("count", len(images)),
		))
	}

	// Check if this is an HTMX request for HTML or regular JSON request
	if r.Header.Get("HX-Request") != "" {
		h.renderHTMLResponse(w, images)
	} else {
		h.renderJSONResponse(w, images)
	}

	if span != nil {
		span.SetStatus(codes.Ok, "")
	}
}

// getStorageImages fetches and filters images from storage
func (h *Handler) getStorageImages(ctx context.Context) ([]ImageResponse, error) {
	var span trace.Span
	if h.tracer != nil {
		ctx, span = h.tracer.Start(ctx, "getStorageImages")
		defer span.End()
	}

	storageServiceImpl, ok := h.storageService.(*implementations.StorageServiceImpl)
	if !ok {
		err := fmt.Errorf("storage service not available")
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "storage service unavailable")
		}
		return nil, err
	}

	objects, err := storageServiceImpl.ListObjects(ctx, "", 100)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to list objects")
		}
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	if span != nil {
		span.SetAttributes(attribute.Int("objects.total", len(objects)))
	}

	images := h.filterImageObjects(objects)

	if span != nil {
		span.SetAttributes(attribute.Int("images.filtered", len(images)))
		span.SetStatus(codes.Ok, "")
	}

	return images, nil
}

// filterImageObjects converts storage objects to image responses
func (h *Handler) filterImageObjects(objects []implementations.ObjectInfo) []ImageResponse {
	images := make([]ImageResponse, 0)
	for _, obj := range objects {
		if isImageFile(obj.Key, obj.ContentType) {
			url := fmt.Sprintf("/api/images/%s/view", obj.Key)
			images = append(images, ImageResponse{
				ID:         obj.Key,
				Name:       extractOriginalFilename(obj.UserMetadata, obj.Key),
				URL:        url,
				Size:       obj.Size,
				UploadTime: obj.LastModified.Format("2006-01-02 15:04:05"),
			})
		}
	}
	return images

}

// renderHTMLResponse renders images as HTML for HTMX requests
func (h *Handler) renderHTMLResponse(w http.ResponseWriter, images []ImageResponse) {
	w.Header().Set("Content-Type", "text/html")
	if len(images) == 0 {
		if _, err := w.Write([]byte(`<div class="col-span-full text-center text-gray-500 py-8">No images found in storage</div>`)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
		}
		return
	}

	for _, img := range images {
		html := fmt.Sprintf(`
			<div class="image-card bg-white rounded-lg shadow-md overflow-hidden">
				<img src="%s" alt="%s" class="gallery-image cursor-pointer"
					 onclick="openModal('%s', '%s', %d)">
				<div class="p-3">
					<h3 class="font-medium text-gray-900 truncate">%s</h3>
					<p class="text-sm text-gray-500">%s</p>
					<p class="text-xs text-gray-400">%s</p>
				</div>
			</div>`,
			img.URL, img.Name, img.URL, img.Name, img.Size,
			img.Name, formatFileSize(img.Size), img.UploadTime)
		if _, err := w.Write([]byte(html)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

// renderJSONResponse renders images as JSON for API requests
func (h *Handler) renderJSONResponse(w http.ResponseWriter, images []ImageResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"images":      images,
		"total_count": len(images),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) getImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	imagePath := chi.URLParam(r, "id")

	// Create child span for this handler
	ctx, span := h.startSpan(ctx, "getImageHandler",
		attribute.String("handler", "get_image"),
		attribute.String("image.path", imagePath),
	)
	defer h.endSpan(span)

	// Check if image exists
	exists, err := h.storageService.Exists(ctx, imagePath)
	if err != nil {
		h.handleError(ctx, span, err, "Error checking image existence", "failed to check existence", imagePath)
		http.Error(w, "Error checking image", http.StatusInternalServerError)
		return
	}

	if !exists {
		h.setSpanStatus(span, codes.Error, "image not found", attribute.Bool("image.found", false))
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	h.setSpanAttributes(span, attribute.Bool("image.found", true))

	// Generate presigned URL
	url, err := h.storageService.GenerateURL(ctx, imagePath, 3600)
	if err != nil {
		h.handleError(ctx, span, err, "Failed to generate image URL", "failed to generate URL", imagePath)
		http.Error(w, "Failed to generate image URL", http.StatusInternalServerError)
		return
	}

	// Get image metadata
	fileInfo, err := h.storageService.GetFileInfo(ctx, imagePath)
	if err != nil {
		h.handleError(ctx, span, err, "Failed to get image info", "failed to get file info", imagePath)
		http.Error(w, "Failed to get image info", http.StatusInternalServerError)
		return
	}

	h.setSpanAttributes(span,
		attribute.Int64("image.size", fileInfo.Size),
		attribute.String("image.content_type", fileInfo.ContentType),
	)
	h.addSpanEvent(span, "image_info_retrieved")
	h.setSpanStatus(span, codes.Ok, "")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ImageResponse{
		ID:         imagePath,
		URL:        url,
		Size:       fileInfo.Size,
		UploadTime: time.Unix(fileInfo.LastModified, 0).Format("2006-01-02 15:04:05"),
	}); err != nil {
		h.handleError(ctx, span, err, "Failed to encode response", "", "")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
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
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	URL        string `json:"url"`
	Size       int64  `json:"size"`
	UploadTime string `json:"upload_time"`
}

func isImageContentType(contentType string) bool {
	supportedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
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
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
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

func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 Bytes"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
