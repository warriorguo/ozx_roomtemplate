package generate

import (
	"fmt"
	"math"
)

// ComputeMainPath finds paths through the room center connecting all required doors,
// then computes per-cell distance metrics for enemy placement.
func ComputeMainPath(ground, bridge [][]int, doorPositions map[DoorPosition]Point, width, height int) (*MainPathData, *MainPathDebugInfo) {
	debug := &MainPathDebugInfo{}

	// Build walkable grid (ground or bridge)
	walkable := make([][]bool, height)
	for y := 0; y < height; y++ {
		walkable[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			walkable[y][x] = ground[y][x] == 1 || bridge[y][x] == 1
		}
	}

	// Collect door positions
	doors := make([]Point, 0, len(doorPositions))
	for _, pos := range doorPositions {
		doors = append(doors, pos)
	}

	if len(doors) < 2 {
		debug.Misses = append(debug.Misses, "fewer than 2 doors, no main path")
		return emptyMainPathData(width, height), debug
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
			path := findCenterBiasedPath(walkable, doors[i], doors[j], centerX, centerY, width, height)
			if path != nil {
				for _, p := range path {
					onMainPath[p.Y][p.X] = true
				}
				debug.PathSegments = append(debug.PathSegments,
					fmt.Sprintf("(%d,%d)->(%d,%d) len=%d", doors[i].X, doors[i].Y, doors[j].X, doors[j].Y, len(path)))
			} else {
				debug.Misses = append(debug.Misses,
					fmt.Sprintf("no path found (%d,%d)->(%d,%d)", doors[i].X, doors[i].Y, doors[j].X, doors[j].Y))
			}
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
		DirectDistance:   directDist,
		WalkingDistance:  walkingDist,
		SquishyScore:    squishyScore,
	}, debug
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
