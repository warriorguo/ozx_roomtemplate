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
	StaticCount    int            `json:"staticCount"`    // Suggested number of statics to place (optional)
	TurretCount    int            `json:"turretCount"`    // Suggested number of turrets to place (optional)
	MobGroundCount int            `json:"mobGroundCount"` // Suggested number of mob ground to place (optional)
}

// BridgeGenerateResponse represents the generated template
type BridgeGenerateResponse struct {
	Payload model.TemplatePayload `json:"payload"`
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

	// Initialize empty ground layer
	ground := make([][]int, req.Height)
	for y := 0; y < req.Height; y++ {
		ground[y] = make([]int, req.Width)
	}

	// Get door center positions
	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Step 1: Connect all doors
	connectDoors(ground, doorPositions, req.Width, req.Height)

	// Step 2: Draw small platforms
	drawPlatforms(ground, req.Width, req.Height, req.Doors, doorPositions)

	// Create empty layers for other layers
	emptyLayer := createEmptyLayer(req.Width, req.Height)

	// Build door states
	doorStates := &model.DoorStates{
		Top:    boolToInt(doorSet[DoorTop]),
		Right:  boolToInt(doorSet[DoorRight]),
		Bottom: boolToInt(doorSet[DoorBottom]),
		Left:   boolToInt(doorSet[DoorLeft]),
	}

	// Step 3: Generate static layer if requested
	staticLayer := copyLayer(emptyLayer)
	if req.StaticCount > 0 {
		generateStaticLayer(staticLayer, ground, emptyLayer, emptyLayer, doorPositions, req.Width, req.Height, req.StaticCount)
	}

	// Step 4: Generate turret layer if requested
	turretLayer := copyLayer(emptyLayer)
	if req.TurretCount > 0 {
		generateTurretLayer(turretLayer, ground, emptyLayer, emptyLayer, staticLayer, doorPositions, req.Width, req.Height, req.TurretCount)
	}

	// Step 5: Generate mob ground layer if requested
	mobGroundLayer := copyLayer(emptyLayer)
	if req.MobGroundCount > 0 {
		generateMobGroundLayer(mobGroundLayer, ground, emptyLayer, emptyLayer, staticLayer, turretLayer, doorPositions, req.Width, req.Height, req.MobGroundCount)
	}

	// Build payload
	roomType := "bridge"
	payload := model.TemplatePayload{
		Ground:    ground,
		SoftEdge:  copyLayer(emptyLayer),
		Bridge:    copyLayer(emptyLayer),
		Static:    staticLayer,
		Turret:    turretLayer,
		MobGround: mobGroundLayer,
		MobAir:    copyLayer(emptyLayer),
		Doors:     doorStates,
		RoomType:  &roomType,
		Meta: model.TemplateMeta{
			Name:    fmt.Sprintf("bridge-%dx%d", req.Width, req.Height),
			Version: 1,
			Width:   req.Width,
			Height:  req.Height,
		},
	}

	return &BridgeGenerateResponse{Payload: payload}, nil
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
			connectTwoPoints(ground, from, to, width, height)
			connected[door] = true
		}
	}
}

// connectTwoPoints connects two points with a straight line or L-shaped path through center
func connectTwoPoints(ground [][]int, from, to Point, width, height int) {
	brush := connectionBrushes[rand.Intn(len(connectionBrushes))]

	// Calculate center point
	centerX, centerY := width/2, height/2

	if from.X != to.X && from.Y != to.Y {
		// L-shaped path through center point
		centerPoint := Point{X: centerX, Y: centerY}
		drawLine(ground, from, centerPoint, brush, width, height)
		drawLine(ground, centerPoint, to, brush, width, height)
	} else {
		// Straight line (works for aligned points or random choice)
		drawLine(ground, from, to, brush, width, height)
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
