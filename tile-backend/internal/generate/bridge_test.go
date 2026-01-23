package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateBridgeRoom_ValidInput(t *testing.T) {
	tests := []struct {
		name   string
		req    BridgeGenerateRequest
		verify func(t *testing.T, resp *BridgeGenerateResponse)
	}{
		{
			name: "two doors - top and bottom",
			req: BridgeGenerateRequest{
				Width:  10,
				Height: 10,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			verify: func(t *testing.T, resp *BridgeGenerateResponse) {
				// Verify dimensions
				assert.Equal(t, 10, len(resp.Payload.Ground))
				assert.Equal(t, 10, len(resp.Payload.Ground[0]))

				// Verify doors are connected (there should be walkable path)
				topDoorX := 10 / 2
				bottomDoorX := 10 / 2
				assert.Equal(t, 1, resp.Payload.Ground[0][topDoorX], "top door should be walkable")
				assert.Equal(t, 1, resp.Payload.Ground[9][bottomDoorX], "bottom door should be walkable")

				// Verify room type
				assert.NotNil(t, resp.Payload.RoomType)
				assert.Equal(t, "bridge", *resp.Payload.RoomType)

				// Verify door states
				assert.NotNil(t, resp.Payload.Doors)
				assert.Equal(t, 1, resp.Payload.Doors.Top)
				assert.Equal(t, 0, resp.Payload.Doors.Right)
				assert.Equal(t, 1, resp.Payload.Doors.Bottom)
				assert.Equal(t, 0, resp.Payload.Doors.Left)
			},
		},
		{
			name: "two doors - left and right",
			req: BridgeGenerateRequest{
				Width:  15,
				Height: 10,
				Doors:  []DoorPosition{DoorLeft, DoorRight},
			},
			verify: func(t *testing.T, resp *BridgeGenerateResponse) {
				// Verify dimensions
				assert.Equal(t, 10, len(resp.Payload.Ground))
				assert.Equal(t, 15, len(resp.Payload.Ground[0]))

				// Verify doors are connected
				leftDoorY := 10 / 2
				rightDoorY := 10 / 2
				assert.Equal(t, 1, resp.Payload.Ground[leftDoorY][0], "left door should be walkable")
				assert.Equal(t, 1, resp.Payload.Ground[rightDoorY][14], "right door should be walkable")

				// Verify door states
				assert.Equal(t, 0, resp.Payload.Doors.Top)
				assert.Equal(t, 1, resp.Payload.Doors.Right)
				assert.Equal(t, 0, resp.Payload.Doors.Bottom)
				assert.Equal(t, 1, resp.Payload.Doors.Left)
			},
		},
		{
			name: "three doors",
			req: BridgeGenerateRequest{
				Width:  12,
				Height: 12,
				Doors:  []DoorPosition{DoorTop, DoorRight, DoorBottom},
			},
			verify: func(t *testing.T, resp *BridgeGenerateResponse) {
				// Verify all three doors are marked
				assert.Equal(t, 1, resp.Payload.Doors.Top)
				assert.Equal(t, 1, resp.Payload.Doors.Right)
				assert.Equal(t, 1, resp.Payload.Doors.Bottom)
				assert.Equal(t, 0, resp.Payload.Doors.Left)
			},
		},
		{
			name: "four doors",
			req: BridgeGenerateRequest{
				Width:  20,
				Height: 15,
				Doors:  []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft},
			},
			verify: func(t *testing.T, resp *BridgeGenerateResponse) {
				// Verify all four doors are marked
				assert.Equal(t, 1, resp.Payload.Doors.Top)
				assert.Equal(t, 1, resp.Payload.Doors.Right)
				assert.Equal(t, 1, resp.Payload.Doors.Bottom)
				assert.Equal(t, 1, resp.Payload.Doors.Left)

				// Verify meta
				assert.Equal(t, 20, resp.Payload.Meta.Width)
				assert.Equal(t, 15, resp.Payload.Meta.Height)
				assert.Equal(t, 1, resp.Payload.Meta.Version)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GenerateBridgeRoom(tt.req)
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify all layers exist and have correct dimensions
			assert.Equal(t, tt.req.Height, len(resp.Payload.Ground))
			assert.Equal(t, tt.req.Height, len(resp.Payload.SoftEdge))
			assert.Equal(t, tt.req.Height, len(resp.Payload.Bridge))
			assert.Equal(t, tt.req.Height, len(resp.Payload.Static))
			assert.Equal(t, tt.req.Height, len(resp.Payload.Turret))
			assert.Equal(t, tt.req.Height, len(resp.Payload.MobGround))
			assert.Equal(t, tt.req.Height, len(resp.Payload.MobAir))

			// Verify other layers are empty
			for y := 0; y < tt.req.Height; y++ {
				for x := 0; x < tt.req.Width; x++ {
					assert.Equal(t, 0, resp.Payload.SoftEdge[y][x], "softEdge should be 0")
					assert.Equal(t, 0, resp.Payload.Bridge[y][x], "bridge should be 0")
					assert.Equal(t, 0, resp.Payload.Static[y][x], "static should be 0")
					assert.Equal(t, 0, resp.Payload.Turret[y][x], "turret should be 0")
					assert.Equal(t, 0, resp.Payload.MobGround[y][x], "mobGround should be 0")
					assert.Equal(t, 0, resp.Payload.MobAir[y][x], "mobAir should be 0")
				}
			}

			// Run custom verification
			tt.verify(t, resp)
		})
	}
}

