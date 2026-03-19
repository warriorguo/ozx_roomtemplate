package generate

import (
	"fmt"
	"math/rand"
	"tile-backend/internal/model"
)

// FullRoomGenerateRequest represents the request for generating a full room
type FullRoomGenerateRequest struct {
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

// FullRoomGenerateResponse represents the generated template
type FullRoomGenerateResponse struct {
	Payload   model.TemplatePayload `json:"payload"`
	DebugInfo *FullRoomDebugInfo    `json:"debugInfo,omitempty"`
}

// FullRoomDebugInfo contains debug information about the full room generation process
type FullRoomDebugInfo struct {
	Ground      *FullRoomGroundDebugInfo `json:"ground,omitempty"`
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

// FullRoomGroundDebugInfo contains debug info for full room ground layer generation
type FullRoomGroundDebugInfo struct {
	CornerErase *CornerEraseDebugInfo `json:"cornerErase,omitempty"`
	CenterPits  *CenterPitsDebugInfo  `json:"centerPits,omitempty"`
}

// CornerEraseDebugInfo contains debug info for corner erasing (step 2)
type CornerEraseDebugInfo struct {
	Skipped    bool              `json:"skipped"`
	SkipReason string            `json:"skipReason,omitempty"`
	BrushType  string            `json:"brushType,omitempty"`
	BrushSize  string            `json:"brushSize,omitempty"`
	Combo      string            `json:"combo,omitempty"`
	Corners    []CornerEraseInfo `json:"corners,omitempty"`
}

// CornerEraseInfo describes a single corner erase operation
type CornerEraseInfo struct {
	Corner     string `json:"corner"`
	Position   string `json:"position"`
	Size       string `json:"size"`
	RolledBack bool   `json:"rolledBack,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// CenterPitsDebugInfo contains debug info for center pit operations (step 3)
type CenterPitsDebugInfo struct {
	Skipped    bool            `json:"skipped"`
	SkipReason string          `json:"skipReason,omitempty"`
	BrushSize  string          `json:"brushSize,omitempty"`
	PitCount   int             `json:"pitCount,omitempty"`
	Symmetry   string          `json:"symmetry,omitempty"`
	Pits       []CenterPitInfo `json:"pits,omitempty"`
}

// CenterPitInfo describes a single center pit operation
type CenterPitInfo struct {
	Position   string `json:"position"`
	Size       string `json:"size"`
	RolledBack bool   `json:"rolledBack,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// Corner identifiers for full room generation
type cornerID int

const (
	cornerTopLeft cornerID = iota
	cornerTopRight
	cornerBottomLeft
	cornerBottomRight
)

func cornerIDToString(c cornerID) string {
	switch c {
	case cornerTopLeft:
		return "top-left"
	case cornerTopRight:
		return "top-right"
	case cornerBottomLeft:
		return "bottom-left"
	case cornerBottomRight:
		return "bottom-right"
	default:
		return "unknown"
	}
}

// Corner combination with probability weight
type cornerCombo struct {
	corners []cornerID
	weight  float64 // cumulative probability
	label   string
}

var cornerCombos = []cornerCombo{
	{[]cornerID{cornerTopLeft, cornerBottomLeft, cornerTopRight, cornerBottomRight}, 0.50, "[TL,BL,TR,BR]"},
	{[]cornerID{cornerTopLeft, cornerBottomLeft}, 0.60, "[TL,BL]"},
	{[]cornerID{cornerTopRight, cornerBottomRight}, 0.70, "[TR,BR]"},
	{[]cornerID{cornerBottomLeft, cornerTopRight}, 0.80, "[BL,TR]"},
	{[]cornerID{cornerTopLeft, cornerBottomRight}, 0.90, "[TL,BR]"},
	{[]cornerID{cornerTopLeft}, 0.925, "[TL]"},
	{[]cornerID{cornerBottomLeft}, 0.95, "[BL]"},
	{[]cornerID{cornerTopRight}, 0.975, "[TR]"},
	{[]cornerID{cornerBottomRight}, 1.0, "[BR]"},
}

// GenerateFullRoom generates a full-type room
func GenerateFullRoom(req FullRoomGenerateRequest) (*FullRoomGenerateResponse, error) {
	// Validate input
	if req.Width < 4 || req.Width > 200 {
		return nil, fmt.Errorf("width must be between 4 and 200")
	}
	if req.Height < 4 || req.Height > 200 {
		return nil, fmt.Errorf("height must be between 4 and 200")
	}
	if len(req.Doors) < 2 {
		return nil, fmt.Errorf("at least 2 doors are required")
	}

	// Check for duplicate doors
	doorSet := make(map[DoorPosition]bool)
	for _, door := range req.Doors {
		if doorSet[door] {
			return nil, fmt.Errorf("duplicate door: %s", door)
		}
		doorSet[door] = true
	}

	debugInfo := &FullRoomDebugInfo{}

	// Create layers
	ground := createEmptyLayer(req.Width, req.Height)
	emptyLayer := createEmptyLayer(req.Width, req.Height)

	// Get door center positions
	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Step 1: Fill all ground tiles
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			ground[y][x] = 1
		}
	}

	// Step 2: Corner erase (40% probability)
	groundDebug := &FullRoomGroundDebugInfo{}
	generateFullRoomCornerErase(ground, req.Width, req.Height, req.Doors, groundDebug)

	// Step 3: Center pits (30% probability)
	generateFullRoomCenterPits(ground, req.Width, req.Height, req.Doors, groundDebug)

	debugInfo.Ground = groundDebug

	// Generate other layers using shared functions
	// Soft edge
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

	// Bridge layer
	bridgeLayer := copyLayer(emptyLayer)
	bridgeLayerDebug := generateBridgeLayerWithDebug(bridgeLayer, ground, softEdgeLayer, req.Width, req.Height)
	debugInfo.BridgeLayer = bridgeLayerDebug

	// Rail layer
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

	// Apply stage rules (validate + override counts if stage type specified)
	stageResult, stageErr := ValidateAndApplyStage(req.StageType, "full", req.Doors, ground, req.Width, req.Height)
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

	// Static layer
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

	// Zoner layer (place before chaser — zoner prefers high squishy score positions)
	zonerLayer := copyLayer(emptyLayer)
	if req.ZonerCount > 0 {
		zonerDebug := GenerateZonerLayer(zonerLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, doorPositions, mainPathData, req.Width, req.Height, req.ZonerCount)
		debugInfo.Zoner = zonerDebug
	} else {
		debugInfo.Zoner = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "zonerCount is 0 or not specified"}
	}

	// Chaser layer
	chaserLayer := copyLayer(emptyLayer)
	if req.ChaserCount > 0 {
		chaserDebug := GenerateChaserLayer(chaserLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, zonerLayer, doorPositions, mainPathData, req.Width, req.Height, req.ChaserCount)
		debugInfo.Chaser = chaserDebug
	} else {
		debugInfo.Chaser = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "chaserCount is 0 or not specified"}
	}

	// DPS layer
	dpsLayer := copyLayer(emptyLayer)
	if req.DPSCount > 0 {
		dpsDebug := GenerateDPSLayer(dpsLayer, ground, softEdgeLayer, bridgeLayer, railLayer, staticLayer, zonerLayer, chaserLayer, doorPositions, mainPathData, req.Width, req.Height, req.DPSCount)
		debugInfo.DPS = dpsDebug
	} else {
		debugInfo.DPS = &EnemyLayerDebugInfo{Skipped: true, SkipReason: "dpsCount is 0 or not specified"}
	}

	// Mob air layer
	mobAirLayer := copyLayer(emptyLayer)
	if req.MobAirCount > 0 {
		mobAirDebug := GenerateMobAirLayerNew(mobAirLayer, ground, softEdgeLayer, bridgeLayer, staticLayer, zonerLayer, chaserLayer, dpsLayer, doorPositions, req.Width, req.Height, req.MobAirCount)
		debugInfo.MobAir = mobAirDebug
	} else {
		debugInfo.MobAir = &MobAirDebugInfo{Skipped: true, SkipReason: "mobAirCount is 0 or not specified"}
	}

	// Create door states
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
	roomType := "full"
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
			Name:    fmt.Sprintf("full-%dx%d", req.Width, req.Height),
			Version: 1,
			Width:   req.Width,
			Height:  req.Height,
		},
	}

	return &FullRoomGenerateResponse{
		Payload:   payload,
		DebugInfo: debugInfo,
	}, nil
}

