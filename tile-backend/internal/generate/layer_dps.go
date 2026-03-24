package generate

import (
	"fmt"
	"sort"
)

// GenerateDPSLayer generates the DPS layer.
// DPS must be on ground, within 0-4 of main path. Can be near chaser/static.
func GenerateDPSLayer(dpsLayer, ground, softEdge, bridge, rail, staticLayer, zonerLayer, chaserLayer [][]int,
	doorPositions map[DoorPosition]Point, mainPath *MainPathData, width, height, targetCount int, regionFilter ...*RegionFilter) *EnemyLayerDebugInfo {
	return generateDPSLayerCore(dpsLayer, ground, softEdge, bridge, rail, staticLayer, zonerLayer, chaserLayer,
		doorPositions, mainPath, width, height, targetCount, false, regionFilter...)
}

// GenerateDPSLayerRelaxed is like GenerateDPSLayer but skips the 8-directional
// spacing constraint and allows overlap with chaser cells. It is used as a
// last-resort fallback when strict placement exhausts all spaced candidates
// but the stage minimum has not been met.
func GenerateDPSLayerRelaxed(dpsLayer, ground, softEdge, bridge, rail, staticLayer, zonerLayer, chaserLayer [][]int,
	doorPositions map[DoorPosition]Point, mainPath *MainPathData, width, height, targetCount int) *EnemyLayerDebugInfo {
	return generateDPSLayerCore(dpsLayer, ground, softEdge, bridge, rail, staticLayer, zonerLayer, chaserLayer,
		doorPositions, mainPath, width, height, targetCount, true)
}

// generateDPSLayerCore is the shared implementation. When relaxSpacing is true the
// 8-directional spacing constraint and chaser-overlap check are not enforced —
// this allows meeting minimum counts in constrained rooms.
func generateDPSLayerCore(dpsLayer, ground, softEdge, bridge, rail, staticLayer, zonerLayer, chaserLayer [][]int,
	doorPositions map[DoorPosition]Point, mainPath *MainPathData, width, height, targetCount int, relaxSpacing bool, regionFilter ...*RegionFilter) *EnemyLayerDebugInfo {

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
			// Cannot overlap zoner (but CAN overlap chaser adjacency — relaxed constraint)
			if zonerLayer[y][x] != 0 {
				continue
			}
			// Must be within dpsMaxPathDist of main path
			if mainPath == nil || mainPath.DirectDistance[y][x] > dpsMaxPathDist {
				continue
			}
			// In relaxed mode, skip candidates that already have a DPS placed
			if relaxSpacing && dpsLayer[y][x] != 0 {
				continue
			}
			candidates = append(candidates, pos)
		}
	}

	if len(candidates) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{Reason: "no valid positions found"})
		return debug
	}

	// Score: prefer positions near chaser or static (support role)
	// Also prefer moderate squishy score
	sort.Slice(candidates, func(i, j int) bool {
		pi, pj := candidates[i], candidates[j]
		si := dpsScore(pi, mainPath, chaserLayer, staticLayer, width, height)
		sj := dpsScore(pj, mainPath, chaserLayer, staticLayer, width, height)
		return si > sj // higher score = better
	})

	remaining := targetCount
	for remaining > 0 && len(candidates) > 0 {
		pos, idx := pickFromTopN(candidates, 0.3, 3)

		if !relaxSpacing {
			// No adjacent existing DPS and cannot overlap chaser
			if touchesLayer(pos, dpsLayer, width, height) || chaserLayer[pos.Y][pos.X] != 0 {
				candidates = append(candidates[:idx], candidates[idx+1:]...)
				continue
			}
		}

		dpsLayer[pos.Y][pos.X] = 1
		candidates = append(candidates[:idx], candidates[idx+1:]...)

		if !relaxSpacing {
			candidates = filterAdjacent(candidates, pos)
		}

		remaining--
		debug.PlacedCount++
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "1x1",
			Reason:   fmt.Sprintf("squishy=%.2f pathDist=%d relaxed=%v", mainPath.SquishyScore[pos.Y][pos.X], mainPath.DirectDistance[pos.Y][pos.X], relaxSpacing),
		})
	}

	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more DPS", remaining),
		})
	}

	return debug
}

// dpsScore computes a placement preference score for DPS.
// Higher score = better. Prefers proximity to chaser/static.
func dpsScore(pos Point, mainPath *MainPathData, chaserLayer, staticLayer [][]int, width, height int) float64 {
	score := 0.0

	// Bonus for being near chaser or static
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if chaserLayer[ny][nx] == 1 {
					score += 3.0
				}
				if staticLayer[ny][nx] == 1 {
					score += 2.0
				}
			}
		}
	}

	// Moderate squishy score bonus
	if mainPath != nil {
		score += mainPath.SquishyScore[pos.Y][pos.X] * 0.5
	}

	return score
}
