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

// RailPlatform represents a walkable area suitable for rail placement.
// Can be a simple rectangle (Cells == nil) or a non-rectangular merged shape.
type RailPlatform struct {
	X, Y          int      // Bounding box top-left corner
	Width, Height int      // Bounding box size
	Cells         [][]bool // nil for simple rectangles; set for merged shapes
}

// isFilled checks if a local coordinate (relative to X,Y) is filled
func (p RailPlatform) isFilled(lx, ly int) bool {
	if lx < 0 || lx >= p.Width || ly < 0 || ly >= p.Height {
		return false
	}
	if p.Cells == nil {
		return true // simple rectangle: all cells filled
	}
	return p.Cells[ly][lx]
}

// isPerimeter checks if a local coordinate is on the outer perimeter
func (p RailPlatform) isPerimeter(lx, ly int) bool {
	if !p.isFilled(lx, ly) {
		return false
	}
	return !p.isFilled(lx-1, ly) || !p.isFilled(lx+1, ly) ||
		!p.isFilled(lx, ly-1) || !p.isFilled(lx, ly+1)
}

// area returns the number of filled cells
func (p RailPlatform) area() int {
	if p.Cells == nil {
		return p.Width * p.Height
	}
	count := 0
	for ly := 0; ly < p.Height; ly++ {
		for lx := 0; lx < p.Width; lx++ {
			if p.Cells[ly][lx] {
				count++
			}
		}
	}
	return count
}