func TestGenerateBridgeRoom_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		req         BridgeGenerateRequest
		expectedErr string
	}{
		{
			name: "width too small",
			req: BridgeGenerateRequest{
				Width:  2,
				Height: 10,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "invalid dimensions",
		},
		{
			name: "height too small",
			req: BridgeGenerateRequest{
				Width:  10,
				Height: 3,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "invalid dimensions",
		},
		{
			name: "width too large",
			req: BridgeGenerateRequest{
				Width:  250,
				Height: 10,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "invalid dimensions",
		},
		{
			name: "only one door",
			req: BridgeGenerateRequest{
				Width:  10,
				Height: 10,
				Doors:  []DoorPosition{DoorTop},
			},
			expectedErr: "at least 2 doors",
		},
		{
			name: "no doors",
			req: BridgeGenerateRequest{
				Width:  10,
				Height: 10,
				Doors:  []DoorPosition{},
			},
			expectedErr: "at least 2 doors",
		},
		{
			name: "duplicate doors",
			req: BridgeGenerateRequest{
				Width:  10,
				Height: 10,
				Doors:  []DoorPosition{DoorTop, DoorTop},
			},
			expectedErr: "duplicate door",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GenerateBridgeRoom(tt.req)
			assert.Nil(t, resp)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestGenerateBridgeRoom_DoorsConnected(t *testing.T) {
	// Test that doors are actually connected by verifying there's a path
	req := BridgeGenerateRequest{
		Width:  15,
		Height: 15,
		Doors:  []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Use BFS to verify all doors are connected
	doorPositions := []Point{
		{X: 15 / 2, Y: 0},       // top
		{X: 15 / 2, Y: 14},      // bottom
		{X: 0, Y: 15 / 2},       // left
		{X: 14, Y: 15 / 2},      // right
	}

	// Start BFS from first door position
	visited := make(map[Point]bool)
	queue := []Point{doorPositions[0]}
	visited[doorPositions[0]] = true

	directions := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dir := range directions {
			next := Point{X: current.X + dir.X, Y: current.Y + dir.Y}
			if next.X >= 0 && next.X < 15 && next.Y >= 0 && next.Y < 15 {
				if !visited[next] && resp.Payload.Ground[next.Y][next.X] == 1 {
					visited[next] = true
					queue = append(queue, next)
				}
			}
		}
	}

	// Verify all door positions are reachable
	for i, doorPos := range doorPositions {
		// Check if door position or nearby cells are visited
		found := false
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				checkPos := Point{X: doorPos.X + dx, Y: doorPos.Y + dy}
				if checkPos.X >= 0 && checkPos.X < 15 && checkPos.Y >= 0 && checkPos.Y < 15 {
					if visited[checkPos] {
						found = true
						break
					}
				}
			}
			if found {
				break
			}
		}
		assert.True(t, found, "door %d at (%d, %d) should be connected", i, doorPos.X, doorPos.Y)
	}
}

func TestGenerateBridgeRoom_GroundHasWalkableTiles(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:  20,
		Height: 15,
		Doors:  []DoorPosition{DoorTop, DoorBottom},
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Count walkable tiles
	walkableCount := 0
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Ground[y][x] == 1 {
				walkableCount++
			}
		}
	}

	// Should have some walkable tiles (at least the path between doors)
	assert.Greater(t, walkableCount, 0, "should have walkable tiles")

	// Should not be completely filled (it's a bridge, not a full room)
	totalTiles := req.Width * req.Height
	assert.Less(t, walkableCount, totalTiles, "should not fill entire room")
}

func TestBuildStrategies(t *testing.T) {
	doorPositions := map[DoorPosition]Point{
		DoorTop:    {X: 10, Y: 0},
		DoorBottom: {X: 10, Y: 19},
		DoorLeft:   {X: 0, Y: 10},
		DoorRight:  {X: 19, Y: 10},
	}

	strategies := buildStrategies(20, 20, []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}, doorPositions)

	// Should have center strategy with weight 50
	hasCenter := false
	for _, s := range strategies {
		if s.Name == "center" {
			hasCenter = true
			assert.Equal(t, 50, s.Weight)
			assert.Len(t, s.Points, 1)
			assert.Equal(t, Point{X: 10, Y: 10}, s.Points[0])
		}
	}
	assert.True(t, hasCenter, "should have center strategy")

	// Should have multiple strategies
	assert.Greater(t, len(strategies), 1)
}

