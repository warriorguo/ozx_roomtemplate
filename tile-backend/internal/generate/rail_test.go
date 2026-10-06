package generate

import (
	"math/rand"
	"strings"
	"testing"
)

func fullGround(width, height int) [][]int {
	g := createEmptyLayer(width, height)
	for y := range g {
		for x := range g[y] {
			g[y][x] = 1
		}
	}
	return g
}

func railASCII(rail, ground [][]int) string {
	var b strings.Builder
	for y := range rail {
		for x := range rail[y] {
			switch {
			case rail[y][x] == 1:
				b.WriteByte('R')
			case ground[y][x] == 1:
				b.WriteByte('.')
			default:
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// railComponents counts 4-connected rail components
func railComponents(rail [][]int, width, height int) int {
	seen := createEmptyLayer(width, height)
	n := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if rail[y][x] != 1 || seen[y][x] == 1 {
				continue
			}
			n++
			queue := []Point{{X: x, Y: y}}
			seen[y][x] = 1
			for len(queue) > 0 {
				c := queue[0]
				queue = queue[1:]
				for _, d := range railDirs {
					p := Point{X: c.X + d.X, Y: c.Y + d.Y}
					if isRail(rail, p, width, height) && seen[p.Y][p.X] == 0 {
						seen[p.Y][p.X] = 1
						queue = append(queue, p)
					}
				}
			}
		}
	}
	return n
}

// assertRailNetwork checks the invariants every generated rail must hold:
// on walkable cells, one connected network, and through every door's rail cell.
func assertRailNetwork(t *testing.T, rail, ground, bridge [][]int, doors []DoorPosition, width, height int) {
	t.Helper()
	if !hasAnyRail(rail, width, height) {
		return
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if rail[y][x] == 1 && ground[y][x] != 1 && bridge[y][x] != 1 {
				t.Errorf("rail (%d,%d) is not on ground or bridge", x, y)
			}
		}
	}
	if n := railComponents(rail, width, height); n != 1 {
		t.Errorf("rail has %d components, want 1\n%s", n, railASCII(rail, ground))
	}
	for _, side := range doors {
		p, ok := railDoorCell(side, ground, bridge, width, height)
		if ok && rail[p.Y][p.X] != 1 {
			t.Errorf("door %s: rail cell %v not on the rail\n%s", side, p, railASCII(rail, ground))
		}
	}
}

func TestGenerateRailLayer_EmptyGround(t *testing.T) {
	width, height := 20, 20
	ground := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	railLayer := createEmptyLayer(width, height)

	debug := GenerateRailLayer(railLayer, ground, bridge, []DoorPosition{DoorLeft, DoorRight}, width, height)
	if !debug.Skipped {
		t.Errorf("expected rail generation to be skipped with empty ground")
	}
}

// A full 16x10 room is one rectangle: one ring, joined to all four doors,
// each door entering through its visual-space doorway cell.
func TestGenerateRailLayer_FullRoomAllDoors(t *testing.T) {
	width, height := 16, 10
	ground := fullGround(width, height)
	bridge := createEmptyLayer(width, height)
	doors := []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}

	for trial := 0; trial < 20; trial++ {
		rail := createEmptyLayer(width, height)
		debug := GenerateRailLayer(rail, ground, bridge, doors, width, height)
		if debug.PlatformsFound != 1 || len(debug.RailLoops) != 1 {
			t.Fatalf("want 1 rectangle and 1 ring, got %d / %d", debug.PlatformsFound, len(debug.RailLoops))
		}
		assertRailNetwork(t, rail, ground, bridge, doors, width, height)
		// visual up/down = data right/left: the cell on the visual right, y=H/2
		// visual left/right = data top/bottom: the cell visually below, x=W/2-1
		for _, p := range []Point{{0, 5}, {15, 5}, {7, 0}, {7, 9}} {
			if rail[p.Y][p.X] != 1 {
				t.Errorf("door rail cell %v missing\n%s", p, railASCII(rail, ground))
			}
		}
	}
}

