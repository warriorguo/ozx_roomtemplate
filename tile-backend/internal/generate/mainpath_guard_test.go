package generate

import (
	"strings"
	"testing"
)

// TestComputeMainPathRejectsSealedDoor verifies that a door with no walkable
// cell anywhere on its wall is an error, not a debug note.
//
// Regression test for ORT-106: snapDoorToEdge returned the unwalkable anchor
// untouched, findCenterBiasedPath then snapped it to the nearest walkable cell,
// and the resulting path ran between two interior cells without ever crossing
// the doorway. Nothing in the return value said so.
func TestComputeMainPathRejectsSealedDoor(t *testing.T) {
	const w, h = 12, 10
	ground := createEmptyLayer(w, h)
	for y := 0; y < h; y++ {
		for x := 1; x < w; x++ { // column 0 left entirely void
			ground[y][x] = 1
		}
	}
	bridge := createEmptyLayer(w, h)
	doors := getDoorCenterPositions(w, h, []DoorPosition{DoorLeft, DoorRight})

	data, _, err := ComputeMainPath(ground, bridge, doors, w, h)
	if err == nil {
		t.Fatalf("expected an error for a sealed left doorway, got none (data=%v)", data != nil)
	}
	if !strings.Contains(err.Error(), "left") || !strings.Contains(err.Error(), "sealed") {
		t.Errorf("error should name the sealed door, got: %v", err)
	}
}

// TestComputeMainPathRejectsDisconnectedDoors verifies that two doors on
// walkable but mutually unreachable ground is an error.
func TestComputeMainPathRejectsDisconnectedDoors(t *testing.T) {
	const w, h = 12, 10
	ground := createEmptyLayer(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x != w/2 { // a full-height void column splits the room in two
				ground[y][x] = 1
			}
		}
	}
	bridge := createEmptyLayer(w, h)
	doors := getDoorCenterPositions(w, h, []DoorPosition{DoorLeft, DoorRight})

	_, _, err := ComputeMainPath(ground, bridge, doors, w, h)
	if err == nil {
		t.Fatal("expected an error for doors separated by a void column, got none")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("error should say the doors are not connected, got: %v", err)
	}
}

// TestComputeMainPathSingleDoorIsNotAnError keeps one-door rooms working: there
// is no pair to connect, so there is no main path and nothing to fail.
func TestComputeMainPathSingleDoorIsNotAnError(t *testing.T) {
	const w, h = 12, 10
	ground := createEmptyLayer(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ground[y][x] = 1
		}
	}
	bridge := createEmptyLayer(w, h)
	doors := getDoorCenterPositions(w, h, []DoorPosition{DoorLeft})

	data, _, err := ComputeMainPath(ground, bridge, doors, w, h)
	if err != nil {
		t.Fatalf("single-door room should not error: %v", err)
	}
	if data == nil {
		t.Fatal("single-door room should still return main path data")
	}
}

// TestGeneratorsNeverFailMainPath is the other half of the guard: with ORT-105
// in place the invariant holds, so the new error must never fire during normal
// generation. A failure here means either a real regression in ground
// generation or a guard that is too eager.
func TestGeneratorsNeverFailMainPath(t *testing.T) {
	doorConfigs := [][]DoorPosition{
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom},
		{DoorTop, DoorLeft},
		{DoorBottom, DoorRight},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}
	stageTypes := []string{"", "teaching", "building"}
	sizes := [][2]int{{10, 8}, {16, 10}, {20, 12}, {24, 14}, {30, 20}}

	for trial := 0; trial < 900; trial++ {
		doors := doorConfigs[trial%len(doorConfigs)]
		size := sizes[trial%len(sizes)]
		stage := stageTypes[trial%len(stageTypes)]
		w, h := size[0], size[1]

		var err error
		switch trial % 3 {
		case 0:
			_, err = GenerateFullRoom(FullRoomGenerateRequest{Width: w, Height: h, Doors: doors, StageType: stage})
		case 1:
			_, err = GenerateBridgeRoom(BridgeGenerateRequest{Width: w, Height: h, Doors: doors, StageType: stage})
		case 2:
			_, err = GeneratePlatformRoom(PlatformGenerateRequest{Width: w, Height: h, Doors: doors, StageType: stage})
		}

		if err != nil && strings.Contains(err.Error(), "main path") {
			t.Fatalf("trial %d (size %dx%d, doors %v, stage %q): %v", trial, w, h, doors, stage, err)
		}
	}
}
