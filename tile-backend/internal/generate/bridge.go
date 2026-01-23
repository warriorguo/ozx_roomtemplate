package generate

import (
	"fmt"
	"math/rand"
	"tile-backend/internal/model"
)

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
	Static      *StaticDebugInfo      `json:"static,omitempty"`
	Turret      *TurretDebugInfo      `json:"turret,omitempty"`
	MobGround   *MobGroundDebugInfo   `json:"mobGround,omitempty"`
	MobAir      *MobAirDebugInfo      `json:"mobAir,omitempty"`
}

// BridgeLayerDebugInfo contains debug info for bridge layer generation
type BridgeLayerDebugInfo struct {
	Skipped        bool              `json:"skipped"`
	SkipReason     string            `json:"skipReason,omitempty"`
	IslandsFound   int               `json:"islandsFound"`
	BridgesPlaced  int               `json:"bridgesPlaced"`
	Connections    []BridgeConnection `json:"connections"`
	Misses         []MissInfo        `json:"misses,omitempty"`
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

// GenerateBridgeRoom generates a bridge-type room template
func GenerateBridgeRoom(req BridgeGenerateRequest) (*BridgeGenerateResponse, error) {
	// Validate input
	if req.Width < 4 || req.Width > 200 || req.Height < 4 || req.Height > 200 {
		return nil, fmt.Errorf("invalid dimensions: width and height must be between 4 and 200")
	}
	if len(req.Doors) < 2 {
		return nil, fmt.Errorf("at least 2 doors are required for bridge generation")
	}

	// Validate doors are unique
	doorSet := make(map[DoorPosition]bool)
	for _, door := range req.Doors {
		if doorSet[door] {
			return nil, fmt.Errorf("duplicate door position: %s", door)
		}
		doorSet[door] = true
	}

	// Initialize debug info
	debugInfo := &GenerateDebugInfo{}

	// Initialize empty ground layer
	ground := make([][]int, req.Height)
	for y := 0; y < req.Height; y++ {
		ground[y] = make([]int, req.Width)
	}

	// Get door center positions
	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Step 1: Connect all doors
	groundDebug := &GroundDebugInfo{}
	connectDoorsWithDebug(ground, doorPositions, req.Width, req.Height, groundDebug)

	// Step 2: Draw small platforms
	drawPlatformsWithDebug(ground, req.Width, req.Height, req.Doors, doorPositions, groundDebug)

	// Step 2.5: Draw floating islands in void areas (50% probability per island)
	drawFloatingIslandsWithDebug(ground, req.Width, req.Height, groundDebug)
	debugInfo.Ground = groundDebug

	// Create empty layers for other layers
	emptyLayer := createEmptyLayer(req.Width, req.Height)

	// Build door states
	doorStates := &model.DoorStates{
		Top:    boolToInt(doorSet[DoorTop]),
		Right:  boolToInt(doorSet[DoorRight]),
		Bottom: boolToInt(doorSet[DoorBottom]),
		Left:   boolToInt(doorSet[DoorLeft]),
	}

	// Step 3: Generate soft edge layer if requested
	softEdgeLayer := copyLayer(emptyLayer)
	if req.SoftEdgeCount > 0 {
		softEdgeDebug := generateSoftEdgeLayerWithDebug(softEdgeLayer, ground, doorPositions, req.Width, req.Height, req.SoftEdgeCount)
		debugInfo.SoftEdge = softEdgeDebug
	} else {
		debugInfo.SoftEdge = &SoftEdgeDebugInfo{
			Skipped:    true,
			SkipReason: "softEdgeCount is 0 or not specified",
		}
	}

	// Step 3.5: Generate bridge layer to connect floating islands
	bridgeLayer := copyLayer(emptyLayer)
	bridgeLayerDebug := generateBridgeLayerWithDebug(bridgeLayer, ground, req.Width, req.Height)
	debugInfo.BridgeLayer = bridgeLayerDebug

	// Step 4: Generate static layer if requested
	staticLayer := copyLayer(emptyLayer)
	if req.StaticCount > 0 {
		staticDebug := generateStaticLayerWithDebug(staticLayer, ground, softEdgeLayer, bridgeLayer, doorPositions, req.Width, req.Height, req.StaticCount)
		debugInfo.Static = staticDebug
	} else {
		debugInfo.Static = &StaticDebugInfo{
			Skipped:    true,
			SkipReason: "staticCount is 0 or not specified",
		}
	}

	// Step 5: Generate turret layer if requested
	turretLayer := copyLayer(emptyLayer)
	if req.TurretCount > 0 {
		turretDebug := generateTurretLayerWithDebug(turretLayer, ground, softEdgeLayer, bridgeLayer, staticLayer, doorPositions, req.Width, req.Height, req.TurretCount)
		debugInfo.Turret = turretDebug
	} else {
		debugInfo.Turret = &TurretDebugInfo{
			Skipped:    true,
			SkipReason: "turretCount is 0 or not specified",
		}
	}

	// Step 6: Generate mob ground layer if requested
	mobGroundLayer := copyLayer(emptyLayer)
	if req.MobGroundCount > 0 {
		mobGroundDebug := generateMobGroundLayerWithDebug(mobGroundLayer, ground, softEdgeLayer, bridgeLayer, staticLayer, turretLayer, doorPositions, req.Width, req.Height, req.MobGroundCount)
		debugInfo.MobGround = mobGroundDebug
	} else {
		debugInfo.MobGround = &MobGroundDebugInfo{
			Skipped:    true,
			SkipReason: "mobGroundCount is 0 or not specified",
		}
	}

	// Step 7: Generate mob air layer if requested
	mobAirLayer := copyLayer(emptyLayer)
	if req.MobAirCount > 0 {
		mobAirDebug := generateMobAirLayerWithDebug(mobAirLayer, ground, softEdgeLayer, bridgeLayer, staticLayer, turretLayer, mobGroundLayer, doorPositions, req.Width, req.Height, req.MobAirCount)
		debugInfo.MobAir = mobAirDebug
	} else {
		debugInfo.MobAir = &MobAirDebugInfo{
			Skipped:    true,
			SkipReason: "mobAirCount is 0 or not specified",
		}
	}

	// Build payload
	roomType := "bridge"
	payload := model.TemplatePayload{
		Ground:    ground,
		SoftEdge:  softEdgeLayer,
		Bridge:    bridgeLayer,
		Static:    staticLayer,
		Turret:    turretLayer,
		MobGround: mobGroundLayer,
		MobAir:    mobAirLayer,
		Doors:     doorStates,
		RoomType:  &roomType,
		Meta: model.TemplateMeta{
			Name:    fmt.Sprintf("bridge-%dx%d", req.Width, req.Height),
			Version: 1,
			Width:   req.Width,
			Height:  req.Height,
		},
	}

	return &BridgeGenerateResponse{Payload: payload, DebugInfo: debugInfo}, nil
}

// getDoorCenterPositions returns the center position of each door
func getDoorCenterPositions(width, height int, doors []DoorPosition) map[DoorPosition]Point {
	positions := make(map[DoorPosition]Point)
	for _, door := range doors {
		switch door {
		case DoorTop:
			positions[door] = Point{X: width / 2, Y: 0}
		case DoorBottom:
			positions[door] = Point{X: width / 2, Y: height - 1}
		case DoorLeft:
			positions[door] = Point{X: 0, Y: height / 2}
		case DoorRight:
			positions[door] = Point{X: width - 1, Y: height / 2}
		}
	}
	return positions
}

// connectDoors connects all doors using random brushes with straight or L-shaped paths
func connectDoors(ground [][]int, doorPositions map[DoorPosition]Point, width, height int) {
	connectDoorsWithDebug(ground, doorPositions, width, height, nil)
}

// connectDoorsWithDebug connects all doors and records debug info
func connectDoorsWithDebug(ground [][]int, doorPositions map[DoorPosition]Point, width, height int, debug *GroundDebugInfo) {
	doors := make([]DoorPosition, 0, len(doorPositions))
	for door := range doorPositions {
		doors = append(doors, door)
	}

	// Connect doors pairwise until all are connected
	connected := make(map[DoorPosition]bool)
	connected[doors[0]] = true

	for len(connected) < len(doors) {
		// Find an unconnected door and connect it to a connected door
		for _, door := range doors {
			if connected[door] {
				continue
			}

			// Find a connected door to connect to
			var targetDoor DoorPosition
			for d := range connected {
				targetDoor = d
				break
			}

			// Connect the two doors
			from := doorPositions[targetDoor]
			to := doorPositions[door]
			connInfo := connectTwoPointsWithDebug(ground, from, to, width, height, string(targetDoor), string(door))
			if debug != nil {
				debug.DoorConnections = append(debug.DoorConnections, connInfo)
			}
			connected[door] = true
		}
	}
}

// connectTwoPoints connects two points with a straight line or L-shaped path through center
func connectTwoPoints(ground [][]int, from, to Point, width, height int) {
	connectTwoPointsWithDebug(ground, from, to, width, height, "", "")
}

// connectTwoPointsWithDebug connects two points and returns debug info
func connectTwoPointsWithDebug(ground [][]int, from, to Point, width, height int, fromName, toName string) DoorConnectionInfo {
	brush := connectionBrushes[rand.Intn(len(connectionBrushes))]

	// Calculate center point
	centerX, centerY := width/2, height/2

	pathType := "direct"
	if from.X != to.X && from.Y != to.Y {
		// L-shaped path through center point
		pathType = "L-shaped via center"
		centerPoint := Point{X: centerX, Y: centerY}
		drawLine(ground, from, centerPoint, brush, width, height)
		drawLine(ground, centerPoint, to, brush, width, height)
	} else {
		// Straight line (works for aligned points or random choice)
		drawLine(ground, from, to, brush, width, height)
	}

	return DoorConnectionInfo{
		From:      fmt.Sprintf("%s (%d,%d)", fromName, from.X, from.Y),
		To:        fmt.Sprintf("%s (%d,%d)", toName, to.X, to.Y),
		PathType:  pathType,
		BrushSize: fmt.Sprintf("%dx%d", brush.Width, brush.Height),
	}
}

// drawLine draws a line between two points using the specified brush
func drawLine(ground [][]int, from, to Point, brush BrushSize, width, height int) {
	// Bresenham-like line drawing
	dx := abs(to.X - from.X)
	dy := abs(to.Y - from.Y)
	sx := 1
	if from.X > to.X {
		sx = -1
	}
	sy := 1
	if from.Y > to.Y {
		sy = -1
	}

	x, y := from.X, from.Y
	err := dx - dy

	for {
		applyBrush(ground, x, y, brush, width, height)

		if x == to.X && y == to.Y {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// applyBrush applies a brush centered at the given point
func applyBrush(ground [][]int, centerX, centerY int, brush BrushSize, width, height int) {
	startX := centerX - brush.Width/2
	startY := centerY - brush.Height/2

	for dy := 0; dy < brush.Height; dy++ {
		for dx := 0; dx < brush.Width; dx++ {
			x := startX + dx
			y := startY + dy
			if x >= 0 && x < width && y >= 0 && y < height {
				ground[y][x] = 1
			}
		}
	}
}

// applyBrushWithMirror applies a brush and its mirror counterpart
func applyBrushWithMirror(ground [][]int, centerX, centerY int, brush BrushSize, width, height int, mirror MirrorAxis) {
	// Apply original brush
	applyBrush(ground, centerX, centerY, brush, width, height)

	// Apply mirrored brush
	switch mirror {
	case MirrorY:
		// Mirror left-right (across vertical center line Y-axis)
		mirroredX := width - 1 - centerX
		applyBrush(ground, mirroredX, centerY, brush, width, height)
	case MirrorX:
		// Mirror top-bottom (across horizontal center line X-axis)
		mirroredY := height - 1 - centerY
		applyBrush(ground, centerX, mirroredY, brush, width, height)
	}
}

// drawPlatforms draws small platforms according to the probability-based strategy
func drawPlatforms(ground [][]int, width, height int, doors []DoorPosition, doorPositions map[DoorPosition]Point) {
	drawPlatformsWithDebug(ground, width, height, doors, doorPositions, nil)
}

// drawPlatformsWithDebug draws platforms and records debug info
func drawPlatformsWithDebug(ground [][]int, width, height int, doors []DoorPosition, doorPositions map[DoorPosition]Point, debug *GroundDebugInfo) {
	// Determine number of draws (1-3)
	drawCount := rand.Intn(3) + 1

	// Build strategies with weights
	strategies := buildStrategies(width, height, doors, doorPositions)

	for i := 0; i < drawCount && len(strategies) > 0; i++ {
		// Select strategy by weight
		selectedIdx := selectByWeight(strategies)
		if selectedIdx < 0 {
			break
		}

		strategy := strategies[selectedIdx]

		// Draw on all points in the strategy with mirroring
		brush := platformBrushes[rand.Intn(len(platformBrushes))]
		for _, point := range strategy.Points {
			applyBrushWithMirror(ground, point.X, point.Y, brush, width, height, strategy.Mirror)
		}

		// Record debug info
		if debug != nil {
			points := make([]string, len(strategy.Points))
			for j, p := range strategy.Points {
				points[j] = fmt.Sprintf("(%d,%d)", p.X, p.Y)
			}
			mirrorStr := "none"
			switch strategy.Mirror {
			case MirrorX:
				mirrorStr = "top-bottom"
			case MirrorY:
				mirrorStr = "left-right"
			}
			debug.Platforms = append(debug.Platforms, PlatformInfo{
				Strategy:  strategy.Name,
				BrushSize: fmt.Sprintf("%dx%d", brush.Width, brush.Height),
				Points:    points,
				Mirror:    mirrorStr,
			})
		}

		// Remove selected strategy
		strategies = append(strategies[:selectedIdx], strategies[selectedIdx+1:]...)
	}
}

// buildStrategies builds the platform placement strategies with weights
func buildStrategies(width, height int, doors []DoorPosition, doorPositions map[DoorPosition]Point) []Strategy {
	strategies := []Strategy{}
	centerX, centerY := width/2, height/2

	// Screen center: weight 50 (no mirror needed, it's at center)
	strategies = append(strategies, Strategy{
		Name:   "center",
		Weight: 50,
		Points: []Point{{X: centerX, Y: centerY}},
		Mirror: MirrorNone,
	})

	// Check for horizontal (left-right) connection
	_, hasLeft := doorPositions[DoorLeft]
	_, hasRight := doorPositions[DoorRight]
	if hasLeft && hasRight {
		// Left and right door positions: weight 10
		// Points on X-axis (horizontal center), mirror top-bottom
		strategies = append(strategies, Strategy{
			Name:   "left_right_doors",
			Weight: 10,
			Points: []Point{doorPositions[DoorLeft], doorPositions[DoorRight]},
			Mirror: MirrorX,
		})
		// Midpoints between center and doors: weight 10
		// Points on X-axis (horizontal center), mirror top-bottom
		strategies = append(strategies, Strategy{
			Name:   "left_right_midpoints",
			Weight: 10,
			Points: []Point{
				{X: (centerX + doorPositions[DoorLeft].X) / 2, Y: centerY},
				{X: (centerX + doorPositions[DoorRight].X) / 2, Y: centerY},
			},
			Mirror: MirrorX,
		})
	}

	// Check for vertical (top-bottom) connection
	_, hasTop := doorPositions[DoorTop]
	_, hasBottom := doorPositions[DoorBottom]
	if hasTop && hasBottom {
		// Top and bottom door positions: weight 10
		// Points on Y-axis (vertical center), mirror left-right
		strategies = append(strategies, Strategy{
			Name:   "top_bottom_doors",
			Weight: 10,
			Points: []Point{doorPositions[DoorTop], doorPositions[DoorBottom]},
			Mirror: MirrorY,
		})
		// Midpoints between center and doors: weight 10
		// Points on Y-axis (vertical center), mirror left-right
		strategies = append(strategies, Strategy{
			Name:   "top_bottom_midpoints",
			Weight: 10,
			Points: []Point{
				{X: centerX, Y: (centerY + doorPositions[DoorTop].Y) / 2},
				{X: centerX, Y: (centerY + doorPositions[DoorBottom].Y) / 2},
			},
			Mirror: MirrorY,
		})
	}

	// All connected doors: weight 10 (no mirror, points already cover all positions)
	allDoorPoints := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		allDoorPoints = append(allDoorPoints, pos)
	}
	if len(allDoorPoints) > 0 {
		strategies = append(strategies, Strategy{
			Name:   "all_doors",
			Weight: 10,
			Points: allDoorPoints,
			Mirror: MirrorNone,
		})
	}

	// Midpoints between center and all doors: weight 10 (no mirror, points already cover all positions)
	midpoints := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		midpoints = append(midpoints, Point{
			X: (centerX + pos.X) / 2,
			Y: (centerY + pos.Y) / 2,
		})
	}
	if len(midpoints) > 0 {
		strategies = append(strategies, Strategy{
			Name:   "all_midpoints",
			Weight: 10,
			Points: midpoints,
			Mirror: MirrorNone,
		})
	}

	return strategies
}

// selectByWeight selects a strategy index by weight
func selectByWeight(strategies []Strategy) int {
	if len(strategies) == 0 {
		return -1
	}

	totalWeight := 0
	for _, s := range strategies {
		totalWeight += s.Weight
	}

	if totalWeight == 0 {
		return -1
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for i, s := range strategies {
		cumulative += s.Weight
		if r < cumulative {
			return i
		}
	}

	return len(strategies) - 1
}

// Helper functions
func createEmptyLayer(width, height int) [][]int {
	layer := make([][]int, height)
	for y := 0; y < height; y++ {
		layer[y] = make([]int, width)
	}
	return layer
}

func copyLayer(layer [][]int) [][]int {
	copied := make([][]int, len(layer))
	for y := range layer {
		copied[y] = make([]int, len(layer[y]))
		copy(copied[y], layer[y])
	}
	return copied
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// EmptyArea represents a rectangular void region
type EmptyArea struct {
	X, Y          int // Top-left corner
	Width, Height int
}

// Minimum empty area size for floating island consideration
const minEmptyAreaSize = 4

// Minimum floating island size
const minIslandSize = 2

// Minimum distance from existing ground
const minIslandGroundDistance = 2

// findEmptyAreas finds all rectangular void regions >= 4x4 in size
func findEmptyAreas(ground [][]int, width, height int) []EmptyArea {
	var areas []EmptyArea

	// Use a visited map to avoid duplicate areas
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Skip if already visited or not void
			if visited[y][x] || ground[y][x] != 0 {
				continue
			}

			// Find the maximum rectangle starting from this point
			area := findMaxRectangle(ground, visited, x, y, width, height)
			if area.Width >= minEmptyAreaSize && area.Height >= minEmptyAreaSize {
				areas = append(areas, area)
			}
		}
	}

	return areas
}

// findMaxRectangle finds the maximum rectangle of void cells starting from (startX, startY)
func findMaxRectangle(ground [][]int, visited [][]bool, startX, startY, width, height int) EmptyArea {
	// Find max width from starting point
	maxWidth := 0
	for x := startX; x < width && ground[startY][x] == 0; x++ {
		maxWidth++
	}

	// Find max height that maintains at least minEmptyAreaSize width
	maxHeight := 0
	currentWidth := maxWidth
	for y := startY; y < height && currentWidth >= minEmptyAreaSize; y++ {
		// Check this row's void width
		rowWidth := 0
		for x := startX; x < width && x < startX+currentWidth && ground[y][x] == 0; x++ {
			rowWidth++
		}
		if rowWidth < minEmptyAreaSize {
			break
		}
		currentWidth = rowWidth
		maxHeight++
	}

	// Mark visited cells for the found rectangle
	for y := startY; y < startY+maxHeight && y < height; y++ {
		for x := startX; x < startX+currentWidth && x < width; x++ {
			visited[y][x] = true
		}
	}

	return EmptyArea{
		X:      startX,
		Y:      startY,
		Width:  currentWidth,
		Height: maxHeight,
	}
}

// drawFloatingIslandsWithDebug draws floating islands in void areas with 50% probability per attempt
func drawFloatingIslandsWithDebug(ground [][]int, width, height int, debug *GroundDebugInfo) {
	// Step 1: Find all empty areas >= 4x4
	emptyAreas := findEmptyAreas(ground, width, height)
	if len(emptyAreas) == 0 {
		if debug != nil {
			debug.FloatingIslands = append(debug.FloatingIslands, FloatingIslandInfo{
				Skipped:    true,
				SkipReason: "no empty areas >= 4x4 found",
			})
		}
		return
	}

	// Shuffle the empty areas
	rand.Shuffle(len(emptyAreas), func(i, j int) {
		emptyAreas[i], emptyAreas[j] = emptyAreas[j], emptyAreas[i]
	})

	// Step 2-4: Loop with 50% probability
	for len(emptyAreas) > 0 {
		// Step 2: 50% chance to continue
		if rand.Float64() >= 0.8 {
			if debug != nil {
				debug.FloatingIslands = append(debug.FloatingIslands, FloatingIslandInfo{
					Skipped:    true,
					SkipReason: "stopped by 50% probability check",
				})
			}
			break
		}

		// Step 3: Pop an empty area
		area := emptyAreas[0]
		emptyAreas = emptyAreas[1:]

		// Try to place a floating island in this area
		placed := tryPlaceFloatingIsland(ground, area, width, height, debug)
		if !placed && debug != nil {
			debug.FloatingIslands = append(debug.FloatingIslands, FloatingIslandInfo{
				FromArea:   fmt.Sprintf("(%d,%d) %dx%d", area.X, area.Y, area.Width, area.Height),
				Skipped:    true,
				SkipReason: "could not find valid position with min distance 2 from ground",
			})
		}
	}
}

// tryPlaceFloatingIsland attempts to place a floating island in the given empty area
func tryPlaceFloatingIsland(ground [][]int, area EmptyArea, gridWidth, gridHeight int, debug *GroundDebugInfo) bool {
	// Collect all valid (position, size) combinations
	// The margin can extend outside the empty area (to grid edge or other void cells)
	// So we try all sizes that fit within the area and let isValidIslandPosition check margins
	type placement struct {
		x, y, w, h int
	}
	var validPlacements []placement

	// Try all possible sizes from minIslandSize up to area size
	maxW := area.Width
	maxH := area.Height

	for islandHeight := minIslandSize; islandHeight <= maxH; islandHeight++ {
		for islandWidth := minIslandSize; islandWidth <= maxW; islandWidth++ {
			// Find valid positions for this size
			for y := area.Y; y <= area.Y+area.Height-islandHeight; y++ {
				for x := area.X; x <= area.X+area.Width-islandWidth; x++ {
					if isValidIslandPosition(ground, x, y, islandWidth, islandHeight, gridWidth, gridHeight) {
						validPlacements = append(validPlacements, placement{x, y, islandWidth, islandHeight})
					}
				}
			}
		}
	}

	if len(validPlacements) == 0 {
		return false
	}

	// Pick a random valid placement
	p := validPlacements[rand.Intn(len(validPlacements))]

	// Draw the island
	for dy := 0; dy < p.h; dy++ {
		for dx := 0; dx < p.w; dx++ {
			ground[p.y+dy][p.x+dx] = 1
		}
	}

	// Record debug info
	if debug != nil {
		centerX := p.x + p.w/2
		centerY := p.y + p.h/2
		debug.FloatingIslands = append(debug.FloatingIslands, FloatingIslandInfo{
			Position: fmt.Sprintf("(%d,%d)", centerX, centerY),
			Size:     fmt.Sprintf("%dx%d", p.w, p.h),
			FromArea: fmt.Sprintf("(%d,%d) %dx%d", area.X, area.Y, area.Width, area.Height),
		})
	}

	return true
}

// isValidIslandPosition checks if an island can be placed at (x, y) with exactly minIslandGroundDistance from existing ground
func isValidIslandPosition(ground [][]int, x, y, islandWidth, islandHeight, gridWidth, gridHeight int) bool {
	// Check that the island area and its surrounding margin are all void
	// Margin is minIslandGroundDistance on each side
	checkStartX := x - minIslandGroundDistance
	checkStartY := y - minIslandGroundDistance
	checkEndX := x + islandWidth + minIslandGroundDistance
	checkEndY := y + islandHeight + minIslandGroundDistance

	for cy := checkStartY; cy < checkEndY; cy++ {
		for cx := checkStartX; cx < checkEndX; cx++ {
			// Skip cells outside the grid (they're fine, considered void)
			if cx < 0 || cx >= gridWidth || cy < 0 || cy >= gridHeight {
				continue
			}

			// If this cell is inside the island area, it should be void (we'll place ground there)
			isInsideIsland := cx >= x && cx < x+islandWidth && cy >= y && cy < y+islandHeight
			if isInsideIsland {
				// The island area must currently be void
				if ground[cy][cx] != 0 {
					return false
				}
			} else {
				// The margin area must be void (no existing ground within minIslandGroundDistance)
				if ground[cy][cx] != 0 {
					return false
				}
			}
		}
	}

	// Additionally, check that there IS ground just outside the margin (at distance exactly minIslandGroundDistance+1)
	// This ensures the island is placed close to existing ground, not in the middle of nowhere
	hasNearbyGround := false
	outerDist := minIslandGroundDistance + 1
	outerStartX := x - outerDist
	outerStartY := y - outerDist
	outerEndX := x + islandWidth + outerDist
	outerEndY := y + islandHeight + outerDist

	for cy := outerStartY; cy < outerEndY; cy++ {
		for cx := outerStartX; cx < outerEndX; cx++ {
			// Skip cells inside the already-checked margin area
			if cx >= checkStartX && cx < checkEndX && cy >= checkStartY && cy < checkEndY {
				continue
			}
			// Skip cells outside the grid
			if cx < 0 || cx >= gridWidth || cy < 0 || cy >= gridHeight {
				continue
			}
			// Check if there's ground at this outer ring
			if ground[cy][cx] == 1 {
				hasNearbyGround = true
				break
			}
		}
		if hasNearbyGround {
			break
		}
	}

	return hasNearbyGround
}

// Bridge layer constants
const bridgeSize = 2 // Bridge is always 2x2

// Island represents a connected region of ground cells
type Island struct {
	ID     int
	Cells  []Point
	MinX   int
	MinY   int
	MaxX   int
	MaxY   int
}

// generateBridgeLayerWithDebug generates bridges to connect floating islands to ground/other islands
func generateBridgeLayerWithDebug(bridgeLayer, ground [][]int, width, height int) *BridgeLayerDebugInfo {
	debug := &BridgeLayerDebugInfo{}

	// Step 1: Find all connected regions (islands) in the ground layer
	islands := findAllIslands(ground, width, height)

	if len(islands) <= 1 {
		debug.Skipped = true
		debug.SkipReason = "no floating islands found (all ground is connected)"
		debug.IslandsFound = len(islands)
		return debug
	}

	debug.IslandsFound = len(islands)

	// The main ground (usually the largest or the one connected to doors) is island 0
	// Other islands need to be connected
	mainIslandID := 0 // Assume the first found island is the main ground

	// Find the largest island as the main ground
	maxSize := 0
	for i, island := range islands {
		if len(island.Cells) > maxSize {
			maxSize = len(island.Cells)
			mainIslandID = i
		}
	}

	// Track which islands are connected to the main ground
	connected := make(map[int]bool)
	connected[mainIslandID] = true

	// Step 2: Connect each floating island to the main ground or another connected island
	for {
		// Find an unconnected island
		unconnectedID := -1
		for i := range islands {
			if !connected[i] {
				unconnectedID = i
				break
			}
		}

		if unconnectedID == -1 {
			// All islands connected
			break
		}

		unconnectedIsland := islands[unconnectedID]

		// Find the nearest connected island/ground
		bestConnection := findBestBridgeConnection(unconnectedIsland, islands, connected, ground, bridgeLayer, width, height)

		if bestConnection == nil {
			// Cannot connect this island
			debug.Misses = append(debug.Misses, MissInfo{
				Reason: fmt.Sprintf("cannot find valid bridge path for island at (%d,%d)-(%d,%d)",
					unconnectedIsland.MinX, unconnectedIsland.MinY, unconnectedIsland.MaxX, unconnectedIsland.MaxY),
			})
			// Mark as connected anyway to avoid infinite loop
			connected[unconnectedID] = true
			continue
		}

		// Place the bridge
		placeBridge(bridgeLayer, bestConnection.bridgeX, bestConnection.bridgeY)
		connected[unconnectedID] = true
		debug.BridgesPlaced++

		debug.Connections = append(debug.Connections, BridgeConnection{
			From:     fmt.Sprintf("island (%d,%d)-(%d,%d)", unconnectedIsland.MinX, unconnectedIsland.MinY, unconnectedIsland.MaxX, unconnectedIsland.MaxY),
			To:       bestConnection.targetDesc,
			Position: fmt.Sprintf("(%d,%d)", bestConnection.bridgeX, bestConnection.bridgeY),
			Size:     "2x2",
		})
	}

	return debug
}

// bridgeConnectionResult holds the result of finding a bridge connection
type bridgeConnectionResult struct {
	bridgeX    int
	bridgeY    int
	targetDesc string
}

// findBestBridgeConnection finds the best position to place a 2x2 bridge connecting an island to existing ground
func findBestBridgeConnection(island Island, allIslands []Island, connected map[int]bool, ground, bridgeLayer [][]int, width, height int) *bridgeConnectionResult {
	// For each edge cell of the island, try to find a valid bridge position
	// The bridge must touch the island (2x2 fully adjacent) and also touch ground or another connected island

	type candidate struct {
		bridgeX, bridgeY int
		distance         int
		targetDesc       string
	}
	var candidates []candidate

	// Check all possible 2x2 bridge positions around the island
	// A bridge at (bx, by) occupies cells (bx, by), (bx+1, by), (bx, by+1), (bx+1, by+1)
	// It must touch the island and must connect to existing ground

	for _, cell := range island.Cells {
		// Try placing bridge in 4 directions from this cell
		// Direction: bridge placed such that it touches this cell

		directions := []struct {
			dx, dy int
			desc   string
		}{
			{-2, 0, "left"},   // Bridge to the left of cell
			{2, 0, "right"},   // Bridge to the right of cell
			{0, -2, "above"},  // Bridge above cell
			{0, 2, "below"},   // Bridge below cell
		}

		for _, dir := range directions {
			bx := cell.X + dir.dx
			by := cell.Y + dir.dy

			// For left/right: bridge needs to span 2 cells vertically to touch island
			// For above/below: bridge needs to span 2 cells horizontally to touch island
			// Let's check if the bridge can be placed and touches both island and ground

			if !canPlaceBridge(bx, by, ground, bridgeLayer, width, height) {
				continue
			}

			// Check if bridge touches the island
			if !bridgeTouchesIsland(bx, by, island, ground) {
				continue
			}

			// Check if bridge touches existing ground (not part of this island)
			touchesGround, targetDesc := bridgeTouchesExistingGround(bx, by, island, allIslands, connected, ground, width, height)
			if !touchesGround {
				continue
			}

			// Calculate distance (for prioritization)
			centerX := (island.MinX + island.MaxX) / 2
			centerY := (island.MinY + island.MaxY) / 2
			dist := abs(bx-centerX) + abs(by-centerY)

			candidates = append(candidates, candidate{bx, by, dist, targetDesc})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by distance and pick the closest
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.distance < best.distance {
			best = c
		}
	}

	return &bridgeConnectionResult{
		bridgeX:    best.bridgeX,
		bridgeY:    best.bridgeY,
		targetDesc: best.targetDesc,
	}
}

// canPlaceBridge checks if a 2x2 bridge can be placed at (x, y)
func canPlaceBridge(x, y int, ground, bridgeLayer [][]int, width, height int) bool {
	// Bridge must be within bounds
	if x < 0 || x+bridgeSize > width || y < 0 || y+bridgeSize > height {
		return false
	}

	// All cells must be void (ground=0) and no existing bridge
	for dy := 0; dy < bridgeSize; dy++ {
		for dx := 0; dx < bridgeSize; dx++ {
			if ground[y+dy][x+dx] != 0 || bridgeLayer[y+dy][x+dx] != 0 {
				return false
			}
		}
	}

	return true
}

// bridgeTouchesIsland checks if a 2x2 bridge at (bx, by) fully touches the island (2x2 contact)
func bridgeTouchesIsland(bx, by int, island Island, ground [][]int) bool {
	// Check if the bridge has at least 2 adjacent cells touching the island
	touchCount := 0

	// Check all 4 sides of the bridge
	bridgeCells := []Point{
		{bx, by}, {bx + 1, by}, {bx, by + 1}, {bx + 1, by + 1},
	}

	for _, bc := range bridgeCells {
		// Check adjacent cells (not diagonal)
		adjacents := []Point{
			{bc.X - 1, bc.Y}, {bc.X + 1, bc.Y}, {bc.X, bc.Y - 1}, {bc.X, bc.Y + 1},
		}
		for _, adj := range adjacents {
			// Check if adjacent cell is part of the island
			for _, ic := range island.Cells {
				if ic.X == adj.X && ic.Y == adj.Y {
					touchCount++
					break
				}
			}
		}
	}

	// Need at least 2 touch points for 2x2 full contact
	return touchCount >= 2
}

// bridgeTouchesExistingGround checks if bridge touches ground that's not part of the given island
func bridgeTouchesExistingGround(bx, by int, excludeIsland Island, allIslands []Island, connected map[int]bool, ground [][]int, width, height int) (bool, string) {
	// Check all cells adjacent to the bridge
	bridgeCells := []Point{
		{bx, by}, {bx + 1, by}, {bx, by + 1}, {bx + 1, by + 1},
	}

	excludeCells := make(map[Point]bool)
	for _, c := range excludeIsland.Cells {
		excludeCells[c] = true
	}

	touchCount := 0
	var targetDesc string

	for _, bc := range bridgeCells {
		adjacents := []Point{
			{bc.X - 1, bc.Y}, {bc.X + 1, bc.Y}, {bc.X, bc.Y - 1}, {bc.X, bc.Y + 1},
		}
		for _, adj := range adjacents {
			if adj.X < 0 || adj.X >= width || adj.Y < 0 || adj.Y >= height {
				continue
			}
			if ground[adj.Y][adj.X] == 1 && !excludeCells[adj] {
				touchCount++
				// Find which island this belongs to
				for i, island := range allIslands {
					if connected[i] {
						for _, ic := range island.Cells {
							if ic.X == adj.X && ic.Y == adj.Y {
								if i == 0 {
									targetDesc = "main ground"
								} else {
									targetDesc = fmt.Sprintf("island %d", i)
								}
								break
							}
						}
					}
				}
			}
		}
	}

	if targetDesc == "" && touchCount >= 2 {
		targetDesc = "ground"
	}

	return touchCount >= 2, targetDesc
}

// placeBridge places a 2x2 bridge at the given position
func placeBridge(bridgeLayer [][]int, x, y int) {
	for dy := 0; dy < bridgeSize; dy++ {
		for dx := 0; dx < bridgeSize; dx++ {
			bridgeLayer[y+dy][x+dx] = 1
		}
	}
}

// findAllIslands finds all connected regions of ground cells using flood fill
func findAllIslands(ground [][]int, width, height int) []Island {
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	var islands []Island
	islandID := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if ground[y][x] == 1 && !visited[y][x] {
				// Found a new island, flood fill to find all connected cells
				island := floodFillIsland(ground, visited, x, y, width, height, islandID)
				islands = append(islands, island)
				islandID++
			}
		}
	}

	return islands
}

// floodFillIsland performs flood fill to find all cells of an island
func floodFillIsland(ground [][]int, visited [][]bool, startX, startY, width, height, id int) Island {
	island := Island{
		ID:   id,
		MinX: startX,
		MinY: startY,
		MaxX: startX,
		MaxY: startY,
	}

	// BFS flood fill
	queue := []Point{{X: startX, Y: startY}}
	visited[startY][startX] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		island.Cells = append(island.Cells, curr)

		// Update bounds
		if curr.X < island.MinX {
			island.MinX = curr.X
		}
		if curr.X > island.MaxX {
			island.MaxX = curr.X
		}
		if curr.Y < island.MinY {
			island.MinY = curr.Y
		}
		if curr.Y > island.MaxY {
			island.MaxY = curr.Y
		}

		// Check 4 neighbors
		neighbors := []Point{
			{curr.X - 1, curr.Y}, {curr.X + 1, curr.Y},
			{curr.X, curr.Y - 1}, {curr.X, curr.Y + 1},
		}

		for _, n := range neighbors {
			if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height &&
				ground[n.Y][n.X] == 1 && !visited[n.Y][n.X] {
				visited[n.Y][n.X] = true
				queue = append(queue, n)
			}
		}
	}

	return island
}

