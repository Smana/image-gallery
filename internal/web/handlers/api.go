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
)

func (h *Handler) listImagesHandler(w http.ResponseWriter, r *http.Request) {
	images, err := h.getStorageImages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request for HTML or regular JSON request
	if r.Header.Get("HX-Request") != "" {
		h.renderHTMLResponse(w, images)
	} else {
		h.renderJSONResponse(w, images)
	}
}

// getStorageImages fetches and filters images from storage
func (h *Handler) getStorageImages() ([]ImageResponse, error) {
	ctx := context.Background()

	storageServiceImpl, ok := h.storageService.(*implementations.StorageServiceImpl)
	if !ok {
		return nil, fmt.Errorf("storage service not available")
	}

	fmt.Printf("DEBUG: Calling ListObjects with prefix='' maxKeys=100\n")
	objects, err := storageServiceImpl.ListObjects(ctx, "", 100)
	if err != nil {
		fmt.Printf("DEBUG: ListObjects error: %v\n", err)
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	fmt.Printf("DEBUG: Found %d objects\n", len(objects))

	return h.filterImageObjects(objects), nil
}

// filterImageObjects converts storage objects to image responses
func (h *Handler) filterImageObjects(objects []implementations.ObjectInfo) []ImageResponse {
	images := make([]ImageResponse, 0)
	for _, obj := range objects {
		fmt.Printf("DEBUG: Checking object %s with content type: %s\n", obj.Key, obj.ContentType)
		if isImageFile(obj.Key, obj.ContentType) {
			fmt.Printf("DEBUG: Object %s passed content type filter\n", obj.Key)
			url := fmt.Sprintf("/api/images/%s/view", obj.Key)

			fmt.Printf("DEBUG: Adding image %s to results\n", obj.Key)
			images = append(images, ImageResponse{
				ID:         obj.Key,
				Name:       extractOriginalFilename(obj.UserMetadata, obj.Key),
				URL:        url,
				Size:       obj.Size,
				UploadTime: obj.LastModified.Format("2006-01-02 15:04:05"),
			})
		} else {
			fmt.Printf("DEBUG: Object %s rejected - unsupported content type: %s\n", obj.Key, obj.ContentType)
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
	ctx := context.Background()
	imagePath := chi.URLParam(r, "id")

	// Check if image exists
	exists, err := h.storageService.Exists(ctx, imagePath)
	if err != nil {
		http.Error(w, "Error checking image", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Generate presigned URL
	url, err := h.storageService.GenerateURL(ctx, imagePath, 3600)
	if err != nil {
		http.Error(w, "Failed to generate image URL", http.StatusInternalServerError)
		return
	}

	// Get image metadata
	fileInfo, err := h.storageService.GetFileInfo(ctx, imagePath)
	if err != nil {
		http.Error(w, "Failed to get image info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ImageResponse{
		ID:         imagePath,
		URL:        url,
		Size:       fileInfo.Size,
		UploadTime: time.Unix(fileInfo.LastModified, 0).Format("2006-01-02 15:04:05"),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
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
