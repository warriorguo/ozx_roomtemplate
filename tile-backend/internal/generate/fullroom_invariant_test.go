package generate

import (
	"testing"
)

// TestFullRoomGroundAlwaysConnected verifies the invariant:
// full room generation must never produce isolated (disconnected) ground islands,
// across all door configurations and sizes.
//
// Regression test for ORT-37: fullroom single top-door config produced disconnected
// ground (10/64 reachable) because corner/pit erasing only rolled back on
// areAllDoorsConnected failure, which does not detect small isolated fragments.
func TestFullRoomGroundAlwaysConnected(t *testing.T) {
	doorConfigs := [][]DoorPosition{
		{DoorTop},
		{DoorBottom},
		{DoorLeft},
		{DoorRight},
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom},
		{DoorTop, DoorLeft},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}

	stageTypes := []string{"", "teaching", "building", "pressure"}

	sizes := [][2]int{
		{8, 8},
		{10, 10},
		{15, 15},
		{20, 20},
		{25, 15},
	}

	dirs := []Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	failures := 0

	for trial := 0; trial < 500; trial++ {
		doorSet := doorConfigs[trial%len(doorConfigs)]
		size := sizes[trial%len(sizes)]
		stage := stageTypes[trial%len(stageTypes)]
		w, h := size[0], size[1]

		req := FullRoomGenerateRequest{
			Width:         w,
			Height:        h,
			Doors:         doorSet,
			StageType:     stage,
			SoftEdgeCount: 3,
			StaticCount:   3,
			RailEnabled:   trial%2 == 0,
		}

		resp, err := GenerateFullRoom(req)
		if err != nil {
			continue
		}

		ground := resp.Payload.Ground

		// Count total ground cells and find a start cell
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

		// BFS flood-fill from the first ground cell
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