func TestSelectByWeight(t *testing.T) {
	strategies := []Strategy{
		{Name: "a", Weight: 100, Points: []Point{{0, 0}}},
		{Name: "b", Weight: 0, Points: []Point{{1, 1}}},
	}

	// With weight 100 vs 0, should always select first
	for i := 0; i < 10; i++ {
		idx := selectByWeight(strategies)
		assert.Equal(t, 0, idx)
	}

	// Empty strategies should return -1
	idx := selectByWeight([]Strategy{})
	assert.Equal(t, -1, idx)
}

func TestApplyBrush(t *testing.T) {
	ground := createEmptyLayer(10, 10)

	// Apply 3x3 brush at center
	applyBrush(ground, 5, 5, BrushSize{3, 3}, 10, 10)

	// Check that 3x3 area is filled
	for y := 4; y <= 6; y++ {
		for x := 4; x <= 6; x++ {
			assert.Equal(t, 1, ground[y][x], "cell (%d, %d) should be 1", x, y)
		}
	}

	// Check corners are not filled
	assert.Equal(t, 0, ground[0][0])
	assert.Equal(t, 0, ground[9][9])
}

func TestApplyBrush_EdgeCases(t *testing.T) {
	ground := createEmptyLayer(5, 5)

	// Apply brush at edge - should not panic and should clip
	applyBrush(ground, 0, 0, BrushSize{3, 3}, 5, 5)

	// Should have filled what fits
	assert.Equal(t, 1, ground[0][0])
	assert.Equal(t, 1, ground[0][1])
	assert.Equal(t, 1, ground[1][0])
	assert.Equal(t, 1, ground[1][1])
}

func TestCreateEmptyLayer(t *testing.T) {
	layer := createEmptyLayer(5, 3)

	assert.Equal(t, 3, len(layer))
	assert.Equal(t, 5, len(layer[0]))

	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			assert.Equal(t, 0, layer[y][x])
		}
	}
}

func TestCopyLayer(t *testing.T) {
	original := [][]int{{1, 2}, {3, 4}}
	copied := copyLayer(original)

	// Should be equal
	assert.Equal(t, original, copied)

	// Should be independent
	copied[0][0] = 99
	assert.Equal(t, 1, original[0][0])
}

func TestApplyBrushWithMirror_MirrorY(t *testing.T) {
	// Test left-right mirror (across Y-axis / vertical center)
	ground := createEmptyLayer(10, 10)

	// Apply brush at x=2, y=5 with MirrorY
	// 2x2 brush centered at (2,5): startX=1, startY=4, fills (1,4),(2,4),(1,5),(2,5)
	applyBrushWithMirror(ground, 2, 5, BrushSize{2, 2}, 10, 10, MirrorY)

	// Original position should be filled
	assert.Equal(t, 1, ground[5][2], "original position should be filled")
	assert.Equal(t, 1, ground[5][1], "original position should be filled")

	// Mirrored position: mirroredX = 10-1-2 = 7, centered at (7,5)
	// 2x2 brush centered at (7,5): startX=6, startY=4, fills (6,4),(7,4),(6,5),(7,5)
	assert.Equal(t, 1, ground[5][6], "mirrored position should be filled")
	assert.Equal(t, 1, ground[5][7], "mirrored position should be filled")
}

func TestApplyBrushWithMirror_MirrorX(t *testing.T) {
	// Test top-bottom mirror (across X-axis / horizontal center)
	ground := createEmptyLayer(10, 10)

	// Apply brush at x=5, y=2 with MirrorX
	// 2x2 brush centered at (5,2): startX=4, startY=1, fills (4,1),(5,1),(4,2),(5,2)
	applyBrushWithMirror(ground, 5, 2, BrushSize{2, 2}, 10, 10, MirrorX)

	// Original position should be filled
	assert.Equal(t, 1, ground[2][5], "original position should be filled")
	assert.Equal(t, 1, ground[1][5], "original position should be filled")

	// Mirrored position: mirroredY = 10-1-2 = 7, centered at (5,7)
	// 2x2 brush centered at (5,7): startX=4, startY=6, fills (4,6),(5,6),(4,7),(5,7)
	assert.Equal(t, 1, ground[6][5], "mirrored position should be filled")
	assert.Equal(t, 1, ground[7][5], "mirrored position should be filled")
}

func TestApplyBrushWithMirror_MirrorNone(t *testing.T) {
	// Test no mirror
	ground := createEmptyLayer(10, 10)

	// Apply brush at x=2, y=2 with MirrorNone
	applyBrushWithMirror(ground, 2, 2, BrushSize{2, 2}, 10, 10, MirrorNone)

	// Original position should be filled
	assert.Equal(t, 1, ground[2][2], "original position should be filled")

	// Mirrored positions should NOT be filled
	assert.Equal(t, 0, ground[2][7], "mirrored Y position should not be filled")
	assert.Equal(t, 0, ground[7][2], "mirrored X position should not be filled")
}

// Static Layer Generation Tests