// With no rectangle large enough for a ring, the doors are joined to each
// other directly, in a straight line where possible.
func TestGenerateRailLayer_CorridorJoinsDoors(t *testing.T) {
	width, height := 16, 10
	ground := createEmptyLayer(width, height)
	for x := 0; x < width; x++ {
		ground[4][x] = 1
		ground[5][x] = 1
	}
	bridge := createEmptyLayer(width, height)
	rail := createEmptyLayer(width, height)

	debug := GenerateRailLayer(rail, ground, bridge, []DoorPosition{DoorLeft, DoorRight}, width, height)
	if debug.Skipped {
		t.Fatalf("expected a door-to-door rail, got skipped: %s", debug.SkipReason)
	}
	for x := 0; x < width; x++ {
		if rail[5][x] != 1 || rail[4][x] != 0 {
			t.Fatalf("want a straight rail along y=5\n%s", railASCII(rail, ground))
		}
	}
}

func TestGenerateRailLayer_SingleDoorNoRing(t *testing.T) {
	width, height := 16, 10
	ground := createEmptyLayer(width, height)
	for x := 0; x < width; x++ {
		ground[5][x] = 1
	}
	bridge := createEmptyLayer(width, height)
	rail := createEmptyLayer(width, height)

	debug := GenerateRailLayer(rail, ground, bridge, []DoorPosition{DoorLeft}, width, height)
	if !debug.Skipped || hasAnyRail(rail, width, height) {
		t.Errorf("one door and no ring should produce no rail\n%s", railASCII(rail, ground))
	}
}

func TestMaxRailInset(t *testing.T) {
	cases := []struct {
		w, h, want int
	}{
		{4, 4, 0},   // d=2
		{7, 20, 0},  // d=3
		{8, 8, 1},   // d=4
		{14, 11, 1}, // d=5
		{12, 12, 2}, // d=6
		{20, 30, 2},
	}
	for _, c := range cases {
		if got := maxRailInset(railRect{W: c.w, H: c.h}); got != c.want {
			t.Errorf("%dx%d: got inset %d, want %d", c.w, c.h, got, c.want)
		}
	}
}

func TestFindRailRects_SplitRoom(t *testing.T) {
	width, height := 24, 12
	ground := fullGround(width, height)
	for y := 0; y < height; y++ {
		ground[y][11] = 0
		ground[y][12] = 0
	}
	rects := findRailRects(ground, createEmptyLayer(width, height), width, height)
	if len(rects) != 2 {
		t.Fatalf("want 2 rectangles, got %v", rects)
	}
	for _, r := range rects {
		if r.W != 10 || r.H != 10 {
			t.Errorf("want 10x10 rectangles inside the border, got %v", r)
		}
		if r.X < 1 || r.Y < 1 || r.X+r.W > width-1 || r.Y+r.H > height-1 {
			t.Errorf("rectangle %v touches the room border", r)
		}
	}
}

func TestFindRailRects_SmallAreaIgnored(t *testing.T) {
	width, height := 10, 10
	ground := createEmptyLayer(width, height)
	for y := 2; y < 5; y++ { // 3x4 = 12, not more than 12
		for x := 2; x < 6; x++ {
			ground[y][x] = 1
		}
	}
	if rects := findRailRects(ground, createEmptyLayer(width, height), width, height); len(rects) != 0 {
		t.Errorf("area 12 must not carry a ring, got %v", rects)
	}
}

// Two rings side by side are joined by two straight connectors, as far apart
// as the overlap allows.
func TestConnectRailRings_TwoStraightConnectorsFarApart(t *testing.T) {
	width, height := 24, 12
	ground := fullGround(width, height)
	rail := createEmptyLayer(width, height)
	ringOf := make([][]int, height)
	for y := range ringOf {
		ringOf[y] = make([]int, width)
		for x := range ringOf[y] {
			ringOf[y][x] = -1
		}
	}
	for i, r := range []railRect{{X: 2, Y: 2, W: 7, H: 8}, {X: 13, Y: 2, W: 8, H: 8}} {
		for _, c := range ringCells(r) {
			rail[c.Y][c.X] = 1
			ringOf[c.Y][c.X] = i
		}
	}
	walkable := func(p Point) bool { return p.X >= 0 && p.X < width && p.Y >= 0 && p.Y < height }
	debug := &RailDebugInfo{}

	connectRailRings(rail, ringOf, 2, walkable, width, height, debug)

	if len(debug.Connectors) != 1 || !debug.Connectors[0].Straight || len(debug.Connectors[0].Lengths) != 2 {
		t.Fatalf("want one straight double connector, got %+v\n%s", debug.Connectors, railASCII(rail, ground))
	}
	for _, y := range []int{2, 9} {
		for x := 9; x <= 12; x++ {
			if rail[y][x] != 1 {
				t.Fatalf("want connectors on rows 2 and 9\n%s", railASCII(rail, ground))
			}
		}
	}
	if n := railComponents(rail, width, height); n != 1 {
		t.Errorf("rings not joined: %d components", n)
	}
}

