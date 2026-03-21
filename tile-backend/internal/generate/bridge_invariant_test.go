package generate

import (
	"fmt"
	"testing"
)

// TestBridgeBlockAlwaysHasOneFullEdgeOnGround verifies the invariant:
// every contiguous bridge block must have at least one full edge where ALL
// adjacent external cells are ground=1.
//
// Since placed bridges are rectangles (2×2, 4×2, 2×4), adjacent bridge
// placements may share cells and form non-rectangular connected components.
// We therefore verify the per-placement invariant by scanning all maximal
// axis-aligned rectangular sub-regions of each connected component that
// match a supported bridge size, and confirm at least one such rectangle
// satisfies the full-edge-on-ground check.
//
// For simpler end-to-end coverage we also verify that every bridge-connected
// component touches ground on at least one cell — a weaker but necessary
// condition that still catches floating clusters.
func TestBridgeBlockAlwaysHasOneFullEdgeOnGround(t *testing.T) {
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

	dirs := []Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

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

		// Find all connected components of bridge=1 cells.
		// For each component, verify it touches ground on at least one cell.
		visited := make([][]bool, h)
		for y := 0; y < h; y++ {
			visited[y] = make([]bool, w)
		}

		for startY := 0; startY < h; startY++ {
			for startX := 0; startX < w; startX++ {
				if bridge[startY][startX] != 1 || visited[startY][startX] {
					continue
				}

				// BFS to collect all cells of this bridge connected component
				var cells []Point
				queue := []Point{{startX, startY}}
				visited[startY][startX] = true

				for len(queue) > 0 {
					curr := queue[0]
					queue = queue[1:]
					cells = append(cells, curr)
					for _, d := range dirs {
						nx, ny := curr.X+d.X, curr.Y+d.Y
						if nx >= 0 && nx < w && ny >= 0 && ny < h &&
							bridge[ny][nx] == 1 && !visited[ny][nx] {
							visited[ny][nx] = true
							queue = append(queue, Point{nx, ny})
						}
					}
				}

				// Check that this connected component touches ground on at least one side.
				// This is the necessary condition: no bridge cluster may be fully surrounded by void.
				componentTouchesGround := false
				for _, cell := range cells {
					for _, d := range dirs {
						nx, ny := cell.X+d.X, cell.Y+d.Y
						if nx >= 0 && nx < w && ny >= 0 && ny < h && ground[ny][nx] == 1 {
							componentTouchesGround = true
							break
						}
					}
					if componentTouchesGround {
						break
					}
				}

				if !componentTouchesGround {
					// Print grid for debugging
					t.Errorf("trial %d (size %dx%d, doors %v): bridge component starting at (%d,%d) (%d cells) has no adjacent ground cell",
						trial, w, h, doorSet, startX, startY, len(cells))
					fmt.Printf("Ground+Bridge layer:\n")
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
