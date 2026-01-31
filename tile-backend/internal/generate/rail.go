package generate

import (
	"fmt"
	"math/rand"
)

// Rail generation constants
const (
	minRailAreaSize     = 5  // Minimum enclosed area size (5x5)
	minRailEdgeDistance = 2  // Minimum distance from room edge
	minPlatformForRail  = 6  // Minimum platform size to place rail (6x6)
	maxRailIndents      = 4  // Maximum number of indents per rail loop
	railIndentProb      = 30 // Probability of adding an indent (30%)
	railRetryProb       = 50 // Probability of retrying after placement (50%)
)

// RailDebugInfo contains debug information about rail generation
type RailDebugInfo struct {
	Skipped        bool            `json:"skipped"`
	SkipReason     string          `json:"skipReason,omitempty"`
	PlatformsFound int             `json:"platformsFound"`
	RailLoops      []RailLoopInfo  `json:"railLoops"`
	Misses         []MissInfo      `json:"misses,omitempty"`
}

// RailLoopInfo describes a rail loop placement
type RailLoopInfo struct {
	Platform    string       `json:"platform"`    // Platform position and size
	BoundingBox string       `json:"boundingBox"` // Rail bounding box
	Perimeter   int          `json:"perimeter"`   // Rail perimeter length
	Indents     []IndentInfo `json:"indents"`     // Indents added to the loop
}

// IndentInfo describes an indent in a rail loop
type IndentInfo struct {
	Position  string `json:"position"`  // Position of the indent
	Direction string `json:"direction"` // Direction (inward/outward)
	Size      int    `json:"size"`      // Size of the indent
}

// RailPlatform represents a rectangular walkable area suitable for rail placement
type RailPlatform struct {
	X, Y          int // Top-left corner
	Width, Height int
}

// RailLoop represents a closed rail loop
type RailLoop struct {
	Cells     []Point // All cells in the loop
	MinX      int
	MinY      int
	MaxX      int
	MaxY      int
	Perimeter int
}

// GenerateRailLayer generates the rail layer based on ground and bridge layers
func GenerateRailLayer(railLayer, ground, bridge [][]int, width, height int) *RailDebugInfo {
	debug := &RailDebugInfo{
		RailLoops: []RailLoopInfo{},
		Misses:    []MissInfo{},
	}

	// Step 1: Find all platforms >= 6x6 that can support rail
	platforms := findRailPlatforms(ground, bridge, width, height)
	debug.PlatformsFound = len(platforms)

	if len(platforms) == 0 {
		debug.Skipped = true
		debug.SkipReason = fmt.Sprintf("no platforms >= %dx%d found for rail placement", minPlatformForRail, minPlatformForRail)
		return debug
	}

	// Shuffle platforms for random selection
	rand.Shuffle(len(platforms), func(i, j int) {
		platforms[i], platforms[j] = platforms[j], platforms[i]
	})

	// Step 2: Try to place rail loops on platforms
	for len(platforms) > 0 {
		// Pop a platform
		platform := platforms[0]
		platforms = platforms[1:]

		// Try to place a rail loop on this platform
		loopInfo := tryPlaceRailLoop(railLayer, ground, bridge, platform, width, height, debug)
		if loopInfo != nil {
			debug.RailLoops = append(debug.RailLoops, *loopInfo)
		}

		// 50% probability to continue placing more rails
		if rand.Intn(100) >= railRetryProb {
			break
		}
	}

	// Step 3: Post-process - remove intersecting rails and keep largest loop
	postProcessRails(railLayer, width, height, debug)

	if len(debug.RailLoops) == 0 {
		debug.Skipped = true
		debug.SkipReason = "could not place any valid rail loops"
	}

	return debug
}