// Static placement size (fixed 2x2)
const staticSize = 2

// PlacementStrategy represents the strategy for placing statics
type PlacementStrategy int

const (
	StrategyCenterOutward PlacementStrategy = iota // Start from center, spread outward
	StrategyEdgeInward                             // Start from edges, spread inward
)

// generateStaticLayer generates the static layer with the given constraints
// staticLayer: output layer to place statics
// ground: ground layer (static requires ground=1)
// softEdge: soft edge layer (static cannot overlap)
// bridge: bridge layer (static cannot overlap)
// doorPositions: positions of doors
// width, height: dimensions
// targetCount: suggested number of statics to place
func generateStaticLayer(staticLayer, ground, softEdge, bridge [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) {
	// Get all door cells and their adjacent cells (forbidden zone)
	forbiddenCells := getDoorForbiddenCells(doorPositions, width, height)

	// Find all valid 2x2 positions for static placement
	validPositions := findValidStaticPositions(ground, softEdge, bridge, staticLayer, forbiddenCells, width, height)
	if len(validPositions) == 0 {
		return
	}

	// Sort positions by strategy (will be re-sorted on each strategy switch)
	centerX, centerY := width/2, height/2
	currentStrategy := StrategyCenterOutward

	remaining := targetCount
	strategyAttempts := 0
	maxStrategyAttempts := 2 * targetCount // Prevent infinite loop

	for remaining > 0 && strategyAttempts < maxStrategyAttempts {
		// Sort valid positions based on current strategy
		sortPositionsByStrategy(validPositions, currentStrategy, centerX, centerY, width, height)

		// Try to place one static
		placed := false
		for i, pos := range validPositions {
			// Check if this position is still valid (may have been invalidated by previous placements)
			if !isValidStaticPosition(pos, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height) {
				continue
			}

			// Check connectivity after placement
			if !checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, pos, width, height) {
				continue
			}

			// Place the static (2x2)
			placeStatic(staticLayer, pos)
			remaining--
			placed = true

			// Remove this position and update valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that now touch this static
			validPositions = filterTouchingPositions(validPositions, pos)
			break
		}

		if !placed {
			// Switch strategy and try again
			strategyAttempts++
		}

		// Alternate strategy after each placement or failed attempt
		if currentStrategy == StrategyCenterOutward {
			currentStrategy = StrategyEdgeInward
		} else {
			currentStrategy = StrategyCenterOutward
		}
	}
}

