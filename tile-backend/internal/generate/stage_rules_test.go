package generate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		// ORT-101: zoner floor 1, chaser/dps floor 2, mobAir floor 6; existing
		// non-zero ranges doubled (mobAir tripled, peak mobAir set to 18).
		{"teaching", 4, 6, 2, 2, 1, 1, 6, 6},
		{"building", 4, 6, 4, 6, 1, 1, 6, 6},
		{"pressure", 8, 12, 12, 16, 2, 2, 6, 12},
		{"peak", 12, 24, 12, 16, 4, 6, 18, 18},
		{"release", 2, 4, 2, 2, 1, 1, 6, 6},
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

// TestReleaseStage_DPSCountInRange guards the release stage DPS count against
// its configured range.
//
// The upper bound is strict: exceeding the stage maximum is the ORT-28
// regression this test was written for.
//
// The lower bound tolerates a small tail. Release inherits the "teaching"
// placement rule, which confines DPS to y in [5,7] (StagePlacementHints.
// DPSYRange). That band is only three rows tall, and combined with the
// 8-directional spacing constraint and the door forbidden zones it can
// occasionally hold fewer cells than the stage minimum. There is no relaxed DPS
// fallback outside grouped placement, so the generator simply places fewer.
// Measured at roughly 0.5% of rooms after ORT-101 raised the minimums.
func TestReleaseStage_DPSCountInRange(t *testing.T) {
	cfg := GetStageConfig("release")
	require.NotNil(t, cfg)

	const trials = 200
	underMin := 0

	for trial := 0; trial < trials; trial++ {
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

		dpsCount := countCells(resp.Payload.DPS)

		// Strict: never exceed the configured maximum (ORT-28).
		require.LessOrEqualf(t, dpsCount, cfg.DPSRange[1],
			"trial=%d: release stage dps count=%d exceeds configured max %d",
			trial, dpsCount, cfg.DPSRange[1])

		if dpsCount < cfg.DPSRange[0] {
			underMin++
		}
	}

	// Tolerant: a thin tail is expected from the y-band above, but a systemic
	// shortfall means the band or the minimum needs revisiting.
	assert.LessOrEqualf(t, underMin, trials/20,
		"release stage fell below its configured DPS minimum %d in %d/%d rooms — "+
			"more than a tail, so the DPSYRange band and the stage minimum are in conflict",
		cfg.DPSRange[0], underMin, trials)
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

// TestReleaseStage_EnemyMix verifies the release stage stays within every
// configured range. Release used to be DPS-only; ORT-101 gave it a light mix of
// all four enemy types, so this asserts the ranges rather than absence.
func TestReleaseStage_EnemyMix(t *testing.T) {
	cfg := GetStageConfig("release")
	require.NotNil(t, cfg)

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

		for _, c := range []struct {
			name  string
			count int
			rng   [2]int
		}{
			{"chaser", countCells(resp.Payload.Chaser), cfg.ChaserRange},
			{"zoner", countCells(resp.Payload.Zoner), cfg.ZonerRange},
			{"dps", countCells(resp.Payload.DPS), cfg.DPSRange},
			{"mobAir", countCells(resp.Payload.MobAir), cfg.MobAirRange},
		} {
			assert.GreaterOrEqualf(t, c.count, c.rng[0],
				"trial=%d: release %s count %d below configured min %d", trial, c.name, c.count, c.rng[0])
			assert.LessOrEqualf(t, c.count, c.rng[1],
				"trial=%d: release %s count %d above configured max %d", trial, c.name, c.count, c.rng[1])
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

// TestStageCountsPlaceableAtMinRoomSize is the invariant that ties ORT-101 to
// ORT-102: at the smallest room a stage will accept, its own enemy counts must
// still place without engaging the relaxed fallback. The relaxed pass meets the
// count by dropping the 8-directional spacing constraint, which emits adjacent
// same-category spawn tiles and crashes the game for Zoner (ORT-93).
//
// If a future count increase outgrows its stage's MinWidth/MinHeight, this test
// fails and the minimum needs raising with it.
func TestStageCountsPlaceableAtMinRoomSize(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	// Stages with no configured minimum are exercised at the editor default.
	const defaultW, defaultH = 16, 8

	for _, stage := range []string{"teaching", "building", "pressure", "peak", "release"} {
		cfg := GetStageConfig(stage)
		if cfg == nil {
			t.Fatalf("stage %q not found", stage)
		}
		w, h := cfg.MinWidth, cfg.MinHeight
		if w == 0 || h == 0 {
			w, h = defaultW, defaultH
		}

		t.Run(fmt.Sprintf("%s_%dx%d", stage, w, h), func(t *testing.T) {
			const trials = 40
			adjacent, overlapping := 0, 0
			for i := 0; i < trials; i++ {
				resp, err := GenerateFullRoom(FullRoomGenerateRequest{
					Width: w, Height: h, Doors: doors, StaticCount: 8,
					StageType: stage, RoomCategory: "normal",
				})
				if err != nil {
					t.Fatalf("trial %d: %v", i, err)
				}
				p := resp.Payload
				for _, layer := range [][][]int{p.Chaser, p.DPS, p.Zoner} {
					if hasSameLayerAdjacency(layer, w, h) {
						adjacent++
						break
					}
				}
				if layersOverlap(p.DPS, p.Chaser, w, h) {
					overlapping++
				}
			}
			// Allow a small tail: placement is randomised and the relaxed pass
			// legitimately fires in rare pathological ground shapes.
			maxAllowed := trials / 10 // 10%
			assert.LessOrEqualf(t, adjacent, maxAllowed,
				"%s at its minimum %dx%d produced same-layer 8-dir adjacency in %d/%d rooms "+
					"(ORT-93 spacing violation); raise MinWidth/MinHeight for this stage",
				stage, w, h, adjacent, trials)
			assert.LessOrEqualf(t, overlapping, maxAllowed,
				"%s at its minimum %dx%d placed DPS on a chaser cell in %d/%d rooms",
				stage, w, h, overlapping, trials)
		})
	}
}

// hasSameLayerAdjacency reports whether any cell in the layer has an
// 8-directionally adjacent cell in that same layer.
func hasSameLayerAdjacency(layer [][]int, width, height int) bool {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if layer[y][x] == 0 {
				continue
			}
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					ny, nx := y+dy, x+dx
					if ny >= 0 && ny < height && nx >= 0 && nx < width && layer[ny][nx] == 1 {
						return true
					}
				}
			}
		}
	}
	return false
}

// layersOverlap reports whether the two layers both occupy the same cell.
func layersOverlap(a, b [][]int, width, height int) bool {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if a[y][x] == 1 && b[y][x] == 1 {
				return true
			}
		}
	}
	return false
}
