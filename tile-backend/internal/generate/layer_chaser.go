package generate

import (
	"fmt"
	"sort"
)

// GenerateChaserLayer generates the chaser layer.
// Chasers must be on ground, within 0-3 of main path, prefer LOW squishy score.
func GenerateChaserLayer(chaserLayer, ground, softEdge, bridge, rail, staticLayer, zonerLayer [][]int,
	doorPositions map[DoorPosition]Point, mainPath *MainPathData, width, height, targetCount int, regionFilter ...*RegionFilter) *EnemyLayerDebugInfo {

	debug := &EnemyLayerDebugInfo{
		TargetCount: targetCount,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	forbidden := getDoorForbiddenCellsRadius(doorPositions, width, height, doorForbiddenRadius)

	// Get optional region filter
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
			// Cannot overlap zoner
			if zonerLayer[y][x] != 0 {
				continue
			}
			// Must be within chaserMaxPathDist of main path
			if mainPath == nil || mainPath.DirectDistance[y][x] > chaserMaxPathDist {
				continue
			}
			candidates = append(candidates, pos)
		}
	}

	if len(candidates) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{Reason: "no valid positions found"})
		return debug
	}

	// Sort by squishy score ascending (prefer LOW score — easy to reach)
	sort.Slice(candidates, func(i, j int) bool {
		si := mainPath.SquishyScore[candidates[i].Y][candidates[i].X]
		sj := mainPath.SquishyScore[candidates[j].Y][candidates[j].X]
		return si < sj
	})

	remaining := targetCount
	for remaining > 0 && len(candidates) > 0 {
		// Pick randomly from top 30% of candidates (min 3)
		pos, idx := pickFromTopN(candidates, 0.3, 3)
		// Re-check: no adjacent existing chaser (8-directional)
		if touchesLayer(pos, chaserLayer, width, height) {
			candidates = append(candidates[:idx], candidates[idx+1:]...)
			continue
		}
		chaserLayer[pos.Y][pos.X] = 1
		candidates = append(candidates[:idx], candidates[idx+1:]...)
		// Filter out adjacent positions
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
			Reason: fmt.Sprintf("could not place %d more chasers", remaining),
		})
	}

	return debug
}
