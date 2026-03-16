package generate

import (
	"fmt"
	"math/rand"
	"sort"
)

// MobAirStrategy represents placement strategy for mob air
type MobAirStrategy int

const (
	MobAirStrategyCenterOutward MobAirStrategy = iota // Place from center outward
	MobAirStrategyEvenlySpaced                        // Distribute with roughly equal spacing
)

// generateMobAirLayer generates the mob air layer with the given constraints
func generateMobAirLayer(mobAirLayer, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) {

	if targetCount <= 0 {
		return
	}

	// Select strategy randomly
	strategy := MobAirStrategy(rand.Intn(2))

	// Find all valid positions
	validPositions := findValidMobAirPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
		doorPositions, width, height)

	if len(validPositions) == 0 {
		return
	}

	// Sort/arrange positions based on strategy
	centerX, centerY := width/2, height/2

	switch strategy {
	case MobAirStrategyCenterOutward:
		sortMobAirPositionsCenterOutward(validPositions, centerX, centerY)
	case MobAirStrategyEvenlySpaced:
		validPositions = arrangeMobAirEvenlySpaced(validPositions, targetCount, width, height)
	}

	// Place mob air
	remaining := targetCount
	for _, pos := range validPositions {
		if remaining <= 0 {
			break
		}

		// Verify position is still valid (may have been invalidated by previous placements)
		if !isValidMobAirPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
			doorPositions, width, height) {
			continue
		}

		// Place mob air
		mobAirLayer[pos.Y][pos.X] = 1
		remaining--
	}
}

// generateMobAirLayerWithDebug generates the mob air layer with debug info
func generateMobAirLayerWithDebug(mobAirLayer, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) *MobAirDebugInfo {

	debug := &MobAirDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Strategy:    "",
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	if targetCount <= 0 {
		debug.Skipped = true
		debug.SkipReason = "targetCount is 0"
		return debug
	}

	// Select strategy randomly
	strategy := MobAirStrategy(rand.Intn(2))

	switch strategy {
	case MobAirStrategyCenterOutward:
		debug.Strategy = "center_outward"
	case MobAirStrategyEvenlySpaced:
		debug.Strategy = "evenly_spaced"
	}

	// Find all valid positions
	validPositions := findValidMobAirPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
		doorPositions, width, height)

	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid positions found (all positions blocked by static/turret/mobGround/doors or too close to room edges)",
		})
		return debug
	}

	initialValidCount := len(validPositions)

	// Sort/arrange positions based on strategy
	centerX, centerY := width/2, height/2

	switch strategy {
	case MobAirStrategyCenterOutward:
		sortMobAirPositionsCenterOutward(validPositions, centerX, centerY)
	case MobAirStrategyEvenlySpaced:
		validPositions = arrangeMobAirEvenlySpaced(validPositions, targetCount, width, height)
	}

	// Place mob air
	remaining := targetCount
	invalidatedCount := 0
	for _, pos := range validPositions {
		if remaining <= 0 {
			break
		}

		// Verify position is still valid (may have been invalidated by previous placements)
		if !isValidMobAirPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
			doorPositions, width, height) {
			invalidatedCount++
			continue
		}

		// Place mob air
		mobAirLayer[pos.Y][pos.X] = 1
		remaining--
		debug.PlacedCount++

		// Determine placement reason
		reason := fmt.Sprintf("placed via %s strategy", debug.Strategy)
		if ground[pos.Y][pos.X] == 0 {
			reason += " (on void, flying mob)"
		} else {
			reason += " (on ground)"
		}

		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "1x1",
			Reason:   reason,
		})
	}

	// Record miss info
	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated by previous placement (already occupied)",
			Count:  invalidatedCount,
		})
	}
	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all %d valid positions, needed %d more", initialValidCount, remaining),
		})
	}

	return debug
}

// sortMobAirPositionsCenterOutward sorts positions by distance from center (closest first)
func sortMobAirPositionsCenterOutward(positions []Point, centerX, centerY int) {
	center := Point{X: centerX, Y: centerY}
	sort.Slice(positions, func(i, j int) bool {
		return manhattanDistance(positions[i], center) < manhattanDistance(positions[j], center)
	})
}

