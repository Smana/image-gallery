package handlers

import (
	"database/sql"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"image-gallery/internal/config"
	"image-gallery/internal/domain/image"
	"image-gallery/internal/faults"
	"image-gallery/internal/observability"
	"image-gallery/internal/services"
	"image-gallery/internal/web/ui"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	// Legacy fields for backward compatibility
	db     *sql.DB
	config *config.Config

	// New service-based dependencies
	container      *services.Container
	imageService   image.ImageService
	tagService     image.TagService
	storageService image.StorageService

	// Demo controls (fault injection)
	demo         *faults.Service
	demoInjector *faults.Injector

	// Observability
	tracer      trace.Tracer
	httpMetrics *observability.HTTPMetrics
	logger      *observability.Logger
}

// NewWithContainer creates a handler with the dependency injection container
func NewWithContainer(container *services.Container) *Handler {
	// Initialize observability components
	tracer := observability.GetTracer()
	meter := observability.GetMeter()

	// Create HTTP metrics (ignore error for graceful degradation)
	httpMetrics, err := observability.NewHTTPMetrics(meter)
	if err != nil {
		httpMetrics = nil
	}

	return &Handler{
		// Legacy fields for backward compatibility
		db:     container.DB(),
		config: container.Config(),

		// New service-based dependencies
		container:      container,
		imageService:   container.ImageService(),
		tagService:     container.TagService(),
		storageService: container.StorageService(),

		// Demo controls (fault injection)
		demo:         container.DemoService(),
		demoInjector: container.DemoInjector(),

		// Observability
		tracer:      tracer,
		httpMetrics: httpMetrics,
		logger:      container.Logger(),
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	// One middleware: trace continuation, route-pattern span names, semconv RED metrics.
	r.Use(observability.Middleware(h.tracer, h.httpMetrics))

	if h.demoInjector != nil {
		r.Use(h.demoInjector.Middleware) // inside the server span, so faults are recorded on it
	}

	// Standard Chi middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Health check endpoints (no additional middleware for performance)
	r.Get("/healthz", h.healthzHandler)
	r.Get("/readyz", h.readyzHandler)

	r.Handle("/static/*", http.StripPrefix("/static/", staticHandler()))

	// Web routes
	r.Get("/", h.indexHandler)
	r.Get("/gallery", h.galleryHandler)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/images", func(r chi.Router) {
			r.Get("/", h.listImagesHandler)
			r.Post("/", h.uploadImagesHandler)                // Upload images endpoint
			r.Get("/{id}/view", h.viewImageHandler)           // Proxy endpoint for viewing images
			r.Get("/{id}/thumbnail", h.thumbnailImageHandler) // Thumbnail, or the original until the worker is done
			r.Delete("/{id}", h.deleteImageHandler)           // Delete image endpoint
		})
		// Settings endpoints
		r.Route("/settings", func(r chi.Router) {
			r.Get("/", h.getSettingsHandler)         // Get user settings
			r.Put("/", h.updateSettingsHandler)      // Update user settings
			r.Post("/reset", h.resetSettingsHandler) // Reset to defaults
			// Demo controls: h.demo is nil when DEMO_CONTROLS_ENABLED=false
			if h.demo != nil {
				r.Get("/demo", h.getDemoHandler)
				r.Put("/demo", h.updateDemoHandler)
				r.Post("/demo/reset", h.resetDemoHandler)
			}
		})
		// Tags endpoints
		r.Route("/tags", func(r chi.Router) {
			r.Get("/predefined", h.getPredefinedTagsHandler) // Get predefined tags
		})
		// Test endpoint for observability validation (generates traces + logs)
		r.Get("/test-db", h.testDatabaseHandler)
	})

	return r
}

// Web handlers for HTML responses

func (h *Handler) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/gallery", http.StatusFound)
}