// generateFullRoomCornerErase performs step 2: erase corners with 40% probability
func generateFullRoomCornerErase(ground [][]int, width, height int, doors []DoorPosition, debug *FullRoomGroundDebugInfo) {
	cornerDebug := &CornerEraseDebugInfo{}

	// 40% probability to execute
	if rand.Float64() >= 0.4 {
		cornerDebug.Skipped = true
		cornerDebug.SkipReason = "did not pass 40% probability check"
		debug.CornerErase = cornerDebug
		return
	}

	cornerDebug.Skipped = false

	// Choose brush type (50/50)
	var brushW, brushH int
	if rand.Float64() < 0.5 {
		// Brush 1: 1 <= x <= M/2, 1 <= y <= 2
		cornerDebug.BrushType = "horizontal"
		maxW := width / 2
		if maxW < 1 {
			maxW = 1
		}
		brushW = 1 + rand.Intn(maxW)
		brushH = 1 + rand.Intn(2)
	} else {
		// Brush 2: 1 <= x <= 2, 1 <= y <= N/2
		cornerDebug.BrushType = "vertical"
		maxH := height / 2
		if maxH < 1 {
			maxH = 1
		}
		brushW = 1 + rand.Intn(2)
		brushH = 1 + rand.Intn(maxH)
	}

	cornerDebug.BrushSize = fmt.Sprintf("%dx%d", brushW, brushH)

	// Select corner combination by probability
	r := rand.Float64()
	var selected cornerCombo
	for _, combo := range cornerCombos {
		if r < combo.weight {
			selected = combo
			break
		}
	}
	cornerDebug.Combo = selected.label

	// Apply each corner erase
	for _, corner := range selected.corners {
		x, y := getCornerPosition(corner, width, height, brushW, brushH)

		info := CornerEraseInfo{
			Corner:   cornerIDToString(corner),
			Position: fmt.Sprintf("(%d,%d)", x, y),
			Size:     fmt.Sprintf("%dx%d", brushW, brushH),
		}

		// Save state for rollback
		backup := copyLayer(ground)

		// Erase the corner
		eraseRect(ground, x, y, brushW, brushH, width, height)

		// Check connectivity
		if !areAllDoorsConnected(ground, width, height, doors) {
			// Rollback
			restoreLayer(ground, backup)

			// Retry once: try a smaller brush
			retryW := brushW
			retryH := brushH
			if retryW > 1 {
				retryW = retryW / 2
				if retryW < 1 {
					retryW = 1
				}
			}
			if retryH > 1 {
				retryH = retryH / 2
				if retryH < 1 {
					retryH = 1
				}
			}

			x2, y2 := getCornerPosition(corner, width, height, retryW, retryH)
			eraseRect(ground, x2, y2, retryW, retryH, width, height)

			if !areAllDoorsConnected(ground, width, height, doors) {
				// Rollback again and skip remaining corners
				restoreLayer(ground, backup)
				info.RolledBack = true
				info.Reason = "broke door connectivity after retry, skipping remaining corners"
				cornerDebug.Corners = append(cornerDebug.Corners, info)
				break
			}

			// Retry succeeded with smaller brush
			info.Size = fmt.Sprintf("%dx%d (retried from %dx%d)", retryW, retryH, brushW, brushH)
			info.Position = fmt.Sprintf("(%d,%d)", x2, y2)
		}

		cornerDebug.Corners = append(cornerDebug.Corners, info)
	}

	debug.CornerErase = cornerDebug
}

