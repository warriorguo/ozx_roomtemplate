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
