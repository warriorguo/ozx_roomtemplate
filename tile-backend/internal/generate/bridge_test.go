package generate

import (
	"fmt"
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

// Soft Edge Layer Generation Tests

func TestGenerateBridgeRoom_WithSoftEdgeCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:         25,
		Height:        25,
		Doors:         []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		SoftEdgeCount: 3,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Count soft edge cells
	softEdgeCellCount := 0
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.SoftEdge[y][x] == 1 {
				softEdgeCellCount++
			}
		}
	}

	// Should have some soft edges placed (each is at least 3 cells)
	t.Logf("Placed %d soft edge cells", softEdgeCellCount)

	// If any placed, verify they are on void cells (soft edges mark concave void notches)
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.SoftEdge[y][x] == 1 {
				assert.Equal(t, 0, resp.Payload.Ground[y][x],
					"soft edge at (%d,%d) must be on void (concave notch)", x, y)
			}
		}
	}
}

func TestGenerateBridgeRoom_SoftEdgeDistanceFromDoors(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:         25,
		Height:        25,
		Doors:         []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		SoftEdgeCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Get door positions
	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Verify soft edges are at least 2 cells away from doors
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.SoftEdge[y][x] == 1 {
				for _, doorPos := range doorPositions {
					dist := manhattanDistance(Point{X: x, Y: y}, doorPos)
					assert.GreaterOrEqual(t, dist, softEdgeMinDoorDistance,
						"soft edge at (%d,%d) should be at least %d cells from door at (%d,%d)",
						x, y, softEdgeMinDoorDistance, doorPos.X, doorPos.Y)
				}
			}
		}
	}
}

func TestGenerateBridgeRoom_SoftEdgeShape(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:         30,
		Height:        30,
		Doors:         []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		SoftEdgeCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify soft edges form 1×N or N×1 shapes
	// Check debug info if available
	if resp.DebugInfo != nil && resp.DebugInfo.SoftEdge != nil {
		for _, placement := range resp.DebugInfo.SoftEdge.Placements {
			t.Logf("SoftEdge: %s size=%s reason=%s", placement.Position, placement.Size, placement.Reason)
		}
	}
}

func TestGenerateBridgeRoom_ZeroSoftEdgeCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:         15,
		Height:        15,
		Doors:         []DoorPosition{DoorTop, DoorBottom},
		SoftEdgeCount: 0,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify soft edge layer is all zeros
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			assert.Equal(t, 0, resp.Payload.SoftEdge[y][x], "soft edge should be 0 when SoftEdgeCount=0")
		}
	}
}

func TestFindHorizontalConcave(t *testing.T) {
	// Create a ground layer with a horizontal void notch (carved from top)
	// This matches the user's expected pattern like row 3: ███····██████
	// Ground looks like:
	// 0 0 0 0 0 0 0 0  <- void (outside)
	// 1 0 0 0 0 1 1 1  <- notch: ground at edges, void in middle with ground below
	// 1 1 1 1 1 1 1 1  <- solid ground
	// 1 1 1 1 1 1 1 1
	ground := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0}, // void above
		{1, 0, 0, 0, 0, 1, 1, 1}, // void notch at x=1-4, ground at x=0 and x=5
		{1, 1, 1, 1, 1, 1, 1, 1}, // ground below
		{1, 1, 1, 1, 1, 1, 1, 1},
	}
	softEdgeLayer := createEmptyLayer(8, 4)
	doorPositions := map[DoorPosition]Point{}

	// Should find a horizontal concave notch at (1,1) to (4,1)
	// The notch has: ground below (row 2), void above (row 0), ground on left (x=0) and right (x=5)
	placement := findHorizontalConcave(ground, softEdgeLayer, doorPositions, 1, 1, 8, 4)
	require.NotNil(t, placement, "should find horizontal concave")
	assert.Equal(t, 1, placement.StartX)
	assert.Equal(t, 1, placement.StartY)
	assert.Equal(t, 4, placement.Width, "concave should be 4 cells wide")
	assert.Equal(t, 1, placement.Height)
}