// arrangeMobAirEvenlySpaced arranges positions to be roughly evenly spaced
// Returns a subset of positions that are well-distributed across the map
// The grid is calculated based on targetCount to ensure even distribution
func arrangeMobAirEvenlySpaced(validPositions []Point, targetCount int, width, height int) []Point {
	if len(validPositions) == 0 || targetCount <= 0 {
		return nil
	}

	if len(validPositions) <= targetCount {
		return validPositions
	}

	// Calculate grid dimensions based on targetCount
	// We want to create a grid where gridCols * gridRows >= targetCount
	// and the grid cells are as square as possible
	gridCols, gridRows := calculateGridDimensions(targetCount, width, height)

	// Create a grid-based selection
	selected := make([]Point, 0, targetCount)
	used := make(map[Point]bool)

	// Calculate cell size
	cellWidth := float64(width) / float64(gridCols)
	cellHeight := float64(height) / float64(gridRows)

	// Iterate through grid cells and find nearest valid position to each cell center
	for row := 0; row < gridRows && len(selected) < targetCount; row++ {
		for col := 0; col < gridCols && len(selected) < targetCount; col++ {
			// Calculate ideal position at cell center
			idealX := int(float64(col)*cellWidth + cellWidth/2)
			idealY := int(float64(row)*cellHeight + cellHeight/2)
			idealPos := Point{X: idealX, Y: idealY}

			// Find nearest valid position to this ideal position
			nearest := findNearestValidPosition(validPositions, idealPos, used)
			if nearest.X >= 0 {
				selected = append(selected, nearest)
				used[nearest] = true
			}
		}
	}

	// If we still need more, fill from remaining valid positions
	for _, pos := range validPositions {
		if len(selected) >= targetCount {
			break
		}
		if !used[pos] {
			selected = append(selected, pos)
			used[pos] = true
		}
	}

	return selected
}

// calculateGridDimensions calculates grid cols and rows based on target count
// The grid is designed to distribute targetCount items evenly across width x height
func calculateGridDimensions(targetCount, width, height int) (cols, rows int) {
	if targetCount <= 0 {
		return 1, 1
	}

	if targetCount == 1 {
		return 1, 1
	}

	// Calculate aspect ratio
	aspectRatio := float64(width) / float64(height)

	// Calculate grid dimensions that:
	// 1. Have cols * rows >= targetCount
	// 2. Maintain aspect ratio similar to room dimensions
	// 3. Create roughly square cells

	// Start with sqrt(targetCount) and adjust for aspect ratio
	sqrtCount := int(float64(targetCount)*0.5 + 0.5)
	if sqrtCount < 1 {
		sqrtCount = 1
	}

	// Adjust cols and rows based on aspect ratio
	if aspectRatio >= 1 {
		// Wider than tall
		cols = int(float64(sqrtCount)*aspectRatio + 0.5)
		if cols < 1 {
			cols = 1
		}
		rows = (targetCount + cols - 1) / cols
		if rows < 1 {
			rows = 1
		}
	} else {
		// Taller than wide
		rows = int(float64(sqrtCount)/aspectRatio + 0.5)
		if rows < 1 {
			rows = 1
		}
		cols = (targetCount + rows - 1) / rows
		if cols < 1 {
			cols = 1
		}
	}

	// Ensure we have enough cells
	for cols*rows < targetCount {
		if float64(width)/float64(cols) > float64(height)/float64(rows) {
			cols++
		} else {
			rows++
		}
	}

	return cols, rows
}

// findNearestValidPosition finds the valid position nearest to the target
func findNearestValidPosition(validPositions []Point, target Point, used map[Point]bool) Point {
	bestPos := Point{X: -1, Y: -1}
	bestDist := 999999

	for _, pos := range validPositions {
		if used[pos] {
			continue
		}

		dist := manhattanDistance(pos, target)
		if dist < bestDist {
			bestDist = dist
			bestPos = pos
		}
	}

	return bestPos
}

