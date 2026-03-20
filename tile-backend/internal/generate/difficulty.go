package generate

import "math"

// DifficultyScore holds the computed difficulty rating for a room
type DifficultyScore struct {
	Terrain float64          `json:"terrain"` // 0-1
	Enemy   float64          `json:"enemy"`   // 0-1
	Overall float64          `json:"overall"` // 0-1
	Details DifficultyDetail `json:"details"`
}

// DifficultyDetail holds per-factor breakdown
type DifficultyDetail struct {
	// Terrain factors
	GroundCoverage    float64 `json:"groundCoverage"`    // ratio 0-1 (lower = harder)
	NarrowPassages    int     `json:"narrowPassages"`    // count of 1-2 wide corridors
	SoftEdgeCount     int     `json:"softEdgeCount"`     // number of soft edge cells
	PathTortuosity    float64 `json:"pathTortuosity"`    // main path actual/straight ratio
	PathStaticBlocks  int     `json:"pathStaticBlocks"`  // statics within 2 of main path
	IslandCount       int     `json:"islandCount"`       // disconnected ground regions

	// Enemy factors
	ChaserCount       int     `json:"chaserCount"`
	ZonerCount        int     `json:"zonerCount"`
	DPSCount          int     `json:"dpsCount"`
	MobAirCount       int     `json:"mobAirCount"`
	EnemyDensity      float64 `json:"enemyDensity"`      // enemies / walkable area
	EnemyConcentration float64 `json:"enemyConcentration"` // spatial clustering 0-1 (higher = more clustered)
}

// ComputeDifficulty calculates a difficulty score for the generated room
func ComputeDifficulty(ground, softEdge, staticLayer, chaserLayer, zonerLayer, dpsLayer, mobAirLayer [][]int,
	mainPath *MainPathData, width, height int) *DifficultyScore {

	details := DifficultyDetail{}

	// === Terrain factors ===

	// 1. Ground coverage
	walkable := 0
	total := width * height
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if ground[y][x] == 1 {
				walkable++
			}
		}
	}
	details.GroundCoverage = float64(walkable) / float64(total)

	// 2. Narrow passages (cells with ground where only 1-2 orthogonal neighbors are ground)
	details.NarrowPassages = countNarrowPassages(ground, width, height)

	// 3. SoftEdge count
	details.SoftEdgeCount = countCells(softEdge)

	// 4. Path tortuosity
	details.PathTortuosity = computePathTortuosity(mainPath, width, height)

	// 5. Statics near main path
	details.PathStaticBlocks = countStaticsNearPath(staticLayer, mainPath, width, height)

	// 6. Island count
	details.IslandCount = countIslands(ground, width, height)

	// === Enemy factors ===
	details.ChaserCount = countCells(chaserLayer)
	details.ZonerCount = countCells(zonerLayer)
	details.DPSCount = countCells(dpsLayer)
	details.MobAirCount = countCells(mobAirLayer)

	totalEnemies := details.ChaserCount + details.ZonerCount + details.DPSCount + details.MobAirCount
	if walkable > 0 {
		details.EnemyDensity = float64(totalEnemies) / float64(walkable)
	}

	details.EnemyConcentration = computeEnemyConcentration(chaserLayer, zonerLayer, dpsLayer, mobAirLayer, width, height)

	// === Score computation ===
	terrain := computeTerrainScore(details, width, height)
	enemy := computeEnemyScore(details, width, height)
	overall := terrain*0.4 + enemy*0.6

	return &DifficultyScore{
		Terrain: clamp01(terrain),
		Enemy:   clamp01(enemy),
		Overall: clamp01(overall),
		Details: details,
	}
}

// computeTerrainScore combines terrain factors into a 0-1 score
func computeTerrainScore(d DifficultyDetail, width, height int) float64 {
	score := 0.0
	area := float64(width * height)

	// Low ground coverage = harder (weight 0.25)
	// coverage 1.0 → 0 difficulty, coverage 0.3 → 1.0 difficulty
	coverageDiff := clamp01((1.0 - d.GroundCoverage) / 0.7)
	score += coverageDiff * 0.25

	// Narrow passages (weight 0.2)
	// 0 passages → 0, 10+ passages → 1.0
	passageDiff := clamp01(float64(d.NarrowPassages) / 10.0)
	score += passageDiff * 0.2

	// SoftEdge count (weight 0.1)
	// normalized by room perimeter
	perimeter := float64(2 * (width + height))
	softEdgeDiff := clamp01(float64(d.SoftEdgeCount) / perimeter)
	score += softEdgeDiff * 0.1

	// Path tortuosity (weight 0.2)
	// 1.0 = straight, 3.0+ = very winding
	tortDiff := clamp01((d.PathTortuosity - 1.0) / 2.0)
	score += tortDiff * 0.2

	// Path static blocks (weight 0.15)
	// normalized by area
	staticBlockDiff := clamp01(float64(d.PathStaticBlocks) / (area * 0.05))
	score += staticBlockDiff * 0.15

	// Island count (weight 0.1)
	// 1 island = 0, 4+ islands = 1.0
	islandDiff := clamp01(float64(d.IslandCount-1) / 3.0)
	score += islandDiff * 0.1

	return score
}