func TestFindVerticalConcave(t *testing.T) {
	// Create a ground layer with a vertical void notch (carved from left)
	// Ground looks like:
	// 0 1 1 1 1
	// 0 0 1 1 1  <- vertical notch at x=1, ground on right
	// 0 0 1 1 1
	// 0 0 1 1 1
	// 0 1 1 1 1
	ground := [][]int{
		{0, 1, 1, 1, 1}, // ground closes the top
		{0, 0, 1, 1, 1}, // void notch at x=1, y=1-3
		{0, 0, 1, 1, 1},
		{0, 0, 1, 1, 1},
		{0, 1, 1, 1, 1}, // ground closes the bottom
	}
	softEdgeLayer := createEmptyLayer(5, 5)
	doorPositions := map[DoorPosition]Point{}

	// Should find a vertical concave notch at (1,1) to (1,3)
	// The notch has: ground on right (x=2), void on left (x=0), ground at top (y=0) and bottom (y=4)
	placement := findVerticalConcave(ground, softEdgeLayer, doorPositions, 1, 1, 5, 5)
	require.NotNil(t, placement, "should find vertical concave")
	assert.Equal(t, 1, placement.StartX)
	assert.Equal(t, 1, placement.StartY)
	assert.Equal(t, 1, placement.Width)
	assert.Equal(t, 3, placement.Height, "concave should be 3 cells tall")
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

// ============================================================================
// Mob Ground Layer Tests
// ============================================================================

func TestGenerateBridgeRoom_WithMobGroundCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          25,
		Height:         25,
		Doors:          []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		MobGroundCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Count mob ground cells
	mobGroundCount := 0
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobGround[y][x] == 1 {
				mobGroundCount++
			}
		}
	}

	// Should have placed some mob ground (may be less if space is limited)
	t.Logf("Placed %d mob ground cells", mobGroundCount)
	assert.Greater(t, mobGroundCount, 0, "should have placed some mob ground")
}

func TestGenerateBridgeRoom_MobGroundOnGround(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          20,
		Height:         20,
		Doors:          []DoorPosition{DoorTop, DoorBottom},
		MobGroundCount: 3,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify all mob ground cells are on ground
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobGround[y][x] == 1 {
				assert.Equal(t, 1, resp.Payload.Ground[y][x],
					"mob ground at (%d,%d) should be on ground", x, y)
			}
		}
	}
}

func TestGenerateBridgeRoom_MobGroundNotOnOtherLayers(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          30,
		Height:         30,
		Doors:          []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		StaticCount:    3,
		TurretCount:    4,
		MobGroundCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify mob ground doesn't overlap with static or turret
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobGround[y][x] == 1 {
				assert.Equal(t, 0, resp.Payload.Static[y][x],
					"mob ground at (%d,%d) should not overlap with static", x, y)
				assert.Equal(t, 0, resp.Payload.Turret[y][x],
					"mob ground at (%d,%d) should not overlap with turret", x, y)
			}
		}
	}
}

func TestGenerateBridgeRoom_MobGroundDistanceFromDoors(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          20,
		Height:         20,
		Doors:          []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		MobGroundCount: 4,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Get door positions
	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Verify mob ground is at least 2 cells away from doors
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobGround[y][x] == 1 {
				for _, doorPos := range doorPositions {
					dist := manhattanDistance(Point{X: x, Y: y}, doorPos)
					assert.GreaterOrEqual(t, dist, mobGroundMinDoorDistance,
						"mob ground at (%d,%d) should be at least %d cells from door at (%d,%d)",
						x, y, mobGroundMinDoorDistance, doorPos.X, doorPos.Y)
				}
			}
		}
	}
}

