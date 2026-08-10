package generate

import "testing"

// ORT-114: a generate request that omits width/height gets the default room
// size. The numbers are the contract with the game — OZX derives the room size
// from the ground array's shape (outer = X, inner = Y), so a 16x10 data-space
// room is the 10-wide, 16-high normal room the camera is framed for. Changing
// them changes every room generated without explicit dimensions.
func TestDefaultRoomSizeConstants(t *testing.T) {
	if DefaultRoomWidth != 16 || DefaultRoomHeight != 10 {
		t.Fatalf("default room size = %dx%d, want 16x10 (data space) = 10 wide x 16 high in OZX",
			DefaultRoomWidth, DefaultRoomHeight)
	}
}

func TestGeneratorsApplyDefaultDimensions(t *testing.T) {
	doors := []DoorPosition{DoorLeft, DoorRight}

	t.Run("fullroom", func(t *testing.T) {
		res, err := GenerateFullRoom(FullRoomGenerateRequest{Doors: doors})
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		assertDefaultSized(t, res.Payload.Ground, res.Payload.Meta.Width, res.Payload.Meta.Height)
	})

	t.Run("bridge", func(t *testing.T) {
		res, err := GenerateBridgeRoom(BridgeGenerateRequest{Doors: doors})
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		assertDefaultSized(t, res.Payload.Ground, res.Payload.Meta.Width, res.Payload.Meta.Height)
	})

	t.Run("platform", func(t *testing.T) {
		res, err := GeneratePlatformRoom(PlatformGenerateRequest{Doors: doors})
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		assertDefaultSized(t, res.Payload.Ground, res.Payload.Meta.Width, res.Payload.Meta.Height)
	})
}

// A zero dimension means "not supplied"; a negative one is still a bad request
// and must not be quietly rewritten into a valid room.
func TestNegativeDimensionsStillRejected(t *testing.T) {
	doors := []DoorPosition{DoorLeft, DoorRight}
	if _, err := GenerateFullRoom(FullRoomGenerateRequest{Width: -1, Doors: doors}); err == nil {
		t.Error("fullroom accepted a negative width")
	}
	if _, err := GenerateBridgeRoom(BridgeGenerateRequest{Height: -1, Doors: doors}); err == nil {
		t.Error("bridge accepted a negative height")
	}
	if _, err := GeneratePlatformRoom(PlatformGenerateRequest{Width: -1, Doors: doors}); err == nil {
		t.Error("platform accepted a negative width")
	}
}

func assertDefaultSized(t *testing.T, ground [][]int, metaW, metaH int) {
	t.Helper()
	if metaW != DefaultRoomWidth || metaH != DefaultRoomHeight {
		t.Errorf("meta = %dx%d, want %dx%d", metaW, metaH, DefaultRoomWidth, DefaultRoomHeight)
	}
	// The game ignores meta and reads the array shape, so that is what actually
	// has to be right.
	if len(ground) != DefaultRoomHeight {
		t.Fatalf("ground has %d rows, want %d", len(ground), DefaultRoomHeight)
	}
	for y, row := range ground {
		if len(row) != DefaultRoomWidth {
			t.Fatalf("ground row %d has %d cells, want %d", y, len(row), DefaultRoomWidth)
		}
	}
}