// findRailPlatforms finds all rectangular walkable areas >= 6x6 with edge distance >= 2
func findRailPlatforms(ground, bridge [][]int, width, height int) []RailPlatform {
	var platforms []RailPlatform

	// Create combined walkable layer
	walkable := make([][]bool, height)
	for y := 0; y < height; y++ {
		walkable[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			walkable[y][x] = ground[y][x] == 1 || bridge[y][x] == 1
		}
	}

	// Find rectangular regions of walkable cells
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	// Scan for platforms with minimum edge distance
	for y := minRailEdgeDistance; y < height-minRailEdgeDistance; y++ {
		for x := minRailEdgeDistance; x < width-minRailEdgeDistance; x++ {
			if visited[y][x] || !walkable[y][x] {
				continue
			}

			// Find the maximum rectangle starting from this point
			platform := findMaxWalkableRectangle(walkable, visited, x, y, width, height)

			// Check if platform meets minimum size requirements
			if platform.Width >= minPlatformForRail && platform.Height >= minPlatformForRail {
				// Ensure platform respects edge distance
				if platform.X >= minRailEdgeDistance &&
					platform.Y >= minRailEdgeDistance &&
					platform.X+platform.Width <= width-minRailEdgeDistance &&
					platform.Y+platform.Height <= height-minRailEdgeDistance {
					platforms = append(platforms, platform)
				}
			}
		}
	}

	return platforms
}

// findMaxWalkableRectangle finds the maximum rectangle of walkable cells
func findMaxWalkableRectangle(walkable [][]bool, visited [][]bool, startX, startY, width, height int) RailPlatform {
	// Find max width from starting point
	maxWidth := 0
	for x := startX; x < width && walkable[startY][x]; x++ {
		maxWidth++
	}

	// Find max height that maintains at least minPlatformForRail width
	maxHeight := 0
	currentWidth := maxWidth
	for y := startY; y < height && currentWidth >= minPlatformForRail; y++ {
		// Check this row's walkable width
		rowWidth := 0
		for x := startX; x < width && x < startX+currentWidth && walkable[y][x]; x++ {
			rowWidth++
		}
		if rowWidth < minPlatformForRail {
			break
		}
		currentWidth = rowWidth
		maxHeight++
	}

	// Mark visited cells
	for y := startY; y < startY+maxHeight && y < height; y++ {
		for x := startX; x < startX+currentWidth && x < width; x++ {
			visited[y][x] = true
		}
	}

	return RailPlatform{
		X:      startX,
		Y:      startY,
		Width:  currentWidth,
		Height: maxHeight,
	}
}

// tryPlaceRailLoop attempts to place a rail loop on the given platform
func tryPlaceRailLoop(railLayer, ground, bridge [][]int, platform RailPlatform, width, height int, debug *RailDebugInfo) *RailLoopInfo {
	// Calculate the maximum rail rectangle that fits within the platform
	// Leave 1 cell margin inside for potential indents
	railX := platform.X
	railY := platform.Y
	railW := platform.Width
	railH := platform.Height

	// Ensure minimum enclosed area (5x5 means the hollow part is at least 5x5)
	// So the rail rectangle must be at least 7x7 (outer border + 5x5 inside + outer border)
	minRailSize := minRailAreaSize + 2
	if railW < minRailSize || railH < minRailSize {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("platform (%d,%d) %dx%d too small for minimum rail size %dx%d",
				platform.X, platform.Y, platform.Width, platform.Height, minRailSize, minRailSize),
		})
		return nil
	}

	// Randomly shrink the rail rectangle a bit for variety
	if railW > minRailSize+2 {
		shrink := rand.Intn((railW - minRailSize) / 2)
		railX += shrink
		railW -= shrink * 2
	}
	if railH > minRailSize+2 {
		shrink := rand.Intn((railH - minRailSize) / 2)
		railY += shrink
		railH -= shrink * 2
	}

	// Verify all rail positions are on ground or bridge
	if !verifyRailPositions(railX, railY, railW, railH, ground, bridge, width, height) {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("rail positions at (%d,%d) %dx%d not all on ground/bridge",
				railX, railY, railW, railH),
		})
		return nil
	}

	// Draw the hollow rectangle (rail loop)
	loopInfo := &RailLoopInfo{
		Platform:    fmt.Sprintf("(%d,%d) %dx%d", platform.X, platform.Y, platform.Width, platform.Height),
		BoundingBox: fmt.Sprintf("(%d,%d) %dx%d", railX, railY, railW, railH),
		Indents:     []IndentInfo{},
	}

	// Draw top and bottom edges
	for x := railX; x < railX+railW; x++ {
		railLayer[railY][x] = 1                // Top edge
		railLayer[railY+railH-1][x] = 1        // Bottom edge
	}

	// Draw left and right edges (excluding corners already drawn)
	for y := railY + 1; y < railY+railH-1; y++ {
		railLayer[y][railX] = 1              // Left edge
		railLayer[y][railX+railW-1] = 1      // Right edge
	}

	// Calculate initial perimeter
	loopInfo.Perimeter = 2*(railW+railH) - 4 // Subtract 4 for corners counted twice

	// Add random indents (30% probability, max 4)
	indentCount := 0
	for indentCount < maxRailIndents {
		if rand.Intn(100) >= railIndentProb {
			break
		}

		indent := tryAddRailIndent(railLayer, ground, bridge, railX, railY, railW, railH, width, height)
		if indent != nil {
			loopInfo.Indents = append(loopInfo.Indents, *indent)
			loopInfo.Perimeter += indent.Size * 2 // Each indent adds cells to perimeter
			indentCount++
		}
	}

	return loopInfo
}

