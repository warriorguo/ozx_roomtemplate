package generate

import "fmt"

// Bridge size constants
const bridgeSize = 2 // Default 2x2 bridge

// bridgeSizes lists all supported bridge dimensions (width, height).
// Supported sizes: 2x2 (square), 4x2 (wide horizontal), 2x4 (tall vertical).
var bridgeSizes = []struct{ w, h int }{
	{2, 2},
	{4, 2},
	{2, 4},
}

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
			placeBridge(bridgeLayer, bestConnection.bridgeX, bestConnection.bridgeY, bestConnection.bridgeW, bestConnection.bridgeH)
			connected[unconnectedID] = true
			debug.BridgesPlaced++

			debug.Connections = append(debug.Connections, BridgeConnection{
				From:     fmt.Sprintf("island (%d,%d)-(%d,%d)", unconnectedIsland.MinX, unconnectedIsland.MinY, unconnectedIsland.MaxX, unconnectedIsland.MaxY),
				To:       bestConnection.targetDesc,
				Position: fmt.Sprintf("(%d,%d)", bestConnection.bridgeX, bestConnection.bridgeY),
				Size:     fmt.Sprintf("%dx%d", bestConnection.bridgeW, bestConnection.bridgeH),
			})
		}
	}

	// Step 3: Fill concave gaps with bridges
	// Look for horizontal gaps where ground exists on both sides but void in middle,
	// and ground exists above the void (creating a concave shape)
	concaveGapBridges := fillConcaveGapsWithBridges(bridgeLayer, ground, softEdgeLayer, width, height)
	debug.ConcaveGapBridges = concaveGapBridges
	debug.BridgesPlaced += len(concaveGapBridges)

	// Step 4: Fallback — if no bridges were placed at all, force at least one.
	// A bridge room with zero bridge tiles is indistinguishable from a flat ground
	// room, which defeats the purpose of the bridge room type.
	if debug.BridgesPlaced == 0 {
		placed := placeAtLeastOneBridge(bridgeLayer, ground, softEdgeLayer, width, height)
		if placed != nil {
			debug.BridgesPlaced++
			debug.Connections = append(debug.Connections, BridgeConnection{
				From:     "fallback",
				To:       "adjacent ground",
				Position: fmt.Sprintf("(%d,%d)", placed.X, placed.Y),
				Size:     fmt.Sprintf("%dx%d", bridgeSize, bridgeSize),
			})
		} else {
			debug.Skipped = true
			debug.SkipReason = "no bridges needed (no floating islands, no concave gaps, no valid void positions)"
		}
	}

	return debug
}

