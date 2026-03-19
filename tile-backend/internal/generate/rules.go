package generate

import "fmt"

// ============================================================================
// Validation constants
// ============================================================================

// Static placement size (fixed 2x2)
const staticSize = 2

// Unified door forbidden radius
const doorForbiddenRadius = 2

// Legacy turret/mobground constants (kept for backward-compatible validation helpers)
const (
	turretMinDoorDistance    = 4
	turretMinTurretDistance  = 2
	mobGroundMinDoorDistance = 2
)

// Enemy placement distance ranges from main path
const (
	chaserMaxPathDist = 3 // Chaser: 0-3 from main path
	zonerMaxPathDist  = 5 // Zoner: 0-5 from main path
	dpsMaxPathDist    = 4 // DPS: 0-4 from main path
)

// Mob air constants
const (
	mobAirMinDoorDistance = 4 // Minimum distance from doors
	mobAirMinEdgeDistance = 2 // Minimum distance from room edges
)

// Soft edge constants
const (
	softEdgeMinDoorDistance = 2 // Minimum distance from doors
)

// ============================================================================
// Static validation
// ============================================================================

// isValidStaticPosition checks if a 2x2 static can be placed at the given top-left corner
func isValidStaticPosition(pos Point, ground, softEdge, bridge, staticLayer [][]int, forbiddenCells map[Point]bool, width, height int) bool {
	// Check all 4 cells of the 2x2 area
	for dy := 0; dy < staticSize; dy++ {
		for dx := 0; dx < staticSize; dx++ {
			x := pos.X + dx
			y := pos.Y + dy

			// Check bounds
			if x >= width || y >= height {
				return false
			}

			// Check ground layer (must be 1)
			if ground[y][x] != 1 {
				return false
			}

			// Check soft edge (must be 0)
			if softEdge[y][x] != 0 {
				return false
			}

			// Check bridge (must be 0)
			if bridge[y][x] != 0 {
				return false
			}

			// Check existing static (must be 0)
			if staticLayer[y][x] != 0 {
				return false
			}

			// Check forbidden cells (door area)
			if forbiddenCells[Point{X: x, Y: y}] {
				return false
			}
		}
	}

	// Check that the static doesn't touch any existing static (including diagonals)
	if touchesExistingStatic(pos, staticLayer, width, height) {
		return false
	}

	return true
}

// touchesExistingStatic checks if placing a 2x2 static at pos would touch any existing static
func touchesExistingStatic(pos Point, staticLayer [][]int, width, height int) bool {
	// Check a 4x4 area around the 2x2 placement (1 cell buffer on each side)
	for dy := -1; dy <= staticSize; dy++ {
		for dx := -1; dx <= staticSize; dx++ {
			// Skip the cells that will be occupied by this static
			if dx >= 0 && dx < staticSize && dy >= 0 && dy < staticSize {
				continue
			}

			x := pos.X + dx
			y := pos.Y + dy

			if x >= 0 && x < width && y >= 0 && y < height {
				if staticLayer[y][x] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// wouldTouch checks if two 2x2 statics would touch (including diagonals)
func wouldTouch(pos1, pos2 Point) bool {
	// Two 2x2 squares touch if their bounding boxes (expanded by 1) overlap
	// pos1 occupies [pos1.X, pos1.X+1] x [pos1.Y, pos1.Y+1]
	// pos2 occupies [pos2.X, pos2.X+1] x [pos2.Y, pos2.Y+1]
	// They touch if the gap between them is <= 1 in both dimensions

	// Check X overlap with 1 cell buffer
	xOverlap := !(pos1.X+staticSize+1 <= pos2.X || pos2.X+staticSize+1 <= pos1.X)
	// Check Y overlap with 1 cell buffer
	yOverlap := !(pos1.Y+staticSize+1 <= pos2.Y || pos2.Y+staticSize+1 <= pos1.Y)

	return xOverlap && yOverlap
}

// checkConnectivityAfterPlacement checks if all doors remain connected after placing a static
func checkConnectivityAfterPlacement(ground, staticLayer [][]int, doorPositions map[DoorPosition]Point, newStaticPos Point, width, height int) bool {
	if len(doorPositions) < 2 {
		return true
	}

	// Create a temporary walkable map: ground=1 and static=0 means walkable
	// Temporarily mark the new static position as blocked
	walkable := make([][]bool, height)
	for y := 0; y < height; y++ {
		walkable[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			walkable[y][x] = ground[y][x] == 1 && staticLayer[y][x] == 0
		}
	}

	// Block the new static position
	for dy := 0; dy < staticSize; dy++ {
		for dx := 0; dx < staticSize; dx++ {
			x := newStaticPos.X + dx
			y := newStaticPos.Y + dy
			if x < width && y < height {
				walkable[y][x] = false
			}
		}
	}

	// Get door positions as a slice
	doors := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		doors = append(doors, pos)
	}

	// Check if all doors are connected using BFS from the first door
	startDoor := doors[0]
	visited := bfsConnectivity(walkable, startDoor, width, height)

	// Check if all other doors are reachable
	for _, door := range doors[1:] {
		if !visited[door.Y][door.X] {
			return false
		}
	}

	return true
}

// bfsConnectivity performs BFS to find all connected cells from a starting point
func bfsConnectivity(walkable [][]bool, start Point, width, height int) [][]bool {
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	// Find the nearest walkable cell to the start point (door might be on edge)
	startCell := findNearestWalkable(walkable, start, width, height)
	if startCell.X < 0 {
		return visited // No walkable cell found
	}

	queue := []Point{startCell}
	visited[startCell.Y][startCell.X] = true

	// 4-directional movement
	dxArr := []int{0, 1, 0, -1}
	dyArr := []int{-1, 0, 1, 0}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for i := 0; i < 4; i++ {
			nx := curr.X + dxArr[i]
			ny := curr.Y + dyArr[i]

			if nx >= 0 && nx < width && ny >= 0 && ny < height && !visited[ny][nx] && walkable[ny][nx] {
				visited[ny][nx] = true
				queue = append(queue, Point{X: nx, Y: ny})
			}
		}
	}

	return visited
}

// findNearestWalkable finds the nearest walkable cell to the given point
func findNearestWalkable(walkable [][]bool, pos Point, width, height int) Point {
	// Check the point itself first
	if pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height && walkable[pos.Y][pos.X] {
		return pos
	}

	// Search in expanding squares
	for radius := 1; radius < max(width, height); radius++ {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if abs(dx) != radius && abs(dy) != radius {
					continue // Only check the perimeter
				}
				nx := pos.X + dx
				ny := pos.Y + dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height && walkable[ny][nx] {
					return Point{X: nx, Y: ny}
				}
			}
		}
	}

	return Point{X: -1, Y: -1} // Not found
}