// getDoorForbiddenCells returns all cells that are doors or adjacent to doors
func getDoorForbiddenCells(doorPositions map[DoorPosition]Point, width, height int) map[Point]bool {
	forbidden := make(map[Point]bool)

	// Door area is typically larger than a single point
	// For each door, mark a 4x4 area centered on the door position as forbidden
	// This ensures statics don't touch door areas
	for _, doorPos := range doorPositions {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				x := doorPos.X + dx
				y := doorPos.Y + dy
				if x >= 0 && x < width && y >= 0 && y < height {
					forbidden[Point{X: x, Y: y}] = true
				}
			}
		}
	}

	return forbidden
}

// findValidStaticPositions finds all valid top-left corners for 2x2 static placement
func findValidStaticPositions(ground, softEdge, bridge, staticLayer [][]int, forbiddenCells map[Point]bool, width, height int) []Point {
	var positions []Point

	// Iterate through all possible top-left corners for 2x2 placement
	for y := 0; y <= height-staticSize; y++ {
		for x := 0; x <= width-staticSize; x++ {
			pos := Point{X: x, Y: y}
			if isValidStaticPosition(pos, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// isValidStaticPosition checks if a 2x2 static can be placed at the given top-left corner
func isValidStaticPosition(pos Point, ground, softEdge, bridge, staticLayer [][]int, forbiddenCells map[Point]bool, width, height int) bool {
	// Check all 4 cells of the 2x2 area
	for dy := 0; dy < staticSize; dy++ {
		for dx := 0; dx < staticSize; dx++ {
			x := pos.X + dx
			y := pos.Y + dy

			// Check bounds
			if x >= width || y >= height {
				return false
			}

			// Check ground layer (must be 1)
			if ground[y][x] != 1 {
				return false
			}

			// Check soft edge (must be 0)
			if softEdge[y][x] != 0 {
				return false
			}

			// Check bridge (must be 0)
			if bridge[y][x] != 0 {
				return false
			}

			// Check existing static (must be 0)
			if staticLayer[y][x] != 0 {
				return false
			}

			// Check forbidden cells (door area)
			if forbiddenCells[Point{X: x, Y: y}] {
				return false
			}
		}
	}

	// Check that the static doesn't touch any existing static (including diagonals)
	if touchesExistingStatic(pos, staticLayer, width, height) {
		return false
	}

	return true
}

// touchesExistingStatic checks if placing a 2x2 static at pos would touch any existing static
func touchesExistingStatic(pos Point, staticLayer [][]int, width, height int) bool {
	// Check a 4x4 area around the 2x2 placement (1 cell buffer on each side)
	for dy := -1; dy <= staticSize; dy++ {
		for dx := -1; dx <= staticSize; dx++ {
			// Skip the cells that will be occupied by this static
			if dx >= 0 && dx < staticSize && dy >= 0 && dy < staticSize {
				continue
			}

			x := pos.X + dx
			y := pos.Y + dy

			if x >= 0 && x < width && y >= 0 && y < height {
				if staticLayer[y][x] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// sortPositionsByStrategy sorts positions based on the placement strategy
func sortPositionsByStrategy(positions []Point, strategy PlacementStrategy, centerX, centerY, width, height int) {
	switch strategy {
	case StrategyCenterOutward:
		// Sort by distance from center (closest first)
		for i := 0; i < len(positions)-1; i++ {
			for j := i + 1; j < len(positions); j++ {
				distI := distanceFromCenter(positions[i], centerX, centerY)
				distJ := distanceFromCenter(positions[j], centerX, centerY)
				if distJ < distI {
					positions[i], positions[j] = positions[j], positions[i]
				}
			}
		}
	case StrategyEdgeInward:
		// Sort by distance from edge (closest to edge first)
		for i := 0; i < len(positions)-1; i++ {
			for j := i + 1; j < len(positions); j++ {
				distI := distanceFromEdge(positions[i], width, height)
				distJ := distanceFromEdge(positions[j], width, height)
				if distJ < distI {
					positions[i], positions[j] = positions[j], positions[i]
				}
			}
		}
	}
}

// distanceFromCenter calculates the Manhattan distance from center
func distanceFromCenter(pos Point, centerX, centerY int) int {
	// Use the center of the 2x2 static
	staticCenterX := pos.X + staticSize/2
	staticCenterY := pos.Y + staticSize/2
	return abs(staticCenterX-centerX) + abs(staticCenterY-centerY)
}

// distanceFromEdge calculates the minimum distance from any edge
func distanceFromEdge(pos Point, width, height int) int {
	// Use the center of the 2x2 static
	staticCenterX := pos.X + staticSize/2
	staticCenterY := pos.Y + staticSize/2

	distLeft := staticCenterX
	distRight := width - 1 - staticCenterX
	distTop := staticCenterY
	distBottom := height - 1 - staticCenterY

	minDist := distLeft
	if distRight < minDist {
		minDist = distRight
	}
	if distTop < minDist {
		minDist = distTop
	}
	if distBottom < minDist {
		minDist = distBottom
	}
	return minDist
}

// placeStatic places a 2x2 static at the given top-left corner
func placeStatic(staticLayer [][]int, pos Point) {
	for dy := 0; dy < staticSize; dy++ {
		for dx := 0; dx < staticSize; dx++ {
			staticLayer[pos.Y+dy][pos.X+dx] = 1
		}
	}
}

// filterTouchingPositions removes positions that would touch the newly placed static
func filterTouchingPositions(positions []Point, placedPos Point) []Point {
	var filtered []Point
	for _, pos := range positions {
		if !wouldTouch(pos, placedPos) {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

// wouldTouch checks if two 2x2 statics would touch (including diagonals)
func wouldTouch(pos1, pos2 Point) bool {
	// Two 2x2 squares touch if their bounding boxes (expanded by 1) overlap
	// pos1 occupies [pos1.X, pos1.X+1] x [pos1.Y, pos1.Y+1]
	// pos2 occupies [pos2.X, pos2.X+1] x [pos2.Y, pos2.Y+1]
	// They touch if the gap between them is <= 1 in both dimensions

	// Check X overlap with 1 cell buffer
	xOverlap := !(pos1.X+staticSize+1 <= pos2.X || pos2.X+staticSize+1 <= pos1.X)
	// Check Y overlap with 1 cell buffer
	yOverlap := !(pos1.Y+staticSize+1 <= pos2.Y || pos2.Y+staticSize+1 <= pos1.Y)

	return xOverlap && yOverlap
}

// checkConnectivityAfterPlacement checks if all doors remain connected after placing a static
func checkConnectivityAfterPlacement(ground, staticLayer [][]int, doorPositions map[DoorPosition]Point, newStaticPos Point, width, height int) bool {
	if len(doorPositions) < 2 {
		return true
	}

	// Create a temporary walkable map: ground=1 and static=0 means walkable
	// Temporarily mark the new static position as blocked
	walkable := make([][]bool, height)
	for y := 0; y < height; y++ {
		walkable[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			walkable[y][x] = ground[y][x] == 1 && staticLayer[y][x] == 0
		}
	}

	// Block the new static position
	for dy := 0; dy < staticSize; dy++ {
		for dx := 0; dx < staticSize; dx++ {
			x := newStaticPos.X + dx
			y := newStaticPos.Y + dy
			if x < width && y < height {
				walkable[y][x] = false
			}
		}
	}

	// Get door positions as a slice
	doors := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		doors = append(doors, pos)
	}

	// Check if all doors are connected using BFS from the first door
	startDoor := doors[0]
	visited := bfsConnectivity(walkable, startDoor, width, height)

	// Check if all other doors are reachable
	for _, door := range doors[1:] {
		if !visited[door.Y][door.X] {
			return false
		}
	}

	return true
}

// bfsConnectivity performs BFS to find all connected cells from a starting point
func bfsConnectivity(walkable [][]bool, start Point, width, height int) [][]bool {
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	// Find the nearest walkable cell to the start point (door might be on edge)
	startCell := findNearestWalkable(walkable, start, width, height)
	if startCell.X < 0 {
		return visited // No walkable cell found
	}

	queue := []Point{startCell}
	visited[startCell.Y][startCell.X] = true

	// 4-directional movement
	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for i := 0; i < 4; i++ {
			nx := curr.X + dx[i]
			ny := curr.Y + dy[i]

			if nx >= 0 && nx < width && ny >= 0 && ny < height && !visited[ny][nx] && walkable[ny][nx] {
				visited[ny][nx] = true
				queue = append(queue, Point{X: nx, Y: ny})
			}
		}
	}

	return visited
}

// findNearestWalkable finds the nearest walkable cell to the given point
func findNearestWalkable(walkable [][]bool, pos Point, width, height int) Point {
	// Check the point itself first
	if pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height && walkable[pos.Y][pos.X] {
		return pos
	}

	// Search in expanding squares
	for radius := 1; radius < max(width, height); radius++ {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if abs(dx) != radius && abs(dy) != radius {
					continue // Only check the perimeter
				}
				nx := pos.X + dx
				ny := pos.Y + dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height && walkable[ny][nx] {
					return Point{X: nx, Y: ny}
				}
			}
		}
	}

	return Point{X: -1, Y: -1} // Not found
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Turret placement constants
const (
	turretMinDoorDistance   = 4 // Minimum distance from doors
	turretMinTurretDistance = 2 // Minimum distance between turrets
	turretEdgePreference    = 2 // Prefer cells within this distance from edge or corners
)

// generateTurretLayer generates the turret layer with the given constraints
// turretLayer: output layer to place turrets
// ground: ground layer (turret requires ground=1)
// softEdge: soft edge layer (turret cannot overlap)
// bridge: bridge layer (turret cannot overlap)
// staticLayer: static layer (turret cannot overlap)
// doorPositions: positions of doors
// width, height: dimensions
// targetCount: suggested number of turrets to place
func generateTurretLayer(turretLayer, ground, softEdge, bridge, staticLayer [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) {
	// Find all valid positions for turret placement
	validPositions := findValidTurretPositions(ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height)
	if len(validPositions) == 0 {
		return
	}

	// Sort positions by preference (ground corners first, then room corners and edges, then by distance from center)
	centerX, centerY := width/2, height/2
	sortTurretPositionsByPreference(validPositions, centerX, centerY, width, height, ground)

	remaining := targetCount
	maxAttempts := 2 * targetCount // Prevent infinite loop

	for remaining > 0 && maxAttempts > 0 {
		placed := false

		for i, pos := range validPositions {
			// Check if this position is still valid
			if !isValidTurretPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height) {
				continue
			}

			// Check connectivity after placement
			if !checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer, doorPositions, pos, width, height) {
				continue
			}

			// Place the turret (1 tile)
			turretLayer[pos.Y][pos.X] = 1
			remaining--
			placed = true

			// Remove this position from valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that are too close to this turret
			validPositions = filterTurretsTooClose(validPositions, pos)
			break
		}

		if !placed {
			maxAttempts--
		}
	}
}

// findValidTurretPositions finds all valid positions for turret placement
func findValidTurretPositions(ground, softEdge, bridge, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) []Point {
	var positions []Point

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidTurretPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// isValidTurretPosition checks if a turret can be placed at the given position
func isValidTurretPosition(pos Point, ground, softEdge, bridge, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) bool {
	x, y := pos.X, pos.Y

	// Check bounds
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Check ground layer (must be 1)
	if ground[y][x] != 1 {
		return false
	}

	// Check soft edge (must be 0)
	if softEdge[y][x] != 0 {
		return false
	}

	// Check bridge (must be 0)
	if bridge[y][x] != 0 {
		return false
	}

	// Check static layer (must be 0)
	if staticLayer[y][x] != 0 {
		return false
	}

	// Check existing turret (must be 0)
	if turretLayer[y][x] != 0 {
		return false
	}

	// Check minimum distance from doors (at least 4 cells)
	for _, doorPos := range doorPositions {
		dist := manhattanDistance(pos, doorPos)
		if dist < turretMinDoorDistance {
			return false
		}
	}

	// Check minimum distance from other turrets (at least 2 cells)
	if tooCloseToExistingTurret(pos, turretLayer, width, height) {
		return false
	}

	return true
}

// manhattanDistance calculates the Manhattan distance between two points
func manhattanDistance(p1, p2 Point) int {
	return abs(p1.X-p2.X) + abs(p1.Y-p2.Y)
}

// tooCloseToExistingTurret checks if the position is too close to any existing turret
func tooCloseToExistingTurret(pos Point, turretLayer [][]int, width, height int) bool {
	for dy := -turretMinTurretDistance + 1; dy < turretMinTurretDistance; dy++ {
		for dx := -turretMinTurretDistance + 1; dx < turretMinTurretDistance; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if turretLayer[ny][nx] != 0 {
					// Check actual Manhattan distance
					if abs(dx)+abs(dy) < turretMinTurretDistance {
						return true
					}
				}
			}
		}
	}
	return false
}

// sortTurretPositionsByPreference sorts positions by placement preference
// Priority: ground corners (90°/270°) first, then room corners and edges, then by distance from center
func sortTurretPositionsByPreference(positions []Point, centerX, centerY, width, height int, ground [][]int) {
	// Calculate preference score for each position
	type scoredPos struct {
		pos   Point
		score int // Lower is better
	}

	scored := make([]scoredPos, len(positions))
	for i, pos := range positions {
		score := calculateTurretPreferenceScore(pos, centerX, centerY, width, height, ground)
		scored[i] = scoredPos{pos: pos, score: score}
	}

	// Sort by score (lower is better)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score < scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Copy back to positions
	for i, sp := range scored {
		positions[i] = sp.pos
	}
}

// calculateTurretPreferenceScore calculates a preference score for turret placement
// Lower score means higher preference
func calculateTurretPreferenceScore(pos Point, centerX, centerY, width, height int, ground [][]int) int {
	// Calculate distance from center
	distToCenter := abs(pos.X-centerX) + abs(pos.Y-centerY)

	// Highest priority: ground right angles (90°) and inner corners (270°)
	// This is where ground forms an L-shape
	groundCornerType := getGroundCornerType(pos, ground, width, height)
	if groundCornerType == CornerType90 || groundCornerType == CornerType270 {
		return -200 + distToCenter // Strongest preference for ground corners
	}

	// Calculate distance from edges
	distToEdge := minDistanceToEdge(pos, width, height)

	// Prefer positions near edges (within turretEdgePreference) or corners
	// But also prefer positions closer to center among valid positions
	edgeBonus := 0
	if distToEdge <= turretEdgePreference {
		edgeBonus = -100 // Strong preference for edge positions
	}

	// Check if it's a corner-like position (near two edges)
	isCornerLike := isNearCorner(pos, width, height, turretEdgePreference)
	if isCornerLike {
		edgeBonus -= 50 // Extra bonus for room corners
	}

	// Combine: edge bonus + distance from center (prefer closer to center among valid positions)
	return edgeBonus + distToCenter
}

// CornerType represents the type of corner a ground tile forms
type CornerType int

const (
	CornerTypeNone CornerType = iota
	CornerType90              // Right angle: 2 adjacent ground tiles at 90°
	CornerType270             // Inner corner: 3 adjacent ground tiles at 270°
)

// getGroundCornerType determines if a position is at a ground corner (90° or 270°)
// A 90° corner has exactly 2 orthogonal neighbors that are adjacent to each other (L-shape)
// A 270° corner has exactly 3 orthogonal neighbors (inverted L-shape)
func getGroundCornerType(pos Point, ground [][]int, width, height int) CornerType {
	x, y := pos.X, pos.Y

	// Count orthogonal ground neighbors
	// 0=top, 1=right, 2=bottom, 3=left
	neighbors := [4]bool{}
	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}

	groundCount := 0
	for i := 0; i < 4; i++ {
		nx, ny := x+dx[i], y+dy[i]
		if nx >= 0 && nx < width && ny >= 0 && ny < height && ground[ny][nx] == 1 {
			neighbors[i] = true
			groundCount++
		}
	}

	// 90° right angle: exactly 2 adjacent neighbors forming an L
	// Valid L-shapes: top+right, right+bottom, bottom+left, left+top
	if groundCount == 2 {
		if (neighbors[0] && neighbors[1]) || // top + right
			(neighbors[1] && neighbors[2]) || // right + bottom
			(neighbors[2] && neighbors[3]) || // bottom + left
			(neighbors[3] && neighbors[0]) { // left + top
			return CornerType90
		}
	}

	// 270° inner corner: exactly 3 neighbors (one side missing)
	if groundCount == 3 {
		return CornerType270
	}

	return CornerTypeNone
}

// minDistanceToEdge calculates the minimum distance to any edge
func minDistanceToEdge(pos Point, width, height int) int {
	distLeft := pos.X
	distRight := width - 1 - pos.X
	distTop := pos.Y
	distBottom := height - 1 - pos.Y

	minDist := distLeft
	if distRight < minDist {
		minDist = distRight
	}
	if distTop < minDist {
		minDist = distTop
	}
	if distBottom < minDist {
		minDist = distBottom
	}
	return minDist
}

// isNearCorner checks if the position is near a corner
func isNearCorner(pos Point, width, height, threshold int) bool {
	// Near top-left
	if pos.X <= threshold && pos.Y <= threshold {
		return true
	}
	// Near top-right
	if pos.X >= width-1-threshold && pos.Y <= threshold {
		return true
	}
	// Near bottom-left
	if pos.X <= threshold && pos.Y >= height-1-threshold {
		return true
	}
	// Near bottom-right
	if pos.X >= width-1-threshold && pos.Y >= height-1-threshold {
		return true
	}
	return false
}

// filterTurretsTooClose removes positions that are too close to the newly placed turret
func filterTurretsTooClose(positions []Point, placedPos Point) []Point {
	var filtered []Point
	for _, pos := range positions {
		dist := manhattanDistance(pos, placedPos)
		if dist >= turretMinTurretDistance {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

// checkTurretConnectivityAfterPlacement checks if all doors remain connected after placing a turret
func checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, newTurretPos Point, width, height int) bool {
	if len(doorPositions) < 2 {
		return true
	}

	// Create a temporary walkable map
	walkable := make([][]bool, height)
	for y := 0; y < height; y++ {
		walkable[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			// Walkable if ground=1 and not blocked by static or turret
			walkable[y][x] = ground[y][x] == 1 && staticLayer[y][x] == 0 && turretLayer[y][x] == 0
		}
	}

	// Block the new turret position
	walkable[newTurretPos.Y][newTurretPos.X] = false

	// Get door positions as a slice
	doors := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		doors = append(doors, pos)
	}

	// Check if all doors are connected using BFS from the first door
	startDoor := doors[0]
	visited := bfsConnectivity(walkable, startDoor, width, height)

	// Check if all other doors are reachable
	for _, door := range doors[1:] {
		if !visited[door.Y][door.X] {
			return false
		}
	}

	return true
}

// ============================================================================
// Mob Ground Layer Generation
// ============================================================================

// MobGroundStrategy represents placement strategy for mob ground
type MobGroundStrategy int

const (
	MobGroundStrategyLargeOpenArea MobGroundStrategy = iota // Place in large open area from center outward
	MobGroundStrategyNearDoors                              // Place near doors
	MobGroundStrategyCenterOutward                          // Place from center outward
)

const (
	mobGroundMinDoorDistance = 2 // Minimum distance from doors
)

// generateMobGroundLayer generates the mob ground layer with the given constraints
func generateMobGroundLayer(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) {

	if targetCount <= 0 {
		return
	}

	// Step 1: Divide count into 2-3 groups
	groups := divideMobGroundIntoGroups(targetCount)
	if len(groups) == 0 {
		return
	}

	// Step 2: Select strategies for each group (no duplicates)
	availableStrategies := []MobGroundStrategy{
		MobGroundStrategyLargeOpenArea,
		MobGroundStrategyNearDoors,
		MobGroundStrategyCenterOutward,
	}

	// Shuffle strategies
	for i := len(availableStrategies) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		availableStrategies[i], availableStrategies[j] = availableStrategies[j], availableStrategies[i]
	}

	// Check if large open area strategy is viable
	largeOpenAreaCenter := findLargeOpenAreaCenter(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height)
	if largeOpenAreaCenter.X < 0 {
		// Remove large open area strategy if not viable
		for i, s := range availableStrategies {
			if s == MobGroundStrategyLargeOpenArea {
				availableStrategies = append(availableStrategies[:i], availableStrategies[i+1:]...)
				break
			}
		}
	}

	// Step 3: Execute placement for each group
	centerX, centerY := width/2, height/2

	for i, groupCount := range groups {
		if i >= len(availableStrategies) {
			break
		}

		strategy := availableStrategies[i]
		placed := executeMobGroundStrategy(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer,
			doorPositions, width, height, groupCount, strategy, centerX, centerY, largeOpenAreaCenter)

		if placed == 0 {
			// Strategy failed, continue with next group
			continue
		}
	}
}

// divideMobGroundIntoGroups divides the target count into 2-3 groups
func divideMobGroundIntoGroups(targetCount int) []int {
	if targetCount <= 0 {
		return nil
	}

	if targetCount == 1 {
		return []int{1}
	}

	if targetCount == 2 {
		return []int{1, 1}
	}

	// Try 3 groups first
	groupCount := 3
	if targetCount < 3 {
		groupCount = 2
	}

	baseSize := targetCount / groupCount
	remainder := targetCount % groupCount

	groups := make([]int, groupCount)
	for i := 0; i < groupCount; i++ {
		groups[i] = baseSize
		if i < remainder {
			groups[i]++
		}
	}

	// Merge groups if any has less than 1
	for i := len(groups) - 1; i >= 0; i-- {
		if groups[i] < 1 && len(groups) > 1 {
			if i > 0 {
				groups[i-1] += groups[i]
			} else if len(groups) > 1 {
				groups[1] += groups[0]
			}
			groups = append(groups[:i], groups[i+1:]...)
		}
	}

	return groups
}

// executeMobGroundStrategy executes a specific placement strategy
func executeMobGroundStrategy(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int,
	strategy MobGroundStrategy, centerX, centerY int, largeOpenAreaCenter Point) int {

	placed := 0
	remaining := targetCount
	maxAttempts := targetCount * 3

	for remaining > 0 && maxAttempts > 0 {
		// Find valid positions based on strategy
		validPositions := findValidMobGroundPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer,
			doorPositions, width, height)

		if len(validPositions) == 0 {
			break
		}

		// Sort positions based on strategy
		sortMobGroundPositionsByStrategy(validPositions, strategy, centerX, centerY, width, height, doorPositions, largeOpenAreaCenter)

		// Try to place (prefer 2x2, fallback to 1x1)
		placedOne := false
		for _, pos := range validPositions {
			// Try 2x2 first
			if canPlace2x2MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				place2x2MobGround(mobGroundLayer, pos)
				placed++
				remaining--
				placedOne = true
				break
			}

			// Try 1x1
			if canPlace1x1MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				mobGroundLayer[pos.Y][pos.X] = 1
				placed++
				remaining--
				placedOne = true
				break
			}
		}

		if !placedOne {
			maxAttempts--
		}
	}

	return placed
}

// findValidMobGroundPositions finds all valid positions for mob ground placement
func findValidMobGroundPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) []Point {

	var positions []Point

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidMobGroundPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// isValidMobGroundPosition checks if a single cell is valid for mob ground
func isValidMobGroundPosition(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {

	x, y := pos.X, pos.Y

	// Check bounds
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Must be on ground
	if ground[y][x] != 1 {
		return false
	}

	// Must not overlap with other layers
	if softEdge[y][x] != 0 || bridge[y][x] != 0 || staticLayer[y][x] != 0 || turretLayer[y][x] != 0 {
		return false
	}

	// Must not already have mob ground
	if mobGroundLayer[y][x] != 0 {
		return false
	}

	// Must be at least 2 cells away from doors
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < mobGroundMinDoorDistance {
			return false
		}
	}

	// Must not touch existing mob ground
	if touchesExistingMobGround(pos, mobGroundLayer, width, height) {
		return false
	}

	return true
}

// touchesExistingMobGround checks if placing mob ground at pos would touch existing mob ground
func touchesExistingMobGround(pos Point, mobGroundLayer [][]int, width, height int) bool {
	// Check all 8 neighbors (including diagonals)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if mobGroundLayer[ny][nx] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// canPlace2x2MobGround checks if a 2x2 mob ground can be placed at the given top-left corner
func canPlace2x2MobGround(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {

	// Check all 4 cells
	for dy := 0; dy < 2; dy++ {
		for dx := 0; dx < 2; dx++ {
			checkPos := Point{X: pos.X + dx, Y: pos.Y + dy}
			if !isValidMobGroundPosition(checkPos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				return false
			}
		}
	}

	// Check that 2x2 doesn't touch existing mob ground (expanded check)
	if touches2x2ExistingMobGround(pos, mobGroundLayer, width, height) {
		return false
	}

	return true
}

// touches2x2ExistingMobGround checks if placing a 2x2 mob ground would touch existing mob ground
func touches2x2ExistingMobGround(pos Point, mobGroundLayer [][]int, width, height int) bool {
	// Check a 4x4 area around the 2x2 placement (1 cell buffer on each side)
	for dy := -1; dy <= 2; dy++ {
		for dx := -1; dx <= 2; dx++ {
			// Skip the cells that will be occupied
			if dx >= 0 && dx < 2 && dy >= 0 && dy < 2 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if mobGroundLayer[ny][nx] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// canPlace1x1MobGround checks if a 1x1 mob ground can be placed
func canPlace1x1MobGround(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {
	return isValidMobGroundPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height)
}

// place2x2MobGround places a 2x2 mob ground at the given top-left corner
func place2x2MobGround(mobGroundLayer [][]int, pos Point) {
	for dy := 0; dy < 2; dy++ {
		for dx := 0; dx < 2; dx++ {
			mobGroundLayer[pos.Y+dy][pos.X+dx] = 1
		}
	}
}

// sortMobGroundPositionsByStrategy sorts positions based on the placement strategy
func sortMobGroundPositionsByStrategy(positions []Point, strategy MobGroundStrategy, centerX, centerY, width, height int,
	doorPositions map[DoorPosition]Point, largeOpenAreaCenter Point) {

	switch strategy {
	case MobGroundStrategyLargeOpenArea:
		// Sort by distance from large open area center (closest first)
		for i := 0; i < len(positions)-1; i++ {
			for j := i + 1; j < len(positions); j++ {
				distI := manhattanDistance(positions[i], largeOpenAreaCenter)
				distJ := manhattanDistance(positions[j], largeOpenAreaCenter)
				if distJ < distI {
					positions[i], positions[j] = positions[j], positions[i]
				}
			}
		}

	case MobGroundStrategyNearDoors:
		// Sort by distance from nearest door (closest first)
		for i := 0; i < len(positions)-1; i++ {
			for j := i + 1; j < len(positions); j++ {
				distI := minDistanceToDoor(positions[i], doorPositions)
				distJ := minDistanceToDoor(positions[j], doorPositions)
				if distJ < distI {
					positions[i], positions[j] = positions[j], positions[i]
				}
			}
		}

	case MobGroundStrategyCenterOutward:
		// Sort by distance from center (closest first)
		center := Point{X: centerX, Y: centerY}
		for i := 0; i < len(positions)-1; i++ {
			for j := i + 1; j < len(positions); j++ {
				distI := manhattanDistance(positions[i], center)
				distJ := manhattanDistance(positions[j], center)
				if distJ < distI {
					positions[i], positions[j] = positions[j], positions[i]
				}
			}
		}
	}
}

// minDistanceToDoor returns the minimum distance to any door
func minDistanceToDoor(pos Point, doorPositions map[DoorPosition]Point) int {
	minDist := 999999
	for _, doorPos := range doorPositions {
		dist := manhattanDistance(pos, doorPos)
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

// findLargeOpenAreaCenter finds the center of a large open area connected to doors
// Returns Point{-1, -1} if no suitable large open area exists
func findLargeOpenAreaCenter(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) Point {

	// Find all connected regions of walkable ground
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	// Find the largest region that is connected to at least one door
	var bestCenter Point
	bestSize := 0
	minSizeThreshold := (width * height) / 10 // At least 10% of total area

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if visited[y][x] {
				continue
			}

			pos := Point{X: x, Y: y}
			if !isWalkableForMobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, width, height) {
				visited[y][x] = true
				continue
			}

			// BFS to find connected region
			region := findConnectedRegion(pos, ground, softEdge, bridge, staticLayer, turretLayer, visited, width, height)
			if len(region) == 0 {
				continue
			}

			// Check if region is connected to any door
			connectedToDoor := false
			for _, doorPos := range doorPositions {
				for _, regionPos := range region {
					if manhattanDistance(regionPos, doorPos) <= 2 {
						connectedToDoor = true
						break
					}
				}
				if connectedToDoor {
					break
				}
			}

			if connectedToDoor && len(region) > bestSize && len(region) >= minSizeThreshold {
				bestSize = len(region)
				bestCenter = calculateRegionCenter(region)
			}
		}
	}

	if bestSize == 0 {
		return Point{X: -1, Y: -1}
	}

	return bestCenter
}

// isWalkableForMobGround checks if a cell is walkable for mob ground placement consideration
func isWalkableForMobGround(pos Point, ground, softEdge, bridge, staticLayer, turretLayer [][]int, width, height int) bool {
	x, y := pos.X, pos.Y
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}
	return ground[y][x] == 1 && softEdge[y][x] == 0 && bridge[y][x] == 0 && staticLayer[y][x] == 0 && turretLayer[y][x] == 0
}

// findConnectedRegion finds all cells connected to the starting point
func findConnectedRegion(start Point, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	visited [][]bool, width, height int) []Point {

	var region []Point
	queue := []Point{start}
	visited[start.Y][start.X] = true

	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		region = append(region, curr)

		for i := 0; i < 4; i++ {
			nx, ny := curr.X+dx[i], curr.Y+dy[i]
			if nx >= 0 && nx < width && ny >= 0 && ny < height && !visited[ny][nx] {
				nextPos := Point{X: nx, Y: ny}
				if isWalkableForMobGround(nextPos, ground, softEdge, bridge, staticLayer, turretLayer, width, height) {
					visited[ny][nx] = true
					queue = append(queue, nextPos)
				}
			}
		}
	}

	return region
}

// calculateRegionCenter calculates the center of a region
func calculateRegionCenter(region []Point) Point {
	if len(region) == 0 {
		return Point{X: -1, Y: -1}
	}

	sumX, sumY := 0, 0
	for _, pos := range region {
		sumX += pos.X
		sumY += pos.Y
	}

	return Point{
		X: sumX / len(region),
		Y: sumY / len(region),
	}
}

// ============================================================================
// Mob Air (Fly) Layer Generation
// ============================================================================

// MobAirStrategy represents placement strategy for mob air
type MobAirStrategy int

const (
	MobAirStrategyCenterOutward MobAirStrategy = iota // Place from center outward
	MobAirStrategyEvenlySpaced                        // Distribute with roughly equal spacing
)

const (
	mobAirMinDoorDistance = 4 // Minimum distance from doors
	mobAirMinEdgeDistance = 2 // Minimum distance from room edges
)

// generateMobAirLayer generates the mob air layer with the given constraints
func generateMobAirLayer(mobAirLayer, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) {

	if targetCount <= 0 {
		return
	}

	// Select strategy randomly
	strategy := MobAirStrategy(rand.Intn(2))

	// Find all valid positions
	validPositions := findValidMobAirPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
		doorPositions, width, height)

	if len(validPositions) == 0 {
		return
	}

	// Sort/arrange positions based on strategy
	centerX, centerY := width/2, height/2

	switch strategy {
	case MobAirStrategyCenterOutward:
		sortMobAirPositionsCenterOutward(validPositions, centerX, centerY)
	case MobAirStrategyEvenlySpaced:
		validPositions = arrangeMobAirEvenlySpaced(validPositions, targetCount, width, height)
	}

	// Place mob air
	remaining := targetCount
	for _, pos := range validPositions {
		if remaining <= 0 {
			break
		}

		// Verify position is still valid (may have been invalidated by previous placements)
		if !isValidMobAirPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
			doorPositions, width, height) {
			continue
		}

		// Place mob air
		mobAirLayer[pos.Y][pos.X] = 1
		remaining--
	}
}

// findValidMobAirPositions finds all valid positions for mob air placement
func findValidMobAirPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) []Point {

	var positions []Point

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidMobAirPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
				doorPositions, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// isValidMobAirPosition checks if a single cell is valid for mob air
// Note: Mob Air (flying mobs) do NOT require ground=1, they can spawn anywhere
func isValidMobAirPosition(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {

	x, y := pos.X, pos.Y

	// Check bounds
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Must be at least 2 cells away from room edges
	if x < mobAirMinEdgeDistance || x >= width-mobAirMinEdgeDistance ||
		y < mobAirMinEdgeDistance || y >= height-mobAirMinEdgeDistance {
		return false
	}

	// No ground requirement - flying mobs can spawn anywhere

	// Must not overlap with other layers
	if softEdge[y][x] != 0 || bridge[y][x] != 0 || staticLayer[y][x] != 0 ||
		turretLayer[y][x] != 0 || mobGroundLayer[y][x] != 0 {
		return false
	}

	// Must not already have mob air
	if mobAirLayer[y][x] != 0 {
		return false
	}

	// Must be at least 4 cells away from doors
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < mobAirMinDoorDistance {
			return false
		}
	}

	// Must not touch existing mob air
	if touchesExistingMobAir(pos, mobAirLayer, width, height) {
		return false
	}

	return true
}

// touchesExistingMobAir checks if placing mob air at pos would touch existing mob air
func touchesExistingMobAir(pos Point, mobAirLayer [][]int, width, height int) bool {
	// Check all 8 neighbors (including diagonals)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if mobAirLayer[ny][nx] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// sortMobAirPositionsCenterOutward sorts positions by distance from center (closest first)
func sortMobAirPositionsCenterOutward(positions []Point, centerX, centerY int) {
	center := Point{X: centerX, Y: centerY}
	for i := 0; i < len(positions)-1; i++ {
		for j := i + 1; j < len(positions); j++ {
			distI := manhattanDistance(positions[i], center)
			distJ := manhattanDistance(positions[j], center)
			if distJ < distI {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}
}

// arrangeMobAirEvenlySpaced arranges positions to be roughly evenly spaced
// Returns a subset of positions that are well-distributed across the map
// The grid is calculated based on targetCount to ensure even distribution
func arrangeMobAirEvenlySpaced(validPositions []Point, targetCount int, width, height int) []Point {
	if len(validPositions) == 0 || targetCount <= 0 {
		return nil
	}

	if len(validPositions) <= targetCount {
		return validPositions
	}

	// Calculate grid dimensions based on targetCount
	// We want to create a grid where gridCols * gridRows >= targetCount
	// and the grid cells are as square as possible
	gridCols, gridRows := calculateGridDimensions(targetCount, width, height)

	// Create a grid-based selection
	selected := make([]Point, 0, targetCount)
	used := make(map[Point]bool)

	// Calculate cell size
	cellWidth := float64(width) / float64(gridCols)
	cellHeight := float64(height) / float64(gridRows)

	// Iterate through grid cells and find nearest valid position to each cell center
	for row := 0; row < gridRows && len(selected) < targetCount; row++ {
		for col := 0; col < gridCols && len(selected) < targetCount; col++ {
			// Calculate ideal position at cell center
			idealX := int(float64(col)*cellWidth + cellWidth/2)
			idealY := int(float64(row)*cellHeight + cellHeight/2)
			idealPos := Point{X: idealX, Y: idealY}

			// Find nearest valid position to this ideal position
			nearest := findNearestValidPosition(validPositions, idealPos, used)
			if nearest.X >= 0 {
				selected = append(selected, nearest)
				used[nearest] = true
			}
		}
	}

	// If we still need more, fill from remaining valid positions
	for _, pos := range validPositions {
		if len(selected) >= targetCount {
			break
		}
		if !used[pos] {
			selected = append(selected, pos)
			used[pos] = true
		}
	}

	return selected
}

// calculateGridDimensions calculates grid cols and rows based on target count
// The grid is designed to distribute targetCount items evenly across width x height
func calculateGridDimensions(targetCount, width, height int) (cols, rows int) {
	if targetCount <= 0 {
		return 1, 1
	}

	if targetCount == 1 {
		return 1, 1
	}

	// Calculate aspect ratio
	aspectRatio := float64(width) / float64(height)

	// Calculate grid dimensions that:
	// 1. Have cols * rows >= targetCount
	// 2. Maintain aspect ratio similar to room dimensions
	// 3. Create roughly square cells

	// Start with sqrt(targetCount) and adjust for aspect ratio
	sqrtCount := int(float64(targetCount)*0.5 + 0.5)
	if sqrtCount < 1 {
		sqrtCount = 1
	}

	// Adjust cols and rows based on aspect ratio
	if aspectRatio >= 1 {
		// Wider than tall
		cols = int(float64(sqrtCount)*aspectRatio + 0.5)
		if cols < 1 {
			cols = 1
		}
		rows = (targetCount + cols - 1) / cols
		if rows < 1 {
			rows = 1
		}
	} else {
		// Taller than wide
		rows = int(float64(sqrtCount)/aspectRatio + 0.5)
		if rows < 1 {
			rows = 1
		}
		cols = (targetCount + rows - 1) / rows
		if cols < 1 {
			cols = 1
		}
	}

	// Ensure we have enough cells
	for cols*rows < targetCount {
		if float64(width)/float64(cols) > float64(height)/float64(rows) {
			cols++
		} else {
			rows++
		}
	}

	return cols, rows
}

// findNearestValidPosition finds the valid position nearest to the target
func findNearestValidPosition(validPositions []Point, target Point, used map[Point]bool) Point {
	bestPos := Point{X: -1, Y: -1}
	bestDist := 999999

	for _, pos := range validPositions {
		if used[pos] {
			continue
		}

		dist := manhattanDistance(pos, target)
		if dist < bestDist {
			bestDist = dist
			bestPos = pos
		}
	}

	return bestPos
}

// ============================================================================
// Soft Edge Layer Generation
// ============================================================================

const (
	softEdgeMinDoorDistance = 2 // Minimum distance from doors
	softEdgeMinLength       = 3 // Minimum length (N > 2, so N >= 3)
)

// SoftEdgePlacement represents a potential soft edge placement
type SoftEdgePlacement struct {
	StartX, StartY int
	Width, Height  int // Either (1, N) or (N, 1) where N >= 3
}

// generateSoftEdgeLayerWithDebug generates the soft edge layer with debug info
// Soft edges are 1×N or N×1 strips (N > 2) placed in ground concave areas
func generateSoftEdgeLayerWithDebug(softEdgeLayer, ground [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *SoftEdgeDebugInfo {
	debug := &SoftEdgeDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	if targetCount <= 0 {
		debug.Skipped = true
		debug.SkipReason = "targetCount is 0"
		return debug
	}

	// Find all valid soft edge placements (concave areas)
	placements := findValidSoftEdgePlacements(ground, softEdgeLayer, doorPositions, width, height)
	if len(placements) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid concave areas found in ground layer",
		})
		return debug
	}

	// Shuffle placements for variety
	for i := len(placements) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		placements[i], placements[j] = placements[j], placements[i]
	}

	// Place soft edges until target count reached or placements exhausted
	remaining := targetCount
	overlapCount := 0
	for _, placement := range placements {
		if remaining <= 0 {
			break
		}

		// Verify placement is still valid (not overlapping with already placed)
		if !canPlaceSoftEdge(placement, softEdgeLayer, width, height) {
			overlapCount++
			continue
		}

		// Place the soft edge
		placeSoftEdge(softEdgeLayer, placement)
		remaining--
		debug.PlacedCount++

		// Record placement
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", placement.StartX, placement.StartY),
			Size:     fmt.Sprintf("%dx%d", placement.Width, placement.Height),
			Reason:   "ground concave area",
		})
	}

	// Record miss info
	if overlapCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "overlapping with already placed soft edge",
			Count:  overlapCount,
		})
	}
	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("only %d valid placements available, needed %d more", len(placements), remaining),
		})
	}

	return debug
}

