package http

import (
	"encoding/json"
	"net/http"
	"tile-backend/internal/model"

	"go.uber.org/zap"
)

// respondJSON sends a JSON response with the given status code
func respondJSON(w http.ResponseWriter, logger *zap.Logger, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, logger *zap.Logger, status int, message, details string) {
	response := model.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}

	if details != "" {
		response.Details = map[string]string{"details": details}
	}

	respondJSON(w, logger, status, response)
}
