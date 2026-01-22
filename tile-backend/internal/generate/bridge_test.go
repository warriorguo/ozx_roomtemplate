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
