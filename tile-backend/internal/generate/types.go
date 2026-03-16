package generate

import "tile-backend/internal/model"

// DoorPosition represents which doors need to be connected
type DoorPosition string

const (
	DoorTop    DoorPosition = "top"
	DoorRight  DoorPosition = "right"
	DoorBottom DoorPosition = "bottom"
	DoorLeft   DoorPosition = "left"
)

// BridgeGenerateRequest represents the request for generating a bridge room
type BridgeGenerateRequest struct {
	Width          int            `json:"width"`
	Height         int            `json:"height"`
	Doors          []DoorPosition `json:"doors"`          // At least 2 doors required
	SoftEdgeCount  int            `json:"softEdgeCount"`  // Suggested number of soft edges to place (optional)
	RailEnabled    bool           `json:"railEnabled"`    // Whether to generate rail layer (optional)
	StaticCount    int            `json:"staticCount"`    // Suggested number of statics to place (optional)
	TurretCount    int            `json:"turretCount"`    // Suggested number of turrets to place (optional)
	MobGroundCount int            `json:"mobGroundCount"` // Suggested number of mob ground to place (optional)
	MobAirCount    int            `json:"mobAirCount"`    // Suggested number of mob air (fly) to place (optional)
}

// BridgeGenerateResponse represents the generated template
type BridgeGenerateResponse struct {
	Payload   model.TemplatePayload `json:"payload"`
	DebugInfo *GenerateDebugInfo    `json:"debugInfo,omitempty"`
}

// GenerateDebugInfo contains debug information about the generation process
type GenerateDebugInfo struct {
	Ground      *GroundDebugInfo      `json:"ground,omitempty"`
	SoftEdge    *SoftEdgeDebugInfo    `json:"softEdge,omitempty"`
	BridgeLayer *BridgeLayerDebugInfo `json:"bridgeLayer,omitempty"`
	Rail        *RailDebugInfo        `json:"rail,omitempty"`
	Static      *StaticDebugInfo      `json:"static,omitempty"`
	Turret      *TurretDebugInfo      `json:"turret,omitempty"`
	MobGround   *MobGroundDebugInfo   `json:"mobGround,omitempty"`
	MobAir      *MobAirDebugInfo      `json:"mobAir,omitempty"`
}

// BridgeLayerDebugInfo contains debug info for bridge layer generation
type BridgeLayerDebugInfo struct {
	Skipped           bool               `json:"skipped"`
	SkipReason        string             `json:"skipReason,omitempty"`
	IslandsFound      int                `json:"islandsFound"`
	BridgesPlaced     int                `json:"bridgesPlaced"`
	Connections       []BridgeConnection `json:"connections"`
	ConcaveGapBridges []BridgeConnection `json:"concaveGapBridges,omitempty"`
	Misses            []MissInfo         `json:"misses,omitempty"`
}

// BridgeConnection describes a bridge connection between island and ground/island
type BridgeConnection struct {
	From     string `json:"from"`     // Island position and size
	To       string `json:"to"`       // Target (ground or another island)
	Position string `json:"position"` // Bridge position
	Size     string `json:"size"`     // Bridge size (2x2)
}

// SoftEdgeDebugInfo contains debug info for soft edge layer generation
type SoftEdgeDebugInfo struct {
	Skipped     bool        `json:"skipped"`
	SkipReason  string      `json:"skipReason,omitempty"`
	TargetCount int         `json:"targetCount"`
	PlacedCount int         `json:"placedCount"`
	Placements  []PlaceInfo `json:"placements"`
	Misses      []MissInfo  `json:"misses,omitempty"`
}

// MissInfo describes a failed placement attempt
type MissInfo struct {
	Reason string `json:"reason"`
	Count  int    `json:"count,omitempty"` // Number of times this reason occurred
}

// GroundDebugInfo contains debug info for ground layer generation
type GroundDebugInfo struct {
	DoorConnections []DoorConnectionInfo `json:"doorConnections"`
	Platforms       []PlatformInfo       `json:"platforms"`
	FloatingIslands []FloatingIslandInfo `json:"floatingIslands,omitempty"`
}

// FloatingIslandInfo describes a floating island placement
type FloatingIslandInfo struct {
	Position   string `json:"position"` // Center position of the island
	Size       string `json:"size"`     // Size of the island (WxH)
	FromArea   string `json:"fromArea"` // The empty area it was placed in
	Skipped    bool   `json:"skipped,omitempty"`
	SkipReason string `json:"skipReason,omitempty"`
}

