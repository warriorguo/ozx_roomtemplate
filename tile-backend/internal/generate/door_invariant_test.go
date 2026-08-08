package generate

import (
	"testing"
	"tile-backend/internal/model"
)

// TestDoorsAlwaysWalkable verifies the invariant:
// every requested door's anchor cell is walkable and 4-connected to the room's
// main region, for all three generators, across door configurations and sizes.
//
// Regression test for ORT-105: fullroom's corner erase and center pits are free
// to carve the door cell itself away — areAllDoorsConnected accepts a void door
// whose neighbour is reachable — and two erases on the same wall could take out
// that wall's entire edge line. The main path then had no walkable cell to aim
// at on that wall, so findCenterBiasedPath snapped its endpoint inward and the
// path never crossed the doorway.
func TestDoorsAlwaysWalkable(t *testing.T) {
	doorConfigs := [][]DoorPosition{
		{DoorLeft, DoorRight},
		{DoorTop, DoorBottom},
		{DoorTop, DoorLeft},
		{DoorBottom, DoorRight},
		{DoorTop, DoorBottom, DoorLeft, DoorRight},
	}
	stageTypes := []string{"", "teaching", "building"}
	sizes := [][2]int{{10, 8}, {16, 10}, {20, 12}, {24, 14}, {30, 20}}

	generators := []struct {
		name string
		gen  func(w, h int, doors []DoorPosition, stage string) (model.TemplatePayload, error)
	}{
		{"fullroom", func(w, h int, doors []DoorPosition, stage string) (model.TemplatePayload, error) {
			r, err := GenerateFullRoom(FullRoomGenerateRequest{Width: w, Height: h, Doors: doors, StageType: stage})
			if err != nil {
				return model.TemplatePayload{}, err
			}
			return r.Payload, nil
		}},
		{"bridge", func(w, h int, doors []DoorPosition, stage string) (model.TemplatePayload, error) {
			r, err := GenerateBridgeRoom(BridgeGenerateRequest{Width: w, Height: h, Doors: doors, StageType: stage})
			if err != nil {
				return model.TemplatePayload{}, err
			}
			return r.Payload, nil
		}},
		{"platform", func(w, h int, doors []DoorPosition, stage string) (model.TemplatePayload, error) {
			r, err := GeneratePlatformRoom(PlatformGenerateRequest{Width: w, Height: h, Doors: doors, StageType: stage})
			if err != nil {
				return model.TemplatePayload{}, err
			}
			return r.Payload, nil
		}},
	}

	for _, g := range generators {
		failures := 0
		for trial := 0; trial < 600; trial++ {
			doors := doorConfigs[trial%len(doorConfigs)]
			size := sizes[trial%len(sizes)]
			stage := stageTypes[trial%len(stageTypes)]
			w, h := size[0], size[1]

			payload, err := g.gen(w, h, doors, stage)
			if err != nil {
				continue // stage/size rejections are not this invariant's business
			}

			ground := payload.Ground
			reachable := reachableFromDoor(ground, payload.Bridge, doors[0], w, h)

			for _, door := range doors {
				p := getDoorCenterPositions(w, h, []DoorPosition{door})[door]
				if ground[p.Y][p.X] != 1 && (len(payload.Bridge) == 0 || payload.Bridge[p.Y][p.X] != 1) {
					failures++
					t.Errorf("%s trial %d (size %dx%d, doors %v, stage %q): door %s at (%d,%d) is not walkable",
						g.name, trial, w, h, doors, stage, door, p.X, p.Y)
				} else if !reachable[p.Y][p.X] {
					failures++
					t.Errorf("%s trial %d (size %dx%d, doors %v, stage %q): door %s at (%d,%d) is walkable but not connected to door %s",
						g.name, trial, w, h, doors, stage, door, p.X, p.Y, doors[0])
				}
				if failures > 5 {
					t.FailNow()
				}
			}
		}
	}
}

// reachableFromDoor flood-fills the walkable grid (ground or bridge) from the
// given door's anchor cell.
func reachableFromDoor(ground, bridge [][]int, from DoorPosition, w, h int) [][]bool {
	walkable := func(x, y int) bool {
		if ground[y][x] == 1 {
			return true
		}
		return len(bridge) > 0 && bridge[y][x] == 1
	}

	visited := make([][]bool, h)
	for y := 0; y < h; y++ {
		visited[y] = make([]bool, w)
	}

	start := getDoorCenterPositions(w, h, []DoorPosition{from})[from]
	if !walkable(start.X, start.Y) {
		return visited // caller reports the unwalkable door itself
	}

	visited[start.Y][start.X] = true
	queue := []Point{start}
	dirs := []Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := curr.X+d.X, curr.Y+d.Y
			if nx >= 0 && nx < w && ny >= 0 && ny < h && !visited[ny][nx] && walkable(nx, ny) {
				visited[ny][nx] = true
				queue = append(queue, Point{nx, ny})
			}
		}
	}
	return visited
}
