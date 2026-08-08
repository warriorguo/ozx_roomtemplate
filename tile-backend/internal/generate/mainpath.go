package generate

import (
	"fmt"
	"math"
)

// doorOrder fixes the order doors are visited in, so the pairs are walked and
// reported the same way every run. Ranging over the doorPositions map directly
// would randomize both the widening order and the text of any error.
var doorOrder = []DoorPosition{DoorTop, DoorRight, DoorBottom, DoorLeft}

// ComputeMainPath finds paths through the room center connecting all required doors,
// then computes per-cell distance metrics for enemy placement.
//
// It fails when a room's doors cannot actually be connected: a door with no
// walkable cell on its wall, or a pair of doors with no route between them.
// Both used to be recorded as debug misses and nothing more, so a room the
// player could not traverse was returned as a successful generation — the room
// even looked plausible, because findCenterBiasedPath snaps an unwalkable
// endpoint to the nearest walkable cell and drew a path between two interior
// cells instead (ORT-106).
func ComputeMainPath(ground, bridge [][]int, doorPositions map[DoorPosition]Point, width, height int) (*MainPathData, *MainPathDebugInfo, error) {
	debug := &MainPathDebugInfo{}

	// Build walkable grid (ground or bridge)
	walkable := make([][]bool, height)
	for y := 0; y < height; y++ {
		walkable[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			walkable[y][x] = ground[y][x] == 1 || bridge[y][x] == 1
		}
	}

	// Collect door positions, anchored onto the door's own edge line so the
	// path actually crosses the doorway (see snapDoorToEdge).
	type doorAnchor struct {
		side DoorPosition
		pos  Point
	}
	doors := make([]doorAnchor, 0, len(doorPositions))
	for _, side := range doorOrder {
		pos, ok := doorPositions[side]
		if !ok {
			continue
		}
		anchor := snapDoorToEdge(walkable, side, pos, width, height)
		if !walkable[anchor.Y][anchor.X] {
			return nil, debug, fmt.Errorf("door %s at (%d,%d): no walkable cell on that wall, the doorway is sealed",
				side, pos.X, pos.Y)
		}
		doors = append(doors, doorAnchor{side: side, pos: anchor})
	}

	if len(doors) < 2 {
		debug.Misses = append(debug.Misses, "fewer than 2 doors, no main path")
		return emptyMainPathData(width, height), debug, nil
	}

	// Find center of room
	centerX, centerY := width/2, height/2

	// Compute main path: for each pair of doors, find path biased toward center
	onMainPath := make([][]bool, height)
	for y := 0; y < height; y++ {
		onMainPath[y] = make([]bool, width)
	}

	// Connect all doors through center using center-biased A*
	for i := 0; i < len(doors); i++ {
		for j := i + 1; j < len(doors); j++ {
			a, b := doors[i], doors[j]
			path := findCenterBiasedPath(walkable, a.pos, b.pos, centerX, centerY, width, height)
			if path == nil {
				return nil, debug, fmt.Errorf("doors %s (%d,%d) and %s (%d,%d) are not connected by walkable ground",
					a.side, a.pos.X, a.pos.Y, b.side, b.pos.X, b.pos.Y)
			}
			markWidenedPath(onMainPath, walkable, path, centerX, centerY, width, height)
			debug.PathSegments = append(debug.PathSegments,
				fmt.Sprintf("(%d,%d)->(%d,%d) len=%d", a.pos.X, a.pos.Y, b.pos.X, b.pos.Y, len(path)))
		}
	}

	// Count path cells
	pathCellCount := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if onMainPath[y][x] {
				pathCellCount++
			}
		}
	}
	debug.PathCellCount = pathCellCount

	// Compute direct distance (Chebyshev) and walking distance (BFS) from each cell to main path
	directDist := computeDirectDistance(onMainPath, width, height)
	walkingDist := computeWalkingDistance(onMainPath, walkable, width, height)

	// Compute squishy score
	squishyScore := make([][]float64, height)
	for y := 0; y < height; y++ {
		squishyScore[y] = make([]float64, width)
		for x := 0; x < width; x++ {
			dd := directDist[y][x]
			wd := walkingDist[y][x]
			if dd > 0 && wd > 0 {
				squishyScore[y][x] = float64(wd) / float64(dd)
			} else if dd == 0 {
				squishyScore[y][x] = 0 // on main path
			} else {
				squishyScore[y][x] = 0 // unreachable
			}
		}
	}

	return &MainPathData{
		Width:           width,
		Height:          height,
		OnMainPath:      onMainPath,
		DirectDistance:  directDist,
		WalkingDistance: walkingDist,
		SquishyScore:    squishyScore,
	}, debug, nil
}