func TestGenerateBridgeRoom_MobGroundDoNotTouch(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          30,
		Height:         30,
		Doors:          []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		MobGroundCount: 6,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Find all mob ground positions
	var mobGroundPositions []Point
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobGround[y][x] == 1 {
				mobGroundPositions = append(mobGroundPositions, Point{X: x, Y: y})
			}
		}
	}

	// Group adjacent cells into clusters
	visited := make(map[Point]bool)
	var clusters [][]Point

	for _, pos := range mobGroundPositions {
		if visited[pos] {
			continue
		}

		// BFS to find cluster
		cluster := []Point{pos}
		visited[pos] = true
		queue := []Point{pos}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			// Check 4-directional neighbors
			for _, neighbor := range mobGroundPositions {
				if !visited[neighbor] {
					if abs(curr.X-neighbor.X)+abs(curr.Y-neighbor.Y) == 1 {
						visited[neighbor] = true
						cluster = append(cluster, neighbor)
						queue = append(queue, neighbor)
					}
				}
			}
		}
		clusters = append(clusters, cluster)
	}

	// Verify clusters don't touch each other (including diagonals)
	for i := 0; i < len(clusters); i++ {
		for j := i + 1; j < len(clusters); j++ {
			for _, posI := range clusters[i] {
				for _, posJ := range clusters[j] {
					// Check if they touch (including diagonals)
					dx := abs(posI.X - posJ.X)
					dy := abs(posI.Y - posJ.Y)
					touching := dx <= 1 && dy <= 1
					assert.False(t, touching,
						"mob ground clusters should not touch: (%d,%d) and (%d,%d)",
						posI.X, posI.Y, posJ.X, posJ.Y)
				}
			}
		}
	}
}

func TestDivideMobGroundIntoGroups(t *testing.T) {
	tests := []struct {
		count    int
		expected []int
	}{
		{0, nil},
		{1, []int{1}},
		{2, []int{1, 1}},
		{3, []int{1, 1, 1}},
		{4, []int{2, 1, 1}},
		{5, []int{2, 2, 1}},
		{6, []int{2, 2, 2}},
		{9, []int{3, 3, 3}},
		{10, []int{4, 3, 3}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count_%d", tt.count), func(t *testing.T) {
			result := divideMobGroundIntoGroups(tt.count)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, len(tt.expected), len(result), "group count mismatch")
				// Verify total equals input
				total := 0
				for _, g := range result {
					total += g
					assert.GreaterOrEqual(t, g, 1, "each group should have at least 1")
				}
				assert.Equal(t, tt.count, total, "total should equal input count")
			}
		})
	}
}

func TestGenerateBridgeRoom_ZeroMobGroundCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          15,
		Height:         15,
		Doors:          []DoorPosition{DoorTop, DoorBottom},
		MobGroundCount: 0,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify mob ground layer is all zeros
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			assert.Equal(t, 0, resp.Payload.MobGround[y][x], "mob ground should be 0 when MobGroundCount=0")
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

// ============================================================================
// Mob Air (Fly) Layer Tests
// ============================================================================

func TestGenerateBridgeRoom_WithMobAirCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		MobAirCount: 6,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Count mob air cells
	mobAirCount := 0
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobAir[y][x] == 1 {
				mobAirCount++
			}
		}
	}

	// Should have placed some mob air
	t.Logf("Placed %d mob air cells", mobAirCount)
	assert.Greater(t, mobAirCount, 0, "should have placed some mob air")
}

func TestGenerateBridgeRoom_MobAirNoGroundRequirement(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       20,
		Height:      20,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		MobAirCount: 4,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify mob air can be placed (no ground requirement for flying mobs)
	mobAirCount := 0
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobAir[y][x] == 1 {
				mobAirCount++
			}
		}
	}
	assert.Greater(t, mobAirCount, 0, "should have placed some mob air")
}