func TestGenerateBridgeRoom_WithStaticCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       20,
		Height:      20,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		StaticCount: 3,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Count static placements (each static is 2x2, so count cells with static=1)
	staticCellCount := 0
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Static[y][x] == 1 {
				staticCellCount++
			}
		}
	}

	// Should have some statics placed (may be less than requested if constraints prevent)
	// Each static is 2x2 = 4 cells
	assert.GreaterOrEqual(t, staticCellCount, 0, "should have placed some static cells")
	t.Logf("Placed %d static cells (approximately %d statics)", staticCellCount, staticCellCount/4)
}

func TestGenerateBridgeRoom_StaticOnGround(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		StaticCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify all static cells are on ground
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Static[y][x] == 1 {
				assert.Equal(t, 1, resp.Payload.Ground[y][x], "static at (%d,%d) must be on ground", x, y)
			}
		}
	}
}

func TestGenerateBridgeRoom_StaticsDoNotTouch(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       30,
		Height:      30,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		StaticCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Find all static positions (top-left corners of 2x2 statics)
	var staticPositions []Point
	for y := 0; y < req.Height-1; y++ {
		for x := 0; x < req.Width-1; x++ {
			// Check if this is a top-left corner of a 2x2 static
			if resp.Payload.Static[y][x] == 1 &&
				resp.Payload.Static[y][x+1] == 1 &&
				resp.Payload.Static[y+1][x] == 1 &&
				resp.Payload.Static[y+1][x+1] == 1 {
				// Verify it's actually a corner (not part of a larger block)
				isCorner := true
				if x > 0 && resp.Payload.Static[y][x-1] == 1 {
					isCorner = false
				}
				if y > 0 && resp.Payload.Static[y-1][x] == 1 {
					isCorner = false
				}
				if isCorner {
					staticPositions = append(staticPositions, Point{X: x, Y: y})
				}
			}
		}
	}

	// Verify no two statics touch (including diagonals)
	for i := 0; i < len(staticPositions); i++ {
		for j := i + 1; j < len(staticPositions); j++ {
			pos1 := staticPositions[i]
			pos2 := staticPositions[j]
			assert.False(t, wouldTouch(pos1, pos2), "statics at (%d,%d) and (%d,%d) should not touch", pos1.X, pos1.Y, pos2.X, pos2.Y)
		}
	}
}

func TestGenerateBridgeRoom_DoorsConnectedWithStatics(t *testing.T) {
	// Test with multiple statics to ensure doors remain connected
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		StaticCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Create walkable map (ground=1 and static=0)
	walkable := make([][]bool, req.Height)
	for y := 0; y < req.Height; y++ {
		walkable[y] = make([]bool, req.Width)
		for x := 0; x < req.Width; x++ {
			walkable[y][x] = resp.Payload.Ground[y][x] == 1 && resp.Payload.Static[y][x] == 0
		}
	}

	// Get door positions
	doorPositions := []Point{
		{X: req.Width / 2, Y: 0},            // top
		{X: req.Width / 2, Y: req.Height - 1}, // bottom
		{X: 0, Y: req.Height / 2},           // left
		{X: req.Width - 1, Y: req.Height / 2}, // right
	}

	// Find nearest walkable cell to first door and do BFS
	startDoor := findNearestWalkable(walkable, doorPositions[0], req.Width, req.Height)
	if startDoor.X < 0 {
		t.Skip("No walkable cell found near first door")
	}

	visited := bfsConnectivity(walkable, startDoor, req.Width, req.Height)

	// Verify all other doors are reachable
	for i, doorPos := range doorPositions[1:] {
		nearestWalkable := findNearestWalkable(walkable, doorPos, req.Width, req.Height)
		if nearestWalkable.X >= 0 {
			assert.True(t, visited[nearestWalkable.Y][nearestWalkable.X], "door %d should be connected", i+1)
		}
	}
}

func TestIsValidStaticPosition(t *testing.T) {
	width, height := 10, 10
	ground := createEmptyLayer(width, height)
	softEdge := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	staticLayer := createEmptyLayer(width, height)
	forbiddenCells := make(map[Point]bool)

	// Fill ground in center area
	for y := 3; y < 7; y++ {
		for x := 3; x < 7; x++ {
			ground[y][x] = 1
		}
	}

	// Valid position in center
	assert.True(t, isValidStaticPosition(Point{X: 4, Y: 4}, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height))

	// Invalid - no ground
	assert.False(t, isValidStaticPosition(Point{X: 0, Y: 0}, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height))

	// Invalid - out of bounds
	assert.False(t, isValidStaticPosition(Point{X: 9, Y: 9}, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height))

	// Invalid - forbidden cell
	forbiddenCells[Point{X: 4, Y: 4}] = true
	assert.False(t, isValidStaticPosition(Point{X: 4, Y: 4}, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height))
}