// mainPathWidth is how many tiles wide the main path corridor is. The A* trace
// itself is a single cell thick; markWidenedPath thickens it to this width.
const mainPathWidth = 2

// markWidenedPath stamps a mainPathWidth-square block at every cell of a
// 4-connected path, producing a corridor mainPathWidth tiles wide. Consecutive
// stamps along a straight run overlap, so the run comes out exactly that wide
// rather than accumulating; bends come out as a corridor-width elbow.
//
// Each block widens toward the room center, so the corridor straddles the
// middle of the room rather than drifting to one side of the trace — a run
// along the center line comes out spanning the center, not sitting beside it.
// The direction flips only when the preferred side would leave the grid.
//
// Every trace cell is stamped as the block's anchor, so the door cells the A*
// trace starts and ends on are always on the resulting path. Cells that aren't
// walkable are skipped — the path stays thin where it hugs a wall or squeezes
// through a one-tile door mouth, since there is nothing to widen into.
func markWidenedPath(onMainPath, walkable [][]bool, path []Point, centerX, centerY, width, height int) {
	span := mainPathWidth - 1

	for _, p := range path {
		// Step toward the center; on the center line itself step back so the
		// pair straddles it (e.g. width 16 -> centerX 8 -> columns 7 and 8).
		stepX, stepY := -1, -1
		if p.X < centerX {
			stepX = 1
		}
		if p.Y < centerY {
			stepY = 1
		}
		if x := p.X + span*stepX; x < 0 || x >= width {
			stepX = -stepX
		}
		if y := p.Y + span*stepY; y < 0 || y >= height {
			stepY = -stepY
		}

		for i := 0; i <= span; i++ {
			for j := 0; j <= span; j++ {
				x, y := p.X+i*stepX, p.Y+j*stepY
				if x < 0 || x >= width || y < 0 || y >= height {
					continue
				}
				if walkable[y][x] {
					onMainPath[y][x] = true
				}
			}
		}
	}
}

// findCenterBiasedPath finds a path from start to end that prefers going through the center.
// Uses weighted BFS (Dijkstra) where cells closer to center have lower cost.
func findCenterBiasedPath(walkable [][]bool, start, end Point, centerX, centerY, width, height int) []Point {
	type node struct {
		pos  Point
		cost float64
	}

	// Cost function: lower cost for cells near center
	maxDist := float64(width + height)
	cellCost := func(p Point) float64 {
		distToCenter := math.Abs(float64(p.X-centerX)) + math.Abs(float64(p.Y-centerY))
		// Cells near center cost 1.0, cells far from center cost up to 3.0
		return 1.0 + 2.0*(distToCenter/maxDist)
	}

	// Find nearest walkable to start and end
	startP := findNearestWalkablePoint(walkable, start, width, height)
	endP := findNearestWalkablePoint(walkable, end, width, height)
	if startP.X < 0 || endP.X < 0 {
		return nil
	}

	// Dijkstra
	dist := make([][]float64, height)
	prev := make([][]Point, height)
	for y := 0; y < height; y++ {
		dist[y] = make([]float64, width)
		prev[y] = make([]Point, width)
		for x := 0; x < width; x++ {
			dist[y][x] = math.Inf(1)
			prev[y][x] = Point{-1, -1}
		}
	}
	dist[startP.Y][startP.X] = 0

	// Simple priority queue using a slice (adequate for room sizes up to 200x200)
	queue := []node{{startP, 0}}

	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}

	for len(queue) > 0 {
		// Find min cost node
		minIdx := 0
		for i := 1; i < len(queue); i++ {
			if queue[i].cost < queue[minIdx].cost {
				minIdx = i
			}
		}
		curr := queue[minIdx]
		queue = append(queue[:minIdx], queue[minIdx+1:]...)

		if curr.pos.X == endP.X && curr.pos.Y == endP.Y {
			break
		}

		if curr.cost > dist[curr.pos.Y][curr.pos.X] {
			continue
		}

		for i := 0; i < 4; i++ {
			nx, ny := curr.pos.X+dx[i], curr.pos.Y+dy[i]
			if nx >= 0 && nx < width && ny >= 0 && ny < height && walkable[ny][nx] {
				newCost := curr.cost + cellCost(Point{nx, ny})
				if newCost < dist[ny][nx] {
					dist[ny][nx] = newCost
					prev[ny][nx] = curr.pos
					queue = append(queue, node{Point{nx, ny}, newCost})
				}
			}
		}
	}

	// Reconstruct path
	if math.IsInf(dist[endP.Y][endP.X], 1) {
		return nil
	}

	var path []Point
	cur := endP
	for cur.X != -1 {
		path = append(path, cur)
		if cur.X == startP.X && cur.Y == startP.Y {
			break
		}
		cur = prev[cur.Y][cur.X]
	}

	// Reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

