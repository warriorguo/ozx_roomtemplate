package generate

import "fmt"

// Bridge layer constants
const bridgeSize = 2 // Bridge is always 2x2

// Minimum gap size for placing a bridge in concave areas
const minConcaveGapSize = 4

// Island represents a connected region of ground cells
type Island struct {
	ID    int
	Cells []Point
	MinX  int
	MinY  int
	MaxX  int
	MaxY  int
}

// generateBridgeLayerWithDebug generates bridges to connect floating islands to ground/other islands
func generateBridgeLayerWithDebug(bridgeLayer, ground, softEdgeLayer [][]int, width, height int) *BridgeLayerDebugInfo {
	debug := &BridgeLayerDebugInfo{}

	// Step 1: Find all connected regions (islands) in the ground layer
	islands := findAllIslands(ground, width, height)
	debug.IslandsFound = len(islands)

	// Step 2: Connect disconnected islands (if any)
	if len(islands) > 1 {
		// The main ground (usually the largest or the one connected to doors) is island 0
		// Other islands need to be connected
		mainIslandID := 0 // Assume the first found island is the main ground

		// Find the largest island as the main ground
		maxSize := 0
		for i, island := range islands {
			if len(island.Cells) > maxSize {
				maxSize = len(island.Cells)
				mainIslandID = i
			}
		}

		// Track which islands are connected to the main ground
		connected := make(map[int]bool)
		connected[mainIslandID] = true

		// Connect each floating island to the main ground or another connected island
		for {
			// Find an unconnected island
			unconnectedID := -1
			for i := range islands {
				if !connected[i] {
					unconnectedID = i
					break
				}
			}

			if unconnectedID == -1 {
				// All islands connected
				break
			}

			unconnectedIsland := islands[unconnectedID]

			// Find the nearest connected island/ground
			bestConnection := findBestBridgeConnection(unconnectedIsland, islands, connected, ground, bridgeLayer, softEdgeLayer, width, height)

			if bestConnection == nil {
				// Cannot connect this island
				debug.Misses = append(debug.Misses, MissInfo{
					Reason: fmt.Sprintf("cannot find valid bridge path for island at (%d,%d)-(%d,%d)",
						unconnectedIsland.MinX, unconnectedIsland.MinY, unconnectedIsland.MaxX, unconnectedIsland.MaxY),
				})
				// Mark as connected anyway to avoid infinite loop
				connected[unconnectedID] = true
				continue
			}

			// Place the bridge
			placeBridge(bridgeLayer, bestConnection.bridgeX, bestConnection.bridgeY)
			connected[unconnectedID] = true
			debug.BridgesPlaced++

			debug.Connections = append(debug.Connections, BridgeConnection{
				From:     fmt.Sprintf("island (%d,%d)-(%d,%d)", unconnectedIsland.MinX, unconnectedIsland.MinY, unconnectedIsland.MaxX, unconnectedIsland.MaxY),
				To:       bestConnection.targetDesc,
				Position: fmt.Sprintf("(%d,%d)", bestConnection.bridgeX, bestConnection.bridgeY),
				Size:     "2x2",
			})
		}
	}

	// Step 3: Fill concave gaps with bridges
	// Look for horizontal gaps where ground exists on both sides but void in middle,
	// and ground exists above the void (creating a concave shape)
	concaveGapBridges := fillConcaveGapsWithBridges(bridgeLayer, ground, softEdgeLayer, width, height)
	debug.ConcaveGapBridges = concaveGapBridges
	debug.BridgesPlaced += len(concaveGapBridges)

	if debug.BridgesPlaced == 0 {
		debug.Skipped = true
		debug.SkipReason = "no bridges needed (no floating islands and no concave gaps)"
	}

	return debug
}

// bridgeConnectionResult holds the result of finding a bridge connection
type bridgeConnectionResult struct {
	bridgeX    int
	bridgeY    int
	targetDesc string
}

