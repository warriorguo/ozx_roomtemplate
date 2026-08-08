package generate

import (
	"fmt"
	"sort"
)

// GenerateZonerLayer generates the zoner layer.
// Zoners must be on ground, within 0-5 of main path, prefer HIGH squishy score,
// and no static between zoner and main path.
//
// A zoner occupies a 2x2 block wherever a valid site exists and falls back to a
// single cell otherwise (ORT-103), so targetCount counts spawns rather than
// cells. Every cell of a block satisfies the same constraints the 1x1 path
// checks, and no two zoners touch in any of the 8 directions — which keeps each
// connected group of zoner cells exactly one spawn, the spacing ORT-93 needs.
func GenerateZonerLayer(zonerLayer, ground, softEdge, bridge, rail, staticLayer [][]int,
	doorPositions map[DoorPosition]Point, mainPath *MainPathData, width, height, targetCount int, regionFilter ...*RegionFilter) *EnemyLayerDebugInfo {

	debug := &EnemyLayerDebugInfo{
		TargetCount: targetCount,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	forbidden := getDoorForbiddenCellsRadius(doorPositions, width, height, doorForbiddenRadius)

	var rf *RegionFilter
	if len(regionFilter) > 0 {
		rf = regionFilter[0]
	}

	// collectCells gathers every cell a zoner may sit on. requireLOS applies the
	// "no static between the zoner and the main path" rule.
	collectCells := func(requireLOS bool) []Point {
		var cells []Point
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if !rf.Contains(x, y) {
					continue
				}
				pos := Point{x, y}
				if !isValidEnemyPosition(pos, ground, softEdge, bridge, rail, staticLayer, forbidden, width, height) {
					continue
				}
				// Must be within zonerMaxPathDist of main path
				if mainPath == nil || mainPath.DirectDistance[y][x] > zonerMaxPathDist {
					continue
				}
				if requireLOS && hasStaticBlockingPath(pos, mainPath, staticLayer, width, height) {
					continue
				}
				cells = append(cells, pos)
			}
		}
		return cells
	}

	candidates := collectCells(true)

	// Fallback: if no candidates found with LOS constraint, retry without static-blocking-path filter.
	// This ensures at least some valid positions exist in heavily-static rooms.
	if len(candidates) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{Reason: "no valid positions with LOS constraint, retrying without static-blocking-path filter"})
		candidates = collectCells(false)
	}

	if len(candidates) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{Reason: "no valid positions found even without LOS constraint"})
		return debug
	}

	// Sort by squishy score descending (prefer HIGH score — hard to reach but has LOS)
	sort.Slice(candidates, func(i, j int) bool {
		si := mainPath.SquishyScore[candidates[i].Y][candidates[i].X]
		sj := mainPath.SquishyScore[candidates[j].Y][candidates[j].X]
		return si > sj
	})

	// Preferred footprint: 2x2 blocks whose whole area is drawn from the valid
	// cells above, ranked by the mean squishy score over that area.
	blocks := zonerBlockAnchors(candidates, width, height)
	sort.Slice(blocks, func(i, j int) bool {
		return blockSquishyScore(blocks[i], mainPath) > blockSquishyScore(blocks[j], mainPath)
	})

	remaining := targetCount

	for remaining > 0 && len(blocks) > 0 {
		pos, idx := pickFromTopN(blocks, 0.3, 3)
		blocks = append(blocks[:idx], blocks[idx+1:]...)
		if blockTouchesLayer(pos, zonerSize, zonerLayer, width, height) {
			continue
		}
		placeBlock(zonerLayer, pos, zonerSize)
		blocks = filterTouchingBlocks(blocks, pos, zonerSize)
		remaining--
		debug.PlacedCount++
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "2x2",
			Reason:   fmt.Sprintf("squishy=%.2f pathDist=%d", blockSquishyScore(pos, mainPath), mainPath.DirectDistance[pos.Y][pos.X]),
		})
	}

	// Documented fallback: a single cell, used only once no 2x2 site is left.
	for remaining > 0 && len(candidates) > 0 {
		pos, idx := pickFromTopN(candidates, 0.3, 3)
		candidates = append(candidates[:idx], candidates[idx+1:]...)
		if zonerLayer[pos.Y][pos.X] != 0 || touchesLayer(pos, zonerLayer, width, height) {
			continue
		}
		zonerLayer[pos.Y][pos.X] = 1
		candidates = filterAdjacent(candidates, pos)
		remaining--
		debug.PlacedCount++
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "1x1",
			Reason:   fmt.Sprintf("no 2x2 site left; squishy=%.2f pathDist=%d", mainPath.SquishyScore[pos.Y][pos.X], mainPath.DirectDistance[pos.Y][pos.X]),
		})
	}

	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more zoners", remaining),
		})
	}

	return debug
}

// zonerBlockAnchors returns the top-left corners of every zonerSize x zonerSize
// block whose entire footprint is drawn from cells. Because cells has already
// been filtered by every per-cell zoner constraint, a block built this way
// satisfies all of them on all four of its cells.
func zonerBlockAnchors(cells []Point, width, height int) []Point {
	valid := make(map[Point]bool, len(cells))
	for _, c := range cells {
		valid[c] = true
	}

	var anchors []Point
	for _, c := range cells {
		if c.X+zonerSize > width || c.Y+zonerSize > height {
			continue
		}
		full := true
		for dy := 0; dy < zonerSize && full; dy++ {
			for dx := 0; dx < zonerSize; dx++ {
				if !valid[Point{c.X + dx, c.Y + dy}] {
					full = false
					break
				}
			}
		}
		if full {
			anchors = append(anchors, c)
		}
	}
	return anchors
}

// blockSquishyScore is the mean squishy score over a block's footprint — the
// block-shaped counterpart of the per-cell score the 1x1 path ranks on.
func blockSquishyScore(pos Point, mainPath *MainPathData) float64 {
	sum := 0.0
	for dy := 0; dy < zonerSize; dy++ {
		for dx := 0; dx < zonerSize; dx++ {
			sum += mainPath.SquishyScore[pos.Y+dy][pos.X+dx]
		}
	}
	return sum / float64(zonerSize*zonerSize)
}

// hasStaticBlockingPath checks if there is a static obstacle between pos and the nearest main path cell.
// Uses Bresenham-style line check.
func hasStaticBlockingPath(pos Point, mainPath *MainPathData, staticLayer [][]int, width, height int) bool {
	if mainPath == nil {
		return false
	}

	// Find nearest main path cell
	nearestDist := width + height
	nearest := Point{-1, -1}
	searchRadius := zonerMaxPathDist + 1
	for dy := -searchRadius; dy <= searchRadius; dy++ {
		for dx := -searchRadius; dx <= searchRadius; dx++ {
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height && mainPath.OnMainPath[ny][nx] {
				d := abs(dx) + abs(dy)
				if d < nearestDist {
					nearestDist = d
					nearest = Point{nx, ny}
				}
			}
		}
	}

	if nearest.X < 0 {
		return false
	}

	// Walk a line from pos to nearest and check for statics
	return lineHasStatic(pos, nearest, staticLayer, width, height)
}

// lineHasStatic checks if a straight line between two points crosses a static cell.
func lineHasStatic(from, to Point, staticLayer [][]int, width, height int) bool {
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

	err := dx - dy
	x, y := from.X, from.Y

	for {
		// Skip start and end points
		if !(x == from.X && y == from.Y) && !(x == to.X && y == to.Y) {
			if x >= 0 && x < width && y >= 0 && y < height && staticLayer[y][x] == 1 {
				return true
			}
		}
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

	return false
}
