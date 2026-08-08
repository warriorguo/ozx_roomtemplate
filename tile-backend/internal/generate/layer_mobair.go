package generate

import (
	"fmt"
	"math"
	"sort"
)

// GenerateMobAirLayerNew generates the mobAir layer.
//
// Placement is centre-seeded and evenly distributed (ORT-104): the first mob
// takes the valid cell nearest the room centre, and the rest fill the slots of
// a grid anchored on the valid area, worked outward from that centre.
//
// Geometry sets the spacing. The zoner/chaser density score only chooses
// between the cells a single slot could legally snap to, so air mobs still lean
// toward ground threats without collapsing into a cluster around them — the
// even-coverage requirement wins, the dense-area preference breaks its ties.
//
// Placement is deterministic: the same room always produces the same layer.
func GenerateMobAirLayerNew(mobAirLayer, ground, softEdge, bridge, staticLayer, zonerLayer, chaserLayer, dpsLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int, regionFilter ...*RegionFilter) *MobAirDebugInfo {

	debug := &MobAirDebugInfo{
		TargetCount: targetCount,
		Strategy:    "center_out_even",
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	if targetCount <= 0 {
		debug.Skipped = true
		debug.SkipReason = "targetCount is 0"
		return debug
	}

	var rf *RegionFilter
	if len(regionFilter) > 0 {
		rf = regionFilter[0]
	}

	// isValid re-checks a cell against the live layer: each placement blocks its
	// own cell and its 8 neighbours, so validity changes as we go.
	isValid := func(pos Point) bool {
		return rf.Contains(pos.X, pos.Y) &&
			isValidMobAirPositionNew(pos, ground, softEdge, bridge, staticLayer, zonerLayer, chaserLayer, dpsLayer,
				mobAirLayer, doorPositions, width, height)
	}

	// Row-major order, which makes every "first best wins" tiebreak below
	// deterministic.
	var candidates []Point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValid(pos) {
				candidates = append(candidates, pos)
			}
		}
	}

	if len(candidates) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{Reason: "no valid positions found"})
		return debug
	}

	density := mobAirDensityScores(candidates, zonerLayer, chaserLayer, width, height)
	placed := make([]Point, 0, targetCount)

	place := func(pos Point, reason string) {
		mobAirLayer[pos.Y][pos.X] = 1
		placed = append(placed, pos)
		debug.PlacedCount++
		if ground[pos.Y][pos.X] == 0 {
			reason += " (on void, flying)"
		}
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "1x1",
			Reason:   reason,
		})
	}

	remaining := targetCount

	// 1. Seed the room centre. Distance decides here, not density — the centre
	//    cell is the anchor everything else is arranged around.
	center := Point{X: width / 2, Y: height / 2}
	if seed, ok := nearestValidTo(center, candidates, density, isValid); ok {
		place(seed, "centre seed")
		remaining--
	}

	// 2. Work the remaining grid slots outward from the centre.
	minX, minY, maxX, maxY := pointsBoundingBox(candidates)
	slots, cellW, cellH := mobAirGridSlots(targetCount, minX, minY, maxX, maxY)
	sort.SliceStable(slots, func(i, j int) bool {
		di, dj := manhattanDistance(slots[i], center), manhattanDistance(slots[j], center)
		if di != dj {
			return di < dj
		}
		if slots[i].Y != slots[j].Y {
			return slots[i].Y < slots[j].Y
		}
		return slots[i].X < slots[j].X
	})
	// The centre-most slot is the seed already placed in step 1.
	if len(slots) > 0 {
		slots = slots[1:]
	}

	// A third of a cell of slack: enough room for the density tiebreak to
	// matter, but comfortably inside the cell so two slots can never reach the
	// same neighbourhood and undo the even spacing. Measured across stages, a
	// half-cell window costs real dispersion (Clark-Evans 1.32 vs 1.56 at
	// teaching); the floor of 1 keeps a 3x3 window in tightly packed rooms.
	snapRadius := int(math.Floor(math.Min(cellW, cellH) / 3))
	if snapRadius < 1 {
		snapRadius = 1
	}

	for _, slot := range slots {
		if remaining <= 0 {
			break
		}
		if pos, ok := snapToSlot(slot, candidates, density, snapRadius, isValid); ok {
			place(pos, fmt.Sprintf("grid slot (%d,%d)", slot.X, slot.Y))
			remaining--
		}
	}

	// 3. Top up. A slot whose whole snap window was consumed by a neighbour's
	//    adjacency block leaves the count short, so fill from whatever is still
	//    valid — farthest from what is already placed, which keeps the spread
	//    even rather than bunching the leftovers together.
	for remaining > 0 {
		pos, ok := farthestValidFrom(placed, candidates, density, isValid)
		if !ok {
			break
		}
		place(pos, "top-up (farthest from placed)")
		remaining--
	}

	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more mob air", remaining),
		})
	}

	return debug
}

