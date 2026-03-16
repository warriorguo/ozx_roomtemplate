package generate

import (
	"fmt"
	"math/rand"
	"sort"
)

// MobGroundStrategy represents placement strategy for mob ground
type MobGroundStrategy int

const (
	MobGroundStrategyLargeOpenArea MobGroundStrategy = iota // Place in large open area from center outward
	MobGroundStrategyNearDoors                              // Place near doors
	MobGroundStrategyCenterOutward                          // Place from center outward
)

// generateMobGroundLayer generates the mob ground layer with the given constraints
func generateMobGroundLayer(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) {

	if targetCount <= 0 {
		return
	}

	// Step 1: Divide count into 2-3 groups
	groups := divideMobGroundIntoGroups(targetCount)
	if len(groups) == 0 {
		return
	}

	// Step 2: Select strategies for each group (no duplicates)
	availableStrategies := []MobGroundStrategy{
		MobGroundStrategyLargeOpenArea,
		MobGroundStrategyNearDoors,
		MobGroundStrategyCenterOutward,
	}

	// Shuffle strategies
	for i := len(availableStrategies) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		availableStrategies[i], availableStrategies[j] = availableStrategies[j], availableStrategies[i]
	}

	// Check if large open area strategy is viable
	largeOpenAreaCenter := findLargeOpenAreaCenter(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height)
	if largeOpenAreaCenter.X < 0 {
		// Remove large open area strategy if not viable
		for i, s := range availableStrategies {
			if s == MobGroundStrategyLargeOpenArea {
				availableStrategies = append(availableStrategies[:i], availableStrategies[i+1:]...)
				break
			}
		}
	}

	// Step 3: Execute placement for each group
	centerX, centerY := width/2, height/2

	for i, groupCount := range groups {
		if i >= len(availableStrategies) {
			break
		}

		strategy := availableStrategies[i]
		placed := executeMobGroundStrategy(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer,
			doorPositions, width, height, groupCount, strategy, centerX, centerY, largeOpenAreaCenter)

		if placed == 0 {
			// Strategy failed, continue with next group
			continue
		}
	}
}

// generateMobGroundLayerWithDebug generates the mob ground layer with debug info
func generateMobGroundLayerWithDebug(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int) *MobGroundDebugInfo {

	debug := &MobGroundDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Groups:      []MobGroupInfo{},
		Misses:      []MissInfo{},
	}

	if targetCount <= 0 {
		debug.Skipped = true
		debug.SkipReason = "targetCount is 0"
		return debug
	}

	// Step 1: Divide count into 2-3 groups
	groups := divideMobGroundIntoGroups(targetCount)
	if len(groups) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "failed to divide target count into groups",
		})
		return debug
	}

	// Step 2: Select strategies for each group (no duplicates)
	availableStrategies := []MobGroundStrategy{
		MobGroundStrategyLargeOpenArea,
		MobGroundStrategyNearDoors,
		MobGroundStrategyCenterOutward,
	}

	// Shuffle strategies
	for i := len(availableStrategies) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		availableStrategies[i], availableStrategies[j] = availableStrategies[j], availableStrategies[i]
	}

	// Check if large open area strategy is viable
	largeOpenAreaCenter := findLargeOpenAreaCenter(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height)
	if largeOpenAreaCenter.X < 0 {
		// Remove large open area strategy if not viable
		for i, s := range availableStrategies {
			if s == MobGroundStrategyLargeOpenArea {
				availableStrategies = append(availableStrategies[:i], availableStrategies[i+1:]...)
				break
			}
		}
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "large_open_area strategy not viable (no 4x4 open area found)",
		})
	}

	// Step 3: Execute placement for each group
	centerX, centerY := width/2, height/2

	for i, groupCount := range groups {
		if i >= len(availableStrategies) {
			debug.Misses = append(debug.Misses, MissInfo{
				Reason: fmt.Sprintf("group %d skipped: no more strategies available", i),
			})
			break
		}

		strategy := availableStrategies[i]
		strategyName := getMobGroundStrategyName(strategy)

		groupDebug := MobGroupInfo{
			GroupIndex:  i,
			Strategy:    strategyName,
			TargetCount: groupCount,
			PlacedCount: 0,
			Placements:  []PlaceInfo{},
			Misses:      []MissInfo{},
		}

		placed := executeMobGroundStrategyWithDebug(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer,
			doorPositions, width, height, groupCount, strategy, centerX, centerY, largeOpenAreaCenter, &groupDebug)

		groupDebug.PlacedCount = placed
		debug.PlacedCount += placed
		debug.Groups = append(debug.Groups, groupDebug)
	}

	return debug
}

