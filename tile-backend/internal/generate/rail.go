package generate

import (
	"container/heap"
	"fmt"
	"math/rand"
)

// Rail generation (2026-10-06 review).
//
// The rail is a network, not a loop:
//
//  1. Rings. The room's largest walkable rectangles (inside a 1-cell border)
//     each carry a rectangular rail ring when their area exceeds
//     railRectMinArea. The ring runs along the rectangle's edge, or 1-2 cells
//     inside it, depending on how far the edge is from the rectangle's centre
//     (maxRailInset).
//  2. Connectors. Rings are joined into one network (cheapest pair first).
//     A join prefers straight lines, and lays two of them, as far apart as
//     possible, when the gap has room for two.
//  3. Door spurs. Every open door is joined to the network. The rail enters
//     through one fixed cell of the doorway, chosen in visual space (see
//     railDoorCell).
//
// The network branches wherever a connector or a spur meets a ring. The old
// closed-loop / max-2-neighbours requirement is gone from generation and from
// both validators.
const (
	railRectMinArea    = 12 // a rectangle must be larger than this to carry a ring
	railRectMinSide    = 3  // narrower rectangles have no room for a ring
	maxRailRects       = 3  // the "few largest" rectangles
	railConnectorSlack = 3  // a second connector may be this much longer than the shortest
	railTurnCost       = 4  // routing: penalty per turn, so straight runs win
	railHugCost        = 2  // routing: penalty per step beside existing rail
	railBorderCost     = 3  // routing: penalty per step on the room's outermost border
)

// RailDebugInfo contains debug information about rail generation
type RailDebugInfo struct {
	Skipped        bool                `json:"skipped"`
	SkipReason     string              `json:"skipReason,omitempty"`
	PlatformsFound int                 `json:"platformsFound"` // rectangles that qualified for a ring
	RailLoops      []RailLoopInfo      `json:"railLoops"`      // one per ring
	Connectors     []RailConnectorInfo `json:"connectors,omitempty"`
	DoorSpurs      []RailDoorSpurInfo  `json:"doorSpurs,omitempty"`
	Misses         []MissInfo          `json:"misses,omitempty"`
}

// RailLoopInfo describes a rail ring drawn around one rectangle
type RailLoopInfo struct {
	Platform    string `json:"platform"`    // Rectangle position and size
	BoundingBox string `json:"boundingBox"` // Ring position and size
	Perimeter   int    `json:"perimeter"`   // Ring length in cells
	Inset       int    `json:"inset"`       // Cells between the rectangle edge and the ring
}

// RailConnectorInfo describes how two rings were joined
type RailConnectorInfo struct {
	From     int   `json:"from"` // ring index
	To       int   `json:"to"`
	Straight bool  `json:"straight"`
	Lengths  []int `json:"lengths"` // one entry per connector laid
}

// RailDoorSpurInfo describes the rail run from a door to the network
type RailDoorSpurInfo struct {
	Door   string `json:"door"`
	Start  string `json:"start"`
	Length int    `json:"length"`
}

// railRect is an axis-aligned rectangle in data space
type railRect struct {
	X, Y, W, H int
}

func (r railRect) String() string { return fmt.Sprintf("(%d,%d) %dx%d", r.X, r.Y, r.W, r.H) }

var railDirs = []Point{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}}