// findValidSoftEdgePlacements finds all valid positions for soft edge placement
// A valid soft edge is a 1×N or N×1 strip (N >= 3) in a ground concave area
func findValidSoftEdgePlacements(ground, softEdgeLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) []SoftEdgePlacement {
	var placements []SoftEdgePlacement

	// Find horizontal soft edges (1×N, height=1, width=N)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Try to find a horizontal concave area starting at (x, y)
			if placement := findHorizontalConcave(ground, softEdgeLayer, doorPositions, x, y, width, height); placement != nil {
				placements = append(placements, *placement)
			}
		}
	}

	// Find vertical soft edges (N×1, height=N, width=1)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Try to find a vertical concave area starting at (x, y)
			if placement := findVerticalConcave(ground, softEdgeLayer, doorPositions, x, y, width, height); placement != nil {
				placements = append(placements, *placement)
			}
		}
	}

	return placements
}

// findHorizontalConcave finds a horizontal concave area (1×N) starting at (x, y)
// A horizontal concave is a void notch: void cells with ground on one horizontal edge (top or bottom)
// and ground cells on both ends (left and right), forming a U-shaped depression
func findHorizontalConcave(ground, softEdgeLayer [][]int, doorPositions map[DoorPosition]Point, startX, startY, width, height int) *SoftEdgePlacement {
	// Check if starting position is valid
	if startX >= width || startY >= height {
		return nil
	}

	// Starting position must be VOID (this is a void notch)
	if ground[startY][startX] != 0 {
		return nil
	}

	// Must have ground immediately to the left (this is the start of the notch)
	if startX == 0 || ground[startY][startX-1] != 1 {
		return nil
	}

	// Check door distance for starting position
	if !isFarEnoughFromDoors(startX, startY, doorPositions, softEdgeMinDoorDistance) {
		return nil
	}

	// Determine if this is a top-notch (ground below) or bottom-notch (ground above)
	hasGroundAbove := startY > 0 && ground[startY-1][startX] == 1
	hasGroundBelow := startY < height-1 && ground[startY+1][startX] == 1

	// Must have ground on exactly one horizontal side (forming a U-shape)
	if !hasGroundAbove && !hasGroundBelow {
		return nil // Not a concave notch - no ground on either horizontal side
	}
	if hasGroundAbove && hasGroundBelow {
		return nil // This is a tunnel, not a notch
	}

	// Find the length of this horizontal notch
	length := 1
	for x := startX + 1; x < width; x++ {
		// Must be void to continue the notch
		if ground[startY][x] != 0 {
			break
		}

		// Must maintain the same concave property
		gAbove := startY > 0 && ground[startY-1][x] == 1
		gBelow := startY < height-1 && ground[startY+1][x] == 1

		if hasGroundAbove && !gAbove {
			break // Ground above ended
		}
		if hasGroundBelow && !gBelow {
			break // Ground below ended
		}

		// Check door distance
		if !isFarEnoughFromDoors(x, startY, doorPositions, softEdgeMinDoorDistance) {
			break
		}

		length++
	}

	// Check if there's ground on the right side (closing the notch)
	endX := startX + length
	if endX >= width || ground[startY][endX] != 1 {
		return nil // Notch is open on the right, not a proper concave
	}

	// Must be at least 3 cells long
	if length < softEdgeMinLength {
		return nil
	}

	return &SoftEdgePlacement{
		StartX: startX,
		StartY: startY,
		Width:  length,
		Height: 1,
	}
}