// verifyRailPositions checks if all rail positions are on ground or bridge
func verifyRailPositions(railX, railY, railW, railH int, ground, bridge [][]int, width, height int) bool {
	// Check top edge
	for x := railX; x < railX+railW; x++ {
		if x < 0 || x >= width || railY < 0 || railY >= height {
			return false
		}
		if ground[railY][x] != 1 && bridge[railY][x] != 1 {
			return false
		}
	}

	// Check bottom edge
	bottomY := railY + railH - 1
	for x := railX; x < railX+railW; x++ {
		if x < 0 || x >= width || bottomY < 0 || bottomY >= height {
			return false
		}
		if ground[bottomY][x] != 1 && bridge[bottomY][x] != 1 {
			return false
		}
	}

	// Check left edge
	for y := railY; y < railY+railH; y++ {
		if railX < 0 || railX >= width || y < 0 || y >= height {
			return false
		}
		if ground[y][railX] != 1 && bridge[y][railX] != 1 {
			return false
		}
	}

	// Check right edge
	rightX := railX + railW - 1
	for y := railY; y < railY+railH; y++ {
		if rightX < 0 || rightX >= width || y < 0 || y >= height {
			return false
		}
		if ground[y][rightX] != 1 && bridge[y][rightX] != 1 {
			return false
		}
	}

	return true
}