// GenerateRailLayer generates the rail layer from the ground and bridge layers
// and the room's open doors.
func GenerateRailLayer(railLayer, ground, bridge [][]int, doors []DoorPosition, width, height int) *RailDebugInfo {
	debug := &RailDebugInfo{RailLoops: []RailLoopInfo{}, Misses: []MissInfo{}}
	walkable := func(p Point) bool {
		return p.X >= 0 && p.X < width && p.Y >= 0 && p.Y < height &&
			(ground[p.Y][p.X] == 1 || bridge[p.Y][p.X] == 1)
	}

	// ringOf[y][x] is the ring index a rail cell was drawn for, -1 otherwise
	ringOf := make([][]int, height)
	for y := range ringOf {
		ringOf[y] = make([]int, width)
		for x := range ringOf[y] {
			ringOf[y][x] = -1
		}
	}

	// Step 1: rings
	rects := findRailRects(ground, bridge, width, height)
	debug.PlatformsFound = len(rects)
	for _, rect := range rects {
		inset := rand.Intn(maxRailInset(rect) + 1)
		ring := railRect{X: rect.X + inset, Y: rect.Y + inset, W: rect.W - 2*inset, H: rect.H - 2*inset}
		cells := ringCells(ring)
		idx := len(debug.RailLoops)
		for _, c := range cells {
			railLayer[c.Y][c.X] = 1
			ringOf[c.Y][c.X] = idx
		}
		debug.RailLoops = append(debug.RailLoops, RailLoopInfo{
			Platform:    rect.String(),
			BoundingBox: ring.String(),
			Perimeter:   len(cells),
			Inset:       inset,
		})
	}

	// Step 2: connectors
	connectRailRings(railLayer, ringOf, len(debug.RailLoops), walkable, width, height, debug)

	// Step 3: door spurs
	seeded := false
	for _, side := range doorOrder {
		if !containsDoor(doors, side) {
			continue
		}
		start, ok := railDoorCell(side, ground, bridge, width, height)
		if !ok {
			debug.Misses = append(debug.Misses, MissInfo{
				Reason: fmt.Sprintf("door %s: no walkable doorway cell for the rail", side),
			})
			continue
		}
		if railLayer[start.Y][start.X] == 1 {
			continue
		}
		if !hasAnyRail(railLayer, width, height) {
			// No ring: the first door seeds the network, the rest route to it.
			railLayer[start.Y][start.X] = 1
			seeded = true
			debug.DoorSpurs = append(debug.DoorSpurs, RailDoorSpurInfo{Door: string(side), Start: fmtPoint(start), Length: 1})
			continue
		}
		path := routeRail(railLayer, walkable, []Point{start}, inwardDir(side),
			func(p Point) bool { return railLayer[p.Y][p.X] == 1 }, width, height)
		if path == nil {
			debug.Misses = append(debug.Misses, MissInfo{
				Reason: fmt.Sprintf("door %s: no route from %s to the rail network", side, fmtPoint(start)),
			})
			continue
		}
		for _, c := range path {
			railLayer[c.Y][c.X] = 1
		}
		debug.DoorSpurs = append(debug.DoorSpurs, RailDoorSpurInfo{Door: string(side), Start: fmtPoint(start), Length: len(path)})
	}

	// A lone seed (one door, no ring) is not a rail.
	if seeded && len(debug.DoorSpurs) == 1 && len(debug.RailLoops) == 0 {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				railLayer[y][x] = 0
			}
		}
		debug.DoorSpurs = nil
	}

	if !hasAnyRail(railLayer, width, height) {
		debug.Skipped = true
		debug.SkipReason = fmt.Sprintf("no rectangle larger than %d cells and fewer than two reachable doors", railRectMinArea)
	}
	return debug
}

// findRailRects returns up to maxRailRects of the largest all-walkable
// rectangles inside the room's 1-cell border, largest first. Each one is
// kept a cell clear of the previous ones, so their rings never touch.
func findRailRects(ground, bridge [][]int, width, height int) []railRect {
	avail := make([][]bool, height)
	for y := 0; y < height; y++ {
		avail[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			inner := x >= 1 && x < width-1 && y >= 1 && y < height-1
			avail[y][x] = inner && (ground[y][x] == 1 || bridge[y][x] == 1)
		}
	}

	var rects []railRect
	for len(rects) < maxRailRects {
		r, ok := largestRailRect(avail, width, height)
		if !ok || r.W*r.H <= railRectMinArea {
			break
		}
		rects = append(rects, r)
		for y := r.Y - 1; y <= r.Y+r.H; y++ {
			for x := r.X - 1; x <= r.X+r.W; x++ {
				if x >= 0 && x < width && y >= 0 && y < height {
					avail[y][x] = false
				}
			}
		}
	}
	return rects
}

// largestRailRect finds the largest rectangle of available cells with both
// sides at least railRectMinSide.
func largestRailRect(avail [][]bool, width, height int) (railRect, bool) {
	heights := make([]int, width)
	var best railRect
	bestArea := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if avail[y][x] {
				heights[x]++
			} else {
				heights[x] = 0
			}
		}
		for l := 0; l < width; l++ {
			minH := heights[l]
			for r := l; r < width && minH > 0; r++ {
				if heights[r] < minH {
					minH = heights[r]
				}
				w := r - l + 1
				if w >= railRectMinSide && minH >= railRectMinSide && w*minH > bestArea {
					bestArea = w * minH
					best = railRect{X: l, Y: y - minH + 1, W: w, H: minH}
				}
			}
		}
	}
	return best, bestArea > 0
}