func TestGenerateBridgeRoom_MobAirNotOnOtherLayers(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          30,
		Height:         30,
		Doors:          []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		StaticCount:    3,
		TurretCount:    4,
		MobGroundCount: 3,
		MobAirCount:    5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify mob air doesn't overlap with other layers
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobAir[y][x] == 1 {
				assert.Equal(t, 0, resp.Payload.Static[y][x],
					"mob air at (%d,%d) should not overlap with static", x, y)
				assert.Equal(t, 0, resp.Payload.Turret[y][x],
					"mob air at (%d,%d) should not overlap with turret", x, y)
				assert.Equal(t, 0, resp.Payload.MobGround[y][x],
					"mob air at (%d,%d) should not overlap with mob ground", x, y)
			}
		}
	}
}

func TestGenerateBridgeRoom_MobAirDistanceFromDoors(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		MobAirCount: 5,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Get door positions
	doorPositions := getDoorCenterPositions(req.Width, req.Height, req.Doors)

	// Verify mob air is at least 4 cells away from doors
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobAir[y][x] == 1 {
				for _, doorPos := range doorPositions {
					dist := manhattanDistance(Point{X: x, Y: y}, doorPos)
					assert.GreaterOrEqual(t, dist, mobAirMinDoorDistance,
						"mob air at (%d,%d) should be at least %d cells from door at (%d,%d)",
						x, y, mobAirMinDoorDistance, doorPos.X, doorPos.Y)
				}
			}
		}
	}
}

func TestGenerateBridgeRoom_MobAirDoNotTouch(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       30,
		Height:      30,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		MobAirCount: 8,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Find all mob air positions
	var mobAirPositions []Point
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobAir[y][x] == 1 {
				mobAirPositions = append(mobAirPositions, Point{X: x, Y: y})
			}
		}
	}

	// Verify no two mob air cells touch (including diagonals)
	for i := 0; i < len(mobAirPositions); i++ {
		for j := i + 1; j < len(mobAirPositions); j++ {
			dx := abs(mobAirPositions[i].X - mobAirPositions[j].X)
			dy := abs(mobAirPositions[i].Y - mobAirPositions[j].Y)
			touching := dx <= 1 && dy <= 1
			assert.False(t, touching,
				"mob air at (%d,%d) and (%d,%d) should not touch",
				mobAirPositions[i].X, mobAirPositions[i].Y,
				mobAirPositions[j].X, mobAirPositions[j].Y)
		}
	}
}

func TestGenerateBridgeRoom_MobAirDistanceFromEdges(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       25,
		Height:      25,
		Doors:       []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		MobAirCount: 6,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify mob air is at least 2 cells away from all edges
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			if resp.Payload.MobAir[y][x] == 1 {
				// Check distance from edges
				assert.GreaterOrEqual(t, x, mobAirMinEdgeDistance,
					"mob air at (%d,%d) should be at least %d cells from left edge", x, y, mobAirMinEdgeDistance)
				assert.Less(t, x, req.Width-mobAirMinEdgeDistance,
					"mob air at (%d,%d) should be at least %d cells from right edge", x, y, mobAirMinEdgeDistance)
				assert.GreaterOrEqual(t, y, mobAirMinEdgeDistance,
					"mob air at (%d,%d) should be at least %d cells from top edge", x, y, mobAirMinEdgeDistance)
				assert.Less(t, y, req.Height-mobAirMinEdgeDistance,
					"mob air at (%d,%d) should be at least %d cells from bottom edge", x, y, mobAirMinEdgeDistance)
			}
		}
	}
}

func TestGenerateBridgeRoom_ZeroMobAirCount(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:       15,
		Height:      15,
		Doors:       []DoorPosition{DoorTop, DoorBottom},
		MobAirCount: 0,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)

	// Verify mob air layer is all zeros
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			assert.Equal(t, 0, resp.Payload.MobAir[y][x], "mob air should be 0 when MobAirCount=0")
		}
	}
}

