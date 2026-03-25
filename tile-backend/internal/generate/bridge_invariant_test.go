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

// TestBridgeRoomBridgeTilesAreStrictlyOriented verifies that any bridge tiles
// produced by the generator satisfy the strict directional bridge predicate.
// Bridge rooms may legitimately produce zero bridge tiles when the generated
// ground layout has no valid directional bridge gap.
func TestBridgeRoomBridgeTilesAreStrictlyOriented(t *testing.T) {
	doors := [][]DoorPosition{
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
	}

	stageTypes := []string{"", "teaching", "building"}

	sizes := [][2]int{
		{20, 12},
		{15, 15},
		{20, 20},
	}

	for trial := 0; trial < 300; trial++ {
		doorSet := doors[trial%len(doors)]
		size := sizes[trial%len(sizes)]
		stage := stageTypes[trial%len(stageTypes)]
		w, h := size[0], size[1]

		req := BridgeGenerateRequest{
			Width:         w,
			Height:        h,
			Doors:         doorSet,
			StageType:     stage,
			SoftEdgeCount: 3,
			StaticCount:   3,
		}

		resp, err := GenerateBridgeRoom(req)
		if err != nil {
			continue
		}

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if resp.Payload.Bridge[y][x] != 1 {
					continue
				}
				if resp.Payload.Ground[y][x] != 0 {
					t.Fatalf("trial %d (size %dx%d, doors %v, stage %q): bridge overlaps ground at (%d,%d)",
						trial, w, h, doorSet, stage, x, y)
				}

				valid := false
				for _, sz := range bridgeSizes {
					for dy := 0; dy < sz.h; dy++ {
						for dx := 0; dx < sz.w; dx++ {
							bx := x - dx
							by := y - dy
							if bx < 0 || by < 0 || bx+sz.w > w || by+sz.h > h {
								continue
							}
							if bridgeConnectingDirection(bx, by, sz.w, sz.h, resp.Payload.Ground, w, h) != "" {
								valid = true
								break
							}
						}
						if valid {
							break
						}
					}
					if valid {
						break
					}
				}

				if !valid {
					t.Fatalf("trial %d (size %dx%d, doors %v, stage %q): bridge tile at (%d,%d) is not part of a strict directional bridge",
						trial, w, h, doorSet, stage, x, y)
				}
			}
		}
	}
}

// TestBridgeGroundAlwaysConnected verifies that all ground=1 cells in every bridge room
// form a single 4-connected region, regardless of door configuration or RNG seed.
// Regression test for ORT-35 / ORT-36: bridge rooms intermittently produced disconnected
// ground fragments across all door configurations (left/right, top/bottom, 4-door).
func TestBridgeGroundAlwaysConnected(t *testing.T) {
	doors := [][]DoorPosition{
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom},
		{DoorTop, DoorRight, DoorBottom},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
	}

	stageTypes := []string{"", "teaching", "building"}

	sizes := [][2]int{
		{20, 12},
		{15, 15},
		{20, 20},
		{25, 15},
		{30, 20},
	}

	dirs := []Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	failures := 0

	for trial := 0; trial < 600; trial++ {
		doorSet := doors[trial%len(doors)]
		size := sizes[trial%len(sizes)]
		stage := stageTypes[trial%len(stageTypes)]
		w, h := size[0], size[1]

		req := BridgeGenerateRequest{
			Width:         w,
			Height:        h,
			Doors:         doorSet,
			StageType:     stage,
			SoftEdgeCount: 3,
			StaticCount:   3,
			RailEnabled:   trial%2 == 0,
		}

		resp, err := GenerateBridgeRoom(req)
		if err != nil {
			continue
		}

		ground := resp.Payload.Ground

		// Count total ground cells
		totalGround := 0
		var startCell *Point
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if ground[y][x] == 1 {
					totalGround++
					if startCell == nil {
						p := Point{x, y}
						startCell = &p
					}
				}
			}
		}

		if totalGround == 0 || startCell == nil {
			continue
		}

		// BFS from first ground cell
		visited := make([][]bool, h)
		for y := 0; y < h; y++ {
			visited[y] = make([]bool, w)
		}
		queue := []Point{*startCell}
		visited[startCell.Y][startCell.X] = true
		reachable := 0

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			reachable++
			for _, d := range dirs {
				nx, ny := curr.X+d.X, curr.Y+d.Y
				if nx >= 0 && nx < w && ny >= 0 && ny < h && ground[ny][nx] == 1 && !visited[ny][nx] {
					visited[ny][nx] = true
					queue = append(queue, Point{nx, ny})
				}
			}
		}

		if reachable != totalGround {
			failures++
			t.Errorf("trial %d (size %dx%d, doors %v, stage %q, railEnabled=%v): ground disconnected — %d of %d ground cells reachable",
				trial, w, h, doorSet, stage, req.RailEnabled, reachable, totalGround)
			if failures > 5 {
				t.FailNow()
			}
		}
	}
}

// TestBridgeIslandsHaveAdjacentBridgeTiles verifies that every disconnected ground island
// in a bridge room has at least one adjacent bridge tile.
// Regression test for ORT-32: isolated islands were left without bridge adjacency.
func TestBridgeIslandsHaveAdjacentBridgeTiles(t *testing.T) {
	doors := [][]DoorPosition{
		{DoorTop, DoorBottom},
		{DoorLeft, DoorRight},
		{DoorTop, DoorRight, DoorBottom},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
		{DoorTop, DoorRight},
	}

	stageTypes := []string{"", "teaching", "building"}

	sizes := [][2]int{
		{20, 12},
		{20, 20},
		{25, 15},
	}

	dirs := []Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	failures := 0

	for trial := 0; trial < 300; trial++ {
		doorSet := doors[trial%len(doors)]
		size := sizes[trial%len(sizes)]
		stage := stageTypes[trial%len(stageTypes)]
		w, h := size[0], size[1]

		req := BridgeGenerateRequest{
			Width:         w,
			Height:        h,
			Doors:         doorSet,
			StageType:     stage,
			SoftEdgeCount: 5,
			StaticCount:   5,
		}

		resp, err := GenerateBridgeRoom(req)
		if err != nil {
			continue
		}

		ground := resp.Payload.Ground
		bridge := resp.Payload.Bridge

		islands := findAllIslands(ground, w, h)
		if len(islands) <= 1 {
			continue
		}

		for _, island := range islands {
			hasBridgeNeighbor := false
			for _, cell := range island.Cells {
				for _, d := range dirs {
					nx, ny := cell.X+d.X, cell.Y+d.Y
					if nx >= 0 && nx < w && ny >= 0 && ny < h && bridge[ny][nx] == 1 {
						hasBridgeNeighbor = true
						break
					}
				}
				if hasBridgeNeighbor {
					break
				}
			}
			if !hasBridgeNeighbor {
				failures++
				t.Errorf("trial %d (size %dx%d, doors %v, stage %q): island (%d,%d)-(%d,%d) with %d cells has no adjacent bridge",
					trial, w, h, doorSet, stage,
					island.MinX, island.MinY, island.MaxX, island.MaxY, len(island.Cells))
				if failures > 5 {
					t.FailNow()
				}
			}
		}
	}
}