// maxRailInset is how far inside its rectangle a ring may run. d is the
// shortest distance from an edge to the rectangle's centre: below 4 the ring
// must sit on the edge, below 6 it may also sit one cell in, otherwise up to
// two cells in. A ring is never squeezed below railRectMinSide.
func maxRailInset(r railRect) int {
	short := r.W
	if r.H < short {
		short = r.H
	}
	d := short / 2
	limit := 2
	switch {
	case d < 4:
		limit = 0
	case d < 6:
		limit = 1
	}
	for limit > 0 && short-2*limit < railRectMinSide {
		limit--
	}
	return limit
}

// ringCells returns the perimeter cells of r
func ringCells(r railRect) []Point {
	var cells []Point
	for x := r.X; x < r.X+r.W; x++ {
		cells = append(cells, Point{X: x, Y: r.Y})
		if r.H > 1 {
			cells = append(cells, Point{X: x, Y: r.Y + r.H - 1})
		}
	}
	for y := r.Y + 1; y < r.Y+r.H-1; y++ {
		cells = append(cells, Point{X: r.X, Y: y})
		if r.W > 1 {
			cells = append(cells, Point{X: r.X + r.W - 1, Y: y})
		}
	}
	return cells
}

// connectRailRings joins every ring into one network, cheapest pair first
// (Kruskal). Straight connectors are preferred, two at a time where the gap
// allows; a pair with no straight connector is joined by a routed path.
func connectRailRings(railLayer, ringOf [][]int, ringCount int, walkable func(Point) bool, width, height int, debug *RailDebugInfo) {
	if ringCount < 2 {
		return
	}
	type edge struct{ a, b, cost int }
	var edges []edge
	for a := 0; a < ringCount; a++ {
		for b := a + 1; b < ringCount; b++ {
			if segs := straightRailConnectors(railLayer, ringOf, a, b, walkable, width, height); len(segs) > 0 {
				edges = append(edges, edge{a, b, len(shortestSeg(segs))})
				continue
			}
			path := routeRail(railLayer, walkable, cellsOfRing(ringOf, a, width, height), -1,
				func(p Point) bool { return ringOf[p.Y][p.X] == b }, width, height)
			if path != nil {
				edges = append(edges, edge{a, b, len(path) + 100}) // bent joins rank after straight ones
			}
		}
	}
	for i := 1; i < len(edges); i++ {
		for j := i; j > 0 && edges[j].cost < edges[j-1].cost; j-- {
			edges[j], edges[j-1] = edges[j-1], edges[j]
		}
	}

	parent := make([]int, ringCount)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}

	for _, e := range edges {
		if find(e.a) == find(e.b) {
			continue
		}
		// Recompute against the rail drawn so far.
		info := RailConnectorInfo{From: e.a, To: e.b}
		segs := straightRailConnectors(railLayer, ringOf, e.a, e.b, walkable, width, height)
		if len(segs) > 0 {
			info.Straight = true
			for _, s := range pickRailConnectors(segs) {
				for _, c := range s {
					railLayer[c.Y][c.X] = 1
				}
				info.Lengths = append(info.Lengths, len(s))
			}
		} else {
			path := routeRail(railLayer, walkable, cellsOfRing(ringOf, e.a, width, height), -1,
				func(p Point) bool { return ringOf[p.Y][p.X] == e.b }, width, height)
			if path == nil {
				debug.Misses = append(debug.Misses, MissInfo{Reason: fmt.Sprintf("rings %d and %d: no connector", e.a, e.b)})
				continue
			}
			path = path[1:] // drop the ring cell the route started from
			for _, c := range path {
				railLayer[c.Y][c.X] = 1
			}
			info.Lengths = []int{len(path)}
		}
		parent[find(e.a)] = find(e.b)
		debug.Connectors = append(debug.Connectors, info)
	}
	root := find(0)
	for i := 1; i < ringCount; i++ {
		if find(i) != root {
			debug.Misses = append(debug.Misses, MissInfo{Reason: fmt.Sprintf("ring %d is not joined to the network", i)})
		}
	}
}

// straightRailConnectors lists every straight run of free walkable cells that
// leaves ring a and arrives on ring b. A run that passes beside other rail is
// rejected, so a connector never doubles up against an existing line.
func straightRailConnectors(railLayer, ringOf [][]int, a, b int, walkable func(Point) bool, width, height int) [][]Point {
	var segs [][]Point
	for _, c := range cellsOfRing(ringOf, a, width, height) {
		for _, d := range railDirs {
			var cells []Point
			p := Point{X: c.X + d.X, Y: c.Y + d.Y}
			ok := true
			for walkable(p) && railLayer[p.Y][p.X] == 0 {
				side1 := Point{X: p.X + d.Y, Y: p.Y + d.X}
				side2 := Point{X: p.X - d.Y, Y: p.Y - d.X}
				if isRail(railLayer, side1, width, height) || isRail(railLayer, side2, width, height) {
					ok = false
					break
				}
				cells = append(cells, p)
				p = Point{X: p.X + d.X, Y: p.Y + d.Y}
			}
			if ok && len(cells) > 0 && walkable(p) && ringOf[p.Y][p.X] == b {
				segs = append(segs, cells)
			}
		}
	}
	return segs
}

