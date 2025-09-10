package handlers

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"image-gallery/internal/config"
	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/storage"
	"image-gallery/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	// Legacy fields for backward compatibility
	db      *sql.DB
	storage *storage.MinIOClient
	config  *config.Config
	
	// New service-based dependencies
	container       *services.Container
	imageService    image.ImageService
	tagService      image.TagService
	storageService  image.StorageService
}

// New creates a handler with legacy dependencies (for backward compatibility)
func New(db *sql.DB, storage *storage.MinIOClient, config *config.Config) *Handler {
	return &Handler{
		db:      db,
		storage: storage,
		config:  config,
	}
}

// NewWithContainer creates a handler with the dependency injection container
func NewWithContainer(container *services.Container) *Handler {
	return &Handler{
		// Legacy fields for backward compatibility
		db:      container.DB(),
		storage: container.StorageClient(),
		config:  container.Config(),
		
		// New service-based dependencies
		container:      container,
		imageService:   container.ImageService(),
		tagService:     container.TagService(),
		storageService: container.StorageService(),
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Web routes
	r.Get("/", h.indexHandler)
	r.Get("/gallery", h.galleryHandler)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/images", func(r chi.Router) {
			r.Get("/", h.listImagesHandler)
			r.Get("/{id}", h.getImageHandler)
			r.Get("/{id}/view", h.viewImageHandler) // Proxy endpoint for viewing images
		})
	})

	return r
}

// Web handlers for HTML responses

func (h *Handler) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/gallery", http.StatusFound)
}

// viewImageHandler serves images directly from storage (proxy endpoint)
func (h *Handler) viewImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	imagePath := chi.URLParam(r, "id")
	
	// Get image from storage
	reader, err := h.storageService.Retrieve(ctx, imagePath)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	defer reader.Close()
	
	// Get file info for content type
	fileInfo, err := h.storageService.GetFileInfo(ctx, imagePath)
	if err == nil && fileInfo.ContentType != "" {
		w.Header().Set("Content-Type", fileInfo.ContentType)
	} else {
		// Fallback content type based on extension
		ext := strings.ToLower(filepath.Ext(imagePath))
		switch ext {
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".webp":
			w.Header().Set("Content-Type", "image/webp")
		default:
			w.Header().Set("Content-Type", "application/octet-stream")
		}
	}
	
	// Set cache headers
	w.Header().Set("Cache-Control", "public, max-age=3600")
	
	// Copy image data to response
	io.Copy(w, reader)
}

func (h *Handler) galleryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Image Gallery</title>
    <script src="https://unpkg.com/htmx.org@2.0.3"></script>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    <style>
        .image-card {
            transition: transform 0.2s ease-in-out;
        }
        .image-card:hover {
            transform: scale(1.02);
        }
        .gallery-image {
            width: 100%;
            height: 200px;
            object-fit: cover;
            border-radius: 0.5rem;
        }
    </style>
</head>
<body class="bg-gray-50">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-4xl font-bold text-center mb-8 text-gray-800">Image Gallery</h1>
        <div id="gallery" 
             hx-get="/api/images" 
             hx-trigger="load" 
             hx-target="this" 
             hx-swap="innerHTML"
             class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
            <div class="text-center text-gray-500">Loading images...</div>
        </div>
    </div>
    
    <!-- Image viewer modal -->
    <div id="imageModal" class="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 hidden">
        <div class="max-w-4xl max-h-full p-4">
            <img id="modalImage" src="" alt="" class="max-w-full max-h-full object-contain">
            <div class="text-white text-center mt-4">
                <p id="modalImageName" class="font-semibold"></p>
                <p id="modalImageSize" class="text-sm text-gray-300"></p>
                <button onclick="closeModal()" class="mt-2 px-4 py-2 bg-gray-700 text-white rounded hover:bg-gray-600">Close</button>
            </div>
        </div>
    </div>
    
    <script>
        function openModal(imageUrl, imageName, imageSize) {
            document.getElementById('modalImage').src = imageUrl;
            document.getElementById('modalImageName').textContent = imageName;
            document.getElementById('modalImageSize').textContent = 'Size: ' + formatFileSize(imageSize);
            document.getElementById('imageModal').classList.remove('hidden');
        }
        
        function closeModal() {
            document.getElementById('imageModal').classList.add('hidden');
        }
        
        function formatFileSize(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }
        
        // Close modal on escape key
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                closeModal();
            }
        });
        
        // Close modal on click outside
        document.getElementById('imageModal').addEventListener('click', function(event) {
            if (event.target === this) {
                closeModal();
            }
        });
    </script>
</body>
</html>
	`))
}