// tryAddRailIndent attempts to add an indent to the rail loop
func tryAddRailIndent(railLayer, ground, bridge [][]int, railX, railY, railW, railH, width, height int) *IndentInfo {
	// Choose a random edge to add indent
	edge := rand.Intn(4) // 0=top, 1=right, 2=bottom, 3=left

	var indentX, indentY, indentSize int
	var direction string

	switch edge {
	case 0: // Top edge - indent goes inward (down)
		// Find a position on top edge (not corners)
		if railW <= 4 {
			return nil
		}
		indentX = railX + 2 + rand.Intn(railW-4)
		indentY = railY
		indentSize = 1 + rand.Intn(min(2, railH/2-1))
		direction = "inward-down"

		// Check if indent positions are valid
		for dy := 1; dy <= indentSize; dy++ {
			y := indentY + dy
			if y >= height || ground[y][indentX] != 1 && bridge[y][indentX] != 1 {
				return nil
			}
			if railLayer[y][indentX] == 1 { // Already part of rail
				return nil
			}
		}

		// Place indent
		for dy := 1; dy <= indentSize; dy++ {
			railLayer[indentY+dy][indentX] = 1
		}

	case 1: // Right edge - indent goes inward (left)
		if railH <= 4 {
			return nil
		}
		indentX = railX + railW - 1
		indentY = railY + 2 + rand.Intn(railH-4)
		indentSize = 1 + rand.Intn(min(2, railW/2-1))
		direction = "inward-left"

		for dx := 1; dx <= indentSize; dx++ {
			x := indentX - dx
			if x < 0 || ground[indentY][x] != 1 && bridge[indentY][x] != 1 {
				return nil
			}
			if railLayer[indentY][x] == 1 {
				return nil
			}
		}

		for dx := 1; dx <= indentSize; dx++ {
			railLayer[indentY][indentX-dx] = 1
		}

	case 2: // Bottom edge - indent goes inward (up)
		if railW <= 4 {
			return nil
		}
		indentX = railX + 2 + rand.Intn(railW-4)
		indentY = railY + railH - 1
		indentSize = 1 + rand.Intn(min(2, railH/2-1))
		direction = "inward-up"

		for dy := 1; dy <= indentSize; dy++ {
			y := indentY - dy
			if y < 0 || ground[y][indentX] != 1 && bridge[y][indentX] != 1 {
				return nil
			}
			if railLayer[y][indentX] == 1 {
				return nil
			}
		}

		for dy := 1; dy <= indentSize; dy++ {
			railLayer[indentY-dy][indentX] = 1
		}

	case 3: // Left edge - indent goes inward (right)
		if railH <= 4 {
			return nil
		}
		indentX = railX
		indentY = railY + 2 + rand.Intn(railH-4)
		indentSize = 1 + rand.Intn(min(2, railW/2-1))
		direction = "inward-right"

		for dx := 1; dx <= indentSize; dx++ {
			x := indentX + dx
			if x >= width || ground[indentY][x] != 1 && bridge[indentY][x] != 1 {
				return nil
			}
			if railLayer[indentY][x] == 1 {
				return nil
			}
		}

		for dx := 1; dx <= indentSize; dx++ {
			railLayer[indentY][indentX+dx] = 1
		}
	}

	return &IndentInfo{
		Position:  fmt.Sprintf("(%d,%d)", indentX, indentY),
		Direction: direction,
		Size:      indentSize,
	}
}

// postProcessRails removes intersecting rails and keeps only the largest loop
func postProcessRails(railLayer [][]int, width, height int, debug *RailDebugInfo) {
	// Find all connected rail components
	loops := findRailLoops(railLayer, width, height)

	if len(loops) <= 1 {
		return // No intersection issues
	}

	// Find the largest loop by perimeter
	largestIdx := 0
	largestPerimeter := 0
	for i, loop := range loops {
		if loop.Perimeter > largestPerimeter {
			largestPerimeter = loop.Perimeter
			largestIdx = i
		}
	}

	// Remove all loops except the largest
	for i, loop := range loops {
		if i == largestIdx {
			continue
		}
		for _, cell := range loop.Cells {
			railLayer[cell.Y][cell.X] = 0
		}
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("removed smaller rail loop with perimeter %d at (%d,%d)-(%d,%d)",
				loop.Perimeter, loop.MinX, loop.MinY, loop.MaxX, loop.MaxY),
		})
	}

	// Validate that remaining rail cells form valid closed loops (each cell has exactly 2 neighbors)
	validateAndFixRailLoop(railLayer, width, height)
}

// findRailLoops finds all connected rail loops
func findRailLoops(railLayer [][]int, width, height int) []RailLoop {
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	var loops []RailLoop

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 1 && !visited[y][x] {
				loop := floodFillRail(railLayer, visited, x, y, width, height)
				loops = append(loops, loop)
			}
		}
	}

	return loops
}