// Rings whose only link is around a void get a routed, bent connector.
func TestConnectRailRings_BentWhenNoStraightLine(t *testing.T) {
	width, height := 24, 16
	ground := createEmptyLayer(width, height)
	// Two 6x6 rooms, offset so no row or column passes between them,
	// joined by an L-shaped strip of ground.
	fill := func(x0, y0, x1, y1 int) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				ground[y][x] = 1
			}
		}
	}
	fill(1, 1, 6, 6)
	fill(14, 9, 19, 14)
	fill(7, 3, 16, 3)
	fill(16, 3, 16, 8)
	bridge := createEmptyLayer(width, height)
	rail := createEmptyLayer(width, height)

	debug := GenerateRailLayer(rail, ground, bridge, nil, width, height)
	if len(debug.RailLoops) != 2 {
		t.Fatalf("want 2 rings, got %d\n%s", len(debug.RailLoops), railASCII(rail, ground))
	}
	if len(debug.Connectors) != 1 || debug.Connectors[0].Straight {
		t.Fatalf("want one bent connector, got %+v\n%s", debug.Connectors, railASCII(rail, ground))
	}
	if n := railComponents(rail, width, height); n != 1 {
		t.Errorf("rings not joined: %d components\n%s", n, railASCII(rail, ground))
	}
}

// Generated rooms: whatever the room, the rail is one network on walkable
// cells that passes through every open door.
func TestRailInvariant_GeneratedRooms(t *testing.T) {
	sides := []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}
	rng := rand.New(rand.NewSource(7))
	for trial := 0; trial < 150; trial++ {
		var doors []DoorPosition
		for _, s := range sides {
			if rng.Intn(2) == 0 {
				doors = append(doors, s)
			}
		}
		if len(doors) < 2 {
			doors = []DoorPosition{DoorLeft, DoorTop}
		}
		w, h := 16+rng.Intn(3)*4, 10+rng.Intn(2)*2

		full, err := GenerateFullRoom(FullRoomGenerateRequest{Width: w, Height: h, Doors: doors, RailEnabled: true, StageType: "building"})
		if err != nil {
			t.Fatalf("fullroom trial %d: %v", trial, err)
		}
		p := full.Payload
		assertRailNetwork(t, p.Rail, p.Ground, p.Bridge, doors, w, h)
		if !hasAnyRail(p.Rail, w, h) {
			t.Errorf("fullroom trial %d: no rail", trial)
		}

		plat, err := GeneratePlatformRoom(PlatformGenerateRequest{Width: w, Height: h, Doors: doors, RailEnabled: true, StageType: "building"})
		if err != nil {
			continue // platform may legitimately refuse some combinations
		}
		p = plat.Payload
		assertRailNetwork(t, p.Rail, p.Ground, p.Bridge, doors, w, h)
	}
}

func TestGetRailIndentCells(t *testing.T) {
	width, height := 20, 20
	railLayer := createEmptyLayer(width, height)
	for _, c := range ringCells(railRect{X: 5, Y: 5, W: 11, H: 11}) {
		railLayer[c.Y][c.X] = 1
	}
	// A spur from the border must not change what is enclosed.
	for x := 0; x < 5; x++ {
		railLayer[10][x] = 1
	}

	cells := GetRailIndentCells(railLayer, width, height)
	if len(cells) != 9*9 {
		t.Errorf("want 81 enclosed cells, got %d", len(cells))
	}
	for _, c := range cells {
		if c.X <= 5 || c.X >= 15 || c.Y <= 5 || c.Y >= 15 {
			t.Errorf("indent cell %v should be inside the ring", c)
		}
	}
}