// pickRailConnectors keeps the shortest straight connector, or two connectors
// as far apart as possible when there are two that do not touch. Only
// connectors within railConnectorSlack of the shortest are considered.
func pickRailConnectors(segs [][]Point) [][]Point {
	shortest := shortestSeg(segs)
	var pool [][]Point
	for _, s := range segs {
		if len(s) <= len(shortest)+railConnectorSlack {
			pool = append(pool, s)
		}
	}
	bestSep, bestLen := 1, 0
	var pair [][]Point
	for i := 0; i < len(pool); i++ {
		for j := i + 1; j < len(pool); j++ {
			sep := segSeparation(pool[i], pool[j])
			total := len(pool[i]) + len(pool[j])
			if sep > bestSep || (sep == bestSep && pair != nil && total < bestLen) {
				bestSep, bestLen = sep, total
				pair = [][]Point{pool[i], pool[j]}
			}
		}
	}
	if pair != nil {
		return pair
	}
	return [][]Point{shortest}
}

func shortestSeg(segs [][]Point) []Point {
	best := segs[0]
	for _, s := range segs[1:] {
		if len(s) < len(best) {
			best = s
		}
	}
	return best
}

// segSeparation is the smallest Manhattan distance between two connectors.
// Two connectors are only laid together when this is at least 2, i.e. they
// do not run side by side.
func segSeparation(a, b []Point) int {
	best := -1
	for _, p := range a {
		for _, q := range b {
			d := abs(p.X-q.X) + abs(p.Y-q.Y)
			if best < 0 || d < best {
				best = d
			}
		}
	}
	return best
}

// railDoorCell is the doorway cell the rail enters through. The rule is in
// visual space (canvas rotated 90° CCW, ORT-111): a visual up/down door — data
// right/left — takes the cell on its visual right, data y = H/2; a visual
// left/right door — data top/bottom — takes the cell visually below, data
// x = W/2-1. Falls back to the door anchor when that cell is not walkable.
func railDoorCell(side DoorPosition, ground, bridge [][]int, width, height int) (Point, bool) {
	var p Point
	switch side {
	case DoorLeft:
		p = Point{X: 0, Y: height / 2}
	case DoorRight:
		p = Point{X: width - 1, Y: height / 2}
	case DoorTop:
		p = Point{X: width/2 - 1, Y: 0}
	case DoorBottom:
		p = Point{X: width/2 - 1, Y: height - 1}
	}
	walk := func(q Point) bool {
		return q.X >= 0 && q.X < width && q.Y >= 0 && q.Y < height &&
			(ground[q.Y][q.X] == 1 || bridge[q.Y][q.X] == 1)
	}
	if walk(p) {
		return p, true
	}
	anchor := getDoorCenterPositions(width, height, []DoorPosition{side})[side]
	if walk(anchor) {
		return anchor, true
	}
	return Point{}, false
}

// inwardDir is the railDirs index pointing from a door into the room
func inwardDir(side DoorPosition) int {
	switch side {
	case DoorLeft:
		return 0
	case DoorRight:
		return 1
	case DoorTop:
		return 2
	default:
		return 3
	}
}

