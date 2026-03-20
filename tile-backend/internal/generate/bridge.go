package generate

import (
	"fmt"
	"math/rand"
	"tile-backend/internal/model"
)

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
	bridgeLayerDebug := generateBridgeLayerWithDebug(bridgeLayer, ground, softEdgeLayer, req.Width, req.Height)
	debugInfo.BridgeLayer = bridgeLayerDebug

	// Step 3.6: Generate rail layer if enabled
	railLayer := copyLayer(emptyLayer)
	if req.RailEnabled {
		railDebug := GenerateRailLayer(railLayer, ground, bridgeLayer, req.Width, req.Height)
		debugInfo.Rail = railDebug
	} else {
		debugInfo.Rail = &RailDebugInfo{
			Skipped:    true,
			SkipReason: "railEnabled is false or not specified",
		}
	}

	// Step 4: Generate static layer if requested
	staticLayer := copyLayer(emptyLayer)
	if req.StaticCount > 0 {
		staticDebug := generateStaticLayerWithDebugAndRail(staticLayer, ground, softEdgeLayer, bridgeLayer, railLayer, doorPositions, req.Width, req.Height, req.StaticCount)
		debugInfo.Static = staticDebug
	} else {
		debugInfo.Static = &StaticDebugInfo{
			Skipped:    true,
			SkipReason: "staticCount is 0 or not specified",
		}
	}

	// Apply stage rules
	stageResult, stageErr := ValidateAndApplyStage(req.StageType, "bridge", req.Doors, ground, req.Width, req.Height)
	if stageErr != nil {
		return nil, stageErr
	}
	if stageResult != nil && stageResult.Valid && req.StageType != "" {
		req.ChaserCount = stageResult.ChaserCount
		req.ZonerCount = stageResult.ZonerCount
		req.DPSCount = stageResult.DPSCount
		req.MobAirCount = stageResult.MobAirCount
	}

	// Main path computation
	mainPathData, mainPathDebug := ComputeMainPath(ground, bridgeLayer, doorPositions, req.Width, req.Height)
	debugInfo.MainPath = mainPathDebug

	// Step 5: Generate zoner layer if requested
	zonerLayer := copyLayer(emptyLayer)
	if req.ZonerCount > 0 {
		zonerDebug := GenerateZonerLayer(zonerLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, doorPositions, mainPathData, req.Width, req.Height, req.ZonerCount)
		debugInfo.Zoner = zonerDebug
	} else {
		debugInfo.Zoner = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "zonerCount is 0 or not specified"}
	}

	// Step 6: Generate chaser layer if requested
	chaserLayer := copyLayer(emptyLayer)
	if req.ChaserCount > 0 {
		chaserDebug := GenerateChaserLayer(chaserLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, zonerLayer, doorPositions, mainPathData, req.Width, req.Height, req.ChaserCount)
		debugInfo.Chaser = chaserDebug
	} else {
		debugInfo.Chaser = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "chaserCount is 0 or not specified"}
	}

	// Step 6.5: Generate DPS layer if requested
	dpsLayer := copyLayer(emptyLayer)
	if req.DPSCount > 0 {
		dpsDebug := GenerateDPSLayer(dpsLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, zonerLayer, chaserLayer, doorPositions, mainPathData, req.Width, req.Height, req.DPSCount)
		debugInfo.DPS = dpsDebug
	} else {
		debugInfo.DPS = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "dpsCount is 0 or not specified"}
	}

	// Step 7: Generate mob air layer if requested
	mobAirLayer := copyLayer(emptyLayer)
	if req.MobAirCount > 0 {
		mobAirDebug := GenerateMobAirLayerNew(mobAirLayer, ground, softEdgeLayer, bridgeLayer, staticLayer, zonerLayer, chaserLayer, dpsLayer, doorPositions, req.Width, req.Height, req.MobAirCount)
		debugInfo.MobAir = mobAirDebug
	} else {
		debugInfo.MobAir = &MobAirDebugInfo{Skipped: true, SkipReason: "mobAirCount is 0 or not specified"}
	}

	// Build main path layer for output
	mainPathLayer := copyLayer(emptyLayer)
	if mainPathData != nil {
		for y := 0; y < req.Height; y++ {
			for x := 0; x < req.Width; x++ {
				if mainPathData.OnMainPath[y][x] {
					mainPathLayer[y][x] = 1
				}
			}
		}
	}

	// Build payload
	roomType := "bridge"
	var stageType *string
	if req.StageType != "" {
		stageType = &req.StageType
	}
	payload := model.TemplatePayload{
		Ground:    ground,
		SoftEdge:  softEdgeLayer,
		Bridge:    bridgeLayer,
		Rail:      railLayer,
		Static:    staticLayer,
		Chaser:    chaserLayer,
		Zoner:     zonerLayer,
		DPS:       dpsLayer,
		MobAir:    mobAirLayer,
		MainPath:  mainPathLayer,
		Doors:     doorStates,
		StageType: stageType,
		RoomType:  &roomType,
		Meta: model.TemplateMeta{
			Name:    fmt.Sprintf("bridge-%dx%d", req.Width, req.Height),
			Version: 1,
			Width:   req.Width,
			Height:  req.Height,
		},
	}

	difficulty := ComputeDifficulty(ground, softEdgeLayer, staticLayer, chaserLayer, zonerLayer, dpsLayer, mobAirLayer, mainPathData, req.Width, req.Height)

	return &BridgeGenerateResponse{Payload: payload, DebugInfo: debugInfo, Difficulty: difficulty}, nil
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
