package generate

import (
	"testing"
)

// TestStaticOnGroundInvariant_Bridge verifies that all static cells in bridge room generation
// are always placed on ground cells, across many parameter combinations.
// Regression test for ORT-33: bridge room 4-door building stage produced static on non-ground.
func TestStaticOnGroundInvariant_Bridge(t *testing.T) {
	doorCombos := [][]DoorPosition{
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorRight},
		{DoorTop, DoorRight, DoorBottom},
		{DoorTop, DoorRight, DoorBottom, DoorLeft},
	}

	stageTypes := []string{"", "teaching", "building"}

	staticCounts := []int{0, 1, 3, 5}

	railOptions := []bool{false, true}

	totalCases := 0
	failures := 0

	for _, doors := range doorCombos {
		for _, stage := range stageTypes {
			for _, staticCount := range staticCounts {
				for _, railEnabled := range railOptions {
					// Run each combination multiple times
					for trial := 0; trial < 10; trial++ {
						totalCases++
						req := BridgeGenerateRequest{
							Width:         20,
							Height:        12,
							Doors:         doors,
							StageType:     stage,
							SoftEdgeCount: 0,
							StaticCount:   staticCount,
							RailEnabled:   railEnabled,
						}

						resp, err := GenerateBridgeRoom(req)
						if err != nil {
							continue
						}

						// Invariant: static[y][x]==1 => ground[y][x]==1
						for y := 0; y < req.Height; y++ {
							for x := 0; x < req.Width; x++ {
								if resp.Payload.Static[y][x] == 1 && resp.Payload.Ground[y][x] != 1 {
									failures++
									t.Errorf(
										"bridge: doors=%v stage=%q staticCount=%d rail=%v trial=%d: "+
											"static[%d][%d]=1 but ground=0",
										doors, stage, staticCount, railEnabled, trial, y, x,
									)
									if failures > 5 {
										t.FailNow()
									}
								}
							}
						}
					}
				}
			}
		}
	}

	t.Logf("Tested %d bridge room combinations, %d failures", totalCases, failures)
}

// TestStaticOnGroundInvariant_B4 is the exact B4 scenario from the test matrix
// that triggered ORT-33: bridge, 4 doors, building stage, staticCount=0, railEnabled=true.
func TestStaticOnGroundInvariant_B4(t *testing.T) {
	failures := 0
	for trial := 0; trial < 200; trial++ {
		req := BridgeGenerateRequest{
			Width:         20,
			Height:        12,
			Doors:         []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft},
			StageType:     "building",
			SoftEdgeCount: 0,
			StaticCount:   0,
			RailEnabled:   true,
		}

		resp, err := GenerateBridgeRoom(req)
		if err != nil {
			continue
		}

		for y := 0; y < req.Height; y++ {
			for x := 0; x < req.Width; x++ {
				if resp.Payload.Static[y][x] == 1 && resp.Payload.Ground[y][x] != 1 {
					failures++
					t.Errorf("B4 trial=%d: static[%d][%d]=1 but ground=0", trial, y, x)
					if failures > 5 {
						t.FailNow()
					}
				}
			}
		}
	}
}

// TestStaticOnGroundInvariant_FullRoom verifies the same invariant for fullroom generation.
func TestStaticOnGroundInvariant_FullRoom(t *testing.T) {
	doorCombos := [][]DoorPosition{
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorRight},
		{DoorTop, DoorRight, DoorBottom},
		{DoorTop, DoorRight, DoorBottom, DoorLeft},
	}

	stageTypes := []string{"", "teaching", "building", "release"}
	staticCounts := []int{0, 2, 5, 8}
	railOptions := []bool{false, true}

	totalCases := 0
	failures := 0

	for _, doors := range doorCombos {
		for _, stage := range stageTypes {
			for _, staticCount := range staticCounts {
				for _, railEnabled := range railOptions {
					for trial := 0; trial < 5; trial++ {
						totalCases++
						req := FullRoomGenerateRequest{
							Width:         20,
							Height:        12,
							Doors:         doors,
							StageType:     stage,
							SoftEdgeCount: 0,
							StaticCount:   staticCount,
							RailEnabled:   railEnabled,
						}

						resp, err := GenerateFullRoom(req)
						if err != nil {
							continue
						}

						for y := 0; y < req.Height; y++ {
							for x := 0; x < req.Width; x++ {
								if resp.Payload.Static[y][x] == 1 && resp.Payload.Ground[y][x] != 1 {
									failures++
									t.Errorf(
										"fullroom: doors=%v stage=%q staticCount=%d rail=%v trial=%d: "+
											"static[%d][%d]=1 but ground=0",
										doors, stage, staticCount, railEnabled, trial, y, x,
									)
									if failures > 5 {
										t.FailNow()
									}
								}
							}
						}
					}
				}
			}
		}
	}

	t.Logf("Tested %d fullroom combinations, %d failures", totalCases, failures)
}

