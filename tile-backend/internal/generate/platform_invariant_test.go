package generate

import (
	"testing"
)

// TestPlatformGroundAlwaysConnected verifies the invariant:
// platform room generation must never produce isolated (disconnected) ground islands,
// across all door configurations and sizes, with and without railEnabled.
//
// Regression test for ORT-30: single-door platform layouts (e.g. doors=[left])
// produced isolated islands because areAllDoorsConnected returned true immediately
// for len(doors) < 2, allowing eraser operations to disconnect ground without rollback.
//
// Regression test for ORT-39: top/bottom doors with railEnabled=true intermittently
// produced disconnected ground.
func TestPlatformGroundAlwaysConnected(t *testing.T) {
	doorConfigs := [][]DoorPosition{
		{DoorLeft},
		{DoorRight},
		{DoorTop},
		{DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom},
		{DoorTop, DoorLeft},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
		{DoorBottom, DoorRight},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}

	stageTypes := []string{"", "teaching", "building"}

	sizes := [][2]int{
		{20, 12},
		{15, 15},
		{20, 20},
	}

	failures := 0

	for _, doors := range doorConfigs {
		for _, stage := range stageTypes {
			for _, size := range sizes {
				w, h := size[0], size[1]
				for trial := 0; trial < 30; trial++ {
					req := PlatformGenerateRequest{
						Width:         w,
						Height:        h,
						Doors:         doors,
						StageType:     stage,
						SoftEdgeCount: 3,
						StaticCount:   3,
						RailEnabled:   trial%2 == 0,
					}

					resp, err := GeneratePlatformRoom(req)
					if err != nil {
						continue
					}

					islands := findAllIslands(resp.Payload.Ground, w, h)
					if len(islands) > 1 {
						failures++
						t.Errorf("doors=%v stage=%q size=%dx%d trial=%d railEnabled=%v: %d disconnected ground islands (expected 1)",
							doors, stage, w, h, trial, req.RailEnabled, len(islands))
						if failures > 5 {
							t.FailNow()
						}
					}
				}
			}
		}
	}

	t.Logf("Platform ground connectivity: tested all door/stage/size combinations, %d failures", failures)
}

// TestPlatformNoBridgeTiles verifies the invariant:
// platform room generation must never produce bridge tiles (bridgeCount == 0 always).
//
// Regression test for ORT-39: the platform generator called generateBridgeLayerWithDebug
// which triggers a "force at least one bridge" fallback, placing spurious bridge tiles
// even when the ground is fully connected and no bridges are semantically needed.
func TestPlatformNoBridgeTiles(t *testing.T) {
	doorConfigs := [][]DoorPosition{
		{DoorLeft},
		{DoorRight},
		{DoorTop},
		{DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom},
		{DoorTop, DoorLeft},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
		{DoorBottom, DoorRight},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}

	stageTypes := []string{"", "teaching", "building"}

	sizes := [][2]int{
		{20, 12},
		{15, 15},
		{20, 20},
	}

	failures := 0

	for _, doors := range doorConfigs {
		for _, stage := range stageTypes {
			for _, size := range sizes {
				w, h := size[0], size[1]
				for trial := 0; trial < 20; trial++ {
					req := PlatformGenerateRequest{
						Width:         w,
						Height:        h,
						Doors:         doors,
						StageType:     stage,
						SoftEdgeCount: 3,
						StaticCount:   3,
						RailEnabled:   trial%2 == 0,
					}

					resp, err := GeneratePlatformRoom(req)
					if err != nil {
						continue
					}

					bridgeCount := countCells(resp.Payload.Bridge)
					if bridgeCount > 0 {
						failures++
						t.Errorf("doors=%v stage=%q size=%dx%d trial=%d railEnabled=%v: bridgeCount=%d (expected 0)",
							doors, stage, w, h, trial, req.RailEnabled, bridgeCount)
						if failures > 5 {
							t.FailNow()
						}
					}
				}
			}
		}
	}

	t.Logf("Platform no-bridge invariant: tested all door/stage/size combinations, %d failures", failures)
}