// generateMobGroundLayerWithDebugAndRail generates the mob ground layer avoiding rail positions
func generateMobGroundLayerWithDebugAndRail(mobGroundLayer, ground, softEdge, bridge, rail, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *MobGroundDebugInfo {
	debug := &MobGroundDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Groups:      []MobGroupInfo{},
		Misses:      []MissInfo{},
	}

	// Find all valid positions (avoiding rail)
	validPositions := findValidMobGroundPositionsWithRail(ground, softEdge, bridge, rail, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height)

	// Get rail indent cells - these are prioritized
	railIndentCells := GetRailIndentCells(rail, width, height)
	priorityPositions := filterSinglePositionsInRailIndent(validPositions, railIndentCells)

	if len(validPositions) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid positions found (all blocked by ground, doors, static, turret, softEdge, bridge, or rail)",
		})
		return debug
	}

	remaining := targetCount
	invalidatedCount := 0

	// First try priority positions
	for remaining > 0 && len(priorityPositions) > 0 {
		placed := false
		for i, pos := range priorityPositions {
			if !isValidMobGroundPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				invalidatedCount++
				continue
			}

			mobGroundLayer[pos.Y][pos.X] = 1
			remaining--
			placed = true
			debug.PlacedCount++

			debug.Groups = append(debug.Groups, MobGroupInfo{
				GroupIndex:  len(debug.Groups),
				Strategy:    "rail_indent_priority",
				TargetCount: 1,
				PlacedCount: 1,
				Placements: []PlaceInfo{{
					Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
					Size:     "1x1",
					Reason:   "placed inside rail indent area",
				}},
			})

			priorityPositions = append(priorityPositions[:i], priorityPositions[i+1:]...)
			validPositions = removeSinglePosition(validPositions, pos)
			break
		}

		if !placed {
			break
		}
	}

	// Then place remaining using cluster strategy
	if remaining > 0 {
		rand.Shuffle(len(validPositions), func(i, j int) {
			validPositions[i], validPositions[j] = validPositions[j], validPositions[i]
		})

		groupInfo := MobGroupInfo{
			GroupIndex:  len(debug.Groups),
			Strategy:    "random_avoiding_rail",
			TargetCount: remaining,
			Placements:  []PlaceInfo{},
		}

		for remaining > 0 && len(validPositions) > 0 {
			pos := validPositions[0]
			validPositions = validPositions[1:]

			if !isValidMobGroundPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				invalidatedCount++
				continue
			}

			mobGroundLayer[pos.Y][pos.X] = 1
			remaining--
			debug.PlacedCount++
			groupInfo.PlacedCount++

			groupInfo.Placements = append(groupInfo.Placements, PlaceInfo{
				Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
				Size:     "1x1",
				Reason:   "random placement avoiding rail",
			})
		}

		debug.Groups = append(debug.Groups, groupInfo)
	}

	if invalidatedCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "position invalidated",
			Count:  invalidatedCount,
		})
	}
	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more mob ground", remaining),
		})
	}

	return debug
}

// divideMobGroundIntoGroups divides the target count into 2-3 groups
func divideMobGroundIntoGroups(targetCount int) []int {
	if targetCount <= 0 {
		return nil
	}

	if targetCount == 1 {
		return []int{1}
	}

	if targetCount == 2 {
		return []int{1, 1}
	}

	// Try 3 groups first
	groupCount := 3
	if targetCount < 3 {
		groupCount = 2
	}

	baseSize := targetCount / groupCount
	remainder := targetCount % groupCount

	groups := make([]int, groupCount)
	for i := 0; i < groupCount; i++ {
		groups[i] = baseSize
		if i < remainder {
			groups[i]++
		}
	}

	// Merge groups if any has less than 1
	for i := len(groups) - 1; i >= 0; i-- {
		if groups[i] < 1 && len(groups) > 1 {
			if i > 0 {
				groups[i-1] += groups[i]
			} else if len(groups) > 1 {
				groups[1] += groups[0]
			}
			groups = append(groups[:i], groups[i+1:]...)
		}
	}

	return groups
}

