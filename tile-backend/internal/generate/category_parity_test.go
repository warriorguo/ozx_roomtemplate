package generate

import (
	"testing"

	"tile-backend/internal/model"
)

// ORT-142: basement is a category — the folder OZX files a room under — not a
// different set of generation rules. A basement room is generated at the same
// default size and with the same spawn layers as a normal room of the same
// stage; roomCategory is only stamped onto the output. The old 20x12 basement
// pressure rooms with no chaser/mobAir cells were the special case this rules
// out.
func TestBasementGeneratesLikeNormal(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}
	for _, stage := range []string{"teaching", "building", "pressure", "peak", "release"} {
		t.Run(stage, func(t *testing.T) {
			cfg := GetStageConfig(stage)
			placed := map[string]int{}
			for i := 0; i < 10; i++ {
				res, err := GenerateFullRoom(FullRoomGenerateRequest{
					Doors: doors, StageType: stage, RoomCategory: "basement",
				})
				if err != nil {
					t.Fatalf("generate: %v", err)
				}
				p := res.Payload
				if p.RoomCategory == nil || *p.RoomCategory != "basement" {
					t.Fatalf("roomCategory = %v, want basement", p.RoomCategory)
				}
				assertDefaultSized(t, p.Ground, p.Meta.Width, p.Meta.Height)
				placed["chaser"] += cellCount(p.Chaser)
				placed["zoner"] += cellCount(p.Zoner)
				placed["dps"] += cellCount(p.DPS)
				placed["mobAir"] += cellCount(p.MobAir)
			}
			want := map[string]int{
				"chaser": cfg.ChaserRange[0], "zoner": cfg.ZonerRange[0],
				"dps": cfg.DPSRange[0], "mobAir": cfg.MobAirRange[0],
			}
			for layer, min := range want {
				if min > 0 && placed[layer] == 0 {
					t.Errorf("%s: no %s cell in 10 basement rooms, but the stage fields %d+",
						stage, layer, min)
				}
			}
		})
	}
}

func cellCount(layer model.Layer) int {
	n := 0
	for _, row := range layer {
		for _, v := range row {
			if v != 0 {
				n++
			}
		}
	}
	return n
}
