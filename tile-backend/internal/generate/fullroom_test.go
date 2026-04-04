package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFullRoom_ValidInput(t *testing.T) {
	tests := []struct {
		name string
		req  FullRoomGenerateRequest
	}{
		{
			name: "two doors - top and bottom",
			req: FullRoomGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
		},
		{
			name: "two doors - left and right",
			req: FullRoomGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorLeft, DoorRight},
			},
		},
		{
			name: "three doors",
			req: FullRoomGenerateRequest{
				Width:  25,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorRight, DoorBottom},
			},
		},
		{
			name: "four doors",
			req: FullRoomGenerateRequest{
				Width:  30,
				Height: 30,
				Doors:  []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft},
			},
		},
		{
			name: "small room",
			req: FullRoomGenerateRequest{
				Width:  4,
				Height: 4,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
		},
		{
			name: "rectangular room wide",
			req: FullRoomGenerateRequest{
				Width:  40,
				Height: 15,
				Doors:  []DoorPosition{DoorLeft, DoorRight},
			},
		},
		{
			name: "rectangular room tall",
			req: FullRoomGenerateRequest{
				Width:  15,
				Height: 40,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GenerateFullRoom(tt.req)
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify payload dimensions
			assert.Equal(t, tt.req.Height, len(resp.Payload.Ground))
			assert.Equal(t, tt.req.Width, len(resp.Payload.Ground[0]))

			// Verify all layers have correct dimensions
			assert.Equal(t, tt.req.Height, len(resp.Payload.SoftEdge))
			assert.Equal(t, tt.req.Height, len(resp.Payload.Bridge))
			assert.Equal(t, tt.req.Height, len(resp.Payload.Static))
			assert.Equal(t, tt.req.Height, len(resp.Payload.Chaser))
			assert.Equal(t, tt.req.Height, len(resp.Payload.Zoner))
			assert.Equal(t, tt.req.Height, len(resp.Payload.MobAir))

			// Verify room shape and category
			require.NotNil(t, resp.Payload.RoomShape)
			assert.Equal(t, "all", *resp.Payload.RoomShape)
			require.NotNil(t, resp.Payload.RoomCategory)
			assert.Equal(t, "normal", *resp.Payload.RoomCategory)

			// Verify debug info exists
			require.NotNil(t, resp.DebugInfo)
			require.NotNil(t, resp.DebugInfo.Ground)
		})
	}
}