// executeMobGroundStrategy executes a specific placement strategy
func executeMobGroundStrategy(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int,
	strategy MobGroundStrategy, centerX, centerY int, largeOpenAreaCenter Point) int {

	placed := 0
	remaining := targetCount
	maxAttempts := targetCount * 3

	for remaining > 0 && maxAttempts > 0 {
		// Find valid positions based on strategy
		validPositions := findValidMobGroundPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer,
			doorPositions, width, height)

		if len(validPositions) == 0 {
			break
		}

		// Sort positions based on strategy
		sortMobGroundPositionsByStrategy(validPositions, strategy, centerX, centerY, width, height, doorPositions, largeOpenAreaCenter)

		// Try to place (prefer 2x2, fallback to 1x1)
		placedOne := false
		for _, pos := range validPositions {
			// Try 2x2 first
			if canPlace2x2MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				place2x2MobGround(mobGroundLayer, pos)
				placed++
				remaining--
				placedOne = true
				break
			}

			// Try 1x1
			if canPlace1x1MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				mobGroundLayer[pos.Y][pos.X] = 1
				placed++
				remaining--
				placedOne = true
				break
			}
		}

		if !placedOne {
			maxAttempts--
		}
	}

	return placed
}

// executeMobGroundStrategyWithDebug executes a specific placement strategy with debug info
func executeMobGroundStrategyWithDebug(mobGroundLayer, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int,
	strategy MobGroundStrategy, centerX, centerY int, largeOpenAreaCenter Point, groupDebug *MobGroupInfo) int {

	placed := 0
	remaining := targetCount
	maxAttempts := targetCount * 3
	noValidPositionsCount := 0
	no2x2Or1x1Count := 0

	for remaining > 0 && maxAttempts > 0 {
		// Find valid positions based on strategy
		validPositions := findValidMobGroundPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer,
			doorPositions, width, height)

		if len(validPositions) == 0 {
			noValidPositionsCount++
			break
		}

		// Sort positions based on strategy
		sortMobGroundPositionsByStrategy(validPositions, strategy, centerX, centerY, width, height, doorPositions, largeOpenAreaCenter)

		// Try to place (prefer 2x2, fallback to 1x1)
		placedOne := false
		for _, pos := range validPositions {
			// Try 2x2 first
			if canPlace2x2MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				place2x2MobGround(mobGroundLayer, pos)
				placed++
				remaining--
				placedOne = true

				groupDebug.Placements = append(groupDebug.Placements, PlaceInfo{
					Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
					Size:     "2x2",
					Reason:   fmt.Sprintf("preferred 2x2 placement via %s strategy", getMobGroundStrategyName(strategy)),
				})
				break
			}

			// Try 1x1
			if canPlace1x1MobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				mobGroundLayer[pos.Y][pos.X] = 1
				placed++
				remaining--
				placedOne = true

				groupDebug.Placements = append(groupDebug.Placements, PlaceInfo{
					Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
					Size:     "1x1",
					Reason:   fmt.Sprintf("fallback 1x1 placement via %s strategy", getMobGroundStrategyName(strategy)),
				})
				break
			}
		}

		if !placedOne {
			no2x2Or1x1Count++
			maxAttempts--
		}
	}

	// Record miss info for this group
	if noValidPositionsCount > 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: "no valid positions available (all blocked by ground/static/turret/doors/existing mobs)",
		})
	}
	if no2x2Or1x1Count > 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: "positions found but neither 2x2 nor 1x1 placement possible",
			Count:  no2x2Or1x1Count,
		})
	}
	if remaining > 0 && maxAttempts <= 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: fmt.Sprintf("reached max attempts, could not place %d more mobs", remaining),
		})
	} else if remaining > 0 {
		groupDebug.Misses = append(groupDebug.Misses, MissInfo{
			Reason: fmt.Sprintf("exhausted all valid positions, needed %d more", remaining),
		})
	}

	return placed
}

// place2x2MobGround places a 2x2 mob ground at the given top-left corner
func place2x2MobGround(mobGroundLayer [][]int, pos Point) {
	for dy := 0; dy < 2; dy++ {
		for dx := 0; dx < 2; dx++ {
			mobGroundLayer[pos.Y+dy][pos.X+dx] = 1
		}
	}
}

