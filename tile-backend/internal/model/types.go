package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Layer represents a 2D grid layer (0 or 1 values)
type Layer [][]int

// TemplateMeta represents the metadata for a template
type TemplateMeta struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

// TemplatePayload represents the complete template data as received from frontend
type TemplatePayload struct {
	Ground    Layer        `json:"ground"`
	Static    Layer        `json:"static"`
	Turret    Layer        `json:"turret"`
	MobGround Layer        `json:"mobGround"`
	MobAir    Layer        `json:"mobAir"`
	Meta      TemplateMeta `json:"meta"`
}

// Template represents a complete template record
type Template struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Version   int             `json:"version"`
	Width     int             `json:"width"`
	Height    int             `json:"height"`
	Payload   TemplatePayload `json:"payload"`
	Thumbnail *string         `json:"thumbnail,omitempty"` // Base64 encoded PNG
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TemplateSummary represents a template summary for list responses
type TemplateSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Thumbnail *string   `json:"thumbnail,omitempty"` // Base64 encoded PNG
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateTemplateRequest represents the request body for creating a template
type CreateTemplateRequest struct {
	Name      string          `json:"name"`
	Payload   TemplatePayload `json:"payload"`
	Thumbnail *string         `json:"thumbnail,omitempty"` // Base64 encoded PNG
}

// CreateTemplateResponse represents the response after creating a template
type CreateTemplateResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListTemplatesResponse represents the response for listing templates
type ListTemplatesResponse struct {
	Total int               `json:"total"`
	Items []TemplateSummary `json:"items"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationError represents a validation error for a specific cell
type ValidationError struct {
	Layer  string `json:"layer"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Reason string `json:"reason"`
}

// ValidationResult represents the result of template validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// Custom JSON marshaling for JSONB storage
func (tp *TemplatePayload) MarshalJSON() ([]byte, error) {
	type Alias TemplatePayload
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(tp),
	})
}

func (tp *TemplatePayload) UnmarshalJSON(data []byte) error {
	type Alias TemplatePayload
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(tp),
	}
	return json.Unmarshal(data, aux)
}