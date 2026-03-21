package generate

import (
	"fmt"
	"sort"
)

// GenerateZonerLayer generates the zoner layer.
// Zoners must be on ground, within 0-5 of main path, prefer HIGH squishy score,
// and no static between zoner and main path.
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

	// Find valid positions
	var candidates []Point
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
			// No static between this position and main path
			if hasStaticBlockingPath(pos, mainPath, staticLayer, width, height) {
				continue
			}
			candidates = append(candidates, pos)
		}
	}

	// Fallback: if no candidates found with LOS constraint, retry without static-blocking-path filter.
	// This ensures at least some valid positions exist in heavily-static rooms.
	if len(candidates) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{Reason: "no valid positions with LOS constraint, retrying without static-blocking-path filter"})
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if !rf.Contains(x, y) {
					continue
				}
				pos := Point{x, y}
				if !isValidEnemyPosition(pos, ground, softEdge, bridge, rail, staticLayer, forbidden, width, height) {
					continue
				}
				if mainPath == nil || mainPath.DirectDistance[y][x] > zonerMaxPathDist {
					continue
				}
				candidates = append(candidates, pos)
			}
		}
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

	remaining := targetCount
	for remaining > 0 && len(candidates) > 0 {
		pos, idx := pickFromTopN(candidates, 0.3, 3)
		if touchesLayer(pos, zonerLayer, width, height) {
			candidates = append(candidates[:idx], candidates[idx+1:]...)
			continue
		}
		zonerLayer[pos.Y][pos.X] = 1
		candidates = append(candidates[:idx], candidates[idx+1:]...)
		candidates = filterAdjacent(candidates, pos)
		remaining--
		debug.PlacedCount++
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "1x1",
			Reason:   fmt.Sprintf("squishy=%.2f pathDist=%d", mainPath.SquishyScore[pos.Y][pos.X], mainPath.DirectDistance[pos.Y][pos.X]),
		})
	}

	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more zoners", remaining),
		})
	}

	return debug
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