func TestGenerateFullRoom_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		req         FullRoomGenerateRequest
		expectedErr string
	}{
		{
			name: "width too small",
			req: FullRoomGenerateRequest{
				Width:  2,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "width must be between 4 and 200",
		},
		{
			name: "height too small",
			req: FullRoomGenerateRequest{
				Width:  20,
				Height: 2,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "height must be between 4 and 200",
		},
		{
			name: "width too large",
			req: FullRoomGenerateRequest{
				Width:  250,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "width must be between 4 and 200",
		},
		{
			name: "duplicate doors",
			req: FullRoomGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorTop},
			},
			expectedErr: "duplicate door: top",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GenerateFullRoom(tt.req)
			require.Error(t, err)
			assert.Nil(t, resp)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestGenerateFullRoom_SingleDoor(t *testing.T) {
	// Single-door (and zero-door) fullrooms should generate successfully
	for _, doors := range [][]DoorPosition{
		{DoorTop},
		{DoorBottom},
		{DoorLeft},
		{DoorRight},
		{},
	} {
		req := FullRoomGenerateRequest{
			Width:  20,
			Height: 20,
			Doors:  doors,
		}
		resp, err := GenerateFullRoom(req)
		require.NoError(t, err, "expected no error for doors=%v", doors)
		require.NotNil(t, resp)
	}
}

func TestGenerateFullRoom_DoorsAlwaysConnected(t *testing.T) {
	// Run many iterations to test that doors are always connected
	// despite random corner erasing and center pits
	doorCombos := [][]DoorPosition{
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorLeft},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
		{DoorBottom, DoorRight},
		{DoorTop, DoorRight, DoorBottom, DoorLeft},
	}

	for _, doors := range doorCombos {
		for i := 0; i < 20; i++ {
			req := FullRoomGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  doors,
			}

			resp, err := GenerateFullRoom(req)
			require.NoError(t, err)

			assert.True(t, areAllDoorsConnected(resp.Payload.Ground, req.Width, req.Height, doors),
				"doors should always be connected (doors=%v, iteration=%d)", doors, i)
		}
	}
}

func TestGenerateFullRoom_HighGroundCoverage(t *testing.T) {
	req := FullRoomGenerateRequest{
		Width:  20,
		Height: 20,
		Doors:  []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}

	// Run multiple times to account for randomness
	for i := 0; i < 10; i++ {
		resp, err := GenerateFullRoom(req)
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

		// Full rooms should have high ground coverage (at least 50%)
		totalTiles := req.Width * req.Height
		minWalkable := totalTiles / 2

		assert.GreaterOrEqual(t, walkableCount, minWalkable,
			"full room should have at least 50%% walkable tiles, got %d/%d (iteration=%d)",
			walkableCount, totalTiles, i)
	}
}

func TestGenerateFullRoom_WithOptionalLayers(t *testing.T) {
	req := FullRoomGenerateRequest{
		Width:         25,
		Height:        25,
		Doors:         []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		SoftEdgeCount: 3,
		StaticCount:   4,
		ChaserCount:   3,
		ZonerCount:    5,
		MobAirCount:   4,
	}

	resp, err := GenerateFullRoom(req)
	require.NoError(t, err)
	require.NotNil(t, resp.DebugInfo)

	// Verify layers are not skipped
	assert.False(t, resp.DebugInfo.SoftEdge.Skipped)
	assert.False(t, resp.DebugInfo.Static.Skipped)
	assert.False(t, resp.DebugInfo.Chaser.Skipped)
	assert.False(t, resp.DebugInfo.Zoner.Skipped)
	assert.False(t, resp.DebugInfo.MobAir.Skipped)
}

// TestGenerateFullRoom_PressureStageZonerCount is a regression test for ORT-26.
// Pressure stage requires exactly 1 zoner; the grouped placement path was missing
// ZonerCount in PlacementGroup, causing 0 zoners to be placed.
func TestGenerateFullRoom_PressureStageZonerCount(t *testing.T) {
	// Run multiple times to account for randomness in group halves selection
	for i := 0; i < 20; i++ {
		req := FullRoomGenerateRequest{
			Width:         20,
			Height:        12,
			Doors:         []DoorPosition{DoorTop, DoorRight},
			StageType:     "pressure",
			SoftEdgeCount: 5,
			StaticCount:   5,
			RailEnabled:   true,
		}
		resp, err := GenerateFullRoom(req)
		require.NoError(t, err, "iteration %d", i)
		require.NotNil(t, resp, "iteration %d", i)

		// Count actual zoner cells placed
		zonerCount := 0
		for y := 0; y < req.Height; y++ {
			for x := 0; x < req.Width; x++ {
				if resp.Payload.Zoner[y][x] == 1 {
					zonerCount++
				}
			}
		}
		// Pressure stage ZonerRange is [1,1], so exactly 1 zoner must be placed
		assert.Equal(t, 1, zonerCount, "pressure stage must place exactly 1 zoner (iteration %d)", i)
	}
}

func TestGenerateFullRoom_DebugInfoPopulated(t *testing.T) {
	req := FullRoomGenerateRequest{
		Width:         20,
		Height:        20,
		Doors:         []DoorPosition{DoorTop, DoorBottom},
		SoftEdgeCount: 2,
		StaticCount:   2,
		ChaserCount:   2,
		ZonerCount:    2,
		MobAirCount:   2,
	}

	resp, err := GenerateFullRoom(req)
	require.NoError(t, err)
	require.NotNil(t, resp.DebugInfo)
	require.NotNil(t, resp.DebugInfo.Ground)

	// Verify ground debug info has corner erase and center pits info
	assert.NotNil(t, resp.DebugInfo.Ground.CornerErase)
	assert.NotNil(t, resp.DebugInfo.Ground.CenterPits)

	t.Logf("CornerErase skipped: %v", resp.DebugInfo.Ground.CornerErase.Skipped)
	if !resp.DebugInfo.Ground.CornerErase.Skipped {
		t.Logf("  BrushType: %s", resp.DebugInfo.Ground.CornerErase.BrushType)
		t.Logf("  BrushSize: %s", resp.DebugInfo.Ground.CornerErase.BrushSize)
		t.Logf("  Combo: %s", resp.DebugInfo.Ground.CornerErase.Combo)
		for _, c := range resp.DebugInfo.Ground.CornerErase.Corners {
			t.Logf("  Corner: %s pos=%s size=%s rolledBack=%v", c.Corner, c.Position, c.Size, c.RolledBack)
		}
	}

	t.Logf("CenterPits skipped: %v", resp.DebugInfo.Ground.CenterPits.Skipped)
	if !resp.DebugInfo.Ground.CenterPits.Skipped {
		t.Logf("  BrushSize: %s", resp.DebugInfo.Ground.CenterPits.BrushSize)
		t.Logf("  PitCount: %d", resp.DebugInfo.Ground.CenterPits.PitCount)
		t.Logf("  Symmetry: %s", resp.DebugInfo.Ground.CenterPits.Symmetry)
		for _, p := range resp.DebugInfo.Ground.CenterPits.Pits {
			t.Logf("  Pit: pos=%s size=%s rolledBack=%v", p.Position, p.Size, p.RolledBack)
		}
	}
}