// DoorConnectionInfo describes a door connection
type DoorConnectionInfo struct {
	From      string `json:"from"`
	To        string `json:"to"`
	PathType  string `json:"pathType"` // "direct" or "L-shaped"
	BrushSize string `json:"brushSize"`
}

// PlatformInfo describes a platform placement
type PlatformInfo struct {
	Strategy  string   `json:"strategy"`
	BrushSize string   `json:"brushSize"`
	Points    []string `json:"points"`
	Mirror    string   `json:"mirror"`
}

// StaticDebugInfo contains debug info for static layer generation
type StaticDebugInfo struct {
	Skipped     bool        `json:"skipped"`
	SkipReason  string      `json:"skipReason,omitempty"`
	TargetCount int         `json:"targetCount"`
	PlacedCount int         `json:"placedCount"`
	Placements  []PlaceInfo `json:"placements"`
	Misses      []MissInfo  `json:"misses,omitempty"`
}

// TurretDebugInfo contains debug info for turret layer generation
type TurretDebugInfo struct {
	Skipped     bool        `json:"skipped"`
	SkipReason  string      `json:"skipReason,omitempty"`
	TargetCount int         `json:"targetCount"`
	PlacedCount int         `json:"placedCount"`
	Placements  []PlaceInfo `json:"placements"`
	Misses      []MissInfo  `json:"misses,omitempty"`
}

// MobGroundDebugInfo contains debug info for mob ground layer generation
type MobGroundDebugInfo struct {
	Skipped     bool           `json:"skipped"`
	SkipReason  string         `json:"skipReason,omitempty"`
	TargetCount int            `json:"targetCount"`
	PlacedCount int            `json:"placedCount"`
	Groups      []MobGroupInfo `json:"groups"`
	Misses      []MissInfo     `json:"misses,omitempty"`
}

// MobGroupInfo describes a placement group
type MobGroupInfo struct {
	GroupIndex  int         `json:"groupIndex"`
	Strategy    string      `json:"strategy"`
	TargetCount int         `json:"targetCount"`
	PlacedCount int         `json:"placedCount"`
	Placements  []PlaceInfo `json:"placements"`
	Misses      []MissInfo  `json:"misses,omitempty"`
}

// MobAirDebugInfo contains debug info for mob air layer generation
type MobAirDebugInfo struct {
	Skipped     bool        `json:"skipped"`
	SkipReason  string      `json:"skipReason,omitempty"`
	TargetCount int         `json:"targetCount"`
	PlacedCount int         `json:"placedCount"`
	Strategy    string      `json:"strategy"`
	Placements  []PlaceInfo `json:"placements"`
	Misses      []MissInfo  `json:"misses,omitempty"`
}

// PlaceInfo describes a single placement
type PlaceInfo struct {
	Position string `json:"position"`
	Size     string `json:"size"`
	Reason   string `json:"reason,omitempty"` // Why this position was chosen
}

// Point represents a coordinate
type Point struct {
	X, Y int
}

// BrushSize represents brush dimensions
type BrushSize struct {
	Width, Height int
}

// MirrorAxis represents the axis to mirror around
type MirrorAxis int

const (
	MirrorNone MirrorAxis = iota
	MirrorX               // Mirror top-bottom (across horizontal center line)
	MirrorY               // Mirror left-right (across vertical center line)
)

// Strategy represents a platform placement strategy with weight
type Strategy struct {
	Name   string
	Weight int
	Points []Point
	Mirror MirrorAxis // Which axis to mirror around
}

var connectionBrushes = []BrushSize{
	{2, 2} /*{3, 3},*/, {4, 4},
}

var platformBrushes = []BrushSize{
	/*{2, 2}, {2, 3}, {3, 3}, {3, 2},
	{4, 3}, {3, 4}, {4, 4}, {4, 5},
	{5, 4}, {5, 5},*/
	{4, 4}, {6, 6},
}

// LayerContext holds all layer state to reduce parameter passing
type LayerContext struct {
	Width, Height int
	Ground        [][]int
	SoftEdge      [][]int
	Bridge        [][]int
	Rail          [][]int // nil when rail not enabled
	Static        [][]int
	Turret        [][]int
	MobGround     [][]int
	MobAir        [][]int
	DoorPositions map[DoorPosition]Point
}