// findVerticalConcave finds a vertical concave area (N×1) starting at (x, y)
// A vertical concave is a void notch: void cells with ground on one vertical edge (left or right)
// and ground cells on both ends (top and bottom), forming a U-shaped depression
func findVerticalConcave(ground, softEdgeLayer [][]int, doorPositions map[DoorPosition]Point, startX, startY, width, height int) *SoftEdgePlacement {
	// Check if starting position is valid
	if startX >= width || startY >= height {
		return nil
	}

	// Starting position must be VOID (this is a void notch)
	if ground[startY][startX] != 0 {
		return nil
	}

	// Must have ground immediately above (this is the start of the notch)
	if startY == 0 || ground[startY-1][startX] != 1 {
		return nil
	}

	// Check door distance for starting position
	if !isFarEnoughFromDoors(startX, startY, doorPositions, softEdgeMinDoorDistance) {
		return nil
	}

	// Determine if this is a left-notch (ground to the right) or right-notch (ground to the left)
	hasGroundLeft := startX > 0 && ground[startY][startX-1] == 1
	hasGroundRight := startX < width-1 && ground[startY][startX+1] == 1

	// Must have ground on exactly one vertical side (forming a U-shape)
	if !hasGroundLeft && !hasGroundRight {
		return nil // Not a concave notch - no ground on either vertical side
	}
	if hasGroundLeft && hasGroundRight {
		return nil // This is a tunnel, not a notch
	}

	// Find the length of this vertical notch
	length := 1
	for y := startY + 1; y < height; y++ {
		// Must be void to continue the notch
		if ground[y][startX] != 0 {
			break
		}

		// Must maintain the same concave property
		gLeft := startX > 0 && ground[y][startX-1] == 1
		gRight := startX < width-1 && ground[y][startX+1] == 1

		if hasGroundLeft && !gLeft {
			break // Ground on left ended
		}
		if hasGroundRight && !gRight {
			break // Ground on right ended
		}

		// Check door distance
		if !isFarEnoughFromDoors(startX, y, doorPositions, softEdgeMinDoorDistance) {
			break
		}

		length++
	}

	// Check if there's ground on the bottom side (closing the notch)
	endY := startY + length
	if endY >= height || ground[endY][startX] != 1 {
		return nil // Notch is open on the bottom, not a proper concave
	}

	// Must be at least 3 cells long
	if length < softEdgeMinLength {
		return nil
	}

	return &SoftEdgePlacement{
		StartX: startX,
		StartY: startY,
		Width:  1,
		Height: length,
	}
}

