package generate

import (
	"fmt"
	"testing"
)

var allFourDoors = []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}

// ORT-144: solidGround skips the corner erase and centre pits, so every cell
// of the room is ground and there is nowhere for a soft edge to go.
func TestSolidGroundLeavesNoVoid(t *testing.T) {
	for _, stage := range []string{"teaching", "building", "release"} { // the airiest carve rates
		for i := 0; i < 10; i++ {
			res, err := GenerateFullRoom(FullRoomGenerateRequest{
				Doors: allFourDoors, StageType: stage, SolidGround: true, SoftEdgeCount: 3, RailEnabled: true,
			})
			if err != nil {
				t.Fatalf("%s: %v", stage, err)
			}
			for y, row := range res.Payload.Ground {
				for x, v := range row {
					if v != 1 {
						t.Fatalf("%s: void cell at (%d,%d)", stage, x, y)
					}
				}
			}
			if n := cellCount(res.Payload.SoftEdge); n != 0 {
				t.Fatalf("%s: %d soft edge cells on a solid floor", stage, n)
			}
		}
	}
}

// staticScale / zonerScale multiply both ends of the stage's range before the
// count is drawn.
func TestCountScaleDoublesStageRanges(t *testing.T) {
	ground := make([][]int, DefaultRoomHeight)
	for y := range ground {
		ground[y] = make([]int, DefaultRoomWidth)
		for x := range ground[y] {
			ground[y][x] = 1
		}
	}
	for _, stage := range []string{"teaching", "pressure", "release"} {
		cfg := GetStageConfig(stage)
		for i := 0; i < 50; i++ {
			res, err := ValidateAndApplyStageScaled(stage, "full", allFourDoors, ground,
				DefaultRoomWidth, DefaultRoomHeight, CountScale{Static: 2, Zoner: 2})
			if err != nil {
				t.Fatal(err)
			}
			if res.StaticCount < 2*cfg.StaticRange[0] || res.StaticCount > 2*cfg.StaticRange[1] {
				t.Fatalf("%s: static %d outside doubled range %v", stage, res.StaticCount, cfg.StaticRange)
			}
			if res.ZonerCount < 2*cfg.ZonerRange[0] || res.ZonerCount > 2*cfg.ZonerRange[1] {
				t.Fatalf("%s: zoner %d outside doubled range %v", stage, res.ZonerCount, cfg.ZonerRange)
			}
		}
	}
	// No scale means the stage's own range, exactly as before.
	if r := scaleRange([2]int{4, 9}, 0); r != [2]int{4, 9} {
		t.Fatalf("scale 0 changed the range to %v", r)
	}
}

// A solid floor has a single interior rectangle, so before per-side insets every
// room got one of two tracks; and before jittered dispersion, ten-plus static
// blocks converged on a handful of packings. Both showed up as identical rooms
// in the ORT-144 basement batch.
func TestSolidGroundRoomsVary(t *testing.T) {
	rails, statics := map[string]bool{}, map[string]bool{}
	for i := 0; i < 30; i++ {
		res, err := GenerateFullRoom(FullRoomGenerateRequest{
			Doors: allFourDoors, StageType: "teaching", SolidGround: true, RailEnabled: true,
			StaticScale: 2, ZonerScale: 2,
		})
		if err != nil {
			t.Fatal(err)
		}
		rails[fmt.Sprint(res.Payload.Rail)] = true
		statics[fmt.Sprint(res.Payload.Static)] = true
	}
	if len(rails) < 6 {
		t.Errorf("only %d distinct rail layouts in 30 solid rooms", len(rails))
	}
	if len(statics) < 25 {
		t.Errorf("only %d distinct static layouts in 30 solid rooms", len(statics))
	}
}