// findValidStaticPositions finds all valid top-left corners for 2x2 static placement
func findValidStaticPositions(ground, softEdge, bridge, staticLayer [][]int, forbiddenCells map[Point]bool, width, height int) []Point {
	var positions []Point

	// Iterate through all possible top-left corners for 2x2 placement
	for y := 0; y <= height-staticSize; y++ {
		for x := 0; x <= width-staticSize; x++ {
			pos := Point{X: x, Y: y}
			if isValidStaticPosition(pos, ground, softEdge, bridge, staticLayer, forbiddenCells, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// ============================================================================
// Turret validation
// ============================================================================

// isValidTurretPosition checks if a turret can be placed at the given position
func isValidTurretPosition(pos Point, ground, softEdge, bridge, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) bool {
	x, y := pos.X, pos.Y

	// Check bounds
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Check ground layer (must be 1)
	if ground[y][x] != 1 {
		return false
	}

	// Check soft edge (must be 0)
	if softEdge[y][x] != 0 {
		return false
	}

	// Check bridge (must be 0)
	if bridge[y][x] != 0 {
		return false
	}

	// Check static layer (must be 0)
	if staticLayer[y][x] != 0 {
		return false
	}

	// Check existing turret (must be 0)
	if turretLayer[y][x] != 0 {
		return false
	}

	// Check minimum distance from doors (at least 4 cells)
	for _, doorPos := range doorPositions {
		dist := manhattanDistance(pos, doorPos)
		if dist < turretMinDoorDistance {
			return false
		}
	}

	// Check minimum distance from other turrets (at least 2 cells)
	if tooCloseToExistingTurret(pos, turretLayer, width, height) {
		return false
	}

	return true
}

// tooCloseToExistingTurret checks if the position is too close to any existing turret
func tooCloseToExistingTurret(pos Point, turretLayer [][]int, width, height int) bool {
	for dy := -turretMinTurretDistance + 1; dy < turretMinTurretDistance; dy++ {
		for dx := -turretMinTurretDistance + 1; dx < turretMinTurretDistance; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if turretLayer[ny][nx] != 0 {
					// Check actual Manhattan distance
					if abs(dx)+abs(dy) < turretMinTurretDistance {
						return true
					}
				}
			}
		}
	}
	return false
}

// checkTurretConnectivityAfterPlacement checks if all doors remain connected after placing a turret
func checkTurretConnectivityAfterPlacement(ground, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, newTurretPos Point, width, height int) bool {
	if len(doorPositions) < 2 {
		return true
	}

	// Create a temporary walkable map
	walkable := make([][]bool, height)
	for y := 0; y < height; y++ {
		walkable[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			// Walkable if ground=1 and not blocked by static or turret
			walkable[y][x] = ground[y][x] == 1 && staticLayer[y][x] == 0 && turretLayer[y][x] == 0
		}
	}

	// Block the new turret position
	walkable[newTurretPos.Y][newTurretPos.X] = false

	// Get door positions as a slice
	doors := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		doors = append(doors, pos)
	}

	// Check if all doors are connected using BFS from the first door
	startDoor := doors[0]
	visited := bfsConnectivity(walkable, startDoor, width, height)

	// Check if all other doors are reachable
	for _, door := range doors[1:] {
		if !visited[door.Y][door.X] {
			return false
		}
	}

	return true
}

// CornerType represents the type of corner a ground tile forms
type CornerType int

const (
	CornerTypeNone CornerType = iota
	CornerType90              // Right angle: 2 adjacent ground tiles at 90°
	CornerType270             // Inner corner: 3 adjacent ground tiles at 270°
)

// getGroundCornerType determines if a position is at a ground corner (90° or 270°)
// A 90° corner has exactly 2 orthogonal neighbors that are adjacent to each other (L-shape)
// A 270° corner has exactly 3 orthogonal neighbors (inverted L-shape)
func getGroundCornerType(pos Point, ground [][]int, width, height int) CornerType {
	x, y := pos.X, pos.Y

	// Count orthogonal ground neighbors
	// 0=top, 1=right, 2=bottom, 3=left
	neighbors := [4]bool{}
	dxArr := []int{0, 1, 0, -1}
	dyArr := []int{-1, 0, 1, 0}

	groundCount := 0
	for i := 0; i < 4; i++ {
		nx, ny := x+dxArr[i], y+dyArr[i]
		if nx >= 0 && nx < width && ny >= 0 && ny < height && ground[ny][nx] == 1 {
			neighbors[i] = true
			groundCount++
		}
	}

	// 90° right angle: exactly 2 adjacent neighbors forming an L
	// Valid L-shapes: top+right, right+bottom, bottom+left, left+top
	if groundCount == 2 {
		if (neighbors[0] && neighbors[1]) || // top + right
			(neighbors[1] && neighbors[2]) || // right + bottom
			(neighbors[2] && neighbors[3]) || // bottom + left
			(neighbors[3] && neighbors[0]) { // left + top
			return CornerType90
		}
	}

	// 270° inner corner: exactly 3 neighbors (one side missing)
	if groundCount == 3 {
		return CornerType270
	}

	return CornerTypeNone
}

// findValidTurretPositions finds all valid positions for turret placement
func findValidTurretPositions(ground, softEdge, bridge, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) []Point {
	var positions []Point

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidTurretPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, doorPositions, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// ============================================================================
// Mob ground validation
// ============================================================================

// isValidMobGroundPosition checks if a single cell is valid for mob ground
func isValidMobGroundPosition(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {

	x, y := pos.X, pos.Y

	// Check bounds
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Must be on ground
	if ground[y][x] != 1 {
		return false
	}

	// Must not overlap with other layers
	if softEdge[y][x] != 0 || bridge[y][x] != 0 || staticLayer[y][x] != 0 || turretLayer[y][x] != 0 {
		return false
	}

	// Must not already have mob ground
	if mobGroundLayer[y][x] != 0 {
		return false
	}

	// Must be at least 2 cells away from doors
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < mobGroundMinDoorDistance {
			return false
		}
	}

	// Must not touch existing mob ground
	if touchesExistingMobGround(pos, mobGroundLayer, width, height) {
		return false
	}

	return true
}

// touchesExistingMobGround checks if placing mob ground at pos would touch existing mob ground
func touchesExistingMobGround(pos Point, mobGroundLayer [][]int, width, height int) bool {
	// Check all 8 neighbors (including diagonals)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if mobGroundLayer[ny][nx] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// canPlace2x2MobGround checks if a 2x2 mob ground can be placed at the given top-left corner
func canPlace2x2MobGround(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {

	// Check all 4 cells
	for dy := 0; dy < 2; dy++ {
		for dx := 0; dx < 2; dx++ {
			checkPos := Point{X: pos.X + dx, Y: pos.Y + dy}
			if !isValidMobGroundPosition(checkPos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				return false
			}
		}
	}

	// Check that 2x2 doesn't touch existing mob ground (expanded check)
	if touches2x2ExistingMobGround(pos, mobGroundLayer, width, height) {
		return false
	}

	return true
}

// touches2x2ExistingMobGround checks if placing a 2x2 mob ground would touch existing mob ground
func touches2x2ExistingMobGround(pos Point, mobGroundLayer [][]int, width, height int) bool {
	// Check a 4x4 area around the 2x2 placement (1 cell buffer on each side)
	for dy := -1; dy <= 2; dy++ {
		for dx := -1; dx <= 2; dx++ {
			// Skip the cells that will be occupied
			if dx >= 0 && dx < 2 && dy >= 0 && dy < 2 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if mobGroundLayer[ny][nx] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// canPlace1x1MobGround checks if a 1x1 mob ground can be placed
func canPlace1x1MobGround(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {
	return isValidMobGroundPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height)
}

// findValidMobGroundPositions finds all valid positions for mob ground placement
func findValidMobGroundPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) []Point {

	var positions []Point

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidMobGroundPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// isWalkableForMobGround checks if a cell is walkable for mob ground placement consideration
func isWalkableForMobGround(pos Point, ground, softEdge, bridge, staticLayer, turretLayer [][]int, width, height int) bool {
	x, y := pos.X, pos.Y
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}
	return ground[y][x] == 1 && softEdge[y][x] == 0 && bridge[y][x] == 0 && staticLayer[y][x] == 0 && turretLayer[y][x] == 0
}

// ============================================================================
// Mob air validation
// ============================================================================

// isValidMobAirPosition checks if a single cell is valid for mob air
// Note: Mob Air (flying mobs) do NOT require ground=1, they can spawn anywhere
func isValidMobAirPosition(pos Point, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {

	x, y := pos.X, pos.Y

	// Check bounds
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Must be at least 2 cells away from room edges
	if x < mobAirMinEdgeDistance || x >= width-mobAirMinEdgeDistance ||
		y < mobAirMinEdgeDistance || y >= height-mobAirMinEdgeDistance {
		return false
	}

	// No ground requirement - flying mobs can spawn anywhere

	// Must not overlap with other layers
	if softEdge[y][x] != 0 || bridge[y][x] != 0 || staticLayer[y][x] != 0 ||
		turretLayer[y][x] != 0 || mobGroundLayer[y][x] != 0 {
		return false
	}

	// Must not already have mob air
	if mobAirLayer[y][x] != 0 {
		return false
	}

	// Must be at least 4 cells away from doors
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < mobAirMinDoorDistance {
			return false
		}
	}

	// Must not touch existing mob air
	if touchesExistingMobAir(pos, mobAirLayer, width, height) {
		return false
	}

	return true
}

// touchesExistingMobAir checks if placing mob air at pos would touch existing mob air
func touchesExistingMobAir(pos Point, mobAirLayer [][]int, width, height int) bool {
	// Check all 8 neighbors (including diagonals)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if mobAirLayer[ny][nx] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// findValidMobAirPositions finds all valid positions for mob air placement
func findValidMobAirPositions(ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) []Point {

	var positions []Point

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidMobAirPosition(pos, ground, softEdge, bridge, staticLayer, turretLayer, mobGroundLayer, mobAirLayer,
				doorPositions, width, height) {
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// ============================================================================
// Soft edge / Island validation
// ============================================================================

// isFarEnoughFromDoors checks if a position is far enough from all doors
func isFarEnoughFromDoors(x, y int, doorPositions map[DoorPosition]Point, minDistance int) bool {
	pos := Point{X: x, Y: y}
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < minDistance {
			return false
		}
	}
	return true
}

// canPlaceSoftEdge checks if a soft edge can be placed (not overlapping)
func canPlaceSoftEdge(placement SoftEdgePlacement, softEdgeLayer [][]int, width, height int) bool {
	for dy := 0; dy < placement.Height; dy++ {
		for dx := 0; dx < placement.Width; dx++ {
			x := placement.StartX + dx
			y := placement.StartY + dy

			if x >= width || y >= height {
				return false
			}

			if softEdgeLayer[y][x] != 0 {
				return false
			}
		}
	}
	return true
}

// isValidIslandPosition checks if an island can be placed at (x, y) with exactly minIslandGroundDistance from existing ground
func isValidIslandPosition(ground [][]int, x, y, islandWidth, islandHeight, gridWidth, gridHeight int) bool {
	// Check that the island area and its surrounding margin are all void
	// Margin is minIslandGroundDistance on each side
	checkStartX := x - minIslandGroundDistance
	checkStartY := y - minIslandGroundDistance
	checkEndX := x + islandWidth + minIslandGroundDistance
	checkEndY := y + islandHeight + minIslandGroundDistance

	for cy := checkStartY; cy < checkEndY; cy++ {
		for cx := checkStartX; cx < checkEndX; cx++ {
			// Skip cells outside the grid (they're fine, considered void)
			if cx < 0 || cx >= gridWidth || cy < 0 || cy >= gridHeight {
				continue
			}

			// If this cell is inside the island area, it should be void (we'll place ground there)
			isInsideIsland := cx >= x && cx < x+islandWidth && cy >= y && cy < y+islandHeight
			if isInsideIsland {
				// The island area must currently be void
				if ground[cy][cx] != 0 {
					return false
				}
			} else {
				// The margin area must be void (no existing ground within minIslandGroundDistance)
				if ground[cy][cx] != 0 {
					return false
				}
			}
		}
	}

	// Additionally, check that there IS ground just outside the margin (at distance exactly minIslandGroundDistance+1)
	// This ensures the island is placed close to existing ground, not in the middle of nowhere
	hasNearbyGround := false
	outerDist := minIslandGroundDistance + 1
	outerStartX := x - outerDist
	outerStartY := y - outerDist
	outerEndX := x + islandWidth + outerDist
	outerEndY := y + islandHeight + outerDist

	for cy := outerStartY; cy < outerEndY; cy++ {
		for cx := outerStartX; cx < outerEndX; cx++ {
			// Skip cells inside the already-checked margin area
			if cx >= checkStartX && cx < checkEndX && cy >= checkStartY && cy < checkEndY {
				continue
			}
			// Skip cells outside the grid
			if cx < 0 || cx >= gridWidth || cy < 0 || cy >= gridHeight {
				continue
			}
			// Check if there's ground at this outer ring
			if ground[cy][cx] == 1 {
				hasNearbyGround = true
				break
			}
		}
		if hasNearbyGround {
			break
		}
	}

	return hasNearbyGround
}

// ============================================================================
// Platform connectivity (used by platform.go eraser validation)
// ============================================================================

// areAllDoorsConnected checks if all doors are connected via walkable ground
func areAllDoorsConnected(ground [][]int, width, height int, doors []DoorPosition) bool {
	if len(doors) < 2 {
		return true
	}

	// Get first door position
	startX, startY := getDoorPosition(doors[0], width, height)

	// BFS from first door
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	queue := []Point{{X: startX, Y: startY}}
	visited[startY][startX] = true

	// Also mark adjacent ground as visited (door might be at edge)
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		neighbors := []Point{
			{curr.X - 1, curr.Y}, {curr.X + 1, curr.Y},
			{curr.X, curr.Y - 1}, {curr.X, curr.Y + 1},
		}

		for _, n := range neighbors {
			if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height &&
				!visited[n.Y][n.X] && ground[n.Y][n.X] == 1 {
				visited[n.Y][n.X] = true
				queue = append(queue, n)
			}
		}
	}

	// Check if all other doors are reachable
	for i := 1; i < len(doors); i++ {
		doorX, doorY := getDoorPosition(doors[i], width, height)

		// Check if door position or any adjacent cell is visited
		reachable := false
		checkPositions := []Point{
			{doorX, doorY},
			{doorX - 1, doorY}, {doorX + 1, doorY},
			{doorX, doorY - 1}, {doorX, doorY + 1},
		}

		for _, pos := range checkPositions {
			if pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height {
				if visited[pos.Y][pos.X] {
					reachable = true
					break
				}
			}
		}

		if !reachable {
			return false
		}
	}

	return true
}

// ============================================================================
// Bridge validation
// ============================================================================

// canPlaceBridge checks if a 2x2 bridge can be placed at (x, y)
func canPlaceBridge(x, y int, ground, bridgeLayer, softEdgeLayer [][]int, width, height int) bool {
	// Bridge must be within bounds
	if x < 0 || x+bridgeSize > width || y < 0 || y+bridgeSize > height {
		return false
	}

	// All cells must be void (ground=0), no existing bridge, and no soft edge
	for dy := 0; dy < bridgeSize; dy++ {
		for dx := 0; dx < bridgeSize; dx++ {
			if ground[y+dy][x+dx] != 0 || bridgeLayer[y+dy][x+dx] != 0 || softEdgeLayer[y+dy][x+dx] != 0 {
				return false
			}
		}
	}

	return true
}

// canPlaceBridgeAt checks if a 2x2 bridge can be placed at the given position
func canPlaceBridgeAt(bridgeLayer, ground, softEdgeLayer [][]int, x, y, width, height int) bool {
	// Bridge must be within bounds
	if x < 0 || x+bridgeSize > width || y < 0 || y+bridgeSize > height {
		return false
	}

	// All cells must be void (ground=0), no existing bridge, and no soft edge
	for dy := 0; dy < bridgeSize; dy++ {
		for dx := 0; dx < bridgeSize; dx++ {
			if ground[y+dy][x+dx] != 0 || bridgeLayer[y+dy][x+dx] != 0 || softEdgeLayer[y+dy][x+dx] != 0 {
				return false
			}
		}
	}

	return true
}

// bridgeTouchesIsland checks if a 2x2 bridge at (bx, by) fully touches the island (2x2 contact)
func bridgeTouchesIsland(bx, by int, island Island, ground [][]int) bool {
	// Check if the bridge has at least 2 adjacent cells touching the island
	touchCount := 0

	// Check all 4 sides of the bridge
	bridgeCells := []Point{
		{bx, by}, {bx + 1, by}, {bx, by + 1}, {bx + 1, by + 1},
	}

	for _, bc := range bridgeCells {
		// Check adjacent cells (not diagonal)
		adjacents := []Point{
			{bc.X - 1, bc.Y}, {bc.X + 1, bc.Y}, {bc.X, bc.Y - 1}, {bc.X, bc.Y + 1},
		}
		for _, adj := range adjacents {
			// Check if adjacent cell is part of the island
			for _, ic := range island.Cells {
				if ic.X == adj.X && ic.Y == adj.Y {
					touchCount++
					break
				}
			}
		}
	}

	// Need at least 2 touch points for 2x2 full contact
	return touchCount >= 2
}

// bridgeTouchesExistingGround checks if bridge touches ground that's not part of the given island
func bridgeTouchesExistingGround(bx, by int, excludeIsland Island, allIslands []Island, connected map[int]bool, ground [][]int, width, height int) (bool, string) {
	// Check all cells adjacent to the bridge
	bridgeCells := []Point{
		{bx, by}, {bx + 1, by}, {bx, by + 1}, {bx + 1, by + 1},
	}

	excludeCells := make(map[Point]bool)
	for _, c := range excludeIsland.Cells {
		excludeCells[c] = true
	}

	touchCount := 0
	var targetDesc string

	for _, bc := range bridgeCells {
		adjacents := []Point{
			{bc.X - 1, bc.Y}, {bc.X + 1, bc.Y}, {bc.X, bc.Y - 1}, {bc.X, bc.Y + 1},
		}
		for _, adj := range adjacents {
			if adj.X < 0 || adj.X >= width || adj.Y < 0 || adj.Y >= height {
				continue
			}
			if ground[adj.Y][adj.X] == 1 && !excludeCells[adj] {
				touchCount++
				// Find which island this belongs to
				for i, island := range allIslands {
					if connected[i] {
						for _, ic := range island.Cells {
							if ic.X == adj.X && ic.Y == adj.Y {
								if i == 0 {
									targetDesc = "main ground"
								} else {
									targetDesc = "island"
								}
								break
							}
						}
					}
				}
			}
		}
	}

	if targetDesc == "" && touchCount >= 2 {
		targetDesc = "ground"
	}

	return touchCount >= 2, targetDesc
}

// ============================================================================
// Rail-aware validation variants
// ============================================================================

func isValidStaticPositionWithRail(pos Point, ground, softEdge, bridge, rail, staticLayer [][]int, forbiddenCells map[Point]bool, width, height int) bool {
	// Check all 4 cells of the 2x2 area
	for dy := 0; dy < staticSize; dy++ {
		for dx := 0; dx < staticSize; dx++ {
			x := pos.X + dx
			y := pos.Y + dy

			// Check bounds
			if x < 0 || x >= width || y < 0 || y >= height {
				return false
			}

			// Must be on ground
			if ground[y][x] != 1 {
				return false
			}

			// Cannot be on bridge
			if bridge[y][x] == 1 {
				return false
			}

			// Cannot be on soft edge
			if softEdge[y][x] == 1 {
				return false
			}

			// Cannot be on rail
			if rail[y][x] == 1 {
				return false
			}

			// Cannot be on existing static
			if staticLayer[y][x] == 1 {
				return false
			}

			// Cannot be in forbidden zone (doors)
			if forbiddenCells[Point{X: x, Y: y}] {
				return false
			}
		}
	}
	return true
}

func findValidStaticPositionsWithRail(ground, softEdge, bridge, rail, staticLayer [][]int, forbiddenCells map[Point]bool, width, height int) []Point {
	var positions []Point
	for y := 0; y <= height-staticSize; y++ {
		for x := 0; x <= width-staticSize; x++ {
			pos := Point{X: x, Y: y}
			if isValidStaticPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, forbiddenCells, width, height) {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}

func isValidTurretPositionWithRail(pos Point, ground, softEdge, bridge, rail, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) bool {
	x, y := pos.X, pos.Y

	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Must be on ground
	if ground[y][x] != 1 {
		return false
	}

	// Cannot be on bridge
	if bridge[y][x] == 1 {
		return false
	}

	// Cannot be on soft edge
	if softEdge[y][x] == 1 {
		return false
	}

	// Cannot be on rail
	if rail[y][x] == 1 {
		return false
	}

	// Cannot be on static
	if staticLayer[y][x] == 1 {
		return false
	}

	// Cannot be on existing turret
	if turretLayer[y][x] == 1 {
		return false
	}

	// Must be at least turretMinDoorDistance cells away from doors
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < turretMinDoorDistance {
			return false
		}
	}

	return true
}

func findValidTurretPositionsWithRail(ground, softEdge, bridge, rail, staticLayer, turretLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) []Point {
	var positions []Point
	forbiddenCells := getDoorForbiddenCells(doorPositions, width, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidTurretPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, turretLayer, doorPositions, width, height) && !forbiddenCells[pos] {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}

func isValidMobGroundPositionWithRail(pos Point, ground, softEdge, bridge, rail, staticLayer, turretLayer, mobGroundLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) bool {
	x, y := pos.X, pos.Y

	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Must be on ground
	if ground[y][x] != 1 {
		return false
	}

	// Cannot be on bridge
	if bridge[y][x] == 1 {
		return false
	}

	// Cannot be on soft edge
	if softEdge[y][x] == 1 {
		return false
	}

	// Cannot be on rail
	if rail[y][x] == 1 {
		return false
	}

	// Cannot be on static
	if staticLayer[y][x] == 1 {
		return false
	}

	// Cannot be on turret
	if turretLayer[y][x] == 1 {
		return false
	}

	// Cannot be on existing mob ground
	if mobGroundLayer[y][x] == 1 {
		return false
	}

	// Must be at least mobGroundMinDoorDistance cells away from doors
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < mobGroundMinDoorDistance {
			return false
		}
	}

	return true
}

func findValidMobGroundPositionsWithRail(ground, softEdge, bridge, rail, staticLayer, turretLayer, mobGroundLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) []Point {
	var positions []Point
	forbiddenCells := getDoorForbiddenCells(doorPositions, width, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{X: x, Y: y}
			if isValidMobGroundPositionWithRail(pos, ground, softEdge, bridge, rail, staticLayer, turretLayer, mobGroundLayer, doorPositions, width, height) && !forbiddenCells[pos] {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}

// filterPositionsInRailIndent returns positions (2x2) that overlap with rail indent cells
func filterPositionsInRailIndent(positions []Point, railIndentCells []Point) []Point {
	if len(railIndentCells) == 0 {
		return nil
	}

	indentSet := make(map[Point]bool)
	for _, cell := range railIndentCells {
		indentSet[cell] = true
	}

	var filtered []Point
	for _, pos := range positions {
		// Check if any cell of the 2x2 area is in the indent
		inIndent := false
		for dy := 0; dy < staticSize; dy++ {
			for dx := 0; dx < staticSize; dx++ {
				if indentSet[Point{X: pos.X + dx, Y: pos.Y + dy}] {
					inIndent = true
					break
				}
			}
			if inIndent {
				break
			}
		}
		if inIndent {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

// filterSinglePositionsInRailIndent returns single-cell positions that are inside rail indent cells
func filterSinglePositionsInRailIndent(positions []Point, railIndentCells []Point) []Point {
	if len(railIndentCells) == 0 {
		return nil
	}

	indentSet := make(map[Point]bool)
	for _, cell := range railIndentCells {
		indentSet[cell] = true
	}

	var filtered []Point
	for _, pos := range positions {
		if indentSet[pos] {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

// removePosition removes all occurrences of toRemove from positions
func removePosition(positions []Point, toRemove Point) []Point {
	var result []Point
	for _, pos := range positions {
		if pos.X != toRemove.X || pos.Y != toRemove.Y {
			result = append(result, pos)
		}
	}
	return result
}

// removeSinglePosition removes the first occurrence of toRemove from positions
func removeSinglePosition(positions []Point, toRemove Point) []Point {
	for i, pos := range positions {
		if pos.X == toRemove.X && pos.Y == toRemove.Y {
			return append(positions[:i], positions[i+1:]...)
		}
	}
	return positions
}

// ============================================================================
// Unified door forbidden zone (radius-based)
// ============================================================================

// getDoorForbiddenCellsRadius returns all cells within Manhattan distance `radius` of any door
func getDoorForbiddenCellsRadius(doorPositions map[DoorPosition]Point, width, height, radius int) map[Point]bool {
	forbidden := make(map[Point]bool)
	for _, doorPos := range doorPositions {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if abs(dx)+abs(dy) > radius {
					continue
				}
				x, y := doorPos.X+dx, doorPos.Y+dy
				if x >= 0 && x < width && y >= 0 && y < height {
					forbidden[Point{X: x, Y: y}] = true
				}
			}
		}
	}
	return forbidden
}

// ============================================================================
// New enemy layer validation (Chaser / Zoner / DPS)
// ============================================================================

// isValidEnemyPosition checks if a cell is valid for enemy placement (Chaser/Zoner/DPS)
func isValidEnemyPosition(pos Point, ground, softEdge, bridge, rail, staticLayer [][]int, forbidden map[Point]bool, width, height int) bool {
	x, y := pos.X, pos.Y
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}
	// Must be on ground
	if ground[y][x] != 1 {
		return false
	}
	// Cannot be on softEdge, bridge, static
	if softEdge[y][x] != 0 || bridge[y][x] != 0 || staticLayer[y][x] != 0 {
		return false
	}
	// Cannot be on rail (if exists)
	if rail != nil && rail[y][x] != 0 {
		return false
	}
	// Cannot be in door forbidden zone
	if forbidden[pos] {
		return false
	}
	return true
}

// touchesLayer checks if a position has any 8-directional neighbor with value != 0
func touchesLayer(pos Point, layer [][]int, width, height int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if layer[ny][nx] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// ============================================================================
// New MobAir validation (updated for new layers)
// ============================================================================

// isValidMobAirPositionNew checks if a cell is valid for mob air with new enemy layers
func isValidMobAirPositionNew(pos Point, ground, softEdge, bridge, staticLayer, zonerLayer, chaserLayer, dpsLayer, mobAirLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height int) bool {

	x, y := pos.X, pos.Y
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	// Must be at least 2 cells away from room edges
	if x < mobAirMinEdgeDistance || x >= width-mobAirMinEdgeDistance ||
		y < mobAirMinEdgeDistance || y >= height-mobAirMinEdgeDistance {
		return false
	}

	// No ground requirement - flying mobs can spawn anywhere

	// Must not overlap with other layers
	if softEdge[y][x] != 0 || bridge[y][x] != 0 || staticLayer[y][x] != 0 ||
		zonerLayer[y][x] != 0 || chaserLayer[y][x] != 0 || dpsLayer[y][x] != 0 {
		return false
	}

	// Must not already have mob air
	if mobAirLayer[y][x] != 0 {
		return false
	}

	// Must be at least 4 cells away from doors
	for _, doorPos := range doorPositions {
		if manhattanDistance(pos, doorPos) < mobAirMinDoorDistance {
			return false
		}
	}

	// Must not touch existing mob air (distance >= 1 means no 8-directional adjacency)
	if touchesLayer(pos, mobAirLayer, width, height) {
		return false
	}

	return true
}

// GenerateMobAirLayerNew generates mob air layer using new enemy layers instead of turret/mobGround
func GenerateMobAirLayerNew(mobAirLayer, ground, softEdge, bridge, staticLayer, zonerLayer, chaserLayer, dpsLayer [][]int,
	doorPositions map[DoorPosition]Point, width, height, targetCount int, regionFilter ...*RegionFilter) *MobAirDebugInfo {

	debug := &MobAirDebugInfo{
		TargetCount: targetCount,
		Strategy:    "density_based",
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	// Find valid positions
	var candidates []Point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Point{x, y}
			if isValidMobAirPositionNew(pos, ground, softEdge, bridge, staticLayer, zonerLayer, chaserLayer, dpsLayer, mobAirLayer,
				doorPositions, width, height) {
				candidates = append(candidates, pos)
			}
		}
	}

	if len(candidates) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{Reason: "no valid positions found"})
		return debug
	}

	// Score: prefer positions near zoner/chaser dense areas
	type scored struct {
		pos   Point
		score float64
	}
	var scoredCandidates []scored
	for _, pos := range candidates {
		s := 0.0
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				nx, ny := pos.X+dx, pos.Y+dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height {
					if zonerLayer[ny][nx] == 1 {
						s += 2.0
					}
					if chaserLayer[ny][nx] == 1 {
						s += 1.5
					}
				}
			}
		}
		scoredCandidates = append(scoredCandidates, scored{pos, s})
	}

	// Sort by score descending
	for i := 0; i < len(scoredCandidates)-1; i++ {
		for j := i + 1; j < len(scoredCandidates); j++ {
			if scoredCandidates[j].score > scoredCandidates[i].score {
				scoredCandidates[i], scoredCandidates[j] = scoredCandidates[j], scoredCandidates[i]
			}
		}
	}

	remaining := targetCount
	for _, sc := range scoredCandidates {
		if remaining <= 0 {
			break
		}
		pos := sc.pos
		// Re-validate (previous placements may have invalidated)
		if !isValidMobAirPositionNew(pos, ground, softEdge, bridge, staticLayer, zonerLayer, chaserLayer, dpsLayer, mobAirLayer,
			doorPositions, width, height) {
			continue
		}
		mobAirLayer[pos.Y][pos.X] = 1
		remaining--
		debug.PlacedCount++

		reason := "density_based placement"
		if ground[pos.Y][pos.X] == 0 {
			reason += " (on void, flying)"
		}
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", pos.X, pos.Y),
			Size:     "1x1",
			Reason:   reason,
		})
	}

	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("could not place %d more mob air", remaining),
		})
	}

	return debug
}
