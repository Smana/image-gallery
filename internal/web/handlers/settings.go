package handlers

import (
	"encoding/json"
	"net/http"

	"image-gallery/internal/domain/settings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// getSettingsHandler retrieves user settings (GET /api/settings)
func (h *Handler) getSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create span for this handler
	ctx, span := h.startSpan(ctx, "GetSettingsHandler",
		attribute.String("handler", "get_settings"),
	)
	defer h.endSpan(span)

	// Get user ID from query parameter (optional)
	userID := r.URL.Query().Get("user_id")
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
		h.setSpanAttributes(span, attribute.String("settings.user_id", userID))
	} else {
		h.setSpanAttributes(span, attribute.String("settings.user_id", defaultUserID))
	}

	// Get settings from service
	req := &settings.GetSettingsRequest{
		UserID: userIDPtr,
	}

	result, err := h.container.SettingsService().GetSettings(ctx, req)
	if err != nil {
		h.handleError(ctx, span, err, "Failed to get settings", "failed to get settings", "")
		http.Error(w, "Failed to retrieve settings", http.StatusInternalServerError)
		return
	}

	h.setSpanAttributes(span, attribute.Int("settings.id", result.ID))
	h.addSpanEvent(span, "settings_retrieved")
	h.setSpanStatus(span, codes.Ok, "")

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.handleError(ctx, span, err, "Failed to encode response", "", "")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if h.logger != nil {
		h.logger.Info(ctx).
			Int("settings_id", result.ID).
			Str("user_id", func() string {
				if userIDPtr != nil {
					return *userIDPtr
				}
				return defaultUserID
			}()).
			Msg("Settings retrieved successfully")
	}
}

// updateSettingsHandler updates user settings (PUT /api/settings)
func (h *Handler) updateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create span for this handler
	ctx, span := h.startSpan(ctx, "UpdateSettingsHandler",
		attribute.String("handler", "update_settings"),
	)
	defer h.endSpan(span)

	// Parse request body
	var req settings.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(ctx, span, err, "Failed to decode request body", "invalid request body", "")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := defaultUserID
	if req.UserID != nil {
		userID = *req.UserID
	}
	h.setSpanAttributes(span, attribute.String("settings.user_id", userID))
	h.addSpanEvent(span, "request_decoded")

	if h.logger != nil {
		h.logger.Info(ctx).
			Str("user_id", userID).
			Msg("Updating settings")
	}

	// Update settings via service
	result, err := h.container.SettingsService().UpdateSettings(ctx, &req)
	if err != nil {
		h.handleError(ctx, span, err, "Failed to update settings", "failed to update settings", "")
		http.Error(w, "Failed to update settings: "+err.Error(), http.StatusBadRequest)
		return
	}

	h.setSpanAttributes(span, attribute.Int("settings.id", result.ID))
	h.addSpanEvent(span, "settings_updated")
	h.setSpanStatus(span, codes.Ok, "")

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.handleError(ctx, span, err, "Failed to encode response", "", "")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if h.logger != nil {
		h.logger.Info(ctx).
			Int("settings_id", result.ID).
			Str("user_id", userID).
			Msg("Settings updated successfully")
	}
}

// resetSettingsHandler resets user settings to defaults (POST /api/settings/reset)
func (h *Handler) resetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create span for this handler
	ctx, span := h.startSpan(ctx, "ResetSettingsHandler",
		attribute.String("handler", "reset_settings"),
	)
	defer h.endSpan(span)

	// Get user ID from query parameter (optional)
	userID := r.URL.Query().Get("user_id")
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
		h.setSpanAttributes(span, attribute.String("settings.user_id", userID))
	} else {
		h.setSpanAttributes(span, attribute.String("settings.user_id", defaultUserID))
	}

	if h.logger != nil {
		h.logger.Info(ctx).
			Str("user_id", func() string {
				if userIDPtr != nil {
					return *userIDPtr
				}
				return defaultUserID
			}()).
			Msg("Resetting settings")
	}

	// Reset settings via service
	result, err := h.container.SettingsService().ResetSettings(ctx, userIDPtr)
	if err != nil {
		h.handleError(ctx, span, err, "Failed to reset settings", "failed to reset settings", "")
		http.Error(w, "Failed to reset settings", http.StatusInternalServerError)
		return
	}

	h.setSpanAttributes(span, attribute.Int("settings.id", result.ID))
	h.addSpanEvent(span, "settings_reset")
	h.setSpanStatus(span, codes.Ok, "")

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.handleError(ctx, span, err, "Failed to encode response", "", "")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if h.logger != nil {
		h.logger.Info(ctx).
			Int("settings_id", result.ID).
			Str("user_id", func() string {
				if userIDPtr != nil {
					return *userIDPtr
				}
				return defaultUserID
			}()).
			Msg("Settings reset successfully")
	}
}