// snapDoorToEdge moves a door anchor onto the nearest walkable cell along that
// door's own edge — the top/bottom row for a top/bottom door, the left/right
// column for a left/right door.
//
// Door anchors are the geometric midpoint of each wall, but the ground
// generators carve pits and erase corners afterwards, so the midpoint is often
// not walkable. The generic ring search in findNearestWalkablePoint would then
// snap inward, one row off the wall, and the main path would stop short of the
// doorway instead of crossing it. Staying on the edge line keeps the path
// running out through the room's actual opening on that wall.
//
// If the whole edge line is unwalkable there is no doorway to reach, so the
// anchor is returned untouched and findNearestWalkablePoint handles it.
func snapDoorToEdge(walkable [][]bool, side DoorPosition, pos Point, width, height int) Point {
	if pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height && walkable[pos.Y][pos.X] {
		return pos
	}

	switch side {
	case DoorTop, DoorBottom:
		for d := 1; d < width; d++ {
			for _, x := range [2]int{pos.X - d, pos.X + d} {
				if x >= 0 && x < width && walkable[pos.Y][x] {
					return Point{X: x, Y: pos.Y}
				}
			}
		}
	case DoorLeft, DoorRight:
		for d := 1; d < height; d++ {
			for _, y := range [2]int{pos.Y - d, pos.Y + d} {
				if y >= 0 && y < height && walkable[y][pos.X] {
					return Point{X: pos.X, Y: y}
				}
			}
		}
	}

	return pos
}

// findNearestWalkablePoint finds the nearest walkable cell to pos
func findNearestWalkablePoint(walkable [][]bool, pos Point, width, height int) Point {
	if pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height && walkable[pos.Y][pos.X] {
		return pos
	}
	for r := 1; r < width+height; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if abs(dx) != r && abs(dy) != r {
					continue
				}
				nx, ny := pos.X+dx, pos.Y+dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height && walkable[ny][nx] {
					return Point{nx, ny}
				}
			}
		}
	}
	return Point{-1, -1}
}

// computeDirectDistance computes Chebyshev distance from each cell to nearest main path cell
func computeDirectDistance(onMainPath [][]bool, width, height int) [][]int {
	dist := make([][]int, height)
	for y := 0; y < height; y++ {
		dist[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if onMainPath[y][x] {
				dist[y][x] = 0
			} else {
				dist[y][x] = width + height // large initial value
			}
		}
	}

	// Multi-source BFS from all main path cells (Manhattan distance)
	type qItem struct{ x, y int }
	queue := make([]qItem, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if onMainPath[y][x] {
				queue = append(queue, qItem{x, y})
			}
		}
	}

	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for i := 0; i < 4; i++ {
			nx, ny := cur.x+dx[i], cur.y+dy[i]
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				newDist := dist[cur.y][cur.x] + 1
				if newDist < dist[ny][nx] {
					dist[ny][nx] = newDist
					queue = append(queue, qItem{nx, ny})
				}
			}
		}
	}

	return dist
}

// computeWalkingDistance computes BFS walking distance from each cell to nearest main path cell
// Only walks through walkable cells. Returns -1 for unreachable cells.
func computeWalkingDistance(onMainPath [][]bool, walkable [][]bool, width, height int) [][]int {
	dist := make([][]int, height)
	for y := 0; y < height; y++ {
		dist[y] = make([]int, width)
		for x := 0; x < width; x++ {
			dist[y][x] = -1
		}
	}

	// Multi-source BFS from all main path cells
	type qItem struct{ x, y int }
	queue := make([]qItem, 0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if onMainPath[y][x] {
				dist[y][x] = 0
				queue = append(queue, qItem{x, y})
			}
		}
	}

	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for i := 0; i < 4; i++ {
			nx, ny := cur.x+dx[i], cur.y+dy[i]
			if nx >= 0 && nx < width && ny >= 0 && ny < height && walkable[ny][nx] && dist[ny][nx] == -1 {
				dist[ny][nx] = dist[cur.y][cur.x] + 1
				queue = append(queue, qItem{nx, ny})
			}
		}
	}

	return dist
}

// emptyMainPathData returns an empty MainPathData
func emptyMainPathData(width, height int) *MainPathData {
	onMainPath := make([][]bool, height)
	directDist := make([][]int, height)
	walkingDist := make([][]int, height)
	squishyScore := make([][]float64, height)
	for y := 0; y < height; y++ {
		onMainPath[y] = make([]bool, width)
		directDist[y] = make([]int, width)
		walkingDist[y] = make([]int, width)
		squishyScore[y] = make([]float64, width)
	}
	return &MainPathData{
		Width: width, Height: height,
		OnMainPath: onMainPath, DirectDistance: directDist,
		WalkingDistance: walkingDist, SquishyScore: squishyScore,
	}
}