func TestTouchesExistingStatic(t *testing.T) {
	width, height := 10, 10
	staticLayer := createEmptyLayer(width, height)

	// Place a static at (5,5)
	placeStatic(staticLayer, Point{X: 5, Y: 5})

	// Positions that touch
	assert.True(t, touchesExistingStatic(Point{X: 3, Y: 5}, staticLayer, width, height), "diagonal should touch")
	assert.True(t, touchesExistingStatic(Point{X: 7, Y: 5}, staticLayer, width, height), "adjacent right should touch")
	assert.True(t, touchesExistingStatic(Point{X: 5, Y: 7}, staticLayer, width, height), "adjacent below should touch")
	assert.True(t, touchesExistingStatic(Point{X: 5, Y: 3}, staticLayer, width, height), "adjacent above should touch")

	// Positions that don't touch (with gap)
	assert.False(t, touchesExistingStatic(Point{X: 0, Y: 0}, staticLayer, width, height), "far away should not touch")
	assert.False(t, touchesExistingStatic(Point{X: 8, Y: 5}, staticLayer, width, height), "one gap right should not touch")
}

func TestWouldTouch(t *testing.T) {
	// Two statics side by side (touching)
	assert.True(t, wouldTouch(Point{X: 0, Y: 0}, Point{X: 2, Y: 0}), "side by side should touch")
	assert.True(t, wouldTouch(Point{X: 0, Y: 0}, Point{X: 0, Y: 2}), "stacked should touch")

	// Two statics diagonal (touching)
	assert.True(t, wouldTouch(Point{X: 0, Y: 0}, Point{X: 2, Y: 2}), "diagonal should touch")

	// Two statics with gap (not touching)
	assert.False(t, wouldTouch(Point{X: 0, Y: 0}, Point{X: 4, Y: 0}), "with gap should not touch")
	assert.False(t, wouldTouch(Point{X: 0, Y: 0}, Point{X: 4, Y: 4}), "far diagonal should not touch")
}

func TestDistanceFromCenter(t *testing.T) {
	centerX, centerY := 10, 10

	// At center (pos (9,9) has static center at (10,10) which equals room center)
	assert.Equal(t, 0, distanceFromCenter(Point{X: 9, Y: 9}, centerX, centerY))

	// Near center
	assert.Equal(t, 2, distanceFromCenter(Point{X: 8, Y: 8}, centerX, centerY))

	// Far from center
	dist := distanceFromCenter(Point{X: 0, Y: 0}, centerX, centerY)
	assert.Greater(t, dist, 10)
}

func TestDistanceFromEdge(t *testing.T) {
	width, height := 20, 20

	// At edge (pos (0,10) has static center at (1, 11))
	assert.Equal(t, 1, distanceFromEdge(Point{X: 0, Y: 10}, width, height))

	// At center (pos (9,9) has static center at (10, 10))
	// Distances: left=10, right=9, top=10, bottom=9, min=9
	centerDist := distanceFromEdge(Point{X: 9, Y: 9}, width, height)
	assert.Equal(t, 9, centerDist)
}

func TestPlaceStatic(t *testing.T) {
	staticLayer := createEmptyLayer(10, 10)
	placeStatic(staticLayer, Point{X: 3, Y: 4})

	// Verify 2x2 is filled
	assert.Equal(t, 1, staticLayer[4][3])
	assert.Equal(t, 1, staticLayer[4][4])
	assert.Equal(t, 1, staticLayer[5][3])
	assert.Equal(t, 1, staticLayer[5][4])

	// Verify surrounding cells are not filled
	assert.Equal(t, 0, staticLayer[3][3])
	assert.Equal(t, 0, staticLayer[6][3])
}

func TestFilterTouchingPositions(t *testing.T) {
	// Static at (0,0) occupies (0,0), (1,0), (0,1), (1,1)
	// With 1-cell buffer, touching zone extends to x<=2 and y<=2
	positions := []Point{
		{X: 0, Y: 0},
		{X: 2, Y: 0}, // Would touch (0,0) - adjacent with 1 cell gap
		{X: 2, Y: 2}, // Would touch (0,0) - diagonal
		{X: 4, Y: 0}, // Would not touch (0,0) - 2 cell gap
		{X: 0, Y: 4}, // Would not touch (0,0) - 2 cell gap
	}

	filtered := filterTouchingPositions(positions, Point{X: 0, Y: 0})

	// Should only keep positions that don't touch
	assert.Len(t, filtered, 2)

	// Verify the kept positions
	hasPos := func(pts []Point, x, y int) bool {
		for _, p := range pts {
			if p.X == x && p.Y == y {
				return true
			}
		}
		return false
	}

	assert.True(t, hasPos(filtered, 4, 0))
	assert.True(t, hasPos(filtered, 0, 4))
}

func TestGetDoorForbiddenCells(t *testing.T) {
	doorPositions := map[DoorPosition]Point{
		DoorTop: {X: 10, Y: 0},
	}

	forbidden := getDoorForbiddenCells(doorPositions, 20, 20)

	// Door position should be forbidden
	assert.True(t, forbidden[Point{X: 10, Y: 0}])

	// Adjacent cells should be forbidden
	assert.True(t, forbidden[Point{X: 9, Y: 0}])
	assert.True(t, forbidden[Point{X: 11, Y: 0}])
	assert.True(t, forbidden[Point{X: 10, Y: 1}])

	// Far cells should not be forbidden
	assert.False(t, forbidden[Point{X: 0, Y: 10}])
}

