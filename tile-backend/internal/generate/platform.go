package generate

import (
	"fmt"
	"math/rand"
	"tile-backend/internal/model"
)

// PlatformGenerateRequest represents the request for generating a platform room
type PlatformGenerateRequest struct {
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
}

// PlatformGenerateResponse represents the generated template
type PlatformGenerateResponse struct {
	Payload    model.TemplatePayload `json:"payload"`
	DebugInfo  *PlatformDebugInfo    `json:"debugInfo,omitempty"`
	Difficulty *DifficultyScore      `json:"difficulty,omitempty"`
}

// PlatformDebugInfo contains debug information about the platform generation process
type PlatformDebugInfo struct {
	Ground      *PlatformGroundDebugInfo `json:"ground,omitempty"`
	SoftEdge    *SoftEdgeDebugInfo       `json:"softEdge,omitempty"`
	BridgeLayer *BridgeLayerDebugInfo    `json:"bridgeLayer,omitempty"`
	Rail        *RailDebugInfo           `json:"rail,omitempty"`
	MainPath    *MainPathDebugInfo       `json:"mainPath,omitempty"`
	Static      *StaticDebugInfo         `json:"static,omitempty"`
	Chaser      *EnemyLayerDebugInfo     `json:"chaser,omitempty"`
	Zoner       *EnemyLayerDebugInfo     `json:"zoner,omitempty"`
	DPS         *EnemyLayerDebugInfo     `json:"dps,omitempty"`
	MobAir      *MobAirDebugInfo         `json:"mobAir,omitempty"`
}

// PlatformGroundDebugInfo contains debug info for platform ground layer generation
type PlatformGroundDebugInfo struct {
	Strategy        string               `json:"strategy"`
	Platforms       []PlatformPlaceInfo  `json:"platforms"`
	DoorConnections []DoorConnectionInfo `json:"doorConnections"`
	EraserOps       []EraserOpInfo       `json:"eraserOps,omitempty"`
}

// PlatformPlaceInfo describes a platform placement
type PlatformPlaceInfo struct {
	Position string `json:"position"`        // Top-left position
	Size     string `json:"size"`            // Size (WxH)
	Group    string `json:"group,omitempty"` // For strategy 2: which corner group
}