// findBestBridgeConnection finds the best position to place a 2x2 bridge connecting an island to existing ground
func findBestBridgeConnection(island Island, allIslands []Island, connected map[int]bool, ground, bridgeLayer, softEdgeLayer [][]int, width, height int) *bridgeConnectionResult {
	// For each edge cell of the island, try to find a valid bridge position
	// The bridge must touch the island (2x2 fully adjacent) and also touch ground or another connected island

	type candidate struct {
		bridgeX, bridgeY int
		distance         int
		targetDesc       string
	}
	var candidates []candidate

	// Check all possible 2x2 bridge positions around the island
	// A bridge at (bx, by) occupies cells (bx, by), (bx+1, by), (bx, by+1), (bx+1, by+1)
	// It must touch the island and must connect to existing ground

	for _, cell := range island.Cells {
		// Try all possible 2x2 bridge positions adjacent to this cell.
		// The bridge top-left corner (bx,by) can be at various offsets so that
		// the 2x2 block is directly adjacent to the island cell.
		offsets := []struct{ dx, dy int }{
			{-2, -1}, {-2, 0}, {-2, 1},  // bridge to left
			{2, -1}, {2, 0}, {2, 1},      // bridge to right (cell.X+2 is bridge left edge)
			{-1, -2}, {0, -2}, {1, -2},   // bridge above
			{-1, 2}, {0, 2}, {1, 2},      // bridge below
			// Also try positions where bridge overlaps the gap directly adjacent
			{-1, -1}, {-1, 0}, {-1, 1},   // bridge shifted left by 1
			{1, -1}, {1, 0}, {1, 1},      // bridge shifted right by 1
			{0, -1}, {0, 1},              // bridge shifted up/down by 1
		}

		for _, off := range offsets {
			bx := cell.X + off.dx
			by := cell.Y + off.dy

			if !canPlaceBridge(bx, by, ground, bridgeLayer, softEdgeLayer, width, height) {
				continue
			}

			// Check if bridge touches the island (at least 2 adjacent cells)
			if !bridgeTouchesIsland(bx, by, island, ground) {
				continue
			}

			// Check if bridge touches existing ground (not part of this island)
			touchesGround, targetDesc := bridgeTouchesExistingGround(bx, by, island, allIslands, connected, ground, width, height)
			if !touchesGround {
				continue
			}

			// Calculate distance (for prioritization)
			centerX := (island.MinX + island.MaxX) / 2
			centerY := (island.MinY + island.MaxY) / 2
			dist := abs(bx-centerX) + abs(by-centerY)

			candidates = append(candidates, candidate{bx, by, dist, targetDesc})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by distance and pick the closest
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.distance < best.distance {
			best = c
		}
	}

	return &bridgeConnectionResult{
		bridgeX:    best.bridgeX,
		bridgeY:    best.bridgeY,
		targetDesc: best.targetDesc,
	}
}

// placeBridge places a 2x2 bridge at the given position
func placeBridge(bridgeLayer [][]int, x, y int) {
	for dy := 0; dy < bridgeSize; dy++ {
		for dx := 0; dx < bridgeSize; dx++ {
			bridgeLayer[y+dy][x+dx] = 1
		}
	}
}

// fillConcaveGapsWithBridges finds horizontal concave gaps and places bridges to fill them
// A concave gap is: ground on both sides of a row, void in middle, and ground exists above the void
func fillConcaveGapsWithBridges(bridgeLayer, ground, softEdgeLayer [][]int, width, height int) []BridgeConnection {
	var connections []BridgeConnection

	// Scan each row for horizontal concave gaps
	for y := 1; y < height-1; y++ { // Skip first and last row
		// Find gaps in this row: segments of void cells between ground cells
		gaps := findHorizontalGaps(ground, y, width)

		for _, gap := range gaps {
			// Check if this is a concave gap (ground exists above the void)
			if !isConcaveGap(ground, gap.startX, gap.endX, y, width, height) {
				continue
			}

			// Gap must be wide enough
			gapWidth := gap.endX - gap.startX
			if gapWidth < minConcaveGapSize {
				continue
			}

			// Try placing bridge adjacent to left ground (gap.startX is first void cell)
			// The bridge should be at gap.startX so it touches ground at gap.startX-1
			placed := false
			// Try positions from left side to right side of gap
			for bx := gap.startX; bx+bridgeSize <= gap.endX; bx++ {
				bridgeY := y
				if !canPlaceBridgeAt(bridgeLayer, ground, softEdgeLayer, bx, bridgeY, width, height) {
					continue
				}
				// Verify bridge actually touches ground on at least one side
				if !bridgeTouchesGroundDirectly(bx, bridgeY, ground, width, height) {
					continue
				}
				placeBridge(bridgeLayer, bx, bridgeY)
				connections = append(connections, BridgeConnection{
					From:     fmt.Sprintf("concave gap at y=%d", y),
					To:       fmt.Sprintf("x=%d to x=%d", gap.startX, gap.endX-1),
					Position: fmt.Sprintf("(%d,%d)", bx, bridgeY),
					Size:     "2x2",
				})
				placed = true
				break
			}
			_ = placed
		}
	}

	return connections
}

// horizontalGap represents a gap in a row
type horizontalGap struct {
	startX int // First void cell (inclusive)
	endX   int // Last void cell (exclusive)
}

// findHorizontalGaps finds all horizontal gaps (void segments between ground) in a row
func findHorizontalGaps(ground [][]int, y, width int) []horizontalGap {
	var gaps []horizontalGap

	inGap := false
	gapStart := 0

	for x := 0; x < width; x++ {
		if ground[y][x] == 0 {
			// Void cell
			if !inGap {
				inGap = true
				gapStart = x
			}
		} else {
			// Ground cell
			if inGap {
				// End of gap - but only count if we started after ground
				if gapStart > 0 {
					gaps = append(gaps, horizontalGap{startX: gapStart, endX: x})
				}
				inGap = false
			}
		}
	}

	// Don't include gaps that extend to the right edge (they're not between ground)

	return gaps
}

// isConcaveGap checks if a gap has ground above it (making it concave)
func isConcaveGap(ground [][]int, startX, endX, y, width, height int) bool {
	if y == 0 {
		return false // No row above
	}

	// Check if there's ground in the row above, within the gap region
	// A concave gap needs ground above the void area
	groundAboveCount := 0
	for x := startX; x < endX; x++ {
		if ground[y-1][x] == 1 {
			groundAboveCount++
		}
	}

	// Need significant ground coverage above (at least 50% of gap width)
	gapWidth := endX - startX
	return groundAboveCount >= gapWidth/2
}

// bridgeTouchesGroundDirectly checks if a 2x2 bridge at (bx,by) has at least one
// orthogonally adjacent ground cell (not diagonal)
func bridgeTouchesGroundDirectly(bx, by int, ground [][]int, width, height int) bool {
	// All cells of the 2x2 bridge
	bridgeCells := []Point{
		{bx, by}, {bx + 1, by}, {bx, by + 1}, {bx + 1, by + 1},
	}

	for _, bc := range bridgeCells {
		// Check 4 orthogonal neighbors
		neighbors := []Point{
			{bc.X - 1, bc.Y}, {bc.X + 1, bc.Y},
			{bc.X, bc.Y - 1}, {bc.X, bc.Y + 1},
		}
		for _, n := range neighbors {
			// Skip if neighbor is part of the bridge itself
			if n.X >= bx && n.X < bx+bridgeSize && n.Y >= by && n.Y < by+bridgeSize {
				continue
			}
			if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height {
				if ground[n.Y][n.X] == 1 {
					return true
				}
			}
		}
	}
	return false
}

// findAllIslands finds all connected regions of ground cells using flood fill
func findAllIslands(ground [][]int, width, height int) []Island {
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	var islands []Island
	islandID := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if ground[y][x] == 1 && !visited[y][x] {
				// Found a new island, flood fill to find all connected cells
				island := floodFillIsland(ground, visited, x, y, width, height, islandID)
				islands = append(islands, island)
				islandID++
			}
		}
	}

	return islands
}

// floodFillIsland performs flood fill to find all cells of an island
func floodFillIsland(ground [][]int, visited [][]bool, startX, startY, width, height, id int) Island {
	island := Island{
		ID:   id,
		MinX: startX,
		MinY: startY,
		MaxX: startX,
		MaxY: startY,
	}

	// BFS flood fill
	queue := []Point{{X: startX, Y: startY}}
	visited[startY][startX] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		island.Cells = append(island.Cells, curr)

		// Update bounds
		if curr.X < island.MinX {
			island.MinX = curr.X
		}
		if curr.X > island.MaxX {
			island.MaxX = curr.X
		}
		if curr.Y < island.MinY {
			island.MinY = curr.Y
		}
		if curr.Y > island.MaxY {
			island.MaxY = curr.Y
		}

		// Check 4 neighbors
		neighbors := []Point{
			{curr.X - 1, curr.Y}, {curr.X + 1, curr.Y},
			{curr.X, curr.Y - 1}, {curr.X, curr.Y + 1},
		}

		for _, n := range neighbors {
			if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height &&
				ground[n.Y][n.X] == 1 && !visited[n.Y][n.X] {
				visited[n.Y][n.X] = true
				queue = append(queue, n)
			}
		}
	}

	return island
}
