package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePlatformRoom_ValidInput(t *testing.T) {
	tests := []struct {
		name string
		req  PlatformGenerateRequest
	}{
		{
			name: "two doors - top and bottom",
			req: PlatformGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
		},
		{
			name: "two doors - left and right",
			req: PlatformGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorLeft, DoorRight},
			},
		},
		{
			name: "three doors",
			req: PlatformGenerateRequest{
				Width:  25,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorRight, DoorBottom},
			},
		},
		{
			name: "four doors",
			req: PlatformGenerateRequest{
				Width:  30,
				Height: 30,
				Doors:  []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft},
			},
		},
		{
			name: "corner group doors - top left",
			req: PlatformGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorLeft},
			},
		},
		{
			name: "corner group doors - bottom right",
			req: PlatformGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorBottom, DoorRight},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GeneratePlatformRoom(tt.req)
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

			// Verify room type
			require.NotNil(t, resp.Payload.RoomType)
			assert.Equal(t, "platform", *resp.Payload.RoomType)

			// Verify debug info exists
			require.NotNil(t, resp.DebugInfo)
			require.NotNil(t, resp.DebugInfo.Ground)
		})
	}
}

func TestGeneratePlatformRoom_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		req         PlatformGenerateRequest
		expectedErr string
	}{
		{
			name: "width too small",
			req: PlatformGenerateRequest{
				Width:  5,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "width must be between 10 and 200",
		},
		{
			name: "height too small",
			req: PlatformGenerateRequest{
				Width:  20,
				Height: 5,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "height must be between 10 and 200",
		},
		{
			name: "width too large",
			req: PlatformGenerateRequest{
				Width:  250,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorBottom},
			},
			expectedErr: "width must be between 10 and 200",
		},
		{
			name: "duplicate doors",
			req: PlatformGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  []DoorPosition{DoorTop, DoorTop},
			},
			expectedErr: "duplicate door: top",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GeneratePlatformRoom(tt.req)
			require.Error(t, err)
			assert.Nil(t, resp)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestGeneratePlatformRoom_SingleDoor(t *testing.T) {
	// Single-door (and zero-door) platform rooms should generate successfully
	for _, doors := range [][]DoorPosition{
		{DoorTop},
		{DoorBottom},
		{DoorLeft},
		{DoorRight},
		{},
	} {
		req := PlatformGenerateRequest{
			Width:  20,
			Height: 20,
			Doors:  doors,
		}
		resp, err := GeneratePlatformRoom(req)
		require.NoError(t, err, "expected no error for doors=%v", doors)
		require.NotNil(t, resp)
	}
}

func TestGeneratePlatformRoom_DoorsConnected(t *testing.T) {
	// Test that all doors are connected via walkable ground
	tests := []struct {
		name  string
		doors []DoorPosition
	}{
		{"top-bottom", []DoorPosition{DoorTop, DoorBottom}},
		{"left-right", []DoorPosition{DoorLeft, DoorRight}},
		{"top-left", []DoorPosition{DoorTop, DoorLeft}},
		{"top-right", []DoorPosition{DoorTop, DoorRight}},
		{"bottom-left", []DoorPosition{DoorBottom, DoorLeft}},
		{"bottom-right", []DoorPosition{DoorBottom, DoorRight}},
		{"all doors", []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := PlatformGenerateRequest{
				Width:  20,
				Height: 20,
				Doors:  tt.doors,
			}

			resp, err := GeneratePlatformRoom(req)
			require.NoError(t, err)

			// Verify doors are connected
			assert.True(t, areAllDoorsConnected(resp.Payload.Ground, req.Width, req.Height, tt.doors),
				"all doors should be connected")
		})
	}
}

func TestGeneratePlatformRoom_GroundHasWalkableTiles(t *testing.T) {
	req := PlatformGenerateRequest{
		Width:  20,
		Height: 20,
		Doors:  []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}

	resp, err := GeneratePlatformRoom(req)
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

	// Platform rooms should have significant ground coverage (at least 25%)
	totalTiles := req.Width * req.Height
	minWalkable := totalTiles / 4 // At least 25%

	assert.GreaterOrEqual(t, walkableCount, minWalkable,
		"platform room should have at least 25%% walkable tiles, got %d/%d",
		walkableCount, totalTiles)
}

func TestGeneratePlatformRoom_WithOptionalLayers(t *testing.T) {
	req := PlatformGenerateRequest{
		Width:         25,
		Height:        25,
		Doors:         []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
		SoftEdgeCount: 3,
		StaticCount:   4,
		ChaserCount:   3,
		ZonerCount:    5,
		MobAirCount:   4,
	}

	resp, err := GeneratePlatformRoom(req)
	require.NoError(t, err)
	require.NotNil(t, resp.DebugInfo)

	// Verify layers are not skipped
	assert.False(t, resp.DebugInfo.SoftEdge.Skipped)
	assert.False(t, resp.DebugInfo.Static.Skipped)
	assert.False(t, resp.DebugInfo.Chaser.Skipped)
	assert.False(t, resp.DebugInfo.Zoner.Skipped)
	assert.False(t, resp.DebugInfo.MobAir.Skipped)
}

func TestCanGroupDoorsIntoCorners(t *testing.T) {
	tests := []struct {
		name     string
		doors    []DoorPosition
		expected bool
	}{
		{"top-left", []DoorPosition{DoorTop, DoorLeft}, true},
		{"top-right", []DoorPosition{DoorTop, DoorRight}, true},
		{"bottom-left", []DoorPosition{DoorBottom, DoorLeft}, true},
		{"bottom-right", []DoorPosition{DoorBottom, DoorRight}, true},
		{"top-bottom only", []DoorPosition{DoorTop, DoorBottom}, false},
		{"left-right only", []DoorPosition{DoorLeft, DoorRight}, false},
		{"all doors", []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}, true},
		{"three doors with corner", []DoorPosition{DoorTop, DoorLeft, DoorBottom}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canGroupDoorsIntoCorners(tt.doors)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePlatformRoom_DebugInfoPopulated(t *testing.T) {
	req := PlatformGenerateRequest{
		Width:         20,
		Height:        20,
		Doors:         []DoorPosition{DoorTop, DoorBottom},
		SoftEdgeCount: 2,
		StaticCount:   2,
		ChaserCount:   2,
		ZonerCount:    2,
		MobAirCount:   2,
	}

	resp, err := GeneratePlatformRoom(req)
	require.NoError(t, err)
	require.NotNil(t, resp.DebugInfo)
	require.NotNil(t, resp.DebugInfo.Ground)

	// Verify ground debug info
	assert.NotEmpty(t, resp.DebugInfo.Ground.Strategy)
	assert.NotEmpty(t, resp.DebugInfo.Ground.Platforms)

	t.Logf("Strategy: %s", resp.DebugInfo.Ground.Strategy)
	t.Logf("Platforms: %d", len(resp.DebugInfo.Ground.Platforms))
	t.Logf("Door connections: %d", len(resp.DebugInfo.Ground.DoorConnections))
	t.Logf("Eraser ops: %d", len(resp.DebugInfo.Ground.EraserOps))

	for _, p := range resp.DebugInfo.Ground.Platforms {
		t.Logf("  Platform: pos=%s size=%s group=%s", p.Position, p.Size, p.Group)
	}

	for _, e := range resp.DebugInfo.Ground.EraserOps {
		t.Logf("  Eraser: method=%s pos=%s size=%s rolledBack=%v", e.Method, e.Position, e.Size, e.RolledBack)
	}
}

func TestAreAllDoorsConnected(t *testing.T) {
	tests := []struct {
		name     string
		ground   [][]int
		doors    []DoorPosition
		expected bool
	}{
		{
			name: "all connected",
			ground: [][]int{
				{0, 0, 1, 1, 1, 0, 0},
				{0, 0, 1, 1, 1, 0, 0},
				{1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1},
				{1, 1, 1, 1, 1, 1, 1},
				{0, 0, 1, 1, 1, 0, 0},
				{0, 0, 1, 1, 1, 0, 0},
			},
			doors:    []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
			expected: true,
		},
		{
			name: "disconnected - vertical split",
			ground: [][]int{
				{1, 1, 0, 0, 0, 1, 1},
				{1, 1, 0, 0, 0, 1, 1},
				{1, 1, 0, 0, 0, 1, 1},
				{1, 1, 0, 0, 0, 1, 1},
				{1, 1, 0, 0, 0, 1, 1},
				{1, 1, 0, 0, 0, 1, 1},
				{1, 1, 0, 0, 0, 1, 1},
			},
			doors:    []DoorPosition{DoorLeft, DoorRight},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := len(tt.ground[0])
			height := len(tt.ground)
			result := areAllDoorsConnected(tt.ground, width, height, tt.doors)
			assert.Equal(t, tt.expected, result)
		})
	}
}
