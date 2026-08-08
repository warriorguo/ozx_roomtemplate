package generate

import (
	"fmt"
	"math/rand"
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
// hints may be nil, in which case the default alternating strategy is used.
func generateStaticLayerWithDebugAndRail(staticLayer, ground, softEdge, bridge, rail [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int, hints *StagePlacementHints) *StaticDebugInfo {
	disperse := hints != nil && hints.StaticDisperse

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
	placedPositions := []Point{}

	// First try to place in priority positions (inside rail loop)
	for remaining > 0 && len(priorityPositions) > 0 {
		if disperse {
			sortPositionsDispersed(priorityPositions, placedPositions, width, height)
		} else {
			sortPositionsByStrategy(priorityPositions, currentStrategy, centerX, centerY, width, height)
		}

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
			placedPositions = append(placedPositions, pos)

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
		strategyName := "center_outward"
		if disperse {
			sortPositionsDispersed(validPositions, placedPositions, width, height)
			strategyName = "edge_first_disperse"
		} else {
			sortPositionsByStrategy(validPositions, currentStrategy, centerX, centerY, width, height)
			if currentStrategy == StrategyEdgeInward {
				strategyName = "edge_inward"
			}
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
			placedPositions = append(placedPositions, pos)

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

		// Dispersion re-scores every candidate against what is already placed,
		// so it has no strategy to alternate.
		if !disperse {
			if currentStrategy == StrategyCenterOutward {
				currentStrategy = StrategyEdgeInward
			} else {
				currentStrategy = StrategyCenterOutward
			}
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

// sortPositionsDispersed orders candidates for the edge-first dispersion
// strategy used by the pressure and peak stages (ORT-99).
//
// With nothing placed yet it degenerates to edge-inward, seeding the first
// block against the perimeter. Afterwards it is a farthest-point (Mitchell)
// selection: each candidate is scored by the distance to its nearest already
// placed static and the largest score wins, so blocks repel each other instead
// of alternating in and out from the centre. Ties fall back to the edge
// preference, which keeps cover on the perimeter once spacing is equal.
func sortPositionsDispersed(positions []Point, placed []Point, width, height int) {
	// Shuffle first so equal-score positions get random order
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	if len(placed) == 0 {
		sort.SliceStable(positions, func(i, j int) bool {
			return distanceFromEdge(positions[i], width, height) < distanceFromEdge(positions[j], width, height)
		})
		return
	}

	spread := make(map[Point]int, len(positions))
	for _, pos := range positions {
		spread[pos] = distanceToNearest(pos, placed)
	}

	sort.SliceStable(positions, func(i, j int) bool {
		if spread[positions[i]] != spread[positions[j]] {
			return spread[positions[i]] > spread[positions[j]]
		}
		return distanceFromEdge(positions[i], width, height) < distanceFromEdge(positions[j], width, height)
	})
}

// distanceToNearest returns the Manhattan distance from pos to the closest of
// the already-placed statics. Points are 2x2 top-left corners; since every
// block carries the same offset, corner-to-corner equals center-to-center.
func distanceToNearest(pos Point, placed []Point) int {
	nearest := -1
	for _, p := range placed {
		d := abs(pos.X-p.X) + abs(pos.Y-p.Y)
		if nearest < 0 || d < nearest {
			nearest = d
		}
	}
	return nearest
}

// sortPositionsByStrategy sorts positions based on the placement strategy
func sortPositionsByStrategy(positions []Point, strategy PlacementStrategy, centerX, centerY, width, height int) {
	// Shuffle first so equal-distance positions get random order
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})
	switch strategy {
	case StrategyCenterOutward:
		sort.SliceStable(positions, func(i, j int) bool {
			return distanceFromCenter(positions[i], centerX, centerY) < distanceFromCenter(positions[j], centerX, centerY)
		})
	case StrategyEdgeInward:
		sort.SliceStable(positions, func(i, j int) bool {
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
	placeBlock(staticLayer, pos, staticSize)
}

// filterTouchingPositions removes positions that would touch the newly placed static
func filterTouchingPositions(positions []Point, placedPos Point) []Point {
	return filterTouchingBlocks(positions, placedPos, staticSize)
}