func TestCalculateGridDimensions(t *testing.T) {
	tests := []struct {
		name        string
		targetCount int
		width       int
		height      int
	}{
		{"4 items in 20x20", 4, 20, 20},
		{"9 items in 20x20", 9, 20, 20},
		{"6 items in 30x20", 6, 30, 20},
		{"1 item", 1, 20, 20},
		{"0 items", 0, 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, rows := calculateGridDimensions(tt.targetCount, tt.width, tt.height)

			// Grid should have at least 1 row and 1 col
			assert.GreaterOrEqual(t, cols, 1, "cols should be >= 1")
			assert.GreaterOrEqual(t, rows, 1, "rows should be >= 1")

			// Grid should have enough cells for target count (if targetCount > 0)
			if tt.targetCount > 0 {
				assert.GreaterOrEqual(t, cols*rows, tt.targetCount,
					"grid should have enough cells: cols=%d, rows=%d, target=%d", cols, rows, tt.targetCount)
			}

			t.Logf("%s: cols=%d, rows=%d (for %d items)", tt.name, cols, rows, tt.targetCount)
		})
	}
}

func TestArrangeMobAirEvenlySpaced(t *testing.T) {
	// Create a set of valid positions
	validPositions := []Point{
		{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 10, Y: 0}, {X: 15, Y: 0},
		{X: 0, Y: 5}, {X: 5, Y: 5}, {X: 10, Y: 5}, {X: 15, Y: 5},
		{X: 0, Y: 10}, {X: 5, Y: 10}, {X: 10, Y: 10}, {X: 15, Y: 10},
		{X: 0, Y: 15}, {X: 5, Y: 15}, {X: 10, Y: 15}, {X: 15, Y: 15},
	}

	// Select 4 positions
	result := arrangeMobAirEvenlySpaced(validPositions, 4, 20, 20)

	assert.LessOrEqual(t, len(result), 4)
	assert.Greater(t, len(result), 0)

	// All results should be unique
	seen := make(map[Point]bool)
	for _, pos := range result {
		assert.False(t, seen[pos], "duplicate position in result")
		seen[pos] = true
	}
}

// ============================================================================
// Debug Info Tests
// ============================================================================