// isFarEnoughFromDoors checks if a position is far enough from all doors
func isFarEnoughFromDoors(x, y int, doorPositions map[DoorPosition]Point, minDistance int) bool {
	pos := Point{X: x, Y: y}
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < minDistance {
			return false
		}
	}
	return true
}

// canPlaceSoftEdge checks if a soft edge can be placed (not overlapping)
func canPlaceSoftEdge(placement SoftEdgePlacement, softEdgeLayer [][]int, width, height int) bool {
	for dy := 0; dy < placement.Height; dy++ {
		for dx := 0; dx < placement.Width; dx++ {
			x := placement.StartX + dx
			y := placement.StartY + dy

			if x >= width || y >= height {
				return false
			}

			if softEdgeLayer[y][x] != 0 {
				return false
			}
		}
	}
	return true
}

// placeSoftEdge places a soft edge on the layer
func placeSoftEdge(softEdgeLayer [][]int, placement SoftEdgePlacement) {
	for dy := 0; dy < placement.Height; dy++ {
		for dx := 0; dx < placement.Width; dx++ {
			x := placement.StartX + dx
			y := placement.StartY + dy
			softEdgeLayer[y][x] = 1
		}
	}
}

// ============================================================================
// Debug-enabled layer generation functions
// ============================================================================