// TestStaticOnGroundInvariant_Platform verifies the same invariant for platform room generation.
func TestStaticOnGroundInvariant_Platform(t *testing.T) {
	doorCombos := [][]DoorPosition{
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom},
		{DoorTop, DoorRight},
		{DoorTop, DoorRight, DoorBottom, DoorLeft},
		{DoorLeft},
	}

	stageTypes := []string{"", "teaching", "building"}
	staticCounts := []int{0, 2, 5}
	railOptions := []bool{false, true}

	totalCases := 0
	failures := 0

	for _, doors := range doorCombos {
		for _, stage := range stageTypes {
			for _, staticCount := range staticCounts {
				for _, railEnabled := range railOptions {
					for trial := 0; trial < 5; trial++ {
						totalCases++
						req := PlatformGenerateRequest{
							Width:         20,
							Height:        12,
							Doors:         doors,
							StageType:     stage,
							SoftEdgeCount: 0,
							StaticCount:   staticCount,
							RailEnabled:   railEnabled,
						}

						resp, err := GeneratePlatformRoom(req)
						if err != nil {
							continue
						}

						for y := 0; y < req.Height; y++ {
							for x := 0; x < req.Width; x++ {
								if resp.Payload.Static[y][x] == 1 && resp.Payload.Ground[y][x] != 1 {
									failures++
									t.Errorf(
										"platform: doors=%v stage=%q staticCount=%d rail=%v trial=%d: "+
											"static[%d][%d]=1 but ground=0",
										doors, stage, staticCount, railEnabled, trial, y, x,
									)
									if failures > 5 {
										t.FailNow()
									}
								}
							}
						}
					}
				}
			}
		}
	}

	t.Logf("Tested %d platform room combinations, %d failures", totalCases, failures)
}

// TestStaticPlacementFallback_P4 is a regression test for ORT-40.
// Platform generation with all 4 doors and staticCount=5 must always place at least 1 static
// block even when the standard door-forbidden-zone radius exhausts valid positions.
func TestStaticPlacementFallback_P4(t *testing.T) {
	zeroCount := 0
	trials := 100

	for trial := 0; trial < trials; trial++ {
		req := PlatformGenerateRequest{
			Width:         20,
			Height:        12,
			Doors:         []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft},
			StageType:     "teaching",
			SoftEdgeCount: 5,
			StaticCount:   5,
			RailEnabled:   false,
		}

		resp, err := GeneratePlatformRoom(req)
		if err != nil {
			// Generation error is not the bug we're testing; skip trial
			continue
		}

		// Count placed statics
		placed := 0
		for y := 0; y < req.Height; y++ {
			for x := 0; x < req.Width; x++ {
				if resp.Payload.Static[y][x] == 1 {
					placed++
				}
			}
		}
		// Static is 2x2, so each block = 4 cells; we just need at least one block (4 cells)
		if placed == 0 {
			zeroCount++
			t.Errorf("ORT-40 trial=%d: staticCount=5 but 0 static cells placed (door forbidden zones exhausted all positions)", trial)
		}
	}

	t.Logf("ORT-40: ran %d trials, %d produced zero statics (expected 0)", trials, zeroCount)
}

// TestDPSMinimum_BridgeBuilding verifies that bridge rooms with building stage
// always place the minimum DPS count (2). Regression test for ORT-34.
func TestDPSMinimum_BridgeBuilding(t *testing.T) {
	doorCombos := [][]DoorPosition{
		{DoorTop, DoorRight, DoorBottom, DoorLeft},
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorRight},
	}

	trials := 50
	failures := 0

	for _, doors := range doorCombos {
		for trial := 0; trial < trials; trial++ {
			req := BridgeGenerateRequest{
				Width:         20,
				Height:        12,
				Doors:         doors,
				StageType:     "building",
				SoftEdgeCount: 0,
				StaticCount:   0,
				RailEnabled:   true,
			}

			resp, err := GenerateBridgeRoom(req)
			if err != nil {
				continue
			}

			dpsCount := 0
			for y := 0; y < req.Height; y++ {
				for x := 0; x < req.Width; x++ {
					if resp.Payload.DPS[y][x] == 1 {
						dpsCount++
					}
				}
			}

			if dpsCount < 2 {
				failures++
				t.Errorf("ORT-34 doors=%v trial=%d: dps count=%d, expected >= 2", doors, trial, dpsCount)
			}
		}
	}

	t.Logf("ORT-34: tested %d combinations, %d failures", len(doorCombos)*trials, failures)
}