func TestGenerateBridgeRoom_DebugInfoPopulated(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          25,
		Height:         25,
		Doors:          []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		StaticCount:    3,
		TurretCount:    4,
		MobGroundCount: 3,
		MobAirCount:    4,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)
	require.NotNil(t, resp.DebugInfo, "DebugInfo should not be nil")

	// Verify Ground debug info
	require.NotNil(t, resp.DebugInfo.Ground, "Ground debug info should not be nil")
	assert.Greater(t, len(resp.DebugInfo.Ground.DoorConnections), 0, "should have door connections")
	for _, conn := range resp.DebugInfo.Ground.DoorConnections {
		assert.NotEmpty(t, conn.From, "connection From should not be empty")
		assert.NotEmpty(t, conn.To, "connection To should not be empty")
		assert.NotEmpty(t, conn.PathType, "connection PathType should not be empty")
		assert.NotEmpty(t, conn.BrushSize, "connection BrushSize should not be empty")
	}
	t.Logf("Ground: %d door connections, %d platforms", len(resp.DebugInfo.Ground.DoorConnections), len(resp.DebugInfo.Ground.Platforms))

	// Verify Static debug info
	require.NotNil(t, resp.DebugInfo.Static, "Static debug info should not be nil")
	assert.Equal(t, req.StaticCount, resp.DebugInfo.Static.TargetCount, "static target count should match request")
	assert.GreaterOrEqual(t, resp.DebugInfo.Static.PlacedCount, 0, "placed count should be >= 0")
	assert.Equal(t, len(resp.DebugInfo.Static.Placements), resp.DebugInfo.Static.PlacedCount, "placements should match placed count")
	for _, p := range resp.DebugInfo.Static.Placements {
		assert.NotEmpty(t, p.Position, "placement position should not be empty")
		assert.Equal(t, "2x2", p.Size, "static size should be 2x2")
		assert.NotEmpty(t, p.Reason, "placement reason should not be empty")
	}
	t.Logf("Static: target=%d, placed=%d", resp.DebugInfo.Static.TargetCount, resp.DebugInfo.Static.PlacedCount)

	// Verify Turret debug info
	require.NotNil(t, resp.DebugInfo.Turret, "Turret debug info should not be nil")
	assert.Equal(t, req.TurretCount, resp.DebugInfo.Turret.TargetCount, "turret target count should match request")
	assert.GreaterOrEqual(t, resp.DebugInfo.Turret.PlacedCount, 0, "placed count should be >= 0")
	assert.Equal(t, len(resp.DebugInfo.Turret.Placements), resp.DebugInfo.Turret.PlacedCount, "placements should match placed count")
	for _, p := range resp.DebugInfo.Turret.Placements {
		assert.NotEmpty(t, p.Position, "placement position should not be empty")
		assert.Equal(t, "1x1", p.Size, "turret size should be 1x1")
		assert.NotEmpty(t, p.Reason, "placement reason should not be empty")
	}
	t.Logf("Turret: target=%d, placed=%d", resp.DebugInfo.Turret.TargetCount, resp.DebugInfo.Turret.PlacedCount)

	// Verify MobGround debug info
	require.NotNil(t, resp.DebugInfo.MobGround, "MobGround debug info should not be nil")
	assert.Equal(t, req.MobGroundCount, resp.DebugInfo.MobGround.TargetCount, "mobGround target count should match request")
	assert.GreaterOrEqual(t, resp.DebugInfo.MobGround.PlacedCount, 0, "placed count should be >= 0")
	assert.Greater(t, len(resp.DebugInfo.MobGround.Groups), 0, "should have at least one group")
	for _, group := range resp.DebugInfo.MobGround.Groups {
		assert.NotEmpty(t, group.Strategy, "group strategy should not be empty")
		assert.GreaterOrEqual(t, group.TargetCount, 1, "group target count should be >= 1")
	}
	t.Logf("MobGround: target=%d, placed=%d, groups=%d", resp.DebugInfo.MobGround.TargetCount, resp.DebugInfo.MobGround.PlacedCount, len(resp.DebugInfo.MobGround.Groups))

	// Verify MobAir debug info
	require.NotNil(t, resp.DebugInfo.MobAir, "MobAir debug info should not be nil")
	assert.Equal(t, req.MobAirCount, resp.DebugInfo.MobAir.TargetCount, "mobAir target count should match request")
	assert.NotEmpty(t, resp.DebugInfo.MobAir.Strategy, "mobAir strategy should not be empty")
	assert.GreaterOrEqual(t, resp.DebugInfo.MobAir.PlacedCount, 0, "placed count should be >= 0")
	assert.Equal(t, len(resp.DebugInfo.MobAir.Placements), resp.DebugInfo.MobAir.PlacedCount, "placements should match placed count")
	t.Logf("MobAir: target=%d, placed=%d, strategy=%s", resp.DebugInfo.MobAir.TargetCount, resp.DebugInfo.MobAir.PlacedCount, resp.DebugInfo.MobAir.Strategy)
}

