package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStageRangeConfig verifies that stageConfigs contains the correct ranges
// per the design specification.
func TestStageRangeConfig(t *testing.T) {
	tests := []struct {
		stage     string
		dpsMin    int
		dpsMax    int
		chaserMin int
		chaserMax int
		zonerMin  int
		zonerMax  int
		mobAirMin int
		mobAirMax int
	}{
		{"teaching", 2, 3, 0, 0, 0, 0, 0, 0},
		{"building", 2, 3, 2, 3, 0, 0, 0, 0},
		{"pressure", 4, 6, 6, 8, 1, 1, 2, 4},
		{"peak", 6, 12, 6, 8, 2, 3, 2, 4},
		{"release", 0, 2, 0, 0, 0, 0, 0, 0},
		{"boss", 0, 0, 0, 0, 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.stage, func(t *testing.T) {
			cfg := GetStageConfig(tt.stage)
			if cfg == nil {
				t.Fatalf("stage %q not found in stageConfigs", tt.stage)
			}
			assert.Equal(t, tt.dpsMin, cfg.DPSRange[0], "dps min for stage %q", tt.stage)
			assert.Equal(t, tt.dpsMax, cfg.DPSRange[1], "dps max for stage %q", tt.stage)
			assert.Equal(t, tt.chaserMin, cfg.ChaserRange[0], "chaser min for stage %q", tt.stage)
			assert.Equal(t, tt.chaserMax, cfg.ChaserRange[1], "chaser max for stage %q", tt.stage)
			assert.Equal(t, tt.zonerMin, cfg.ZonerRange[0], "zoner min for stage %q", tt.stage)
			assert.Equal(t, tt.zonerMax, cfg.ZonerRange[1], "zoner max for stage %q", tt.stage)
			assert.Equal(t, tt.mobAirMin, cfg.MobAirRange[0], "mobair min for stage %q", tt.stage)
			assert.Equal(t, tt.mobAirMax, cfg.MobAirRange[1], "mobair max for stage %q", tt.stage)
		})
	}
}

// TestReleaseStage_DPSCountInRange verifies the release stage DPS count is in [0,2].
// Regression test for ORT-28: generator was producing 3 DPS units for release stage.
func TestReleaseStage_DPSCountInRange(t *testing.T) {
	failures := 0
	for trial := 0; trial < 200; trial++ {
		req := FullRoomGenerateRequest{
			Width:         20,
			Height:        12,
			Doors:         []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft},
			StageType:     "release",
			SoftEdgeCount: 5,
			StaticCount:   8,
			RailEnabled:   true,
		}

		resp, err := GenerateFullRoom(req)
		if err != nil {
			continue
		}

		// Count DPS cells
		dpsCount := 0
		for y := 0; y < req.Height; y++ {
			for x := 0; x < req.Width; x++ {
				dpsCount += resp.Payload.DPS[y][x]
			}
		}

		if dpsCount > 2 {
			failures++
			t.Errorf("trial=%d: release stage dps count=%d, expected [0,2]", trial, dpsCount)
			if failures > 5 {
				t.FailNow()
			}
		}
	}

	if failures == 0 {
		t.Logf("All 200 trials produced release stage DPS count in [0,2]")
	}
}

// TestPressureStage_ChaserCountInRange verifies that pressure stage always places at
// least 6 chasers (min of [6,8]) even when grouped placement exhausts valid positions
// in one half-room region.
//
// Regression test for ORT-38: generator produced 5 chasers for pressure stage because
// the grouped placement split 6 chasers across two half-regions; when one region had
// insufficient valid positions the total fell below the minimum.
func TestPressureStage_ChaserCountInRange(t *testing.T) {
	doorConfigs := [][]DoorPosition{
		{DoorTop, DoorRight},
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorRight, DoorBottom},
		{DoorTop, DoorRight, DoorBottom, DoorLeft},
	}
	failures := 0
	for trial := 0; trial < 300; trial++ {
		doors := doorConfigs[trial%len(doorConfigs)]
		req := FullRoomGenerateRequest{
			Width:         20,
			Height:        12,
			Doors:         doors,
			StageType:     "pressure",
			SoftEdgeCount: 5,
			StaticCount:   5,
			RailEnabled:   trial%2 == 0,
		}
		resp, err := GenerateFullRoom(req)
		if err != nil {
			continue
		}
		chaserCount := countCells(resp.Payload.Chaser)
		cfg := GetStageConfig("pressure")
		if chaserCount < cfg.ChaserRange[0] {
			failures++
			t.Errorf("trial=%d (doors=%v): pressure stage chaser count=%d, expected [%d,%d]",
				trial, doors, chaserCount, cfg.ChaserRange[0], cfg.ChaserRange[1])
			if failures > 5 {
				t.FailNow()
			}
		}
	}
	if failures == 0 {
		t.Logf("All 300 trials produced pressure stage chaser count >= %d", GetStageConfig("pressure").ChaserRange[0])
	}
}