// perimeterCells returns all cells on the outer perimeter in global coordinates
func (p RailPlatform) perimeterCells() []Point {
	var cells []Point
	for ly := 0; ly < p.Height; ly++ {
		for lx := 0; lx < p.Width; lx++ {
			if p.isPerimeter(lx, ly) {
				cells = append(cells, Point{X: p.X + lx, Y: p.Y + ly})
			}
		}
	}
	return cells
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

// findRailPlatforms finds solid rectangular walkable areas >= 6x6 with edge distance >= 2.
// Searches from all 4 corners, picks the 2 largest, and merges them into a combined outline.
func findRailPlatforms(ground, bridge [][]int, width, height int) []RailPlatform {
	isWalkable := func(x, y int) bool {
		return ground[y][x] == 1 || bridge[y][x] == 1
	}

	minX, minY := minRailEdgeDistance, minRailEdgeDistance
	maxX, maxY := width-minRailEdgeDistance, height-minRailEdgeDistance

	// findMaxRect finds the largest solid rectangle expanding from a corner direction.
	// dirX/dirY: +1 = expand right/down, -1 = expand left/up
	findMaxRect := func(startX, startY, dirX, dirY int) RailPlatform {
		var best RailPlatform
		bestArea := 0

		// Find max width in the starting row
		maxW := 0
		for x := startX; x >= minX && x < maxX; x += dirX {
			if !isWalkable(x, startY) {
				break
			}
			maxW++
		}
		if maxW < minPlatformForRail {
			return best
		}

		curW := maxW
		h := 0
		for y := startY; y >= minY && y < maxY; y += dirY {
			rowW := 0
			if dirX > 0 {
				for x := startX; x < startX+curW && x < maxX && isWalkable(x, y); x++ {
					rowW++
				}
			} else {
				for x := startX; x > startX-curW && x >= minX && isWalkable(x, y); x-- {
					rowW++
				}
			}
			if rowW < minPlatformForRail {
				break
			}
			curW = rowW
			h++

			if h >= minPlatformForRail && curW*h > bestArea {
				bestArea = curW * h
				// Normalize to top-left origin
				rx := startX
				ry := startY
				if dirX < 0 {
					rx = startX - curW + 1
				}
				if dirY < 0 {
					ry = y
				}
				best = RailPlatform{X: rx, Y: ry, Width: curW, Height: h}
			}
		}
		return best
	}

	// Search from 4 corners
	rects := []RailPlatform{
		findMaxRect(minX, minY, +1, +1),     // top-left
		findMaxRect(maxX-1, minY, -1, +1),   // top-right
		findMaxRect(minX, maxY-1, +1, -1),   // bottom-left
		findMaxRect(maxX-1, maxY-1, -1, -1), // bottom-right
	}

	// Fallback: if corners didn't find enough, scan for the first walkable point and expand
	hasValid := false
	for _, r := range rects {
		if r.Width*r.Height >= minPlatformForRail*minPlatformForRail {
			hasValid = true
			break
		}
	}
	if !hasValid {
		// Scan top-left to bottom-right for any walkable starting point
		for sy := minY; sy <= maxY-minPlatformForRail; sy++ {
			for sx := minX; sx <= maxX-minPlatformForRail; sx++ {
				if isWalkable(sx, sy) {
					r := findMaxRect(sx, sy, +1, +1)
					if r.Width*r.Height >= minPlatformForRail*minPlatformForRail {
						rects = append(rects, r)
						hasValid = true
						break
					}
				}
			}
			if hasValid {
				break
			}
		}
	}

	// Sort by area descending, pick top 2
	type rectWithArea struct {
		r    RailPlatform
		area int
	}
	var valid []rectWithArea
	for _, r := range rects {
		a := r.Width * r.Height
		if a >= minPlatformForRail*minPlatformForRail {
			valid = append(valid, rectWithArea{r, a})
		}
	}
	if len(valid) == 0 {
		return nil
	}

	// Sort descending by area
	for i := 0; i < len(valid)-1; i++ {
		for j := i + 1; j < len(valid); j++ {
			if valid[j].area > valid[i].area {
				valid[i], valid[j] = valid[j], valid[i]
			}
		}
	}

	// If only 1 valid or the two largest don't overlap, return the biggest
	if len(valid) == 1 {
		return []RailPlatform{valid[0].r}
	}

	r1, r2 := valid[0].r, valid[1].r

	// Try to merge: compute union bounding box
	merged := mergeRailPlatforms(r1, r2)
	if merged != nil {
		return []RailPlatform{*merged}
	}

	// No useful merge, return the largest
	return []RailPlatform{r1}
}

// mergeRailPlatforms merges two overlapping/adjacent rectangles into their union shape.
// The result is a non-rectangular RailPlatform with Cells set to the union of both rects.
// Returns nil if the two rectangles don't overlap or aren't adjacent.
func mergeRailPlatforms(a, b RailPlatform) *RailPlatform {
	aRight, aBottom := a.X+a.Width, a.Y+a.Height
	bRight, bBottom := b.X+b.Width, b.Y+b.Height

	// No overlap and not adjacent
	if a.X > bRight || b.X > aRight || a.Y > bBottom || b.Y > aBottom {
		return nil
	}

	// Union bounding box
	ux := a.X
	if b.X < ux {
		ux = b.X
	}
	uy := a.Y
	if b.Y < uy {
		uy = b.Y
	}
	ur := aRight
	if bRight > ur {
		ur = bRight
	}
	ub := aBottom
	if bBottom > ub {
		ub = bBottom
	}

	w, h := ur-ux, ub-uy
	cells := make([][]bool, h)
	for ly := 0; ly < h; ly++ {
		cells[ly] = make([]bool, w)
		for lx := 0; lx < w; lx++ {
			gx, gy := ux+lx, uy+ly
			inA := gx >= a.X && gx < aRight && gy >= a.Y && gy < aBottom
			inB := gx >= b.X && gx < bRight && gy >= b.Y && gy < bBottom
			cells[ly][lx] = inA || inB
		}
	}

	return &RailPlatform{
		X:      ux,
		Y:      uy,
		Width:  w,
		Height: h,
		Cells:  cells,
	}
}

// tryPlaceRailLoop attempts to place a rail loop on the given platform.
// For simple rectangles: draws a hollow rectangle perimeter with optional indents.
// For merged shapes (Cells set): draws the outer perimeter of the union shape.
func tryPlaceRailLoop(railLayer, ground, bridge [][]int, platform RailPlatform, width, height int, debug *RailDebugInfo) *RailLoopInfo {
	if platform.Cells != nil {
		return tryPlaceMergedRailLoop(railLayer, ground, bridge, platform, width, height, debug)
	}
	return tryPlaceRectRailLoop(railLayer, ground, bridge, platform, width, height, debug)
}

// tryPlaceRectRailLoop places a rail loop for a simple rectangular platform
func tryPlaceRectRailLoop(railLayer, ground, bridge [][]int, platform RailPlatform, width, height int, debug *RailDebugInfo) *RailLoopInfo {
	railX := platform.X
	railY := platform.Y
	railW := platform.Width
	railH := platform.Height

	minRailSize := minRailAreaSize + 2
	if railW < minRailSize || railH < minRailSize {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("platform (%d,%d) %dx%d too small for minimum rail size %dx%d",
				platform.X, platform.Y, platform.Width, platform.Height, minRailSize, minRailSize),
		})
		return nil
	}

	// Randomly shrink for variety
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

	if !verifyRailPositions(railX, railY, railW, railH, ground, bridge, width, height) {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("rail positions at (%d,%d) %dx%d not all on ground/bridge",
				railX, railY, railW, railH),
		})
		return nil
	}

	loopInfo := &RailLoopInfo{
		Platform:    fmt.Sprintf("(%d,%d) %dx%d", platform.X, platform.Y, platform.Width, platform.Height),
		BoundingBox: fmt.Sprintf("(%d,%d) %dx%d", railX, railY, railW, railH),
		Indents:     []IndentInfo{},
	}

	// Draw hollow rectangle
	for x := railX; x < railX+railW; x++ {
		railLayer[railY][x] = 1
		railLayer[railY+railH-1][x] = 1
	}
	for y := railY + 1; y < railY+railH-1; y++ {
		railLayer[y][railX] = 1
		railLayer[y][railX+railW-1] = 1
	}
	loopInfo.Perimeter = 2*(railW+railH) - 4

	// Add random indents
	indentCount := 0
	for indentCount < maxRailIndents {
		if rand.Intn(100) >= railIndentProb {
			break
		}
		indent := tryAddRailIndent(railLayer, ground, bridge, railX, railY, railW, railH, width, height)
		if indent != nil {
			loopInfo.Indents = append(loopInfo.Indents, *indent)
			loopInfo.Perimeter += indent.Size * 2
			indentCount++
		}
	}

	return loopInfo
}

