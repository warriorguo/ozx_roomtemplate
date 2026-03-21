package generate

import (
	"fmt"
	"math/rand"
)

// createEmptyLayer creates a new empty (all-zero) layer of the given dimensions
func createEmptyLayer(width, height int) [][]int {
	layer := make([][]int, height)
	for y := 0; y < height; y++ {
		layer[y] = make([]int, width)
	}
	return layer
}

// copyLayer returns a deep copy of the given layer
func copyLayer(layer [][]int) [][]int {
	copied := make([][]int, len(layer))
	for y := range layer {
		copied[y] = make([]int, len(layer[y]))
		copy(copied[y], layer[y])
	}
	return copied
}

// boolToInt converts a bool to 0 or 1
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getDoorCenterPositions returns the center position of each door
func getDoorCenterPositions(width, height int, doors []DoorPosition) map[DoorPosition]Point {
	positions := make(map[DoorPosition]Point)
	for _, door := range doors {
		switch door {
		case DoorTop:
			positions[door] = Point{X: width / 2, Y: 0}
		case DoorBottom:
			positions[door] = Point{X: width / 2, Y: height - 1}
		case DoorLeft:
			positions[door] = Point{X: 0, Y: height / 2}
		case DoorRight:
			positions[door] = Point{X: width - 1, Y: height / 2}
		}
	}
	return positions
}