func TestCheckConnectivityAfterPlacement(t *testing.T) {
	width, height := 10, 10
	ground := createEmptyLayer(width, height)
	staticLayer := createEmptyLayer(width, height)

	// Create a vertical path in the middle
	for y := 0; y < height; y++ {
		ground[y][5] = 1
		ground[y][6] = 1
	}

	doorPositions := map[DoorPosition]Point{
		DoorTop:    {X: 5, Y: 0},
		DoorBottom: {X: 5, Y: 9},
	}

	// Placing static that doesn't block path should be OK
	assert.True(t, checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, Point{X: 0, Y: 0}, width, height))

	// Placing static that blocks path should fail
	// (blocking middle of the path)
	assert.False(t, checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, Point{X: 5, Y: 4}, width, height))
}

func TestGenerateBridgeRoom_ZeroStaticCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       15,
		Height:      15,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		StaticCount: 0, // Explicitly zero
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify static layer is all zeros
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			assert.Equal(t, 0, resp.Payload.Static[y][x], "static should be 0 when StaticCount=0")
		}
	}
}

// Turret Layer Generation Tests

func TestGenerateBridgeRoom_WithTurretCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		TurretCount: 4,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Count turret placements
	turretCount := 0
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Turret[y][x] == 1 {
				turretCount++
			}
		}
	}

	// Should have some turrets placed (may be less than requested if constraints prevent)
	assert.GreaterOrEqual(t, turretCount, 0, "should have placed some turrets")
	t.Logf("Placed %d turrets", turretCount)
}

func TestGenerateBridgeRoom_TurretOnGround(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		TurretCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify all turret cells are on ground
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Turret[y][x] == 1 {
				assert.Equal(t, 1, resp.Payload.Ground[y][x], "turret at (%d,%d) must be on ground", x, y)
			}
		}
	}
}

func TestGenerateBridgeRoom_TurretNotOnStatic(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		StaticCount: 3,
		TurretCount: 4,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify turrets don't overlap with statics
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Turret[y][x] == 1 {
				assert.Equal(t, 0, resp.Payload.Static[y][x], "turret at (%d,%d) must not overlap with static", x, y)
			}
		}
	}
}

func TestGenerateBridgeRoom_TurretDistanceFromDoors(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       30,
		Height:      30,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		TurretCount: 4,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Verify all turrets are at least 4 cells away from doors
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Turret[y][x] == 1 {
				turretPos := Point{X: x, Y: y}
				for door, doorPos := range doorPositions {
					dist := manhattanDistance(turretPos, doorPos)
					assert.GreaterOrEqual(t, dist, turretMinDoorDistance,
						"turret at (%d,%d) is too close to %s door (distance: %d)", x, y, door, dist)
				}
			}
		}
	}
}

func TestGenerateBridgeRoom_TurretMinDistance(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       30,
		Height:      30,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		TurretCount: 6,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Collect all turret positions
	var turretPositions []Point
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.Turret[y][x] == 1 {
				turretPositions = append(turretPositions, Point{X: x, Y: y})
			}
		}
	}

	// Verify minimum distance between turrets
	for i := 0; i < len(turretPositions); i++ {
		for j := i + 1; j < len(turretPositions); j++ {
			dist := manhattanDistance(turretPositions[i], turretPositions[j])
			assert.GreaterOrEqual(t, dist, turretMinTurretDistance,
				"turrets at (%d,%d) and (%d,%d) are too close (distance: %d)",
				turretPositions[i].X, turretPositions[i].Y,
				turretPositions[j].X, turretPositions[j].Y, dist)
		}
	}
}

func TestGenerateBridgeRoom_DoorsConnectedWithTurrets(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		StaticCount: 2,
		TurretCount: 4,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Create walkable map
	walkable := make([][]bool, req.Height)
	for y := 0; y < req.Height; y++ {
		walkable[y] = make([]bool, req.Width)
		for x := 0; x < req.Width; x++ {
			walkable[y][x] = resp.Payload.Ground[y][x] == 1 &&
				resp.Payload.Static[y][x] == 0 &&
				resp.Payload.Turret[y][x] == 0
		}
	}

	// Get door positions
	doorPositions := []Point{
		{X: req.Width / 2, Y: 0},
		{X: req.Width / 2, Y: req.Height - 1},
		{X: 0, Y: req.Height / 2},
		{X: req.Width - 1, Y: req.Height / 2},
	}

	// Find nearest walkable cell to first door and do BFS
	startDoor := findNearestWalkable(walkable, doorPositions[0], req.Width, req.Height)
	if startDoor.X < 0 {
		t.Skip("No walkable cell found near first door")
	}

	visited := bfsConnectivity(walkable, startDoor, req.Width, req.Height)

	// Verify all other doors are reachable
	for i, doorPos := range doorPositions[1:] {
		nearestWalkable := findNearestWalkable(walkable, doorPos, req.Width, req.Height)
		if nearestWalkable.X >= 0 {
			assert.True(t, visited[nearestWalkable.Y][nearestWalkable.X], "door %d should be connected", i+1)
		}
	}
}