// generateStaticLayerWithDebug generates the static layer with debug info
func generateStaticLayerWithDebug(staticLayer, ground, softEdge, bridge [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *StaticDebugInfo {
	debug := &StaticDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	// Get all door cells and their adjacent cells (forbidden zone)
	forbiddenCells := getDoorForbiddenCells(doorPositions, width, height)

	// Find all valid 2x2 positions for static placement
	validPositions := findValidStaticPositions(ground, softEdge, bridge, staticLayer, forbiddenCells, width, height)
	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid 2x2 positions found (all positions blocked by ground, doors, softEdge, or bridge)",
		})
		return debug
	}

	// Sort positions by strategy (will be re-sorted on each strategy switch)
	centerX, centerY := width/2, height/2
	currentStrategy := StrategyCenterOutward

	remaining := targetCount
	strategyAttempts := 0
	maxStrategyAttempts := 2 * targetCount // Prevent infinite loop
	invalidatedCount := 0
	connectivityBlockedCount := 0

	for remaining > 0 && strategyAttempts < maxStrategyAttempts {
		// Sort valid positions based on current strategy
		sortPositionsByStrategy(validPositions, currentStrategy, centerX, centerY, width, height)

		strategyName := "center_outward"
		if currentStrategy == StrategyEdgeInward {
			strategyName = "edge_inward"
		}

		// Try to place one static
		placed := false
		for i, pos := range validPositions {
			// Check if this position is still valid (may have been invalidated by previous placements)
			if !isValidStaticPosition(pos, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height) {
				invalidatedCount++
				continue
			}

			// Check connectivity after placement
			if !checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			// Place the static (2x2)
			placeStatic(staticLayer, pos)
			remaining--
			placed = true
			debug.PlacedCount++

			// Record placement info
			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "2x2",
				Reason:   fmt.Sprintf("strategy: %s, valid position with connectivity preserved", strategyName),
			})

			// Remove this position and update valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that now touch this static
			validPositions = filterTouchingPositions(validPositions, pos)
			break
		}

		if !placed {
			// Switch strategy and try again
			strategyAttempts++
		}

		// Alternate strategy after each placement or failed attempt
		if currentStrategy == StrategyCenterOutward {
			currentStrategy = StrategyEdgeInward
		} else {
			currentStrategy = StrategyCenterOutward
		}
	}

	// Record miss info
	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated by previous placement (touching existing static)",
			Count:  invalidatedCount,
		})
	}
	if connectivityBlockedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position would block door connectivity",
			Count:  connectivityBlockedCount,
		})
	}
	if remaining > 0 && strategyAttempts >= maxStrategyAttempts {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("reached max strategy attempts (%d), could not place %d more statics", maxStrategyAttempts, remaining),
		})
	} else if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all %d valid positions, needed %d more", len(validPositions), remaining),
		})
	}

	return debug
}

