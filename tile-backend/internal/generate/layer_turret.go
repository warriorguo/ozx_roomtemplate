package generate

import (
	"fmt"
	"sort"
)

// generateTurretLayer generates the turret layer with the given constraints
// turretLayer: output layer to place turrets
// ground: ground layer (turret requires ground=1)
// softEdge: soft edge layer (turret cannot overlap)
// bridge: bridge layer (turret cannot overlap)
// staticLayer: static layer (turret cannot overlap)
// doorPositions: positions of doors
// width, height: dimensions
// targetCount: suggested number of turrets to place
func generateTurretLayer(turretLayer, ground, softEdge, bridge, staticLayer [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) {
	// Find all valid positions for turret placement
	validPositions := findValidTurretPositions(ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height)
	if len(validPositions) == 0 {
		return
	}

	// Sort positions by preference (ground corners first, then room corners and edges, then by distance from center)
	centerX, centerY := width/2, height/2
	sortTurretPositionsByPreference(validPositions, centerX, centerY, width, height, ground)

	remaining := targetCount
	maxAttempts := 2 * targetCount // Prevent infinite loop

	for remaining > 0 && maxAttempts > 0 {
		placed := false

		for i, pos := range validPositions {
			// Check if this position is still valid
			if !isValidTurretPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height) {
				continue
			}

			// Check connectivity after placement
			if !checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer, doorPositions, pos, width, height) {
				continue
			}

			// Place the turret (1 tile)
			turretLayer[pos.Y][pos.X] = 1
			remaining--
			placed = true

			// Remove this position from valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that are too close to this turret
			validPositions = filterTurretsTooClose(validPositions, pos)
			break
		}

		if !placed {
			maxAttempts--
		}
	}
}

// generateTurretLayerWithDebug generates the turret layer with debug info
func generateTurretLayerWithDebug(turretLayer, ground, softEdge, bridge, staticLayer [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *TurretDebugInfo {
	debug := &TurretDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	// Find all valid positions for turret placement
	validPositions := findValidTurretPositions(ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height)
	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid positions found (all positions blocked by ground, doors, static, softEdge, or bridge)",
		})
		return debug
	}

	initialValidCount := len(validPositions)

	// Sort positions by preference (ground corners first, then room corners and edges, then by distance from center)
	centerX, centerY := width/2, height/2
	sortTurretPositionsByPreference(validPositions, centerX, centerY, width, height, ground)

	remaining := targetCount
	maxAttempts := 2 * targetCount // Prevent infinite loop
	invalidatedCount := 0
	connectivityBlockedCount := 0

	for remaining > 0 && maxAttempts > 0 {
		placed := false

		for i, pos := range validPositions {
			// Check if this position is still valid
			if !isValidTurretPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height) {
				invalidatedCount++
				continue
			}

			// Check connectivity after placement
			if !checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			// Determine reason based on position characteristics
			reason := getTurretPlacementReason(pos, centerX, centerY, width, height, ground)

			// Place the turret (1 tile)
			turretLayer[pos.Y][pos.X] = 1
			remaining--
			placed = true
			debug.PlacedCount++

			// Record placement info
			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "1x1",
				Reason:   reason,
			})

			// Remove this position from valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that are too close to this turret
			validPositions = filterTurretsTooClose(validPositions, pos)
			break
		}

		if !placed {
			maxAttempts--
		}
	}

	// Record miss info
	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated (too close to existing turret or blocked)",
			Count:  invalidatedCount,
		})
	}
	if connectivityBlockedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position would block door connectivity",
			Count:  connectivityBlockedCount,
		})
	}
	if remaining > 0 && maxAttempts <= 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("reached max attempts, could not place %d more turrets", remaining),
		})
	} else if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all %d valid positions, needed %d more", initialValidCount, remaining),
		})
	}

	return debug
}

// generateTurretLayerWithDebugAndRail generates the turret layer avoiding rail positions
func generateTurretLayerWithDebugAndRail(turretLayer, ground, softEdge, bridge, rail, staticLayer [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *TurretDebugInfo {
	debug := &TurretDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	// Find all valid positions for turret placement (avoiding rail)
	validPositions := findValidTurretPositionsWithRail(ground, softEdge, bridge, rail, staticLayer, turretLayer, doorPositions, width, height)

	// Get rail indent cells - these are prioritized
	railIndentCells := GetRailIndentCells(rail, width, height)
	priorityPositions := filterSinglePositionsInRailIndent(validPositions, railIndentCells)

	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid positions found (all positions blocked by ground, doors, static, softEdge, bridge, or rail)",
		})
		return debug
	}

	centerX, centerY := width/2, height/2
	sortTurretPositionsByPreference(validPositions, centerX, centerY, width, height, ground)

	remaining := targetCount
	maxAttempts := 2 * targetCount
	invalidatedCount := 0
	connectivityBlockedCount := 0

	// First try priority positions
	for remaining > 0 && len(priorityPositions) > 0 {
		placed := false
		for i, pos := range priorityPositions {
			if !isValidTurretPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, turretLayer, doorPositions, width, height) {
				invalidatedCount++
				continue
			}

			if !checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			turretLayer[pos.Y][pos.X] = 1
			remaining--
			placed = true
			debug.PlacedCount++

			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "1x1",
				Reason:   "placed inside rail indent area (prioritized)",
			})

			priorityPositions = append(priorityPositions[:i], priorityPositions[i+1:]...)
			priorityPositions = filterTurretsTooClose(priorityPositions, pos)
			validPositions = removeSinglePosition(validPositions, pos)
			validPositions = filterTurretsTooClose(validPositions, pos)
			break
		}

		if !placed {
			break
		}
	}

	// Then place remaining
	for remaining > 0 && maxAttempts > 0 {
		placed := false

		for i, pos := range validPositions {
			if !isValidTurretPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, turretLayer, doorPositions, width, height) {
				invalidatedCount++
				continue
			}

			if !checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			reason := getTurretPlacementReason(pos, centerX, centerY, width, height, ground) + " (avoiding rail)"

			turretLayer[pos.Y][pos.X] = 1
			remaining--
			placed = true
			debug.PlacedCount++

			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "1x1",
				Reason:   reason,
			})

			validPositions = append(validPositions[:i], validPositions[i+1:]...)
			validPositions = filterTurretsTooClose(validPositions, pos)
			break
		}

		if !placed {
			maxAttempts--
		}
	}

	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated",
			Count:  invalidatedCount,
		})
	}
	if connectivityBlockedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position would block door connectivity",
			Count:  connectivityBlockedCount,
		})
	}
	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more turrets", remaining),
		})
	}

	return debug
}

