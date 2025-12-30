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

// DoorStates represents door open/closed states
type DoorStates struct {
	Top    int `json:"top"`    // 0 or 1
	Right  int `json:"right"`  // 0 or 1
	Bottom int `json:"bottom"` // 0 or 1
	Left   int `json:"left"`   // 0 or 1
}

// RoomAttributes represents room attributes
type RoomAttributes struct {
	Boss     bool `json:"boss"`
	Elite    bool `json:"elite"`
	Mob      bool `json:"mob"`
	Treasure bool `json:"treasure"`
	Teleport bool `json:"teleport"`
	Story    bool `json:"story"`
}

// DoorsConnected represents door connectivity (whether each door connects walkable areas)
type DoorsConnected struct {
	Top    bool `json:"top"`
	Right  bool `json:"right"`
	Bottom bool `json:"bottom"`
	Left   bool `json:"left"`
}

// TemplatePayload represents the complete template data as received from frontend
type TemplatePayload struct {
	Ground         Layer           `json:"ground"`
	Static         Layer           `json:"static"`
	Turret         Layer           `json:"turret"`
	MobGround      Layer           `json:"mobGround"`
	MobAir         Layer           `json:"mobAir"`
	Doors          *DoorStates     `json:"doors,omitempty"`
	Attributes     *RoomAttributes `json:"attributes,omitempty"`
	RoomType       *string         `json:"roomType,omitempty"` // "full", "bridge", or "platform"
	Meta           TemplateMeta    `json:"meta"`
}

// Template represents a complete template record
type Template struct {
	ID              uuid.UUID        `json:"id"`
	Name            string           `json:"name"`
	Version         int              `json:"version"`
	Width           int              `json:"width"`
	Height          int              `json:"height"`
	Payload         TemplatePayload  `json:"payload"`
	Thumbnail       *string          `json:"thumbnail,omitempty"` // Base64 encoded PNG
	WalkableRatio   *float64         `json:"walkable_ratio,omitempty"`
	RoomType        *string          `json:"room_type,omitempty"`
	RoomAttributes  *RoomAttributes  `json:"room_attributes,omitempty"`
	DoorsConnected  *DoorsConnected  `json:"doors_connected,omitempty"`
	StaticCount     *int             `json:"static_count,omitempty"`
	TurretCount     *int             `json:"turret_count,omitempty"`
	MobGroundCount  *int             `json:"mobground_count,omitempty"`
	MobAirCount     *int             `json:"mobair_count,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// TemplateSummary represents a template summary for list responses
type TemplateSummary struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Version        int             `json:"version"`
	Width          int             `json:"width"`
	Height         int             `json:"height"`
	Thumbnail      *string         `json:"thumbnail,omitempty"` // Base64 encoded PNG
	WalkableRatio  *float64        `json:"walkable_ratio,omitempty"`
	RoomType       *string         `json:"room_type,omitempty"`
	RoomAttributes *RoomAttributes `json:"room_attributes,omitempty"`
	DoorsConnected *DoorsConnected `json:"doors_connected,omitempty"`
	StaticCount    *int            `json:"static_count,omitempty"`
	TurretCount    *int            `json:"turret_count,omitempty"`
	MobGroundCount *int            `json:"mobground_count,omitempty"`
	MobAirCount    *int            `json:"mobair_count,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ListTemplatesQueryParams represents query parameters for listing templates
type ListTemplatesQueryParams struct {
	Limit           int
	Offset          int
	NameLike        string
	RoomType        string
	MinWalkableRatio *float64
	MaxWalkableRatio *float64
	MinStaticCount   *int
	MaxStaticCount   *int
	MinTurretCount   *int
	MaxTurretCount   *int
	MinMobGroundCount *int
	MaxMobGroundCount *int
	MinMobAirCount    *int
	MaxMobAirCount    *int
	// Room attributes filters
	HasBoss     *bool
	HasElite    *bool
	HasMob      *bool
	HasTreasure *bool
	HasTeleport *bool
	HasStory    *bool
	// Door connectivity filters
	TopDoorConnected    *bool
	RightDoorConnected  *bool
	BottomDoorConnected *bool
	LeftDoorConnected   *bool
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