// generateTurretLayerWithDebug generates the turret layer with debug info
func generateTurretLayerWithDebug(turretLayer, ground, softEdge, bridge, staticLayer [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *TurretDebugInfo {
	debug := &TurretDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	// Find all valid positions for turret placement
	validPositions := findValidTurretPositions(ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height)
	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid positions found (all positions blocked by ground, doors, static, softEdge, or bridge)",
		})
		return debug
	}

	initialValidCount := len(validPositions)

	// Sort positions by preference (ground corners first, then room corners and edges, then by distance from center)
	centerX, centerY := width/2, height/2
	sortTurretPositionsByPreference(validPositions, centerX, centerY, width, height, ground)

	remaining := targetCount
	maxAttempts := 2 * targetCount // Prevent infinite loop
	invalidatedCount := 0
	connectivityBlockedCount := 0

	for remaining > 0 && maxAttempts > 0 {
		placed := false

		for i, pos := range validPositions {
			// Check if this position is still valid
			if !isValidTurretPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height) {
				invalidatedCount++
				continue
			}

			// Check connectivity after placement
			if !checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			// Determine reason based on position characteristics
			reason := getTurretPlacementReason(pos, centerX, centerY, width, height, ground)

			// Place the turret (1 tile)
			turretLayer[pos.Y][pos.X] = 1
			remaining--
			placed = true
			debug.PlacedCount++

			// Record placement info
			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "1x1",
				Reason:   reason,
			})

			// Remove this position from valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that are too close to this turret
			validPositions = filterTurretsTooClose(validPositions, pos)
			break
		}

		if !placed {
			maxAttempts--
		}
	}

	// Record miss info
	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated (too close to existing turret or blocked)",
			Count:  invalidatedCount,
		})
	}
	if connectivityBlockedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position would block door connectivity",
			Count:  connectivityBlockedCount,
		})
	}
	if remaining > 0 && maxAttempts <= 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("reached max attempts, could not place %d more turrets", remaining),
		})
	} else if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all %d valid positions, needed %d more", initialValidCount, remaining),
		})
	}

	return debug
}

// getTurretPlacementReason returns a human-readable reason for turret placement
func getTurretPlacementReason(pos Point, centerX, centerY, width, height int, ground [][]int) string {
	cornerType := getGroundCornerType(pos, ground, width, height)
	if cornerType == CornerType90 {
		return "ground corner (90° right angle)"
	}
	if cornerType == CornerType270 {
		return "ground corner (270° inner corner)"
	}

	if isNearCorner(pos, width, height, turretEdgePreference) {
		return "near room corner"
	}

	distToEdge := minDistanceToEdge(pos, width, height)
	if distToEdge <= turretEdgePreference {
		return "near room edge"
	}

	return "center outward placement"
}

// generateMobGroundLayerWithDebug generates the mob ground layer with debug info
func generateMobGroundLayerWithDebug(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) *MobGroundDebugInfo {

	debug := &MobGroundDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Groups:      []MobGroupInfo{},
		Misses:      []MissInfo{},
	}

	if targetCount <= 0 {
		debug.Skipped = true
		debug.SkipReason = "targetCount is 0"
		return debug
	}

	// Step 1: Divide count into 2-3 groups
	groups := divideMobGroundIntoGroups(targetCount)
	if len(groups) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "failed to divide target count into groups",
		})
		return debug
	}

	// Step 2: Select strategies for each group (no duplicates)
	availableStrategies := []MobGroundStrategy{
		MobGroundStrategyLargeOpenArea,
		MobGroundStrategyNearDoors,
		MobGroundStrategyCenterOutward,
	}

	// Shuffle strategies
	for i := len(availableStrategies) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		availableStrategies[i], availableStrategies[j] = availableStrategies[j], availableStrategies[i]
	}

	// Check if large open area strategy is viable
	largeOpenAreaCenter := findLargeOpenAreaCenter(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height)
	if largeOpenAreaCenter.X < 0 {
		// Remove large open area strategy if not viable
		for i, s := range availableStrategies {
			if s == MobGroundStrategyLargeOpenArea {
				availableStrategies = append(availableStrategies[:i], availableStrategies[i+1:]...)
				break
			}
		}
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "large_open_area strategy not viable (no 4x4 open area found)",
		})
	}

	// Step 3: Execute placement for each group
	centerX, centerY := width/2, height/2

	for i, groupCount := range groups {
		if i >= len(availableStrategies) {
			debug.Misses = append(debug.Misses, MissInfo{
				Reason: fmt.Sprintf("group %d skipped: no more strategies available", i),
			})
			break
		}

		strategy := availableStrategies[i]
		strategyName := getMobGroundStrategyName(strategy)

		groupDebug := MobGroupInfo{
			GroupIndex:  i,
			Strategy:    strategyName,
			TargetCount: groupCount,
			PlacedCount: 0,
			Placements:  []PlaceInfo{},
			Misses:      []MissInfo{},
		}

		placed := executeMobGroundStrategyWithDebug(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer,
			doorPositions, width, height, groupCount, strategy, centerX, centerY, largeOpenAreaCenter, &groupDebug)

		groupDebug.PlacedCount = placed
		debug.PlacedCount += placed
		debug.Groups = append(debug.Groups, groupDebug)
	}

	return debug
}

// getMobGroundStrategyName returns the name of a mob ground strategy
func getMobGroundStrategyName(strategy MobGroundStrategy) string {
	switch strategy {
	case MobGroundStrategyLargeOpenArea:
		return "large_open_area"
	case MobGroundStrategyNearDoors:
		return "near_doors"
	case MobGroundStrategyCenterOutward:
		return "center_outward"
	default:
		return "unknown"
	}
}

// executeMobGroundStrategyWithDebug executes a specific placement strategy with debug info
func executeMobGroundStrategyWithDebug(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int,
	strategy MobGroundStrategy, centerX, centerY int, largeOpenAreaCenter Point, groupDebug *MobGroupInfo) int {

	placed := 0
	remaining := targetCount
	maxAttempts := targetCount * 3
	noValidPositionsCount := 0
	no2x2Or1x1Count := 0

	for remaining > 0 && maxAttempts > 0 {
		// Find valid positions based on strategy
		validPositions := findValidMobGroundPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer,
			doorPositions, width, height)

		if len(validPositions) == 0 {
			noValidPositionsCount++
			break
		}

		// Sort positions based on strategy
		sortMobGroundPositionsByStrategy(validPositions, strategy, centerX, centerY, width, height, doorPositions, largeOpenAreaCenter)

		// Try to place (prefer 2x2, fallback to 1x1)
		placedOne := false
		for _, pos := range validPositions {
			// Try 2x2 first
			if canPlace2x2MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				place2x2MobGround(mobGroundLayer, pos)
				placed++
				remaining--
				placedOne = true

				groupDebug.Placements = append(groupDebug.Placements, PlaceInfo{
					Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
					Size:     "2x2",
					Reason:   fmt.Sprintf("preferred 2x2 placement via %s strategy", getMobGroundStrategyName(strategy)),
				})
				break
			}

			// Try 1x1
			if canPlace1x1MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				mobGroundLayer[pos.Y][pos.X] = 1
				placed++
				remaining--
				placedOne = true

				groupDebug.Placements = append(groupDebug.Placements, PlaceInfo{
					Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
					Size:     "1x1",
					Reason:   fmt.Sprintf("fallback 1x1 placement via %s strategy", getMobGroundStrategyName(strategy)),
				})
				break
			}
		}

		if !placedOne {
			no2x2Or1x1Count++
			maxAttempts--
		}
	}

	// Record miss info for this group
	if noValidPositionsCount > 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: "no valid positions available (all blocked by ground/static/turret/doors/existing mobs)",
		})
	}
	if no2x2Or1x1Count > 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: "positions found but neither 2x2 nor 1x1 placement possible",
			Count:  no2x2Or1x1Count,
		})
	}
	if remaining > 0 && maxAttempts <= 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: fmt.Sprintf("reached max attempts, could not place %d more mobs", remaining),
		})
	} else if remaining > 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all valid positions, needed %d more", remaining),
		})
	}

	return placed
}

// generateMobAirLayerWithDebug generates the mob air layer with debug info
func generateMobAirLayerWithDebug(mobAirLayer, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) *MobAirDebugInfo {

	debug := &MobAirDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Strategy:    "",
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	if targetCount <= 0 {
		debug.Skipped = true
		debug.SkipReason = "targetCount is 0"
		return debug
	}

	// Select strategy randomly
	strategy := MobAirStrategy(rand.Intn(2))

	switch strategy {
	case MobAirStrategyCenterOutward:
		debug.Strategy = "center_outward"
	case MobAirStrategyEvenlySpaced:
		debug.Strategy = "evenly_spaced"
	}

	// Find all valid positions
	validPositions := findValidMobAirPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
		doorPositions, width, height)

	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid positions found (all positions blocked by static/turret/mobGround/doors or too close to room edges)",
		})
		return debug
	}

	initialValidCount := len(validPositions)

	// Sort/arrange positions based on strategy
	centerX, centerY := width/2, height/2

	switch strategy {
	case MobAirStrategyCenterOutward:
		sortMobAirPositionsCenterOutward(validPositions, centerX, centerY)
	case MobAirStrategyEvenlySpaced:
		validPositions = arrangeMobAirEvenlySpaced(validPositions, targetCount, width, height)
	}

	// Place mob air
	remaining := targetCount
	invalidatedCount := 0
	for _, pos := range validPositions {
		if remaining <= 0 {
			break
		}

		// Verify position is still valid (may have been invalidated by previous placements)
		if !isValidMobAirPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
			doorPositions, width, height) {
			invalidatedCount++
			continue
		}

		// Place mob air
		mobAirLayer[pos.Y][pos.X] = 1
		remaining--
		debug.PlacedCount++

		// Determine placement reason
		reason := fmt.Sprintf("placed via %s strategy", debug.Strategy)
		if ground[pos.Y][pos.X] == 0 {
			reason += " (on void, flying mob)"
		} else {
			reason += " (on ground)"
		}

		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "1x1",
			Reason:   reason,
		})
	}

	// Record miss info
	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated by previous placement (already occupied)",
			Count:  invalidatedCount,
		})
	}
	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all %d valid positions, needed %d more", initialValidCount, remaining),
		})
	}

	return debug
}