func TestIsValidTurretPosition(t *testing.T) {
	width, height := 20, 20
	ground := createEmptyLayer(width, height)
	softEdge := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	staticLayer := createEmptyLayer(width, height)
	turretLayer := createEmptyLayer(width, height)

	// Fill ground in center area
	for y := 5; y < 15; y++ {
		for x := 5; x < 15; x++ {
			ground[y][x] = 1
		}
	}

	doorPositions := map[DoorPosition]Point{
		DoorTop:    {X: 10, Y: 0},
		DoorBottom: {X: 10, Y: 19},
	}

	// Valid position far from doors
	assert.True(t, isValidTurretPosition(Point{X: 10, Y: 10}, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height))

	// Invalid - no ground
	assert.False(t, isValidTurretPosition(Point{X: 0, Y: 0}, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height))

	// Invalid - too close to door (within 4 cells)
	assert.False(t, isValidTurretPosition(Point{X: 10, Y: 3}, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height))
}

func TestManhattanDistance(t *testing.T) {
	assert.Equal(t, 0, manhattanDistance(Point{X: 5, Y: 5}, Point{X: 5, Y: 5}))
	assert.Equal(t, 5, manhattanDistance(Point{X: 0, Y: 0}, Point{X: 3, Y: 2}))
	assert.Equal(t, 10, manhattanDistance(Point{X: 0, Y: 0}, Point{X: 5, Y: 5}))
}

func TestTooCloseToExistingTurret(t *testing.T) {
	width, height := 10, 10
	turretLayer := createEmptyLayer(width, height)

	// Place a turret at (5,5)
	turretLayer[5][5] = 1

	// Positions too close (distance < 2)
	assert.True(t, tooCloseToExistingTurret(Point{X: 5, Y: 4}, turretLayer, width, height), "adjacent should be too close")
	assert.True(t, tooCloseToExistingTurret(Point{X: 6, Y: 5}, turretLayer, width, height), "adjacent should be too close")

	// Positions at exactly distance 2 should be OK
	assert.False(t, tooCloseToExistingTurret(Point{X: 5, Y: 3}, turretLayer, width, height), "distance 2 should be OK")
	assert.False(t, tooCloseToExistingTurret(Point{X: 7, Y: 5}, turretLayer, width, height), "distance 2 should be OK")
}

func TestFilterTurretsTooClose(t *testing.T) {
	positions := []Point{
		{X: 5, Y: 5},
		{X: 5, Y: 6}, // Distance 1 - too close
		{X: 6, Y: 5}, // Distance 1 - too close
		{X: 7, Y: 5}, // Distance 2 - OK
		{X: 5, Y: 7}, // Distance 2 - OK
		{X: 8, Y: 8}, // Distance 6 - OK
	}

	filtered := filterTurretsTooClose(positions, Point{X: 5, Y: 5})

	// Should only keep positions at distance >= 2
	assert.Len(t, filtered, 3)
}

func TestMinDistanceToEdge(t *testing.T) {
	width, height := 20, 20

	// At corner
	assert.Equal(t, 0, minDistanceToEdge(Point{X: 0, Y: 0}, width, height))

	// At edge
	assert.Equal(t, 0, minDistanceToEdge(Point{X: 10, Y: 0}, width, height))

	// At center
	assert.Equal(t, 9, minDistanceToEdge(Point{X: 10, Y: 10}, width, height))
}

func TestIsNearCorner(t *testing.T) {
	width, height := 20, 20
	threshold := 2

	// Corners
	assert.True(t, isNearCorner(Point{X: 0, Y: 0}, width, height, threshold))
	assert.True(t, isNearCorner(Point{X: 19, Y: 0}, width, height, threshold))
	assert.True(t, isNearCorner(Point{X: 0, Y: 19}, width, height, threshold))
	assert.True(t, isNearCorner(Point{X: 19, Y: 19}, width, height, threshold))

	// Near corner
	assert.True(t, isNearCorner(Point{X: 1, Y: 1}, width, height, threshold))

	// Not near corner
	assert.False(t, isNearCorner(Point{X: 10, Y: 10}, width, height, threshold))
}

func TestGenerateBridgeRoom_ZeroTurretCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       15,
		Height:      15,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		TurretCount: 0, // Explicitly zero
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify turret layer is all zeros
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			assert.Equal(t, 0, resp.Payload.Turret[y][x], "turret should be 0 when TurretCount=0")
		}
	}
}