// computeEnemyScore combines enemy factors into a 0-1 score
func computeEnemyScore(d DifficultyDetail, width, height int) float64 {
	score := 0.0

	// Weighted enemy count (weight 0.5)
	// Chaser=1.5, Zoner=2.0, DPS=1.0, MobAir=0.8
	weightedEnemies := float64(d.ChaserCount)*1.5 + float64(d.ZonerCount)*2.0 +
		float64(d.DPSCount)*1.0 + float64(d.MobAirCount)*0.8
	// Normalize: 20 weighted enemies in a 20x12 room ≈ difficulty 1.0
	area := float64(width * height)
	expectedMax := area * 0.08 // ~8% of area as max weighted enemies
	if expectedMax < 1 {
		expectedMax = 1
	}
	enemyCountDiff := clamp01(weightedEnemies / expectedMax)
	score += enemyCountDiff * 0.5

	// Enemy density (weight 0.3)
	// 0.1 density ≈ 1.0 difficulty
	densityDiff := clamp01(d.EnemyDensity / 0.1)
	score += densityDiff * 0.3

	// Enemy concentration (weight 0.2)
	score += d.EnemyConcentration * 0.2

	return score
}

// countNarrowPassages counts ground cells that form narrow corridors (≤2 orthogonal neighbors)
func countNarrowPassages(ground [][]int, width, height int) int {
	count := 0
	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			if ground[y][x] != 1 {
				continue
			}
			neighbors := 0
			for i := 0; i < 4; i++ {
				nx, ny := x+dx[i], y+dy[i]
				if nx >= 0 && nx < width && ny >= 0 && ny < height && ground[ny][nx] == 1 {
					neighbors++
				}
			}
			// Exactly 2 opposite neighbors = corridor
			if neighbors <= 2 {
				// Check if it's a straight corridor (opposite neighbors)
				if neighbors == 2 {
					h := ground[y][x-1] == 1 && ground[y][x+1] == 1
					v := ground[y-1][x] == 1 && ground[y+1][x] == 1
					if h || v {
						count++
					}
				} else if neighbors == 1 {
					// Dead end
					count++
				}
			}
		}
	}
	return count
}

// computePathTortuosity calculates main path actual distance / straight line distance
func computePathTortuosity(mainPath *MainPathData, width, height int) float64 {
	if mainPath == nil {
		return 1.0
	}

	// Count path cells
	pathCells := 0
	var minX, minY, maxX, maxY int
	minX, minY = width, height
	first := true

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if mainPath.OnMainPath[y][x] {
				pathCells++
				if first || x < minX {
					minX = x
				}
				if first || y < minY {
					minY = y
				}
				if first || x > maxX {
					maxX = x
				}
				if first || y > maxY {
					maxY = y
				}
				first = false
			}
		}
	}

	if pathCells == 0 {
		return 1.0
	}

	// Straight line distance between path extremes
	straightDist := math.Sqrt(float64((maxX-minX)*(maxX-minX) + (maxY-minY)*(maxY-minY)))
	if straightDist < 1 {
		straightDist = 1
	}

	return float64(pathCells) / straightDist
}

// countStaticsNearPath counts static cells within 2 Manhattan distance of main path
func countStaticsNearPath(staticLayer [][]int, mainPath *MainPathData, width, height int) int {
	if mainPath == nil {
		return 0
	}

	count := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if staticLayer[y][x] == 1 && mainPath.DirectDistance[y][x] <= 2 {
				count++
			}
		}
	}
	return count
}

// countIslands counts disconnected ground regions using flood fill
func countIslands(ground [][]int, width, height int) int {
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	islands := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if ground[y][x] == 1 && !visited[y][x] {
				islands++
				// Flood fill
				queue := []Point{{x, y}}
				visited[y][x] = true
				for len(queue) > 0 {
					cur := queue[0]
					queue = queue[1:]
					for _, d := range []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
						nx, ny := cur.X+d.X, cur.Y+d.Y
						if nx >= 0 && nx < width && ny >= 0 && ny < height && ground[ny][nx] == 1 && !visited[ny][nx] {
							visited[ny][nx] = true
							queue = append(queue, Point{nx, ny})
						}
					}
				}
			}
		}
	}
	return islands
}

// computeEnemyConcentration measures how clustered enemies are (0 = spread out, 1 = very clustered)
func computeEnemyConcentration(chaserLayer, zonerLayer, dpsLayer, mobAirLayer [][]int, width, height int) float64 {
	// Collect all enemy positions
	var positions []Point
	layers := [][][]int{chaserLayer, zonerLayer, dpsLayer, mobAirLayer}
	for _, layer := range layers {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if layer[y][x] == 1 {
					positions = append(positions, Point{x, y})
				}
			}
		}
	}

	if len(positions) <= 1 {
		return 0
	}

	// Compute centroid
	sumX, sumY := 0.0, 0.0
	for _, p := range positions {
		sumX += float64(p.X)
		sumY += float64(p.Y)
	}
	n := float64(len(positions))
	cx, cy := sumX/n, sumY/n

	// Compute variance (average squared distance from centroid)
	variance := 0.0
	for _, p := range positions {
		dx := float64(p.X) - cx
		dy := float64(p.Y) - cy
		variance += dx*dx + dy*dy
	}
	variance /= n

	// Normalize: max possible variance ≈ (width/2)^2 + (height/2)^2
	maxVariance := float64(width*width+height*height) / 4.0
	if maxVariance < 1 {
		maxVariance = 1
	}

	// Low variance = high concentration
	spreadRatio := variance / maxVariance
	concentration := 1.0 - clamp01(spreadRatio)

	return concentration
}

// clamp01 clamps a value to [0, 1]
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
