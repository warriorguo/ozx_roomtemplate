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
	Width         int            `json:"width"`
	Height        int            `json:"height"`
	Doors         []DoorPosition `json:"doors"`         // At least 2 doors required
	SoftEdgeCount int            `json:"softEdgeCount"` // Suggested number of soft edges to place (optional)
	RailEnabled   bool           `json:"railEnabled"`   // Whether to generate rail layer (optional)
	StaticCount   int            `json:"staticCount"`   // Suggested number of statics to place (optional)
	ChaserCount   int            `json:"chaserCount"`   // Suggested number of chasers to place (optional)
	ZonerCount    int            `json:"zonerCount"`    // Suggested number of zoners to place (optional)
	DPSCount      int            `json:"dpsCount"`      // Suggested number of DPS to place (optional)
	MobAirCount   int            `json:"mobAirCount"`   // Suggested number of mob air (fly) to place (optional)
	StageType     string         `json:"stageType"`     // Room stage type (optional)
	RoomCategory  string         `json:"roomCategory"`  // Room category: normal, basement, test, cave (optional, default: normal)
}

// BridgeGenerateResponse represents the generated template
type BridgeGenerateResponse struct {
	Payload    model.TemplatePayload `json:"payload"`
	DebugInfo  *GenerateDebugInfo    `json:"debugInfo,omitempty"`
	Difficulty *DifficultyScore      `json:"difficulty,omitempty"`
}

// GenerateDebugInfo contains debug information about the generation process
type GenerateDebugInfo struct {
	Ground      *GroundDebugInfo      `json:"ground,omitempty"`
	SoftEdge    *SoftEdgeDebugInfo    `json:"softEdge,omitempty"`
	BridgeLayer *BridgeLayerDebugInfo `json:"bridgeLayer,omitempty"`
	Rail        *RailDebugInfo        `json:"rail,omitempty"`
	MainPath    *MainPathDebugInfo    `json:"mainPath,omitempty"`
	Static      *StaticDebugInfo      `json:"static,omitempty"`
	Chaser      *EnemyLayerDebugInfo  `json:"chaser,omitempty"`
	Zoner       *EnemyLayerDebugInfo  `json:"zoner,omitempty"`
	DPS         *EnemyLayerDebugInfo  `json:"dps,omitempty"`
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

// EnemyLayerDebugInfo contains debug info for enemy layer generation (Chaser/Zoner/DPS)
type EnemyLayerDebugInfo struct {
	Skipped     bool        `json:"skipped"`
	SkipReason  string      `json:"skipReason,omitempty"`
	TargetCount int         `json:"targetCount"`
	PlacedCount int         `json:"placedCount"`
	Placements  []PlaceInfo `json:"placements"`
	Misses      []MissInfo  `json:"misses,omitempty"`
}

// MainPathDebugInfo contains debug info for main path computation
type MainPathDebugInfo struct {
	PathCellCount int      `json:"pathCellCount"`
	PathSegments  []string `json:"pathSegments,omitempty"`
	Misses        []string `json:"misses,omitempty"`
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

// RegionFilter restricts placement to a rectangular region
type RegionFilter struct {
	MinY, MaxY int // y range [MinY, MaxY)
	MinX, MaxX int // x range [MinX, MaxX)
}

// Contains checks if a point is within this region
func (r *RegionFilter) Contains(x, y int) bool {
	if r == nil {
		return true
	}
	return x >= r.MinX && x < r.MaxX && y >= r.MinY && y < r.MaxY
}

// MainPathData holds per-cell computed main path metrics
type MainPathData struct {
	Width, Height   int
	OnMainPath      [][]bool    // true if cell is on main path
	DirectDistance  [][]int     // straight-line distance to nearest main path cell
	WalkingDistance [][]int     // BFS walking distance to nearest main path cell (-1 if unreachable)
	SquishyScore    [][]float64 // walking_distance / direct_distance (higher = better for ranged)
}

// LayerContext holds all layer state to reduce parameter passing
type LayerContext struct {
	Width, Height int
	Ground        [][]int
	SoftEdge      [][]int
	Bridge        [][]int
	Rail          [][]int // nil when rail not enabled
	Static        [][]int
	Chaser        [][]int
	Zoner         [][]int
	DPS           [][]int
	MobAir        [][]int
	MainPath      *MainPathData
	DoorPositions map[DoorPosition]Point
}
