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
