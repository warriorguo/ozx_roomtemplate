package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Layer represents a 2D grid layer (0 or 1 values)
type Layer [][]int

// Point represents a 2D coordinate
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// LineSegment represents a line segment with start and end points
type LineSegment struct {
	Start Point `json:"start"`
	End   Point `json:"end"`
}

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

// StageType represents the room stage type
type StageType = string

const (
	StageStart    StageType = "start"
	StageTeaching StageType = "teaching"
	StageBuilding StageType = "building"
	StagePressure StageType = "pressure"
	StagePeak     StageType = "peak"
	StageRelease  StageType = "release"
	StageBoss     StageType = "boss"
)

// RoomAttributes represents room attributes (deprecated, kept for backward compatibility)
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
	Ground        Layer           `json:"ground"`
	SoftEdge      Layer           `json:"softEdge,omitempty"`      // Optional for backward compatibility
	Bridge        Layer           `json:"bridge,omitempty"`        // Optional for backward compatibility
	Pipeline      Layer           `json:"pipeline,omitempty"`      // Optional for backward compatibility
	PipelineLines []LineSegment   `json:"pipelineLines,omitempty"` // Line segments describing pipeline paths
	Rail          Layer           `json:"rail,omitempty"`          // Optional for backward compatibility
	RailLines     []LineSegment   `json:"railLines,omitempty"`     // Line segments describing rail paths
	Static        Layer           `json:"static"`
	Chaser        Layer           `json:"chaser,omitempty"`
	Zoner         Layer           `json:"zoner,omitempty"`
	DPS           Layer           `json:"dps,omitempty"`
	MobAir        Layer           `json:"mobAir"`
	MainPath      Layer           `json:"mainPath,omitempty"` // Main path through room center
	Doors         *DoorStates     `json:"doors,omitempty"`
	Attributes    *RoomAttributes `json:"attributes,omitempty"` // Deprecated
	StageType    *string         `json:"stageType,omitempty"`    // none, start, teaching, building, pressure, peak, release, boss
	RoomShape    *string         `json:"roomShape,omitempty"`    // "all", "bridge", or "platform"
	RoomCategory *string         `json:"roomCategory,omitempty"` // "normal", "basement", "test", "cave"
	OpenDoors    *int            `json:"openDoors,omitempty"`    // Bitmask: Top=1, Right=2, Bottom=4, Left=8
	Meta          TemplateMeta    `json:"meta"`
}

// Template represents a complete template record
type Template struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Version        int             `json:"version"`
	Width          int             `json:"width"`
	Height         int             `json:"height"`
	Payload        TemplatePayload `json:"payload"`
	Thumbnail      *string         `json:"thumbnail,omitempty"` // Base64 encoded PNG
	WalkableRatio  *float64        `json:"walkable_ratio,omitempty"`
	RoomType       *string         `json:"room_type,omitempty"`
	RoomCategory   *string         `json:"room_category,omitempty"`
	RoomAttributes *RoomAttributes `json:"room_attributes,omitempty"`
	DoorsConnected *DoorsConnected `json:"doors_connected,omitempty"`
	OpenDoors      *int            `json:"open_doors,omitempty"` // Bitmask: Top=1, Right=2, Bottom=4, Left=8
	StaticCount    *int            `json:"static_count,omitempty"`
	ChaserCount    *int            `json:"chaser_count,omitempty"`
	ZonerCount     *int            `json:"zoner_count,omitempty"`
	DPSCount       *int            `json:"dps_count,omitempty"`
	MobAirCount    *int            `json:"mobair_count,omitempty"`
	StageType      *string         `json:"stage_type,omitempty"`
	ProjectID      *uuid.UUID      `json:"project_id,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
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
	RoomCategory   *string         `json:"room_category,omitempty"`
	RoomAttributes *RoomAttributes `json:"room_attributes,omitempty"`
	DoorsConnected *DoorsConnected `json:"doors_connected,omitempty"`
	OpenDoors      *int            `json:"open_doors,omitempty"` // Bitmask: Top=1, Right=2, Bottom=4, Left=8
	StaticCount    *int            `json:"static_count,omitempty"`
	ChaserCount    *int            `json:"chaser_count,omitempty"`
	ZonerCount     *int            `json:"zoner_count,omitempty"`
	DPSCount       *int            `json:"dps_count,omitempty"`
	MobAirCount    *int            `json:"mobair_count,omitempty"`
	StageType      *string         `json:"stage_type,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ListTemplatesQueryParams represents query parameters for listing templates
type ListTemplatesQueryParams struct {
	Limit            int
	Offset           int
	NameLike         string
	RoomType         string
	MinWalkableRatio *float64
	MaxWalkableRatio *float64
	MinStaticCount   *int
	MaxStaticCount   *int
	MinChaserCount   *int
	MaxChaserCount   *int
	MinZonerCount    *int
	MaxZonerCount    *int
	MinDPSCount      *int
	MaxDPSCount      *int
	MinMobAirCount   *int
	MaxMobAirCount   *int
	StageType        string
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

// ComputeOpenDoors returns the openDoors bitmask from DoorStates (Top=1, Right=2, Bottom=4, Left=8)
func ComputeOpenDoors(doors *DoorStates) *int {
	if doors == nil {
		return nil
	}
	bitmask := doors.Top*1 + doors.Right*2 + doors.Bottom*4 + doors.Left*8
	return &bitmask
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
		// Backward compat: old payloads stored "roomType" instead of "roomShape"
		RoomType *string `json:"roomType,omitempty"`
	}{
		Alias: (*Alias)(tp),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Migrate roomType → roomShape for old payloads
	if tp.RoomShape == nil && aux.RoomType != nil {
		shape := *aux.RoomType
		if shape == "full" {
			shape = "all"
		}
		tp.RoomShape = &shape
	}

	// Compute openDoors from doors if not already set
	if tp.OpenDoors == nil && tp.Doors != nil {
		bitmask := tp.Doors.Top*1 + tp.Doors.Right*2 + tp.Doors.Bottom*4 + tp.Doors.Left*8
		tp.OpenDoors = &bitmask
	}

	return nil
}
