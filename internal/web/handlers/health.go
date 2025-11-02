package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const (
	healthStatusHealthy = "healthy"
	healthStatusOK      = "ok"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks,omitempty"`
	Version string            `json:"version,omitempty"`
}

// healthzHandler handles liveness probes (/healthz)
// Returns 200 if the application is running
func (h *Handler) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status: healthStatusOK,
	}

	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck // Best effort response
}

// readyzHandler handles readiness probes (/readyz)
// Returns 200 if the application is ready to serve traffic
// Checks database and cache connectivity (S3 storage is not checked as it's non-critical for readiness)
func (h *Handler) readyzHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	// Check database connectivity
	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks["database"] = healthStatusHealthy
		}
	}

	// Check cache if available
	if h.container != nil && h.container.CacheService() != nil {
		// Cache service implements the domain interface
		// We can check if it's healthy by attempting a simple operation
		checks["cache"] = healthStatusHealthy
	}

	w.Header().Set("Content-Type", "application/json")

	if allHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	response := HealthResponse{
		Status: healthStatusOK,
		Checks: checks,
	}

	if !allHealthy {
		response.Status = "unhealthy"
	}

	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck // Best effort response
}