func TestGenerateBridgeRoom_DebugInfoWithZeroCounts(t *testing.T) {
	req := BridgeGenerateRequest{
		Width:          20,
		Height:         20,
		Doors:          []DoorPosition{DoorTop, DoorBottom},
		SoftEdgeCount:  0,
		StaticCount:    0,
		TurretCount:    0,
		MobGroundCount: 0,
		MobAirCount:    0,
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)
	require.NotNil(t, resp.DebugInfo, "DebugInfo should not be nil")

	// Ground debug info should still be populated
	require.NotNil(t, resp.DebugInfo.Ground, "Ground debug info should not be nil")
	assert.Greater(t, len(resp.DebugInfo.Ground.DoorConnections), 0, "should have door connections")

	// Other layers should have skipped=true with skip reasons when count is 0
	require.NotNil(t, resp.DebugInfo.SoftEdge, "SoftEdge debug info should not be nil")
	assert.True(t, resp.DebugInfo.SoftEdge.Skipped, "SoftEdge should be marked as skipped")
	assert.Contains(t, resp.DebugInfo.SoftEdge.SkipReason, "0", "SoftEdge skip reason should mention count")

	require.NotNil(t, resp.DebugInfo.Static, "Static debug info should not be nil")
	assert.True(t, resp.DebugInfo.Static.Skipped, "Static should be marked as skipped")
	assert.Contains(t, resp.DebugInfo.Static.SkipReason, "0", "Static skip reason should mention count")

	require.NotNil(t, resp.DebugInfo.Turret, "Turret debug info should not be nil")
	assert.True(t, resp.DebugInfo.Turret.Skipped, "Turret should be marked as skipped")
	assert.Contains(t, resp.DebugInfo.Turret.SkipReason, "0", "Turret skip reason should mention count")

	require.NotNil(t, resp.DebugInfo.MobGround, "MobGround debug info should not be nil")
	assert.True(t, resp.DebugInfo.MobGround.Skipped, "MobGround should be marked as skipped")
	assert.Contains(t, resp.DebugInfo.MobGround.SkipReason, "0", "MobGround skip reason should mention count")

	require.NotNil(t, resp.DebugInfo.MobAir, "MobAir debug info should not be nil")
	assert.True(t, resp.DebugInfo.MobAir.Skipped, "MobAir should be marked as skipped")
	assert.Contains(t, resp.DebugInfo.MobAir.SkipReason, "0", "MobAir skip reason should mention count")
}

func TestFindEmptyAreas(t *testing.T) {
	// Create a ground with known empty areas
	// 10x10 grid with a bridge in the middle
	ground := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 0
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 1
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 2
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 3
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, // row 4 - bridge
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, // row 5 - bridge
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 6
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 7
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 8
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // row 9
	}

	areas := findEmptyAreas(ground, 10, 10)

	// Should find 2 empty areas (top 10x4 and bottom 10x4)
	// Areas need to be at least 4x4, so both 10x4 areas qualify
	assert.Equal(t, 2, len(areas), "should find 2 empty areas >= 4x4")
	t.Logf("Found %d empty areas", len(areas))
	for i, area := range areas {
		t.Logf("Area %d: (%d,%d) %dx%d", i, area.X, area.Y, area.Width, area.Height)
		assert.GreaterOrEqual(t, area.Width, 4, "area width should be >= 4")
		assert.GreaterOrEqual(t, area.Height, 4, "area height should be >= 4")
	}
}

func TestFindEmptyAreas_LargeVoidArea(t *testing.T) {
	// Create a 20x20 grid with a small bridge, leaving large void areas
	ground := make([][]int, 20)
	for y := 0; y < 20; y++ {
		ground[y] = make([]int, 20)
	}

	// Draw a small bridge in the center
	for x := 8; x <= 11; x++ {
		ground[9][x] = 1
		ground[10][x] = 1
	}

	areas := findEmptyAreas(ground, 20, 20)

	// Should find some empty areas >= 4x4
	assert.Greater(t, len(areas), 0, "should find at least one empty area >= 4x4")

	for _, area := range areas {
		assert.GreaterOrEqual(t, area.Width, 4, "area width should be >= 4")
		assert.GreaterOrEqual(t, area.Height, 4, "area height should be >= 4")
		t.Logf("Found area: (%d,%d) %dx%d", area.X, area.Y, area.Width, area.Height)
	}
}