// EraserOpInfo describes an eraser operation
type EraserOpInfo struct {
	Method     string `json:"method"`
	Position   string `json:"position"`
	Size       string `json:"size"`
	RolledBack bool   `json:"rolledBack,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// Door grouping for strategy 2
type DoorGroup string

const (
	GroupTopLeft     DoorGroup = "top-left"
	GroupTopRight    DoorGroup = "top-right"
	GroupBottomLeft  DoorGroup = "bottom-left"
	GroupBottomRight DoorGroup = "bottom-right"
)

// GeneratePlatformRoom generates a platform-type room
func GeneratePlatformRoom(req PlatformGenerateRequest) (*PlatformGenerateResponse, error) {
	// Validate input
	if req.Width < 10 || req.Width > 200 {
		return nil, fmt.Errorf("width must be between 10 and 200")
	}
	if req.Height < 10 || req.Height > 200 {
		return nil, fmt.Errorf("height must be between 10 and 200")
	}
	// Check for duplicate doors
	doorSet := make(map[DoorPosition]bool)
	for _, door := range req.Doors {
		if doorSet[door] {
			return nil, fmt.Errorf("duplicate door: %s", door)
		}
		doorSet[door] = true
	}

	debugInfo := &PlatformDebugInfo{}

	// Create empty layers
	ground := createEmptyLayer(req.Width, req.Height)
	emptyLayer := createEmptyLayer(req.Width, req.Height)

	// Get door center positions
	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Step 1: Generate ground layer with platforms
	groundDebug := generatePlatformGround(ground, req.Width, req.Height, req.Doors)
	debugInfo.Ground = groundDebug

	// Step 2: Generate soft edge layer
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

	// Step 3: Generate bridge layer
	bridgeLayer := copyLayer(emptyLayer)
	bridgeLayerDebug := generateBridgeLayerWithDebug(bridgeLayer, ground, softEdgeLayer, req.Width, req.Height)
	debugInfo.BridgeLayer = bridgeLayerDebug

	// Step 3.5: Generate rail layer
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

	// Step 4: Generate static layer
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
	stageResult, stageErr := ValidateAndApplyStage(req.StageType, "platform", req.Doors, ground, req.Width, req.Height)
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

	// Step 5: Generate zoner layer
	zonerLayer := copyLayer(emptyLayer)
	if req.ZonerCount > 0 {
		zonerDebug := GenerateZonerLayer(zonerLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, doorPositions, mainPathData, req.Width, req.Height, req.ZonerCount)
		debugInfo.Zoner = zonerDebug
	} else {
		debugInfo.Zoner = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "zonerCount is 0 or not specified"}
	}

	// Step 6: Generate chaser layer
	chaserLayer := copyLayer(emptyLayer)
	if req.ChaserCount > 0 {
		chaserDebug := GenerateChaserLayer(chaserLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, zonerLayer, doorPositions, mainPathData, req.Width, req.Height, req.ChaserCount)
		debugInfo.Chaser = chaserDebug
	} else {
		debugInfo.Chaser = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "chaserCount is 0 or not specified"}
	}

	// Step 6.5: Generate DPS layer
	dpsLayer := copyLayer(emptyLayer)
	if req.DPSCount > 0 {
		dpsDebug := GenerateDPSLayer(dpsLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, zonerLayer, chaserLayer, doorPositions, mainPathData, req.Width, req.Height, req.DPSCount)
		debugInfo.DPS = dpsDebug
	} else {
		debugInfo.DPS = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "dpsCount is 0 or not specified"}
	}

	// Step 7: Generate mob air layer
	mobAirLayer := copyLayer(emptyLayer)
	if req.MobAirCount > 0 {
		mobAirDebug := GenerateMobAirLayerNew(mobAirLayer, ground, softEdgeLayer, bridgeLayer, staticLayer, zonerLayer, chaserLayer, dpsLayer, doorPositions, req.Width, req.Height, req.MobAirCount)
		debugInfo.MobAir = mobAirDebug
	} else {
		debugInfo.MobAir = &MobAirDebugInfo{Skipped: true, SkipReason: "mobAirCount is 0 or not specified"}
	}

	// Create doors configuration
	doorStates := &model.DoorStates{
		Top:    0,
		Right:  0,
		Bottom: 0,
		Left:   0,
	}
	for _, door := range req.Doors {
		switch door {
		case DoorTop:
			doorStates.Top = 1
		case DoorRight:
			doorStates.Right = 1
		case DoorBottom:
			doorStates.Bottom = 1
		case DoorLeft:
			doorStates.Left = 1
		}
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

	// Create payload
	roomType := "platform"
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
			Name:    fmt.Sprintf("platform-%dx%d", req.Width, req.Height),
			Version: 1,
			Width:   req.Width,
			Height:  req.Height,
		},
	}

	difficulty := ComputeDifficulty(ground, softEdgeLayer, staticLayer, chaserLayer, zonerLayer, dpsLayer, mobAirLayer, mainPathData, req.Width, req.Height)

	return &PlatformGenerateResponse{
		Payload:    payload,
		DebugInfo:  debugInfo,
		Difficulty: difficulty,
	}, nil
}

// generatePlatformGround generates the ground layer for platform rooms
func generatePlatformGround(ground [][]int, width, height int, doors []DoorPosition) *PlatformGroundDebugInfo {
	debug := &PlatformGroundDebugInfo{}

	// Check if strategy 2 is possible (doors can be grouped into corner pairs)
	canUseStrategy2 := canGroupDoorsIntoCorners(doors)

	// Choose strategy
	var useStrategy2 bool
	if canUseStrategy2 {
		useStrategy2 = rand.Float64() < 0.5 // 50% chance to use strategy 2 if possible
	}

	if useStrategy2 {
		debug.Strategy = "strategy2_corner_groups"
		generatePlatformStrategy2(ground, width, height, doors, debug)
	} else {
		debug.Strategy = "strategy1_center_platform"
		generatePlatformStrategy1(ground, width, height, doors, debug)
	}

	// Apply eraser operations
	applyEraserOperations(ground, width, height, doors, useStrategy2, debug)

	return debug
}

// canGroupDoorsIntoCorners checks if doors can be grouped into corner pairs
// Valid groups: top-left (top+left), top-right (top+right), bottom-left (bottom+left), bottom-right (bottom+right)
func canGroupDoorsIntoCorners(doors []DoorPosition) bool {
	doorSet := make(map[DoorPosition]bool)
	for _, door := range doors {
		doorSet[door] = true
	}

	// Check for at least one valid corner pair
	hasTopLeft := doorSet[DoorTop] && doorSet[DoorLeft]
	hasTopRight := doorSet[DoorTop] && doorSet[DoorRight]
	hasBottomLeft := doorSet[DoorBottom] && doorSet[DoorLeft]
	hasBottomRight := doorSet[DoorBottom] && doorSet[DoorRight]

	return hasTopLeft || hasTopRight || hasBottomLeft || hasBottomRight
}

// generatePlatformStrategy1 generates a large center platform and connects all doors
func generatePlatformStrategy1(ground [][]int, width, height int, doors []DoorPosition, debug *PlatformGroundDebugInfo) {
	// Generate large center platform: L > width/2, W > height/2
	minPlatformW := width/2 + 1
	minPlatformH := height/2 + 1
	maxPlatformW := width - 4 // Leave some margin
	maxPlatformH := height - 4

	if maxPlatformW < minPlatformW {
		maxPlatformW = minPlatformW
	}
	if maxPlatformH < minPlatformH {
		maxPlatformH = minPlatformH
	}

	platformW := minPlatformW + rand.Intn(maxPlatformW-minPlatformW+1)
	platformH := minPlatformH + rand.Intn(maxPlatformH-minPlatformH+1)

	// Center the platform
	platformX := (width - platformW) / 2
	platformY := (height - platformH) / 2

	// Draw the platform
	for y := platformY; y < platformY+platformH && y < height; y++ {
		for x := platformX; x < platformX+platformW && x < width; x++ {
			ground[y][x] = 1
		}
	}

	debug.Platforms = append(debug.Platforms, PlatformPlaceInfo{
		Position: fmt.Sprintf("(%d,%d)", platformX, platformY),
		Size:     fmt.Sprintf("%dx%d", platformW, platformH),
	})

	// Connect doors using 2x2, 3x3, or 4x4 brush
	brushSizes := []int{2, 3, 4}
	brushSize := brushSizes[rand.Intn(len(brushSizes))]

	centerX := width / 2
	centerY := height / 2

	for _, door := range doors {
		doorX, doorY := getDoorPosition(door, width, height)

		// Choose path type: direct or via center
		viaCenterProb := 0.5
		viaCenter := rand.Float64() < viaCenterProb

		var pathType string
		if viaCenter {
			pathType = "via center"
			// Draw from door to center
			drawPath(ground, doorX, doorY, centerX, centerY, brushSize)
		} else {
			pathType = "direct to platform"
			// Draw from door to nearest platform edge
			targetX, targetY := getNearestPlatformEdge(doorX, doorY, platformX, platformY, platformW, platformH)
			drawPath(ground, doorX, doorY, targetX, targetY, brushSize)
		}

		debug.DoorConnections = append(debug.DoorConnections, DoorConnectionInfo{
			From:      fmt.Sprintf("%s (%d,%d)", door, doorX, doorY),
			To:        "platform",
			PathType:  pathType,
			BrushSize: fmt.Sprintf("%dx%d", brushSize, brushSize),
		})
	}
}

// generatePlatformStrategy2 generates platforms for corner groups
func generatePlatformStrategy2(ground [][]int, width, height int, doors []DoorPosition, debug *PlatformGroundDebugInfo) {
	doorSet := make(map[DoorPosition]bool)
	for _, door := range doors {
		doorSet[door] = true
	}

	// Find valid corner groups
	type cornerGroup struct {
		group            DoorGroup
		doors            []DoorPosition
		anchorX, anchorY int // Corner anchor point
	}

	var groups []cornerGroup

	if doorSet[DoorTop] && doorSet[DoorLeft] {
		groups = append(groups, cornerGroup{GroupTopLeft, []DoorPosition{DoorTop, DoorLeft}, 0, 0})
	}
	if doorSet[DoorTop] && doorSet[DoorRight] {
		groups = append(groups, cornerGroup{GroupTopRight, []DoorPosition{DoorTop, DoorRight}, width, 0})
	}
	if doorSet[DoorBottom] && doorSet[DoorLeft] {
		groups = append(groups, cornerGroup{GroupBottomLeft, []DoorPosition{DoorBottom, DoorLeft}, 0, height})
	}
	if doorSet[DoorBottom] && doorSet[DoorRight] {
		groups = append(groups, cornerGroup{GroupBottomRight, []DoorPosition{DoorBottom, DoorRight}, width, height})
	}

	// Generate platform for each group
	for _, group := range groups {
		// Platform size: L > width/2, W > height/2
		minPlatformW := width/2 + 1
		minPlatformH := height/2 + 1
		maxPlatformW := width * 4 / 5
		maxPlatformH := height * 4 / 5

		platformW := minPlatformW + rand.Intn(maxPlatformW-minPlatformW+1)
		platformH := minPlatformH + rand.Intn(maxPlatformH-minPlatformH+1)

		// Position platform based on corner group
		var platformX, platformY int
		switch group.group {
		case GroupTopLeft:
			platformX = 0
			platformY = 0
		case GroupTopRight:
			platformX = width - platformW
			platformY = 0
		case GroupBottomLeft:
			platformX = 0
			platformY = height - platformH
		case GroupBottomRight:
			platformX = width - platformW
			platformY = height - platformH
		}

		// Clamp to bounds
		if platformX < 0 {
			platformX = 0
		}
		if platformY < 0 {
			platformY = 0
		}

		// Draw the platform
		for y := platformY; y < platformY+platformH && y < height; y++ {
			for x := platformX; x < platformX+platformW && x < width; x++ {
				ground[y][x] = 1
			}
		}

		debug.Platforms = append(debug.Platforms, PlatformPlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", platformX, platformY),
			Size:     fmt.Sprintf("%dx%d", platformW, platformH),
			Group:    string(group.group),
		})

		// Connect the two doors in this group
		brushSizes := []int{2, 3, 4}
		brushSize := brushSizes[rand.Intn(len(brushSizes))]

		for _, door := range group.doors {
			doorX, doorY := getDoorPosition(door, width, height)

			// Draw path from door to platform
			targetX := platformX + platformW/2
			targetY := platformY + platformH/2
			drawPath(ground, doorX, doorY, targetX, targetY, brushSize)

			debug.DoorConnections = append(debug.DoorConnections, DoorConnectionInfo{
				From:      fmt.Sprintf("%s (%d,%d)", door, doorX, doorY),
				To:        fmt.Sprintf("platform %s", group.group),
				PathType:  "direct",
				BrushSize: fmt.Sprintf("%dx%d", brushSize, brushSize),
			})
		}
	}
}

// getNearestPlatformEdge returns the nearest edge point on the platform from the given position
func getNearestPlatformEdge(fromX, fromY, platformX, platformY, platformW, platformH int) (int, int) {
	// Clamp the point to platform bounds
	targetX := fromX
	targetY := fromY

	if targetX < platformX {
		targetX = platformX
	} else if targetX >= platformX+platformW {
		targetX = platformX + platformW - 1
	}

	if targetY < platformY {
		targetY = platformY
	} else if targetY >= platformY+platformH {
		targetY = platformY + platformH - 1
	}

	return targetX, targetY
}

// Eraser methods
type eraserMethod int

const (
	eraserCenterSingle eraserMethod = iota
	eraserCenterSymmetric2
	eraserCenterSymmetric3
	eraserCorners
	eraserUnconnectedDoorDirection
	eraserStrategy2Corner
)

// applyEraserOperations applies eraser operations to create void areas
func applyEraserOperations(ground [][]int, width, height int, doors []DoorPosition, isStrategy2 bool, debug *PlatformGroundDebugInfo) {
	// Randomly select 0-3 erase operations
	eraseCount := rand.Intn(4)

	if eraseCount == 0 {
		return
	}

	// Build available eraser methods based on strategy
	availableMethods := []eraserMethod{
		eraserCenterSingle,
		eraserCenterSymmetric2,
		eraserCenterSymmetric3,
	}

	if !isStrategy2 {
		availableMethods = append(availableMethods, eraserCorners)

		// Check for unconnected doors
		allDoors := []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}
		doorSet := make(map[DoorPosition]bool)
		for _, d := range doors {
			doorSet[d] = true
		}
		for _, d := range allDoors {
			if !doorSet[d] {
				availableMethods = append(availableMethods, eraserUnconnectedDoorDirection)
				break
			}
		}
	} else {
		availableMethods = append(availableMethods, eraserStrategy2Corner)
	}

	usedMethods := make(map[eraserMethod]bool)

	for i := 0; i < eraseCount; i++ {
		// Filter out already used methods
		var remaining []eraserMethod
		for _, m := range availableMethods {
			if !usedMethods[m] {
				remaining = append(remaining, m)
			}
		}

		if len(remaining) == 0 {
			break
		}

		// Select random method
		method := remaining[rand.Intn(len(remaining))]
		usedMethods[method] = true

		// Save ground state for potential rollback
		groundBackup := copyLayer(ground)

		// Apply eraser method
		opInfo := applyEraserMethod(ground, width, height, doors, isStrategy2, method, debug)

		// Check connectivity: multi-door connectivity AND no isolated ground islands
		disconnected := !areAllDoorsConnected(ground, width, height, doors) ||
			len(findAllIslands(ground, width, height)) > 1
		if disconnected {
			// Rollback
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					ground[y][x] = groundBackup[y][x]
				}
			}
			opInfo.RolledBack = true
			opInfo.Reason = "would break door connectivity or create isolated ground islands"
		}

		debug.EraserOps = append(debug.EraserOps, opInfo)
	}
}

// applyEraserMethod applies a specific eraser method
func applyEraserMethod(ground [][]int, width, height int, doors []DoorPosition, isStrategy2 bool, method eraserMethod, debug *PlatformGroundDebugInfo) EraserOpInfo {
	centerX := width / 2
	centerY := height / 2

	switch method {
	case eraserCenterSingle:
		// Erase single area in center: 2x2, 3x3, 3x4, 4x4, 4x5
		sizes := []struct{ w, h int }{
			{2, 2}, {3, 3}, {3, 4}, {4, 4}, {4, 5},
		}
		size := sizes[rand.Intn(len(sizes))]
		x := centerX - size.w/2
		y := centerY - size.h/2
		eraseRect(ground, x, y, size.w, size.h, width, height)
		return EraserOpInfo{
			Method:   "center_single",
			Position: fmt.Sprintf("(%d,%d)", x, y),
			Size:     fmt.Sprintf("%dx%d", size.w, size.h),
		}

	case eraserCenterSymmetric2:
		// Erase two symmetric areas: 2x2, 3x3, 3x4
		sizes := []struct{ w, h int }{
			{2, 2}, {3, 3}, {3, 4},
		}
		size := sizes[rand.Intn(len(sizes))]

		// Random symmetry: left-right or top-bottom
		isLeftRight := rand.Float64() < 0.5

		var positions string
		if isLeftRight {
			offset := width/4 + rand.Intn(width/4)
			x1 := centerX - offset - size.w/2
			x2 := centerX + offset - size.w/2
			y := centerY - size.h/2
			eraseRect(ground, x1, y, size.w, size.h, width, height)
			eraseRect(ground, x2, y, size.w, size.h, width, height)
			positions = fmt.Sprintf("(%d,%d) and (%d,%d)", x1, y, x2, y)
		} else {
			offset := height/4 + rand.Intn(height/4)
			x := centerX - size.w/2
			y1 := centerY - offset - size.h/2
			y2 := centerY + offset - size.h/2
			eraseRect(ground, x, y1, size.w, size.h, width, height)
			eraseRect(ground, x, y2, size.w, size.h, width, height)
			positions = fmt.Sprintf("(%d,%d) and (%d,%d)", x, y1, x, y2)
		}
		return EraserOpInfo{
			Method:   "center_symmetric_2",
			Position: positions,
			Size:     fmt.Sprintf("%dx%d", size.w, size.h),
		}

	case eraserCenterSymmetric3:
		// Erase three symmetric areas
		sizes := []struct{ w, h int }{
			{2, 2}, {3, 3}, {3, 4},
		}
		size := sizes[rand.Intn(len(sizes))]

		// One in center, two symmetric
		x := centerX - size.w/2
		y := centerY - size.h/2
		eraseRect(ground, x, y, size.w, size.h, width, height)

		isLeftRight := rand.Float64() < 0.5
		var positions string
		if isLeftRight {
			offset := width/3 + rand.Intn(width/6)
			x1 := centerX - offset - size.w/2
			x2 := centerX + offset - size.w/2
			eraseRect(ground, x1, y, size.w, size.h, width, height)
			eraseRect(ground, x2, y, size.w, size.h, width, height)
			positions = fmt.Sprintf("(%d,%d), (%d,%d), (%d,%d)", x, y, x1, y, x2, y)
		} else {
			offset := height/3 + rand.Intn(height/6)
			y1 := centerY - offset - size.h/2
			y2 := centerY + offset - size.h/2
			eraseRect(ground, x, y1, size.w, size.h, width, height)
			eraseRect(ground, x, y2, size.w, size.h, width, height)
			positions = fmt.Sprintf("(%d,%d), (%d,%d), (%d,%d)", x, y, x, y1, x, y2)
		}
		return EraserOpInfo{
			Method:   "center_symmetric_3",
			Position: positions,
			Size:     fmt.Sprintf("%dx%d", size.w, size.h),
		}

	case eraserCorners:
		// Erase platform corners: 5% for 1, 20% for 2, 5% for 3, 70% for 4
		r := rand.Float64()
		var cornerCount int
		if r < 0.05 {
			cornerCount = 1
		} else if r < 0.25 {
			cornerCount = 2
		} else if r < 0.30 {
			cornerCount = 3
		} else {
			cornerCount = 4
		}

		sizes := []struct{ w, h int }{
			{2, 2}, {3, 3}, {3, 4},
		}
		size := sizes[rand.Intn(len(sizes))]

		corners := []struct{ x, y int }{
			{0, 0},                            // top-left
			{width - size.w, 0},               // top-right
			{0, height - size.h},              // bottom-left
			{width - size.w, height - size.h}, // bottom-right
		}

		// Shuffle corners
		rand.Shuffle(len(corners), func(i, j int) {
			corners[i], corners[j] = corners[j], corners[i]
		})

		var positions []string
		for i := 0; i < cornerCount && i < len(corners); i++ {
			eraseRect(ground, corners[i].x, corners[i].y, size.w, size.h, width, height)
			positions = append(positions, fmt.Sprintf("(%d,%d)", corners[i].x, corners[i].y))
		}

		return EraserOpInfo{
			Method:   fmt.Sprintf("corners_%d", cornerCount),
			Position: fmt.Sprintf("%v", positions),
			Size:     fmt.Sprintf("%dx%d", size.w, size.h),
		}

	case eraserUnconnectedDoorDirection:
		// Find unconnected door and erase in that direction
		allDoors := []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}
		doorSet := make(map[DoorPosition]bool)
		for _, d := range doors {
			doorSet[d] = true
		}

		var unconnected []DoorPosition
		for _, d := range allDoors {
			if !doorSet[d] {
				unconnected = append(unconnected, d)
			}
		}

		if len(unconnected) == 0 {
			return EraserOpInfo{Method: "unconnected_door_direction", Reason: "no unconnected doors"}
		}

		door := unconnected[rand.Intn(len(unconnected))]
		sizes := []struct{ w, h int }{
			{3, 3}, {3, 4}, {4, 4},
		}
		size := sizes[rand.Intn(len(sizes))]

		// Position based on door direction (inside platform)
		var x, y int
		switch door {
		case DoorTop:
			x = centerX - size.w/2
			y = centerY - height/4 - size.h/2
		case DoorBottom:
			x = centerX - size.w/2
			y = centerY + height/4 - size.h/2
		case DoorLeft:
			x = centerX - width/4 - size.w/2
			y = centerY - size.h/2
		case DoorRight:
			x = centerX + width/4 - size.w/2
			y = centerY - size.h/2
		}

		eraseRect(ground, x, y, size.w, size.h, width, height)
		return EraserOpInfo{
			Method:   "unconnected_door_direction",
			Position: fmt.Sprintf("(%d,%d) towards %s", x, y, door),
			Size:     fmt.Sprintf("%dx%d", size.w, size.h),
		}

	case eraserStrategy2Corner:
		// Erase a corner of a random platform (strategy 2 only)
		sizes := []struct{ w, h int }{
			{2, 2}, {3, 3},
		}
		size := sizes[rand.Intn(len(sizes))]

		// Pick a random corner
		corners := []struct{ x, y int }{
			{rand.Intn(width / 3), rand.Intn(height / 3)},                                // top-left area
			{width - rand.Intn(width/3) - size.w, rand.Intn(height / 3)},                 // top-right area
			{rand.Intn(width / 3), height - rand.Intn(height/3) - size.h},                // bottom-left area
			{width - rand.Intn(width/3) - size.w, height - rand.Intn(height/3) - size.h}, // bottom-right area
		}
		corner := corners[rand.Intn(len(corners))]

		eraseRect(ground, corner.x, corner.y, size.w, size.h, width, height)
		return EraserOpInfo{
			Method:   "strategy2_corner",
			Position: fmt.Sprintf("(%d,%d)", corner.x, corner.y),
			Size:     fmt.Sprintf("%dx%d", size.w, size.h),
		}
	}

	return EraserOpInfo{}
}

// eraseRect erases a rectangle (sets ground to 0)
func eraseRect(ground [][]int, x, y, w, h, width, height int) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < width && py >= 0 && py < height {
				ground[py][px] = 0
			}
		}
	}
}

// drawPath draws a path between two points using the given brush size
func drawPath(ground [][]int, fromX, fromY, toX, toY, brushSize int) {
	height := len(ground)
	width := len(ground[0])

	// Draw L-shaped path
	// First horizontal, then vertical
	if rand.Float64() < 0.5 {
		// Horizontal first
		drawHorizontalLine(ground, fromX, toX, fromY, brushSize, width, height)
		drawVerticalLine(ground, fromY, toY, toX, brushSize, width, height)
	} else {
		// Vertical first
		drawVerticalLine(ground, fromY, toY, fromX, brushSize, width, height)
		drawHorizontalLine(ground, fromX, toX, toY, brushSize, width, height)
	}
}

// drawHorizontalLine draws a horizontal line with the given brush size
func drawHorizontalLine(ground [][]int, fromX, toX, y, brushSize, width, height int) {
	if fromX > toX {
		fromX, toX = toX, fromX
	}

	halfBrush := brushSize / 2
	for x := fromX; x <= toX; x++ {
		for dy := -halfBrush; dy < brushSize-halfBrush; dy++ {
			for dx := -halfBrush; dx < brushSize-halfBrush; dx++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					ground[py][px] = 1
				}
			}
		}
	}
}

// drawVerticalLine draws a vertical line with the given brush size
func drawVerticalLine(ground [][]int, fromY, toY, x, brushSize, width, height int) {
	if fromY > toY {
		fromY, toY = toY, fromY
	}

	halfBrush := brushSize / 2
	for y := fromY; y <= toY; y++ {
		for dy := -halfBrush; dy < brushSize-halfBrush; dy++ {
			for dx := -halfBrush; dx < brushSize-halfBrush; dx++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					ground[py][px] = 1
				}
			}
		}
	}
}

// getDoorPosition returns the center position for a door
func getDoorPosition(door DoorPosition, width, height int) (int, int) {
	switch door {
	case DoorTop:
		return width / 2, 0
	case DoorBottom:
		return width / 2, height - 1
	case DoorLeft:
		return 0, height / 2
	case DoorRight:
		return width - 1, height / 2
	default:
		return width / 2, height / 2
	}
}