// routeRail finds the cheapest path of free walkable cells from any source to
// a cell where isTarget holds, charging railTurnCost per turn so straight runs
// win, plus railHugCost beside existing rail and railBorderCost on the room's
// border. startDir (a railDirs index, or -1) is the heading at the sources, so
// leaving them any other way counts as a turn. The returned path starts at the
// source and stops before the target cell; nil when no target is reachable.
func routeRail(railLayer [][]int, walkable func(Point) bool, sources []Point, startDir int, isTarget func(Point) bool, width, height int) []Point {
	type state struct {
		p   Point
		dir int
	}
	const noDir = 4
	dist := map[state]int{}
	prev := map[state]state{}
	pq := &railPQ{}
	for _, s := range sources {
		d := noDir
		if startDir >= 0 {
			d = startDir
		}
		st := state{s, d}
		dist[st] = 0
		heap.Push(pq, railPQItem{cost: 0, key: [3]int{s.X, s.Y, d}})
	}
	sourceSet := map[Point]bool{}
	for _, s := range sources {
		sourceSet[s] = true
	}

	for pq.Len() > 0 {
		it := heap.Pop(pq).(railPQItem)
		cur := state{Point{X: it.key[0], Y: it.key[1]}, it.key[2]}
		if it.cost > dist[cur] {
			continue
		}
		for di, d := range railDirs {
			q := Point{X: cur.p.X + d.X, Y: cur.p.Y + d.Y}
			if !walkable(q) || sourceSet[q] {
				continue
			}
			cost := it.cost + 1
			if cur.dir != noDir && cur.dir != di {
				cost += railTurnCost
			}
			if isTarget(q) {
				// Walk back and return.
				path := []Point{}
				for s := cur; ; {
					path = append(path, s.p)
					ps, ok := prev[s]
					if !ok {
						break
					}
					s = ps
				}
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				return path
			}
			if railLayer[q.Y][q.X] == 1 {
				continue
			}
			for _, n := range railDirs {
				r := Point{X: q.X + n.X, Y: q.Y + n.Y}
				if r != cur.p && isRail(railLayer, r, width, height) {
					cost += railHugCost
					break
				}
			}
			if q.X == 0 || q.Y == 0 || q.X == width-1 || q.Y == height-1 {
				cost += railBorderCost
			}
			ns := state{q, di}
			if old, seen := dist[ns]; !seen || cost < old {
				dist[ns] = cost
				prev[ns] = cur
				heap.Push(pq, railPQItem{cost: cost, key: [3]int{q.X, q.Y, di}})
			}
		}
	}
	return nil
}

type railPQItem struct {
	cost int
	key  [3]int
}

type railPQ []railPQItem

func (h railPQ) Len() int { return len(h) }
func (h railPQ) Less(i, j int) bool {
	if h[i].cost != h[j].cost {
		return h[i].cost < h[j].cost
	}
	// Deterministic tiebreak
	for k := 0; k < 3; k++ {
		if h[i].key[k] != h[j].key[k] {
			return h[i].key[k] < h[j].key[k]
		}
	}
	return false
}
func (h railPQ) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *railPQ) Push(x interface{}) { *h = append(*h, x.(railPQItem)) }
func (h *railPQ) Pop() interface{} {
	old := *h
	it := old[len(old)-1]
	*h = old[:len(old)-1]
	return it
}

func cellsOfRing(ringOf [][]int, idx, width, height int) []Point {
	var cells []Point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if ringOf[y][x] == idx {
				cells = append(cells, Point{X: x, Y: y})
			}
		}
	}
	return cells
}

func isRail(railLayer [][]int, p Point, width, height int) bool {
	return p.X >= 0 && p.X < width && p.Y >= 0 && p.Y < height && railLayer[p.Y][p.X] == 1
}

func containsDoor(doors []DoorPosition, side DoorPosition) bool {
	for _, d := range doors {
		if d == side {
			return true
		}
	}
	return false
}

func fmtPoint(p Point) string { return fmt.Sprintf("(%d,%d)", p.X, p.Y) }

// countRailNeighborsLocal counts adjacent rail cells
func countRailNeighborsLocal(railLayer [][]int, x, y, width, height int) int {
	count := 0
	for _, d := range railDirs {
		if isRail(railLayer, Point{X: x + d.X, Y: y + d.Y}, width, height) {
			count++
		}
	}
	return count
}

// GetRailIndentCells returns the non-rail cells enclosed by rail — not
// reachable from the room border without crossing rail. These are preferred
// positions for static placement.
func GetRailIndentCells(railLayer [][]int, width, height int) []Point {
	if !hasAnyRail(railLayer, width, height) {
		return nil
	}
	outside := make([][]bool, height)
	for y := range outside {
		outside[y] = make([]bool, width)
	}
	var queue []Point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			border := x == 0 || y == 0 || x == width-1 || y == height-1
			if border && railLayer[y][x] == 0 {
				outside[y][x] = true
				queue = append(queue, Point{X: x, Y: y})
			}
		}
	}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range railDirs {
			n := Point{X: c.X + d.X, Y: c.Y + d.Y}
			if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height &&
				!outside[n.Y][n.X] && railLayer[n.Y][n.X] == 0 {
				outside[n.Y][n.X] = true
				queue = append(queue, n)
			}
		}
	}
	var cells []Point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 0 && !outside[y][x] {
				cells = append(cells, Point{X: x, Y: y})
			}
		}
	}
	return cells
}

// hasAnyRail checks if there's any rail in the layer
func hasAnyRail(railLayer [][]int, width, height int) bool {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 1 {
				return true
			}
		}
	}
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