// placeAtLeastOneBridge scans for any valid 2x2 void area adjacent to ground and places
// a bridge there. Returns the top-left corner of the placed bridge, or nil if none found.
func placeAtLeastOneBridge(bridgeLayer, ground, softEdgeLayer [][]int, width, height int) *Point {
	// Prefer positions near the vertical center of the room
	centerY := height / 2

	type candidate struct {
		pos  Point
		dist int // distance from center row
	}
	var candidates []candidate

	for y := 0; y <= height-bridgeSize; y++ {
		for x := 0; x <= width-bridgeSize; x++ {
			if canPlaceBridge(x, y, bridgeSize, bridgeSize, ground, bridgeLayer, softEdgeLayer, width, height) {
				dist := abs(y - centerY)
				if dist > height-centerY {
					dist = height - centerY - dist
				}
				candidates = append(candidates, candidate{Point{x, y}, dist})
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Pick the candidate closest to center
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.dist < best.dist {
			best = c
		}
	}

	placeBridge(bridgeLayer, best.pos.X, best.pos.Y, bridgeSize, bridgeSize)
	return &best.pos
}

// bridgeConnectionResult holds the result of finding a bridge connection
type bridgeConnectionResult struct {
	bridgeX    int
	bridgeY    int
	bridgeW    int
	bridgeH    int
	targetDesc string
}

// findBestBridgeConnection finds the best position to place a bridge connecting an island to existing ground.
// It tries all supported bridge sizes (2x2, 4x2, 2x4) and picks the closest valid candidate.
func findBestBridgeConnection(island Island, allIslands []Island, connected map[int]bool, ground, bridgeLayer, softEdgeLayer [][]int, width, height int) *bridgeConnectionResult {
	type candidate struct {
		bridgeX, bridgeY int
		bridgeW, bridgeH int
		distance         int
		targetDesc       string
	}
	var candidates []candidate

	for _, sz := range bridgeSizes {
		bw, bh := sz.w, sz.h

		for _, cell := range island.Cells {
			// Try bridge positions adjacent to this island cell.
			// Generate offsets so that the bw×bh block is directly adjacent.
			var offsets []struct{ dx, dy int }

			// Left side: right edge of bridge (bx+bw-1) at cell.X-1 → bx = cell.X - bw
			for dy := -(bh - 1); dy <= (bh - 1); dy++ {
				offsets = append(offsets, struct{ dx, dy int }{-bw, dy})
			}
			// Right side: left edge of bridge (bx) at cell.X+1 → bx = cell.X + 1
			for dy := -(bh - 1); dy <= (bh - 1); dy++ {
				offsets = append(offsets, struct{ dx, dy int }{1, dy})
			}
			// Above: bottom edge of bridge (by+bh-1) at cell.Y-1 → by = cell.Y - bh
			for dx := -(bw - 1); dx <= (bw - 1); dx++ {
				offsets = append(offsets, struct{ dx, dy int }{dx, -bh})
			}
			// Below: top edge of bridge (by) at cell.Y+1 → by = cell.Y + 1
			for dx := -(bw - 1); dx <= (bw - 1); dx++ {
				offsets = append(offsets, struct{ dx, dy int }{dx, 1})
			}
			// Shifted positions (bridge partially overlaps the gap)
			for dy := -(bh - 1); dy <= (bh - 1); dy++ {
				for dx := -(bw - 1); dx <= (bw - 1); dx++ {
					offsets = append(offsets, struct{ dx, dy int }{dx, dy})
				}
			}

			for _, off := range offsets {
				bx := cell.X + off.dx
				by := cell.Y + off.dy

				if !canPlaceBridge(bx, by, bw, bh, ground, bridgeLayer, softEdgeLayer, width, height) {
					continue
				}

				// Check if bridge touches the island (at least 2 adjacent cells)
				if !bridgeTouchesIsland(bx, by, bw, bh, island, ground) {
					continue
				}

				// Check if bridge touches existing ground (not part of this island)
				touchesGround, targetDesc := bridgeTouchesExistingGround(bx, by, bw, bh, island, allIslands, connected, ground, width, height)
				if !touchesGround {
					continue
				}

				// Calculate distance (for prioritization)
				centerX := (island.MinX + island.MaxX) / 2
				centerY := (island.MinY + island.MaxY) / 2
				dist := abs(bx-centerX) + abs(by-centerY)

				candidates = append(candidates, candidate{bx, by, bw, bh, dist, targetDesc})
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Pick the candidate with the smallest distance
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.distance < best.distance {
			best = c
		}
	}

	return &bridgeConnectionResult{
		bridgeX:    best.bridgeX,
		bridgeY:    best.bridgeY,
		bridgeW:    best.bridgeW,
		bridgeH:    best.bridgeH,
		targetDesc: best.targetDesc,
	}
}

// placeBridge places a bridge of given size (bw x bh) at position (x, y).
func placeBridge(bridgeLayer [][]int, x, y, bw, bh int) {
	for dy := 0; dy < bh; dy++ {
		for dx := 0; dx < bw; dx++ {
			bridgeLayer[y+dy][x+dx] = 1
		}
	}
}

// fillConcaveGapsWithBridges finds horizontal concave gaps and places bridges to fill them.
// A concave gap is: ground on both sides of a row, void in middle, and ground exists above the void.
// The bridge size is chosen based on gap direction:
//   - horizontal gap (width > 2) → prefer 4x2
//   - small or square gap         → 2x2
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

			// Choose bridge size based on gap geometry.
			// For a horizontal gap prefer the wider 4x2 brush; fall back to 2x2.
			var sizePriority []struct{ w, h int }
			if gapWidth >= 4 {
				sizePriority = []struct{ w, h int }{{4, 2}, {2, 2}}
			} else {
				sizePriority = []struct{ w, h int }{{2, 2}}
			}

			placed := false
			for _, sz := range sizePriority {
				if placed {
					break
				}
				// Try positions from left side to right side of gap
				for bx := gap.startX; bx+sz.w <= gap.endX; bx++ {
					bridgeY := y
					if !canPlaceBridgeAt(bridgeLayer, ground, softEdgeLayer, bx, bridgeY, sz.w, sz.h, width, height) {
						continue
					}
					// Verify bridge actually touches ground on at least one side
					if !bridgeTouchesGroundDirectly(bx, bridgeY, sz.w, sz.h, ground, width, height) {
						continue
					}
					placeBridge(bridgeLayer, bx, bridgeY, sz.w, sz.h)
					connections = append(connections, BridgeConnection{
						From:     fmt.Sprintf("concave gap at y=%d", y),
						To:       fmt.Sprintf("x=%d to x=%d", gap.startX, gap.endX-1),
						Position: fmt.Sprintf("(%d,%d)", bx, bridgeY),
						Size:     fmt.Sprintf("%dx%d", sz.w, sz.h),
					})
					placed = true
					break
				}
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

// bridgeTouchesGroundDirectly checks if a bw×bh bridge at (bx,by) has at least one
// orthogonally adjacent ground cell (not diagonal).
func bridgeTouchesGroundDirectly(bx, by, bw, bh int, ground [][]int, width, height int) bool {
	for dy := 0; dy < bh; dy++ {
		for dx := 0; dx < bw; dx++ {
			cx, cy := bx+dx, by+dy
			neighbors := []Point{
				{cx - 1, cy}, {cx + 1, cy},
				{cx, cy - 1}, {cx, cy + 1},
			}
			for _, n := range neighbors {
				// Skip if neighbor is part of the bridge itself
				if n.X >= bx && n.X < bx+bw && n.Y >= by && n.Y < by+bh {
					continue
				}
				if n.X >= 0 && n.X < width && n.Y >= 0 && n.Y < height {
					if ground[n.Y][n.X] == 1 {
						return true
					}
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
