package handlers

import (
	"database/sql"
	"net/http"

	"image-gallery/internal/config"
	"image-gallery/internal/platform/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	db      *sql.DB
	storage *storage.MinIOClient
	config  *config.Config
}

func New(db *sql.DB, storage *storage.MinIOClient, config *config.Config) *Handler {
	return &Handler{
		db:      db,
		storage: storage,
		config:  config,
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
			r.Post("/", h.uploadImageHandler)
			r.Get("/", h.listImagesHandler)
			r.Get("/{id}", h.getImageHandler)
			r.Get("/{id}/download", h.downloadImageHandler)
			r.Delete("/{id}", h.deleteImageHandler)
			r.Put("/{id}", h.updateImageHandler)
		})
	})

	return r
}

func (h *Handler) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/gallery", http.StatusFound)
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
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold mb-8">Image Gallery</h1>
        <div class="mb-8">
            <h2 class="text-xl font-semibold mb-4">Upload Images</h2>
            <form hx-post="/api/images" hx-target="#gallery" hx-swap="beforeend" enctype="multipart/form-data">
                <input type="file" name="image" accept="image/*" multiple class="mb-4">
                <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">Upload</button>
            </form>
        </div>
        <div id="gallery" hx-get="/api/images" hx-trigger="load" class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
        </div>
    </div>
</body>
</html>
	`))
}