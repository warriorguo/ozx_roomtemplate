package generate

import (
	"fmt"
	"testing"
)

// TestBridgeCellsAlwaysAdjacentToGround verifies the invariant:
// every bridge=1 cell must have at least one orthogonally adjacent ground=1 cell.
func TestBridgeCellsAlwaysAdjacentToGround(t *testing.T) {
	doors := [][]DoorPosition{
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
	}

	sizes := [][2]int{
		{10, 10},
		{15, 15},
		{20, 20},
		{30, 20},
		{20, 30},
	}

	dirs := []struct{ dx, dy int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for trial := 0; trial < 200; trial++ {
		doorSet := doors[trial%len(doors)]
		size := sizes[trial%len(sizes)]
		w, h := size[0], size[1]

		req := BridgeGenerateRequest{
			Width:  w,
			Height: h,
			Doors:  doorSet,
		}

		resp, err := GenerateBridgeRoom(req)
		if err != nil {
			t.Fatalf("trial %d: unexpected error: %v", trial, err)
		}

		ground := resp.Payload.Ground
		bridge := resp.Payload.Bridge

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if bridge[y][x] != 1 {
					continue
				}
				// Every bridge cell must have at least one adjacent ground cell
				hasGround := false
				for _, d := range dirs {
					nx, ny := x+d.dx, y+d.dy
					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						if ground[ny][nx] == 1 {
							hasGround = true
							break
						}
					}
				}
				if !hasGround {
					// Print grid for debugging
					t.Errorf("trial %d (size %dx%d, doors %v): bridge cell (%d,%d) has no adjacent ground cell",
						trial, w, h, doorSet, x, y)
					fmt.Printf("Ground layer:\n")
					for gy := 0; gy < h; gy++ {
						for gx := 0; gx < w; gx++ {
							if ground[gy][gx] == 1 {
								fmt.Print("G")
							} else if bridge[gy][gx] == 1 {
								fmt.Print("B")
							} else {
								fmt.Print(".")
							}
						}
						fmt.Println()
					}
					return
				}
			}
		}
	}
}