// mobAirDensityScores weights each candidate by the zoner/chaser presence in the
// 7x7 window around it. It is the "prefer areas dense with ground threats" rule,
// demoted to a tiebreak by ORT-104.
func mobAirDensityScores(candidates []Point, zonerLayer, chaserLayer [][]int, width, height int) map[Point]float64 {
	scores := make(map[Point]float64, len(candidates))
	for _, pos := range candidates {
		s := 0.0
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				nx, ny := pos.X+dx, pos.Y+dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height {
					if zonerLayer[ny][nx] == 1 {
						s += 2.0
					}
					if chaserLayer[ny][nx] == 1 {
						s += 1.5
					}
				}
			}
		}
		scores[pos] = s
	}
	return scores
}

// mobAirGridSlots lays out count ideal positions over the given box as a grid of
// cell centres, and returns the cell size alongside them. Because the cell
// centres are symmetric within the box, the grid is anchored on the box centre
// rather than its top-left corner.
func mobAirGridSlots(count, minX, minY, maxX, maxY int) (slots []Point, cellW, cellH float64) {
	boxW, boxH := maxX-minX+1, maxY-minY+1
	cols, rows := mobAirGridDimensions(count, boxW, boxH)

	cellW = float64(boxW) / float64(cols)
	cellH = float64(boxH) / float64(rows)

	slots = make([]Point, 0, cols*rows)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			slots = append(slots, Point{
				X: minX + int(math.Round((float64(c)+0.5)*cellW-0.5)),
				Y: minY + int(math.Round((float64(r)+0.5)*cellH-0.5)),
			})
		}
	}
	return slots, cellW, cellH
}

// mobAirGridDimensions picks a cols x rows grid with at least count cells, as
// close to square as the box aspect allows.
func mobAirGridDimensions(count, boxW, boxH int) (cols, rows int) {
	if count <= 1 {
		return 1, 1
	}
	if boxW < 1 {
		boxW = 1
	}
	if boxH < 1 {
		boxH = 1
	}

	aspect := float64(boxW) / float64(boxH)
	rows = int(math.Round(math.Sqrt(float64(count) / aspect)))
	if rows < 1 {
		rows = 1
	}
	cols = (count + rows - 1) / rows
	if cols < 1 {
		cols = 1
	}

	// Grow along whichever axis currently has the coarser cells.
	for cols*rows < count {
		if float64(boxW)/float64(cols) >= float64(boxH)/float64(rows) {
			cols++
		} else {
			rows++
		}
	}
	return cols, rows
}

// snapToSlot picks the cell a grid slot resolves to: the highest density score
// within radius of the slot, ties broken by proximity to the slot and then by
// row-major order. Returns false when the slot's window holds nothing valid.
func snapToSlot(slot Point, candidates []Point, density map[Point]float64, radius int, isValid func(Point) bool) (Point, bool) {
	var best Point
	var bestScore float64
	bestDist := 0
	found := false

	for _, c := range candidates {
		d := chebyshevDistance(c, slot)
		if d > radius || !isValid(c) {
			continue
		}
		s := density[c]
		if !found || s > bestScore || (s == bestScore && d < bestDist) {
			best, bestScore, bestDist, found = c, s, d, true
		}
	}
	return best, found
}

// nearestValidTo returns the valid candidate closest to target, preferring the
// denser one where two are equidistant.
func nearestValidTo(target Point, candidates []Point, density map[Point]float64, isValid func(Point) bool) (Point, bool) {
	var best Point
	var bestScore float64
	bestDist := 0
	found := false

	for _, c := range candidates {
		if !isValid(c) {
			continue
		}
		d := manhattanDistance(c, target)
		s := density[c]
		if !found || d < bestDist || (d == bestDist && s > bestScore) {
			best, bestScore, bestDist, found = c, s, d, true
		}
	}
	return best, found
}

// farthestValidFrom returns the valid candidate whose nearest already-placed
// neighbour is the most distant — a farthest-point pick, so top-ups keep
// spreading out instead of bunching. Density breaks ties.
func farthestValidFrom(placed []Point, candidates []Point, density map[Point]float64, isValid func(Point) bool) (Point, bool) {
	var best Point
	var bestScore float64
	bestDist := 0
	found := false

	for _, c := range candidates {
		if !isValid(c) {
			continue
		}
		d := distanceToNearest(c, placed) // -1 when nothing is placed yet
		s := density[c]
		if !found || d > bestDist || (d == bestDist && s > bestScore) {
			best, bestScore, bestDist, found = c, s, d, true
		}
	}
	return best, found
}

// pointsBoundingBox returns the inclusive bounds of a non-empty point set.
func pointsBoundingBox(points []Point) (minX, minY, maxX, maxY int) {
	minX, minY = points[0].X, points[0].Y
	maxX, maxY = minX, minY
	for _, p := range points {
		minX, maxX = min(minX, p.X), max(maxX, p.X)
		minY, maxY = min(minY, p.Y), max(maxY, p.Y)
	}
	return minX, minY, maxX, maxY
}

// chebyshevDistance is the number of king moves between two cells.
func chebyshevDistance(a, b Point) int {
	dx, dy := abs(a.X-b.X), abs(a.Y-b.Y)
	return max(dx, dy)
}