// TestReleaseStage_OnlyDPSEnemies verifies release stage only spawns DPS enemies (no chaser, zoner, mobAir).
func TestReleaseStage_OnlyDPSEnemies(t *testing.T) {
	for trial := 0; trial < 50; trial++ {
		req := FullRoomGenerateRequest{
			Width:     20,
			Height:    12,
			Doors:     []DoorPosition{DoorTop, DoorBottom},
			StageType: "release",
		}

		resp, err := GenerateFullRoom(req)
		if err != nil {
			continue
		}

		// Count non-DPS enemy cells
		for y := 0; y < req.Height; y++ {
			for x := 0; x < req.Width; x++ {
				assert.Equal(t, 0, resp.Payload.Chaser[y][x], "trial=%d: release stage should have no chaser at (%d,%d)", trial, x, y)
				assert.Equal(t, 0, resp.Payload.Zoner[y][x], "trial=%d: release stage should have no zoner at (%d,%d)", trial, x, y)
				assert.Equal(t, 0, resp.Payload.MobAir[y][x], "trial=%d: release stage should have no mobAir at (%d,%d)", trial, x, y)
			}
		}
	}
}

// TestStageMinRoomSize verifies the per-stage minimum room dimensions (ORT-102).
// A room smaller than the minimum must be rejected before any layer is built,
// rather than silently falling back to relaxed placement that violates the
// 8-directional spacing constraint (ORT-93).
func TestStageMinRoomSize(t *testing.T) {
	t.Run("configured minimums", func(t *testing.T) {
		tests := []struct {
			stage             string
			minWidth, minHght int
		}{
			{"start", 0, 0},
			{"teaching", 0, 0},
			{"building", 0, 0},
			{"pressure", 18, 10},
			{"peak", 20, 12},
			{"release", 0, 0},
			{"boss", 0, 0},
		}
		for _, tt := range tests {
			cfg := GetStageConfig(tt.stage)
			if cfg == nil {
				t.Fatalf("stage %q not found", tt.stage)
			}
			assert.Equal(t, tt.minWidth, cfg.MinWidth, "min width for stage %q", tt.stage)
			assert.Equal(t, tt.minHght, cfg.MinHeight, "min height for stage %q", tt.stage)
		}
	})

	// Doors that satisfy every stage's door restrictions used below.
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	t.Run("undersized rooms are rejected", func(t *testing.T) {
		tests := []struct {
			stage         string
			roomType      string
			width, height int
			wantIn        string
		}{
			{"pressure", "full", 16, 8, "at least 18x10"},
			{"pressure", "full", 17, 10, "at least 18x10"}, // width short
			{"pressure", "full", 18, 9, "at least 18x10"},  // height short
			{"peak", "full", 16, 8, "at least 20x12"},
			{"peak", "full", 18, 12, "at least 20x12"}, // width short
			{"peak", "full", 20, 11, "at least 20x12"}, // height short
		}
		for _, tt := range tests {
			ground := makeFullGround(tt.width, tt.height)
			_, err := ValidateAndApplyStage(tt.stage, tt.roomType, doors, ground, tt.width, tt.height)
			if assert.Error(t, err, "%s %dx%d should be rejected", tt.stage, tt.width, tt.height) {
				assert.Contains(t, err.Error(), tt.wantIn)
				assert.Contains(t, err.Error(), "room size")
			}
		}
	})

	t.Run("rooms at or above the minimum are accepted", func(t *testing.T) {
		tests := []struct {
			stage         string
			width, height int
		}{
			{"pressure", 18, 10}, // exactly at the minimum
			{"pressure", 20, 12},
			{"peak", 20, 12}, // exactly at the minimum
			{"peak", 24, 16},
		}
		for _, tt := range tests {
			ground := makeFullGround(tt.width, tt.height)
			res, err := ValidateAndApplyStage(tt.stage, "full", doors, ground, tt.width, tt.height)
			if assert.NoError(t, err, "%s %dx%d should be accepted", tt.stage, tt.width, tt.height) {
				assert.True(t, res.Valid)
			}
		}
	})

	t.Run("unconstrained stages accept small rooms", func(t *testing.T) {
		for _, stage := range []string{"teaching", "building", "release"} {
			ground := makeFullGround(16, 8)
			res, err := ValidateAndApplyStage(stage, "full", doors, ground, 16, 8)
			if assert.NoError(t, err, "stage %q should accept 16x8", stage) {
				assert.True(t, res.Valid)
			}
		}
	})

	t.Run("empty stage type skips the check", func(t *testing.T) {
		ground := makeFullGround(4, 4)
		res, err := ValidateAndApplyStage("", "full", doors, ground, 4, 4)
		assert.NoError(t, err)
		assert.True(t, res.Valid)
	})
}

// makeFullGround returns a fully-walkable ground layer of the given size.
func makeFullGround(width, height int) [][]int {
	g := make([][]int, height)
	for y := range g {
		g[y] = make([]int, width)
		for x := range g[y] {
			g[y][x] = 1
		}
	}
	return g
}
