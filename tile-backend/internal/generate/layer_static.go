package generate

import (
	"fmt"
	"sort"
)

// PlacementStrategy represents the strategy for placing statics
type PlacementStrategy int

const (
	StrategyCenterOutward PlacementStrategy = iota // Start from center, spread outward
	StrategyEdgeInward                             // Start from edges, spread inward
)

// generateStaticLayer generates the static layer with the given constraints
// staticLayer: output layer to place statics
// ground: ground layer (static requires ground=1)
// softEdge: soft edge layer (static cannot overlap)
// bridge: bridge layer (static cannot overlap)
// doorPositions: positions of doors
// width, height: dimensions
// targetCount: suggested number of statics to place
func generateStaticLayer(staticLayer, ground, softEdge, bridge [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) {
	// Get all cells within doorForbiddenRadius of any door (forbidden zone)
	forbiddenCells := getDoorForbiddenCells(doorPositions, width, height)

	// Find all valid 2x2 positions for static placement
	validPositions := findValidStaticPositions(ground, softEdge, bridge, staticLayer, forbiddenCells, width, height)
	if len(validPositions) == 0 {
		return
	}

	// Sort positions by strategy (will be re-sorted on each strategy switch)
	centerX, centerY := width/2, height/2
	currentStrategy := StrategyCenterOutward

	remaining := targetCount
	strategyAttempts := 0
	maxStrategyAttempts := 2 * targetCount // Prevent infinite loop

	for remaining > 0 && strategyAttempts < maxStrategyAttempts {
		// Sort valid positions based on current strategy
		sortPositionsByStrategy(validPositions, currentStrategy, centerX, centerY, width, height)

		// Try to place one static
		placed := false
		for i, pos := range validPositions {
			// Check if this position is still valid (may have been invalidated by previous placements)
			if !isValidStaticPosition(pos, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height) {
				continue
			}

			// Check connectivity after placement
			if !checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, pos, width, height) {
				continue
			}

			// Place the static (2x2)
			placeStatic(staticLayer, pos)
			remaining--
			placed = true

			// Remove this position and update valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that now touch this static
			validPositions = filterTouchingPositions(validPositions, pos)
			break
		}

		if !placed {
			// Switch strategy and try again
			strategyAttempts++
		}

		// Alternate strategy after each placement or failed attempt
		if currentStrategy == StrategyCenterOutward {
			currentStrategy = StrategyEdgeInward
		} else {
			currentStrategy = StrategyCenterOutward
		}
	}
}

// generateStaticLayerWithDebug generates the static layer with debug info
func generateStaticLayerWithDebug(staticLayer, ground, softEdge, bridge [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *StaticDebugInfo {
	debug := &StaticDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	// Get all cells within doorForbiddenRadius of any door (forbidden zone)
	forbiddenCells := getDoorForbiddenCells(doorPositions, width, height)

	// Find all valid 2x2 positions for static placement
	validPositions := findValidStaticPositions(ground, softEdge, bridge, staticLayer, forbiddenCells, width, height)
	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid 2x2 positions found (all positions blocked by ground, doors, softEdge, or bridge)",
		})
		return debug
	}

	// Sort positions by strategy (will be re-sorted on each strategy switch)
	centerX, centerY := width/2, height/2
	currentStrategy := StrategyCenterOutward

	remaining := targetCount
	strategyAttempts := 0
	maxStrategyAttempts := 2 * targetCount // Prevent infinite loop
	invalidatedCount := 0
	connectivityBlockedCount := 0

	for remaining > 0 && strategyAttempts < maxStrategyAttempts {
		// Sort valid positions based on current strategy
		sortPositionsByStrategy(validPositions, currentStrategy, centerX, centerY, width, height)

		strategyName := "center_outward"
		if currentStrategy == StrategyEdgeInward {
			strategyName = "edge_inward"
		}

		// Try to place one static
		placed := false
		for i, pos := range validPositions {
			// Check if this position is still valid (may have been invalidated by previous placements)
			if !isValidStaticPosition(pos, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height) {
				invalidatedCount++
				continue
			}

			// Check connectivity after placement
			if !checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			// Place the static (2x2)
			placeStatic(staticLayer, pos)
			remaining--
			placed = true
			debug.PlacedCount++

			// Record placement info
			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "2x2",
				Reason:   fmt.Sprintf("strategy: %s, valid position with connectivity preserved", strategyName),
			})

			// Remove this position and update valid positions
			validPositions = append(validPositions[:i], validPositions[i+1:]...)

			// Filter out positions that now touch this static
			validPositions = filterTouchingPositions(validPositions, pos)
			break
		}

		if !placed {
			// Switch strategy and try again
			strategyAttempts++
		}

		// Alternate strategy after each placement or failed attempt
		if currentStrategy == StrategyCenterOutward {
			currentStrategy = StrategyEdgeInward
		} else {
			currentStrategy = StrategyCenterOutward
		}
	}

	// Record miss info
	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated by previous placement (touching existing static)",
			Count:  invalidatedCount,
		})
	}
	if connectivityBlockedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position would block door connectivity",
			Count:  connectivityBlockedCount,
		})
	}
	if remaining > 0 && strategyAttempts >= maxStrategyAttempts {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("reached max strategy attempts (%d), could not place %d more statics", maxStrategyAttempts, remaining),
		})
	} else if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all %d valid positions, needed %d more", len(validPositions), remaining),
		})
	}

	return debug
}