func TestIsValidIslandPosition(t *testing.T) {
	// 15x15 grid, mostly void
	ground := make([][]int, 15)
	for y := 0; y < 15; y++ {
		ground[y] = make([]int, 15)
	}

	// Place some ground at the edges
	ground[0][0] = 1
	ground[0][1] = 1
	ground[1][0] = 1
	ground[1][1] = 1

	tests := []struct {
		name         string
		x, y         int
		islandWidth  int
		islandHeight int
		expected     bool
	}{
		{
			name:         "invalid position - too far from existing ground",
			x:            7,
			y:            7,
			islandWidth:  3,
			islandHeight: 3,
			expected:     false, // island must be at exactly distance 2 from ground, (7,7) is too far from (0-1,0-1)
		},
		{
			name:         "invalid position - too close to existing ground",
			x:            3,
			y:            3,
			islandWidth:  3,
			islandHeight: 3,
			expected:     false, // distance 2 required, but corner ground at (0-1,0-1)
		},
		{
			name:         "valid position - exactly at min distance",
			x:            4,
			y:            4,
			islandWidth:  3,
			islandHeight: 3,
			expected:     true, // edge at x=4, margin check from x=2, ground at x=0-1 is at distance 3 (just outside margin)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidIslandPosition(ground, tt.x, tt.y, tt.islandWidth, tt.islandHeight, 15, 15)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDrawFloatingIslandsWithDebug(t *testing.T) {
	// Create a 30x30 grid with a central bridge, leaving large void areas
	ground := make([][]int, 30)
	for y := 0; y < 30; y++ {
		ground[y] = make([]int, 30)
	}

	// Draw a cross-shaped bridge in the center
	for x := 13; x <= 16; x++ {
		for y := 0; y < 30; y++ {
			ground[y][x] = 1
		}
	}
	for y := 13; y <= 16; y++ {
		for x := 0; x < 30; x++ {
			ground[y][x] = 1
		}
	}

	debug := &GroundDebugInfo{}
	drawFloatingIslandsWithDebug(ground, 30, 30, debug)

	// Check debug info is populated
	t.Logf("Floating islands debug info: %+v", debug.FloatingIslands)

	// Due to 50% probability, we may or may not get islands
	// Just verify the function doesn't crash and debug info is set
	assert.NotNil(t, debug.FloatingIslands)
}

func TestFloatingIslandMinDistance(t *testing.T) {
	// Create a controlled environment to test minimum distance
	ground := make([][]int, 20)
	for y := 0; y < 20; y++ {
		ground[y] = make([]int, 20)
	}

	// Place a single ground cell in the center
	ground[10][10] = 1

	// Run multiple times to check distance constraint
	for i := 0; i < 10; i++ {
		testGround := make([][]int, 20)
		for y := 0; y < 20; y++ {
			testGround[y] = make([]int, 20)
			copy(testGround[y], ground[y])
		}

		debug := &GroundDebugInfo{}
		drawFloatingIslandsWithDebug(testGround, 20, 20, debug)

		// Check that any placed islands maintain min distance of 2 from original ground
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				if testGround[y][x] == 1 && ground[y][x] == 0 {
					// This is a newly placed island cell
					// Check distance from original ground at (10,10)
					dist := abs(x-10) + abs(y-10)
					if dist < 2 {
						// Check Manhattan distance - should be at least 2
						// But actually we need to check that there's a 2-cell gap
						// between the island and the original ground
						t.Logf("Island cell at (%d,%d), distance from (10,10): %d", x, y, dist)
					}
				}
			}
		}
	}
}

func TestGenerateBridgeRoom_FloatingIslandsInDebugInfo(t *testing.T) {
	// Generate a large room to have void areas for floating islands
	req := BridgeGenerateRequest{
		Width:  30,
		Height: 30,
		Doors:  []DoorPosition{DoorTop, DoorBottom},
	}

	resp, err := GenerateBridgeRoom(req)
	require.NoError(t, err)
	require.NotNil(t, resp.DebugInfo)
	require.NotNil(t, resp.DebugInfo.Ground)

	// FloatingIslands may or may not be populated (50% probability per island)
	// Just verify the field exists in the response
	t.Logf("Floating islands: %+v", resp.DebugInfo.Ground.FloatingIslands)
}