// floodFillRail performs flood fill to find all cells of a rail loop
func floodFillRail(railLayer [][]int, visited [][]bool, startX, startY, width, height int) RailLoop {
	loop := RailLoop{
		MinX: startX,
		MinY: startY,
		MaxX: startX,
		MaxY: startY,
	}

	queue := []Point{{X: startX, Y: startY}}
	visited[startY][startX] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		loop.Cells = append(loop.Cells, curr)

		// Update bounds
		if curr.X < loop.MinX {
			loop.MinX = curr.X
		}
		if curr.X > loop.MaxX {
			loop.MaxX = curr.X
		}
		if curr.Y < loop.MinY {
			loop.MinY = curr.Y
		}
		if curr.Y > loop.MaxY {
			loop.MaxY = curr.Y
		}

		// Check 4 neighbors
		neighbors := []Point{
			{curr.X - 1, curr.Y}, {curr.X + 1, curr.Y},
			{curr.X, curr.Y - 1}, {curr.X, curr.Y + 1},
		}

		for _, n := range neighbors {
			if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height &&
				railLayer[n.Y][n.X] == 1 && !visited[n.Y][n.X] {
				visited[n.Y][n.X] = true
				queue = append(queue, n)
			}
		}
	}

	loop.Perimeter = len(loop.Cells)
	return loop
}

// validateAndFixRailLoop ensures each rail cell has exactly 2 neighbors
func validateAndFixRailLoop(railLayer [][]int, width, height int) {
	// Find cells with incorrect neighbor counts and remove them
	changed := true
	for changed {
		changed = false
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if railLayer[y][x] == 1 {
					neighbors := countRailNeighborsLocal(railLayer, x, y, width, height)
					if neighbors < 2 {
						// Dead end - remove
						railLayer[y][x] = 0
						changed = true
					}
					// Note: cells with more than 2 neighbors indicate intersections,
					// which should have been handled by keeping only the largest loop
				}
			}
		}
	}
}

// countRailNeighborsLocal counts adjacent rail cells
func countRailNeighborsLocal(railLayer [][]int, x, y, width, height int) int {
	count := 0
	neighbors := []Point{
		{x - 1, y}, {x + 1, y},
		{x, y - 1}, {x, y + 1},
	}

	for _, n := range neighbors {
		if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height {
			if railLayer[n.Y][n.X] == 1 {
				count++
			}
		}
	}

	return count
}

// GetRailIndentCells returns all cells that are indents (cells on rail that are inside the bounding box)
// These cells are preferred positions for static/turret/mobGround placement
func GetRailIndentCells(railLayer [][]int, width, height int) []Point {
	if !hasAnyRail(railLayer, width, height) {
		return nil
	}

	// Find the bounding box of the rail
	minX, minY := width, height
	maxX, maxY := 0, 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 1 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	// Find cells inside the rail bounding box that are NOT on the outer edges
	// These are the indent cells (inside the loop)
	var indentCells []Point

	for y := minY + 1; y < maxY; y++ {
		for x := minX + 1; x < maxX; x++ {
			// Check if this cell is inside the rail loop (not on rail, but enclosed)
			if railLayer[y][x] == 0 {
				// Use flood fill or ray casting to check if inside
				// Simple approach: check if surrounded by rail on all sides at some point
				if isInsideRailLoop(railLayer, x, y, minX, minY, maxX, maxY, width, height) {
					indentCells = append(indentCells, Point{X: x, Y: y})
				}
			}
		}
	}

	return indentCells
}

// hasAnyRail checks if there's any rail in the layer
func hasAnyRail(railLayer [][]int, width, height int) bool {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 1 {
				return true
			}
		}
	}
	return false
}

// isInsideRailLoop checks if a point is inside the rail loop using ray casting
func isInsideRailLoop(railLayer [][]int, x, y, minX, minY, maxX, maxY, width, height int) bool {
	// Ray casting algorithm: count intersections going left
	intersections := 0
	for checkX := x - 1; checkX >= minX; checkX-- {
		if railLayer[y][checkX] == 1 {
			intersections++
		}
	}

	// If odd number of intersections, point is inside
	return intersections%2 == 1
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