// generateStaticLayerWithDebugAndRail generates the static layer avoiding rail positions
func generateStaticLayerWithDebugAndRail(staticLayer, ground, softEdge, bridge, rail [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *StaticDebugInfo {
	debug := &StaticDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	// Get all cells within doorForbiddenRadius of any door (forbidden zone)
	forbiddenCells := getDoorForbiddenCells(doorPositions, width, height)

	// Find all valid 2x2 positions for static placement (avoiding rail)
	validPositions := findValidStaticPositionsWithRail(ground, softEdge, bridge, rail, staticLayer, forbiddenCells, width, height)

	// Get rail indent cells (inside rail loop) - these are prioritized
	railIndentCells := GetRailIndentCells(rail, width, height)
	priorityPositions := filterPositionsInRailIndent(validPositions, railIndentCells)

	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid 2x2 positions found (all positions blocked by ground, doors, softEdge, bridge, or rail)",
		})
		return debug
	}

	// Sort positions: prioritize rail indent positions first, then by strategy
	centerX, centerY := width/2, height/2
	currentStrategy := StrategyCenterOutward

	remaining := targetCount
	strategyAttempts := 0
	maxStrategyAttempts := 2 * targetCount
	invalidatedCount := 0
	connectivityBlockedCount := 0
	railBlockedCount := 0

	// First try to place in priority positions (inside rail loop)
	for remaining > 0 && len(priorityPositions) > 0 {
		sortPositionsByStrategy(priorityPositions, currentStrategy, centerX, centerY, width, height)

		placed := false
		for i, pos := range priorityPositions {
			if !isValidStaticPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, forbiddenCells, width, height) {
				invalidatedCount++
				continue
			}

			if !checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			placeStatic(staticLayer, pos)
			remaining--
			placed = true
			debug.PlacedCount++

			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "2x2",
				Reason:   "placed inside rail indent area (prioritized)",
			})

			priorityPositions = append(priorityPositions[:i], priorityPositions[i+1:]...)
			priorityPositions = filterTouchingPositions(priorityPositions, pos)

			// Also remove from validPositions
			validPositions = removePosition(validPositions, pos)
			validPositions = filterTouchingPositions(validPositions, pos)
			break
		}

		if !placed {
			break
		}
	}

	// Then place remaining in regular positions
	for remaining > 0 && strategyAttempts < maxStrategyAttempts {
		sortPositionsByStrategy(validPositions, currentStrategy, centerX, centerY, width, height)

		strategyName := "center_outward"
		if currentStrategy == StrategyEdgeInward {
			strategyName = "edge_inward"
		}

		placed := false
		for i, pos := range validPositions {
			if !isValidStaticPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, forbiddenCells, width, height) {
				invalidatedCount++
				continue
			}

			if !checkConnectivityAfterPlacement(ground, staticLayer, doorPositions, pos, width, height) {
				connectivityBlockedCount++
				continue
			}

			placeStatic(staticLayer, pos)
			remaining--
			placed = true
			debug.PlacedCount++

			debug.Placements = append(debug.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "2x2",
				Reason:   fmt.Sprintf("strategy: %s, avoiding rail positions", strategyName),
			})

			validPositions = append(validPositions[:i], validPositions[i+1:]...)
			validPositions = filterTouchingPositions(validPositions, pos)
			break
		}

		if !placed {
			strategyAttempts++
		}

		if currentStrategy == StrategyCenterOutward {
			currentStrategy = StrategyEdgeInward
		} else {
			currentStrategy = StrategyCenterOutward
		}
	}

	// Record miss info
	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated by previous placement",
			Count:  invalidatedCount,
		})
	}
	if connectivityBlockedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position would block door connectivity",
			Count:  connectivityBlockedCount,
		})
	}
	if railBlockedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position conflicts with rail",
			Count:  railBlockedCount,
		})
	}
	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more statics", remaining),
		})
	}

	return debug
}

// sortPositionsByStrategy sorts positions based on the placement strategy
func sortPositionsByStrategy(positions []Point, strategy PlacementStrategy, centerX, centerY, width, height int) {
	switch strategy {
	case StrategyCenterOutward:
		sort.Slice(positions, func(i, j int) bool {
			return distanceFromCenter(positions[i], centerX, centerY) < distanceFromCenter(positions[j], centerX, centerY)
		})
	case StrategyEdgeInward:
		sort.Slice(positions, func(i, j int) bool {
			return distanceFromEdge(positions[i], width, height) < distanceFromEdge(positions[j], width, height)
		})
	}
}

// distanceFromCenter calculates the Manhattan distance from center
func distanceFromCenter(pos Point, centerX, centerY int) int {
	// Use the center of the 2x2 static
	staticCenterX := pos.X + staticSize/2
	staticCenterY := pos.Y + staticSize/2
	return abs(staticCenterX-centerX) + abs(staticCenterY-centerY)
}

// distanceFromEdge calculates the minimum distance from any edge
func distanceFromEdge(pos Point, width, height int) int {
	// Use the center of the 2x2 static
	staticCenterX := pos.X + staticSize/2
	staticCenterY := pos.Y + staticSize/2

	distLeft := staticCenterX
	distRight := width - 1 - staticCenterX
	distTop := staticCenterY
	distBottom := height - 1 - staticCenterY

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

// placeStatic places a 2x2 static at the given top-left corner
func placeStatic(staticLayer [][]int, pos Point) {
	for dy := 0; dy < staticSize; dy++ {
		for dx := 0; dx < staticSize; dx++ {
			staticLayer[pos.Y+dy][pos.X+dx] = 1
		}
	}
}

// filterTouchingPositions removes positions that would touch the newly placed static
func filterTouchingPositions(positions []Point, placedPos Point) []Point {
	var filtered []Point
	for _, pos := range positions {
		if !wouldTouch(pos, placedPos) {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