// sortTurretPositionsByPreference sorts positions by placement preference
// Priority: ground corners (90°/270°) first, then room corners and edges, then by distance from center
func sortTurretPositionsByPreference(positions []Point, centerX, centerY, width, height int, ground [][]int) {
	sort.Slice(positions, func(i, j int) bool {
		si := calculateTurretPreferenceScore(positions[i], centerX, centerY, width, height, ground)
		sj := calculateTurretPreferenceScore(positions[j], centerX, centerY, width, height, ground)
		return si < sj
	})
}

// calculateTurretPreferenceScore calculates a preference score for turret placement
// Lower score means higher preference
func calculateTurretPreferenceScore(pos Point, centerX, centerY, width, height int, ground [][]int) int {
	// Calculate distance from center
	distToCenter := abs(pos.X-centerX) + abs(pos.Y-centerY)

	// Highest priority: ground right angles (90°) and inner corners (270°)
	// This is where ground forms an L-shape
	groundCornerType := getGroundCornerType(pos, ground, width, height)
	if groundCornerType == CornerType90 || groundCornerType == CornerType270 {
		return -200 + distToCenter // Strongest preference for ground corners
	}

	// Calculate distance from edges
	distToEdge := minDistanceToEdge(pos, width, height)

	// Prefer positions near edges (within turretEdgePreference) or corners
	// But also prefer positions closer to center among valid positions
	edgeBonus := 0
	if distToEdge <= turretEdgePreference {
		edgeBonus = -100 // Strong preference for edge positions
	}

	// Check if it's a corner-like position (near two edges)
	isCornerLike := isNearCorner(pos, width, height, turretEdgePreference)
	if isCornerLike {
		edgeBonus -= 50 // Extra bonus for room corners
	}

	// Combine: edge bonus + distance from center (prefer closer to center among valid positions)
	return edgeBonus + distToCenter
}

// minDistanceToEdge calculates the minimum distance to any edge
func minDistanceToEdge(pos Point, width, height int) int {
	distLeft := pos.X
	distRight := width - 1 - pos.X
	distTop := pos.Y
	distBottom := height - 1 - pos.Y

	minDist := distLeft
	if distRight < minDist {
		minDist = distRight
	}
	if distTop < minDist {
		minDist = distTop
	}
	if distBottom < minDist {
		minDist = distBottom
	}
	return minDist
}

// isNearCorner checks if the position is near a corner
func isNearCorner(pos Point, width, height, threshold int) bool {
	// Near top-left
	if pos.X <= threshold && pos.Y <= threshold {
		return true
	}
	// Near top-right
	if pos.X >= width-1-threshold && pos.Y <= threshold {
		return true
	}
	// Near bottom-left
	if pos.X <= threshold && pos.Y >= height-1-threshold {
		return true
	}
	// Near bottom-right
	if pos.X >= width-1-threshold && pos.Y >= height-1-threshold {
		return true
	}
	return false
}

// filterTurretsTooClose removes positions that are too close to the newly placed turret
func filterTurretsTooClose(positions []Point, placedPos Point) []Point {
	var filtered []Point
	for _, pos := range positions {
		dist := manhattanDistance(pos, placedPos)
		if dist >= turretMinTurretDistance {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

// getTurretPlacementReason returns a human-readable reason for turret placement
func getTurretPlacementReason(pos Point, centerX, centerY, width, height int, ground [][]int) string {
	cornerType := getGroundCornerType(pos, ground, width, height)
	if cornerType == CornerType90 {
		return "ground corner (90° right angle)"
	}
	if cornerType == CornerType270 {
		return "ground corner (270° inner corner)"
	}

	if isNearCorner(pos, width, height, turretEdgePreference) {
		return "near room corner"
	}

	distToEdge := minDistanceToEdge(pos, width, height)
	if distToEdge <= turretEdgePreference {
		return "near room edge"
	}

	return "center outward placement"
}

