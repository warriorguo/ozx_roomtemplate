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
	Width  int            `json:"width"`
	Height int            `json:"height"`
	Doors  []DoorPosition `json:"doors"` // At least 2 doors required
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

// Strategy represents a platform placement strategy with weight
type Strategy struct {
	Name   string
	Weight int
	Points []Point
}

var connectionBrushes = []BrushSize{
	{2, 2}, {3, 3}, {4, 4},
}

var platformBrushes = []BrushSize{
	{2, 2}, {2, 3}, {3, 3}, {3, 2},
	{4, 3}, {3, 4}, {4, 4}, {4, 5},
	{5, 4}, {5, 5},
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

	// Build payload
	roomType := "bridge"
	payload := model.TemplatePayload{
		Ground:    ground,
		SoftEdge:  copyLayer(emptyLayer),
		Bridge:    copyLayer(emptyLayer),
		Static:    copyLayer(emptyLayer),
		Turret:    copyLayer(emptyLayer),
		MobGround: copyLayer(emptyLayer),
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

	// Decide whether to use straight or L-shaped path through center
	useLShape := rand.Float32() < 0.5

	if useLShape && from.X != to.X && from.Y != to.Y {
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

		// Draw on all points in the strategy
		brush := platformBrushes[rand.Intn(len(platformBrushes))]
		for _, point := range strategy.Points {
			applyBrush(ground, point.X, point.Y, brush, width, height)
		}

		// Remove selected strategy
		strategies = append(strategies[:selectedIdx], strategies[selectedIdx+1:]...)
	}
}

// buildStrategies builds the platform placement strategies with weights
func buildStrategies(width, height int, doors []DoorPosition, doorPositions map[DoorPosition]Point) []Strategy {
	strategies := []Strategy{}
	centerX, centerY := width/2, height/2

	// Screen center: weight 50
	strategies = append(strategies, Strategy{
		Name:   "center",
		Weight: 50,
		Points: []Point{{X: centerX, Y: centerY}},
	})

	// Check for horizontal (left-right) connection
	_, hasLeft := doorPositions[DoorLeft]
	_, hasRight := doorPositions[DoorRight]
	if hasLeft && hasRight {
		// Left and right door positions: weight 10
		strategies = append(strategies, Strategy{
			Name:   "left_right_doors",
			Weight: 10,
			Points: []Point{doorPositions[DoorLeft], doorPositions[DoorRight]},
		})
		// Midpoints between center and doors: weight 10
		strategies = append(strategies, Strategy{
			Name:   "left_right_midpoints",
			Weight: 10,
			Points: []Point{
				{X: (centerX + doorPositions[DoorLeft].X) / 2, Y: centerY},
				{X: (centerX + doorPositions[DoorRight].X) / 2, Y: centerY},
			},
		})
	}

	// Check for vertical (top-bottom) connection
	_, hasTop := doorPositions[DoorTop]
	_, hasBottom := doorPositions[DoorBottom]
	if hasTop && hasBottom {
		// Top and bottom door positions: weight 10
		strategies = append(strategies, Strategy{
			Name:   "top_bottom_doors",
			Weight: 10,
			Points: []Point{doorPositions[DoorTop], doorPositions[DoorBottom]},
		})
		// Midpoints between center and doors: weight 10
		strategies = append(strategies, Strategy{
			Name:   "top_bottom_midpoints",
			Weight: 10,
			Points: []Point{
				{X: centerX, Y: (centerY + doorPositions[DoorTop].Y) / 2},
				{X: centerX, Y: (centerY + doorPositions[DoorBottom].Y) / 2},
			},
		})
	}

	// All connected doors: weight 10
	allDoorPoints := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		allDoorPoints = append(allDoorPoints, pos)
	}
	if len(allDoorPoints) > 0 {
		strategies = append(strategies, Strategy{
			Name:   "all_doors",
			Weight: 10,
			Points: allDoorPoints,
		})
	}

	// Midpoints between center and all doors: weight 10
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