// viewImageHandler serves images directly from storage (proxy endpoint)
//
//nolint:gocyclo // Handler with error handling and content type detection
func (h *Handler) viewImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	imageIDStr := chi.URLParam(r, "id")

	// Parse image ID
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	// Get image metadata from database
	if h.imageService == nil {
		http.Error(w, "Image service not available", http.StatusInternalServerError)
		return
	}

	img, err := h.imageService.GetImage(ctx, imageID)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Get image from storage using storage path
	reader, err := h.storageService.Retrieve(ctx, img.StoragePath)
	if err != nil {
		http.Error(w, "Failed to retrieve image", http.StatusInternalServerError)
		return
	}
	defer func() { _ = reader.Close() }() //nolint:errcheck // Resource cleanup

	// Set content type from database metadata
	if img.ContentType != "" {
		w.Header().Set("Content-Type", img.ContentType)
	} else {
		// Fallback content type based on extension
		ext := strings.ToLower(filepath.Ext(img.StoragePath))
		switch ext {
		case extJPG, extJPEG:
			w.Header().Set("Content-Type", contentTypeJPEG)
		case extPNG:
			w.Header().Set("Content-Type", contentTypePNG)
		case extGIF:
			w.Header().Set("Content-Type", contentTypeGIF)
		case extWebP:
			w.Header().Set("Content-Type", contentTypeWebP)
		default:
			w.Header().Set("Content-Type", "application/octet-stream")
		}
	}

	// Set cache headers
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Copy image data to response
	if _, err := io.Copy(w, reader); err != nil {
		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
}

// thumbnailImageHandler serves the worker's thumbnail, or the original while
// the image is pending, processing or failed.
func (h *Handler) thumbnailImageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid image id", http.StatusBadRequest)
		return
	}
	img, err := h.imageService.GetImage(r.Context(), id)
	if err != nil {
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}
	path, contentType := img.StoragePath, img.ContentType
	if img.ThumbnailPath != nil && *img.ThumbnailPath != "" {
		path, contentType = *img.ThumbnailPath, thumbnailContentType(*img.ThumbnailPath)
	}
	rc, err := h.storageService.Retrieve(r.Context(), path)
	if err != nil {
		http.Error(w, "image unavailable", http.StatusBadGateway)
		return
	}
	defer func() { _ = rc.Close() }() //nolint:errcheck // Resource cleanup
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = io.Copy(w, rc) //nolint:errcheck // response already committed; nothing actionable on copy failure
}

// Content types thumbnailContentType can return, kept as constants so adding
// this lookup does not push shared MIME-type literals over goconst's threshold.
const (
	contentTypePNG  = "image/png"
	contentTypeGIF  = "image/gif"
	contentTypeJPEG = "image/jpeg"
)

// File extensions matched both here and in viewImageHandler's fallback
// content-type switch above, kept as constants for the same goconst reason.
const (
	extJPG  = ".jpg"
	extJPEG = ".jpeg"
	extPNG  = ".png"
	extGIF  = ".gif"
	extWebP = ".webp"
)

func thumbnailContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case extPNG:
		return contentTypePNG
	case extGIF:
		return contentTypeGIF
	default:
		return contentTypeJPEG
	}
}

// galleryCSP admits https background images (a user setting) and Google
// Fonts; everything else, scripts included, must come from the binary.
const galleryCSP = "default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; " +
	"font-src https://fonts.gstatic.com; img-src 'self' https: data: blob:; connect-src 'self'; " +
	"object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'"

func (h *Handler) galleryHandler(w http.ResponseWriter, _ *http.Request) {
	hdr := w.Header()
	hdr.Set("Content-Type", "text/html; charset=utf-8")
	hdr.Set("Content-Security-Policy", galleryCSP)
	hdr.Set("X-Content-Type-Options", "nosniff")
	hdr.Set("Referrer-Policy", "same-origin")
	hdr.Set("Cache-Control", "no-cache")
	_, _ = w.Write(ui.IndexHTML()) //nolint:errcheck // response already committed; nothing actionable on write failure
}

func staticHandler() http.Handler {
	files := http.FileServerFS(ui.Static())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Embedded files carry no modification time, so revalidate rather than cache blind.
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		files.ServeHTTP(w, r)
	})
}