// tryPlaceMergedRailLoop places a rail loop for a merged (non-rectangular) platform.
// Draws the outer perimeter of the union shape.
func tryPlaceMergedRailLoop(railLayer, ground, bridge [][]int, platform RailPlatform, width, height int, debug *RailDebugInfo) *RailLoopInfo {
	// Shrink the platform by 1 cell on each side internally for the rail perimeter
	shrunk := shrinkPlatform(platform)
	if shrunk == nil {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("merged platform (%d,%d) %dx%d too small after shrinking",
				platform.X, platform.Y, platform.Width, platform.Height),
		})
		return nil
	}

	// Get perimeter cells
	perimCells := shrunk.perimeterCells()
	if len(perimCells) < minRailAreaSize*4-4 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("merged platform perimeter too small (%d cells)", len(perimCells)),
		})
		return nil
	}

	// Verify all perimeter cells are on ground or bridge
	for _, cell := range perimCells {
		if cell.X < 0 || cell.X >= width || cell.Y < 0 || cell.Y >= height {
			debug.Misses = append(debug.Misses, MissInfo{
				Reason: "merged rail perimeter extends out of bounds",
			})
			return nil
		}
		if ground[cell.Y][cell.X] != 1 && bridge[cell.Y][cell.X] != 1 {
			debug.Misses = append(debug.Misses, MissInfo{
				Reason: fmt.Sprintf("merged rail cell (%d,%d) not on ground/bridge", cell.X, cell.Y),
			})
			return nil
		}
	}

	// Draw perimeter
	for _, cell := range perimCells {
		railLayer[cell.Y][cell.X] = 1
	}

	loopInfo := &RailLoopInfo{
		Platform:    fmt.Sprintf("(%d,%d) %dx%d (merged)", platform.X, platform.Y, platform.Width, platform.Height),
		BoundingBox: fmt.Sprintf("(%d,%d) %dx%d", shrunk.X, shrunk.Y, shrunk.Width, shrunk.Height),
		Perimeter:   len(perimCells),
		Indents:     []IndentInfo{},
	}

	return loopInfo
}

// shrinkPlatform shrinks a merged platform by 1 cell on all internal edges,
// producing a slightly smaller shape for the rail perimeter.
func shrinkPlatform(p RailPlatform) *RailPlatform {
	if p.Width < minRailAreaSize+2 || p.Height < minRailAreaSize+2 {
		return nil
	}

	// Shrink by 1 on each side of the bounding box
	newW, newH := p.Width-2, p.Height-2
	if newW < minRailAreaSize || newH < minRailAreaSize {
		return nil
	}

	cells := make([][]bool, newH)
	for ly := 0; ly < newH; ly++ {
		cells[ly] = make([]bool, newW)
		for lx := 0; lx < newW; lx++ {
			// Map back to original: offset by 1
			cells[ly][lx] = p.isFilled(lx+1, ly+1)
		}
	}

	return &RailPlatform{
		X:      p.X + 1,
		Y:      p.Y + 1,
		Width:  newW,
		Height: newH,
		Cells:  cells,
	}
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