// drawLine draws a line between two points using the specified brush
func drawLine(ground [][]int, from, to Point, brush BrushSize, width, height int) {
	// Bresenham-like line drawing
	dx := abs(to.X - from.X)
	dy := abs(to.Y - from.Y)
	sx := 1
	if from.X > to.X {
		sx = -1
	}
	sy := 1
	if from.Y > to.Y {
		sy = -1
	}

	x, y := from.X, from.Y
	err := dx - dy

	for {
		applyBrush(ground, x, y, brush, width, height)

		if x == to.X && y == to.Y {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// applyBrush applies a brush centered at the given point
func applyBrush(ground [][]int, centerX, centerY int, brush BrushSize, width, height int) {
	startX := centerX - brush.Width/2
	startY := centerY - brush.Height/2

	for dy := 0; dy < brush.Height; dy++ {
		for dx := 0; dx < brush.Width; dx++ {
			x := startX + dx
			y := startY + dy
			if x >= 0 && x < width && y >= 0 && y < height {
				ground[y][x] = 1
			}
		}
	}
}

// applyBrushWithMirror applies a brush and its mirror counterpart
func applyBrushWithMirror(ground [][]int, centerX, centerY int, brush BrushSize, width, height int, mirror MirrorAxis) {
	// Apply original brush
	applyBrush(ground, centerX, centerY, brush, width, height)

	// Apply mirrored brush
	switch mirror {
	case MirrorY:
		// Mirror left-right (across vertical center line Y-axis)
		mirroredX := width - 1 - centerX
		applyBrush(ground, mirroredX, centerY, brush, width, height)
	case MirrorX:
		// Mirror top-bottom (across horizontal center line X-axis)
		mirroredY := height - 1 - centerY
		applyBrush(ground, centerX, mirroredY, brush, width, height)
	}
}

// manhattanDistance calculates the Manhattan distance between two points
func manhattanDistance(p1, p2 Point) int {
	return abs(p1.X-p2.X) + abs(p1.Y-p2.Y)
}

// getDoorForbiddenCells returns all cells within Manhattan distance doorForbiddenRadius of any door.
// This is the same radius used for enemy placement forbidden zones, ensuring statics
// also respect the door exclusion area.
func getDoorForbiddenCells(doorPositions map[DoorPosition]Point, width, height int) map[Point]bool {
	return getDoorForbiddenCellsRadius(doorPositions, width, height, doorForbiddenRadius)
}

// restoreLayer copies backup data back into target
func restoreLayer(target, backup [][]int) {
	for y := 0; y < len(target); y++ {
		for x := 0; x < len(target[y]); x++ {
			target[y][x] = backup[y][x]
		}
	}
}

// TryMutateWithRollback saves a backup of the layer, applies a mutation,
// then validates. If validation fails, it restores the backup and returns false.
func TryMutateWithRollback(layer [][]int, mutate func(), validate func() bool) bool {
	backup := copyLayer(layer)
	mutate()
	if !validate() {
		restoreLayer(layer, backup)
		return false
	}
	return true
}

// countCells counts cells with value 1 in a layer
func countCells(layer [][]int) int {
	count := 0
	for _, row := range layer {
		for _, v := range row {
			if v == 1 {
				count++
			}
		}
	}
	return count
}

// countLayerDebug creates a simple debug info by counting placed cells
func countLayerDebug(layer [][]int, target int, name string) *EnemyLayerDebugInfo {
	placed := countCells(layer)
	return &EnemyLayerDebugInfo{
		TargetCount: target,
		PlacedCount: placed,
		Placements:  []PlaceInfo{{Reason: fmt.Sprintf("grouped placement for %s", name)}},
	}
}

// drawLPath draws a 4-connected L-shaped path between from and to on the ground layer.
// It first moves horizontally to (to.X, from.Y) then vertically to (to.X, to.Y).
// Because each step is axis-aligned the resulting path is always 4-connected.
func drawLPath(ground [][]int, from, to Point, width, height int) {
	// Horizontal segment: from.Y stays fixed, x moves from from.X to to.X
	x := from.X
	stepX := 1
	if to.X < from.X {
		stepX = -1
	}
	for x != to.X {
		if x >= 0 && x < width && from.Y >= 0 && from.Y < height {
			ground[from.Y][x] = 1
		}
		x += stepX
	}
	// Ensure the corner cell is marked
	if to.X >= 0 && to.X < width && from.Y >= 0 && from.Y < height {
		ground[from.Y][to.X] = 1
	}

	// Vertical segment: x is now to.X, y moves from from.Y to to.Y
	y := from.Y
	stepY := 1
	if to.Y < from.Y {
		stepY = -1
	}
	for y != to.Y {
		if to.X >= 0 && to.X < width && y >= 0 && y < height {
			ground[y][to.X] = 1
		}
		y += stepY
	}
	// Ensure destination cell is marked
	if to.X >= 0 && to.X < width && to.Y >= 0 && to.Y < height {
		ground[to.Y][to.X] = 1
	}
}

// ensureGroundConnectivity guarantees that all ground=1 cells form a single
// 4-connected region. It finds every disconnected island, then connects each one
// to the largest island using a 4-connected L-shaped path on the ground layer.
//
// This is a post-processing repair step called after ground generation that may
// produce disconnected fragments (bridge room, fullroom corner/pit carving, platform).
func ensureGroundConnectivity(ground [][]int, width, height int) {
	islands := findAllIslands(ground, width, height)
	if len(islands) <= 1 {
		return // already connected (or no ground at all)
	}

	// Find the largest island — treat it as the main connected component.
	mainIdx := 0
	for i, isl := range islands {
		if len(isl.Cells) > len(islands[mainIdx].Cells) {
			mainIdx = i
		}
	}
	main := islands[mainIdx]

	// For each smaller island, connect its closest cell to the closest cell of
	// the main island using an L-shaped (4-connected) path.
	for i, isl := range islands {
		if i == mainIdx {
			continue
		}

		// Find the pair (island cell, main cell) with the smallest Manhattan distance.
		bestDist := -1
		var bestFrom, bestTo Point
		for _, ic := range isl.Cells {
			for _, mc := range main.Cells {
				d := abs(ic.X-mc.X) + abs(ic.Y-mc.Y)
				if bestDist < 0 || d < bestDist {
					bestDist = d
					bestFrom = ic
					bestTo = mc
				}
			}
		}

		if bestDist <= 0 {
			continue // already adjacent or empty
		}

		drawLPath(ground, bestFrom, bestTo, width, height)
	}
	// One pass is sufficient: every island is now directly connected to the main
	// island, so the result is a single 4-connected region.
}

// selectByWeight selects a strategy index by weight
func selectByWeight(strategies []Strategy) int {
	if len(strategies) == 0 {
		return -1
	}

	totalWeight := 0
	for _, s := range strategies {
		totalWeight += s.Weight
	}

	if totalWeight == 0 {
		return -1
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for i, s := range strategies {
		cumulative += s.Weight
		if r < cumulative {
			return i
		}
	}

	return len(strategies) - 1
}
