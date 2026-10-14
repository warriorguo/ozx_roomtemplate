package generate

import "testing"

// ORT-129: boss used to demand an empty 6x6 arena, which needs height >= 14, so
// every boss request at the 16x10 default failed. It now follows the shipped
// all_boss_15_01: a default-size room with every spawn layer populated (ORT-128).
// DoorLeft is data space for OZX's bottom door - the one door ch3-ch6 boss rooms
// open.
func TestBossGeneratesAtDefaultSize(t *testing.T) {
	for i := 0; i < 10; i++ {
		res, err := GenerateFullRoom(FullRoomGenerateRequest{
			Doors: []DoorPosition{DoorLeft}, StageType: "boss",
		})
		if err != nil {
			t.Fatalf("boss at the default size: %v", err)
		}
		p := res.Payload
		for name, layer := range map[string][][]int{
			"static": p.Static, "chaser": p.Chaser, "zoner": p.Zoner, "dps": p.DPS, "mobAir": p.MobAir,
		} {
			if cellCount(layer) == 0 {
				t.Fatalf("boss room came back with an empty %s layer", name)
			}
		}
	}
}