// sortMobGroundPositionsByStrategy sorts positions based on the placement strategy
func sortMobGroundPositionsByStrategy(positions []Point, strategy MobGroundStrategy, centerX, centerY, width, height int,
	doorPositions map[DoorPosition]Point, largeOpenAreaCenter Point) {

	switch strategy {
	case MobGroundStrategyLargeOpenArea:
		// Sort by distance from large open area center (closest first)
		sort.Slice(positions, func(i, j int) bool {
			return manhattanDistance(positions[i], largeOpenAreaCenter) < manhattanDistance(positions[j], largeOpenAreaCenter)
		})

	case MobGroundStrategyNearDoors:
		// Sort by distance from nearest door (closest first)
		sort.Slice(positions, func(i, j int) bool {
			return minDistanceToDoor(positions[i], doorPositions) < minDistanceToDoor(positions[j], doorPositions)
		})

	case MobGroundStrategyCenterOutward:
		// Sort by distance from center (closest first)
		center := Point{X: centerX, Y: centerY}
		sort.Slice(positions, func(i, j int) bool {
			return manhattanDistance(positions[i], center) < manhattanDistance(positions[j], center)
		})
	}
}

// minDistanceToDoor returns the minimum distance to any door
func minDistanceToDoor(pos Point, doorPositions map[DoorPosition]Point) int {
	minDist := 999999
	for _, doorPos := range doorPositions {
		dist := manhattanDistance(pos, doorPos)
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

// findLargeOpenAreaCenter finds the center of a large open area connected to doors
// Returns Point{-1, -1} if no suitable large open area exists
func findLargeOpenAreaCenter(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) Point {

	// Find all connected regions of walkable ground
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	// Find the largest region that is connected to at least one door
	var bestCenter Point
	bestSize := 0
	minSizeThreshold := (width * height) / 10 // At least 10% of total area

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if visited[y][x] {
				continue
			}

			pos := Point{X: x, Y: y}
			if !isWalkableForMobGround(pos, ground, softEdge, bridge, staticLayer, turretLayer, width, height) {
				visited[y][x] = true
				continue
			}

			// BFS to find connected region
			region := findConnectedRegion(pos, ground, softEdge, bridge, staticLayer, turretLayer, visited, width, height)
			if len(region) == 0 {
				continue
			}

			// Check if region is connected to any door
			connectedToDoor := false
			for _, doorPos := range doorPositions {
				for _, regionPos := range region {
					if manhattanDistance(regionPos, doorPos) <= 2 {
						connectedToDoor = true
						break
					}
				}
				if connectedToDoor {
					break
				}
			}

			if connectedToDoor && len(region) > bestSize && len(region) >= minSizeThreshold {
				bestSize = len(region)
				bestCenter = calculateRegionCenter(region)
			}
		}
	}

	if bestSize == 0 {
		return Point{X: -1, Y: -1}
	}

	return bestCenter
}

// findConnectedRegion finds all cells connected to the starting point
func findConnectedRegion(start Point, ground, softEdge, bridge, staticLayer, turretLayer [][]int,
	visited [][]bool, width, height int) []Point {

	var region []Point
	queue := []Point{start}
	visited[start.Y][start.X] = true

	dxArr := []int{0, 1, 0, -1}
	dyArr := []int{-1, 0, 1, 0}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		region = append(region, curr)

		for i := 0; i < 4; i++ {
			nx, ny := curr.X+dxArr[i], curr.Y+dyArr[i]
			if nx >= 0 && nx < width && ny >= 0 && ny < height && !visited[ny][nx] {
				nextPos := Point{X: nx, Y: ny}
				if isWalkableForMobGround(nextPos, ground, softEdge, bridge, staticLayer, turretLayer, width, height) {
					visited[ny][nx] = true
					queue = append(queue, nextPos)
				}
			}
		}
	}

	return region
}

// calculateRegionCenter calculates the center of a region
func calculateRegionCenter(region []Point) Point {
	if len(region) == 0 {
		return Point{X: -1, Y: -1}
	}

	sumX, sumY := 0, 0
	for _, pos := range region {
		sumX += pos.X
		sumY += pos.Y
	}

	return Point{
		X: sumX / len(region),
		Y: sumY / len(region),
	}
}

// getMobGroundStrategyName returns the name of a mob ground strategy
func getMobGroundStrategyName(strategy MobGroundStrategy) string {
	switch strategy {
	case MobGroundStrategyLargeOpenArea:
		return "large_open_area"
	case MobGroundStrategyNearDoors:
		return "near_doors"
	case MobGroundStrategyCenterOutward:
		return "center_outward"
	default:
		return "unknown"
	}
}