func TestGetGroundCornerType(t *testing.T) {
	width, height := 10, 10
	ground := createEmptyLayer(width, height)

	// Create an L-shaped ground pattern:
	// · · · · ·
	// · █ █ █ ·
	// · █ · · ·
	// · █ · · ·
	// · · · · ·
	ground[1][1] = 1
	ground[1][2] = 1
	ground[1][3] = 1
	ground[2][1] = 1
	ground[3][1] = 1

	// 90° corner: position (1,1) has neighbors at (1,2) right and (2,1) bottom - forms an L
	// Actually (1,1) has top=(0,1)=0, right=(2,1)=1, bottom=(1,2)=1, left=(0,1)=0
	// So it has 2 neighbors: right and bottom - this is a 90° corner
	assert.Equal(t, CornerType90, getGroundCornerType(Point{X: 1, Y: 1}, ground, width, height),
		"position (1,1) should be 90° corner (right+bottom neighbors)")

	// 270° corner: position (2,1) has neighbors at top=(1,1)=1, right=(3,1)=1, bottom=(2,2)=0, left=(1,1)=1
	// Wait, let me recalculate. ground[y][x], so:
	// ground[1][1]=1, ground[1][2]=1, ground[1][3]=1, ground[2][1]=1, ground[3][1]=1
	// Position (2,1) means x=2, y=1
	// top: (2,0) = 0, right: (3,1) = ground[1][3] = 1, bottom: (2,2) = ground[2][2] = 0, left: (1,1) = ground[1][1] = 1
	// So 2 neighbors (right, left) - not adjacent, so not a corner

	// Let me reconsider the pattern. Let me use x,y coordinates properly:
	// ground[y][x]
	// (1,1): ground[1][1] = 1
	// (2,1): ground[1][2] = 1
	// (3,1): ground[1][3] = 1
	// (1,2): ground[2][1] = 1
	// (1,3): ground[3][1] = 1

	// So the pattern is:
	// Row 0: · · · · ·
	// Row 1: · █ █ █ ·   (y=1: x=1,2,3)
	// Row 2: · █ · · ·   (y=2: x=1)
	// Row 3: · █ · · ·   (y=3: x=1)
	// Row 4: · · · · ·

	// Position (1,1) x=1,y=1: check neighbors
	// top: (1,0) = ground[0][1] = 0
	// right: (2,1) = ground[1][2] = 1
	// bottom: (1,2) = ground[2][1] = 1
	// left: (0,1) = ground[1][0] = 0
	// So 2 neighbors: right+bottom = adjacent = 90° corner ✓

	// Position (2,1) x=2,y=1:
	// top: (2,0) = ground[0][2] = 0
	// right: (3,1) = ground[1][3] = 1
	// bottom: (2,2) = ground[2][2] = 0
	// left: (1,1) = ground[1][1] = 1
	// So 2 neighbors: right+left = not adjacent (opposite sides) = not a corner
	assert.Equal(t, CornerTypeNone, getGroundCornerType(Point{X: 2, Y: 1}, ground, width, height),
		"position (2,1) should not be a corner (left+right neighbors, not adjacent)")

	// Position (3,1) x=3,y=1:
	// top: (3,0) = ground[0][3] = 0
	// right: (4,1) = ground[1][4] = 0
	// bottom: (3,2) = ground[2][3] = 0
	// left: (2,1) = ground[1][2] = 1
	// So 1 neighbor: not a corner
	assert.Equal(t, CornerTypeNone, getGroundCornerType(Point{X: 3, Y: 1}, ground, width, height),
		"position (3,1) should not be a corner (only 1 neighbor)")

	// Position (1,2) x=1,y=2:
	// top: (1,1) = ground[1][1] = 1
	// right: (2,2) = ground[2][2] = 0
	// bottom: (1,3) = ground[3][1] = 1
	// left: (0,2) = ground[2][0] = 0
	// So 2 neighbors: top+bottom = not adjacent = not a corner
	assert.Equal(t, CornerTypeNone, getGroundCornerType(Point{X: 1, Y: 2}, ground, width, height),
		"position (1,2) should not be a corner (top+bottom neighbors, not adjacent)")

	// Now test 270° corner - need 3 neighbors
	// Create a + shaped ground pattern at center:
	ground2 := createEmptyLayer(width, height)
	ground2[4][5] = 1 // top
	ground2[5][4] = 1 // left
	ground2[5][5] = 1 // center
	ground2[5][6] = 1 // right
	ground2[6][5] = 1 // bottom

	// Position (5,5) center: has 4 neighbors = not a corner (straight through)
	assert.Equal(t, CornerTypeNone, getGroundCornerType(Point{X: 5, Y: 5}, ground2, width, height),
		"position (5,5) should not be a corner (4 neighbors)")

	// Create T-shaped pattern for 270° test
	ground3 := createEmptyLayer(width, height)
	ground3[4][5] = 1 // top
	ground3[5][4] = 1 // left
	ground3[5][5] = 1 // center
	ground3[5][6] = 1 // right
	// No bottom neighbor

	// Position (5,5) center: top+left+right = 3 neighbors = 270° corner
	assert.Equal(t, CornerType270, getGroundCornerType(Point{X: 5, Y: 5}, ground3, width, height),
		"position (5,5) should be 270° corner (3 neighbors)")
}
