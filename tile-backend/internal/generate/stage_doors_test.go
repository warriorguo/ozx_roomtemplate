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

// ORT-127: the game asks for peak corner rooms (TL/TR in OZX space), so the
// generator must be able to make one.
func TestPeakAcceptsCornerDoorPair(t *testing.T) {
	if _, err := GenerateFullRoom(FullRoomGenerateRequest{
		Doors: []DoorPosition{DoorRight, DoorBottom}, StageType: "peak",
	}); err != nil {
		t.Fatalf("peak with a corner door pair: %v", err)
	}
}

// ORT-127: under grouped placement (pressure/peak), group 2's zoner used to
// land on group 1's dps - it was only told about the chaser layer.
func TestGroupedZonerNeverSharesACellWithAnEnemy(t *testing.T) {
	for _, stage := range []string{"pressure", "peak"} {
		for i := 0; i < 100; i++ {
			res, err := GenerateFullRoom(FullRoomGenerateRequest{Doors: allFourDoors, StageType: stage})
			if err != nil {
				t.Fatal(err)
			}
			p := res.Payload
			for y := range p.Zoner {
				for x := range p.Zoner[y] {
					if p.Zoner[y][x] != 0 && (p.DPS[y][x] != 0 || p.Chaser[y][x] != 0) {
						t.Fatalf("%s: zoner shares (%d,%d) with a chaser or dps", stage, x, y)
					}
				}
			}
		}
	}
}