// generateFullRoomCenterPits performs step 3: center pits with 30% probability
func generateFullRoomCenterPits(ground [][]int, width, height int, doors []DoorPosition, debug *FullRoomGroundDebugInfo) {
	pitsDebug := &CenterPitsDebugInfo{}

	// 30% probability to execute
	if rand.Float64() >= 0.3 {
		pitsDebug.Skipped = true
		pitsDebug.SkipReason = "did not pass 30% probability check"
		debug.CenterPits = pitsDebug
		return
	}

	pitsDebug.Skipped = false

	// Select brush: 1 <= x <= M/3, 1 <= y <= N/2
	maxBrushW := width / 3
	if maxBrushW < 1 {
		maxBrushW = 1
	}
	maxBrushH := height / 2
	if maxBrushH < 1 {
		maxBrushH = 1
	}
	brushW := 1 + rand.Intn(maxBrushW)
	brushH := 1 + rand.Intn(maxBrushH)

	pitsDebug.BrushSize = fmt.Sprintf("%dx%d", brushW, brushH)

	// Select 1~4 pits
	pitCount := 1 + rand.Intn(4)
	pitsDebug.PitCount = pitCount

	// Choose symmetry: left-right or top-bottom
	isLeftRight := rand.Float64() < 0.5
	if isLeftRight {
		pitsDebug.Symmetry = "left-right"
	} else {
		pitsDebug.Symmetry = "top-bottom"
	}

	centerX := width / 2
	centerY := height / 2

	// Generate pit positions symmetrically around center
	type pitPos struct {
		x, y int
	}
	var pitPositions []pitPos

	for i := 0; i < pitCount; i++ {
		if isLeftRight {
			// Left-right symmetric: spread pits vertically, mirror horizontally
			// Random offset from center on the X axis
			offsetX := 1 + rand.Intn(maxBrushW+1)
			// Random Y position spread
			offsetY := rand.Intn(height/3+1) - height/6
			pitY := centerY + offsetY - brushH/2

			// Left pit
			leftX := centerX - offsetX - brushW
			pitPositions = append(pitPositions, pitPos{leftX, pitY})
			// Right pit (mirror)
			rightX := centerX + offsetX
			pitPositions = append(pitPositions, pitPos{rightX, pitY})
		} else {
			// Top-bottom symmetric: spread pits horizontally, mirror vertically
			offsetY := 1 + rand.Intn(maxBrushH/2+1)
			offsetX := rand.Intn(width/3+1) - width/6
			pitX := centerX + offsetX - brushW/2

			// Top pit
			topY := centerY - offsetY - brushH
			pitPositions = append(pitPositions, pitPos{pitX, topY})
			// Bottom pit (mirror)
			bottomY := centerY + offsetY
			pitPositions = append(pitPositions, pitPos{pitX, bottomY})
		}
	}

	// Apply pits, rollback each pair if connectivity breaks
	for i := 0; i < len(pitPositions); i += 2 {
		backup := copyLayer(ground)

		// Apply the pair (or single if odd)
		pit1 := pitPositions[i]
		eraseRect(ground, pit1.x, pit1.y, brushW, brushH, width, height)

		info1 := CenterPitInfo{
			Position: fmt.Sprintf("(%d,%d)", pit1.x, pit1.y),
			Size:     fmt.Sprintf("%dx%d", brushW, brushH),
		}

		var info2 *CenterPitInfo
		if i+1 < len(pitPositions) {
			pit2 := pitPositions[i+1]
			eraseRect(ground, pit2.x, pit2.y, brushW, brushH, width, height)
			info2 = &CenterPitInfo{
				Position: fmt.Sprintf("(%d,%d)", pit2.x, pit2.y),
				Size:     fmt.Sprintf("%dx%d", brushW, brushH),
			}
		}

		// Check connectivity
		if !areAllDoorsConnected(ground, width, height, doors) {
			restoreLayer(ground, backup)
			info1.RolledBack = true
			info1.Reason = "broke door connectivity"
			if info2 != nil {
				info2.RolledBack = true
				info2.Reason = "broke door connectivity (rolled back with pair)"
			}
		}

		pitsDebug.Pits = append(pitsDebug.Pits, info1)
		if info2 != nil {
			pitsDebug.Pits = append(pitsDebug.Pits, *info2)
		}
	}

	debug.CenterPits = pitsDebug
}

// getCornerPosition returns the top-left position for erasing a corner
func getCornerPosition(corner cornerID, width, height, brushW, brushH int) (int, int) {
	switch corner {
	case cornerTopLeft:
		return 0, 0
	case cornerTopRight:
		return width - brushW, 0
	case cornerBottomLeft:
		return 0, height - brushH
	case cornerBottomRight:
		return width - brushW, height - brushH
	default:
		return 0, 0
	}
}

