package generate

import (
	"testing"
)

func TestGenerateRailLayer_EmptyGround(t *testing.T) {
	width, height := 20, 20
	ground := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	railLayer := createEmptyLayer(width, height)

	debug := GenerateRailLayer(railLayer, ground, bridge, width, height)

	if !debug.Skipped {
		t.Errorf("Expected rail generation to be skipped with empty ground")
	}
}

func TestGenerateRailLayer_SmallPlatform(t *testing.T) {
	width, height := 10, 10
	ground := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	railLayer := createEmptyLayer(width, height)

	// Create a small 5x5 platform (too small for rail)
	for y := 2; y < 7; y++ {
		for x := 2; x < 7; x++ {
			ground[y][x] = 1
		}
	}

	debug := GenerateRailLayer(railLayer, ground, bridge, width, height)

	// Should either skip or find no platforms large enough
	if debug.PlatformsFound > 0 && len(debug.RailLoops) > 0 {
		// Verify rail forms a valid loop
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if railLayer[y][x] == 1 {
					neighbors := countRailNeighborsLocal(railLayer, x, y, width, height)
					if neighbors != 2 {
						t.Errorf("Rail cell at (%d,%d) has %d neighbors, expected 2", x, y, neighbors)
					}
				}
			}
		}
	}
}

func TestGenerateRailLayer_LargePlatform(t *testing.T) {
	width, height := 30, 30
	ground := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	railLayer := createEmptyLayer(width, height)

	// Create a large 20x20 platform (should support rail)
	for y := 5; y < 25; y++ {
		for x := 5; x < 25; x++ {
			ground[y][x] = 1
		}
	}

	debug := GenerateRailLayer(railLayer, ground, bridge, width, height)

	t.Logf("Platforms found: %d", debug.PlatformsFound)
	t.Logf("Rail loops placed: %d", len(debug.RailLoops))

	if debug.PlatformsFound == 0 {
		t.Errorf("Expected to find at least one platform")
	}

	// Count rail cells
	railCount := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 1 {
				railCount++
			}
		}
	}

	t.Logf("Total rail cells: %d", railCount)

	// If rail was placed, verify it forms a valid loop
	if railCount > 0 {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if railLayer[y][x] == 1 {
					// Verify on ground or bridge
					if ground[y][x] != 1 && bridge[y][x] != 1 {
						t.Errorf("Rail cell at (%d,%d) is not on ground or bridge", x, y)
					}

					// Verify each rail cell has exactly 2 neighbors (closed loop)
					neighbors := countRailNeighborsLocal(railLayer, x, y, width, height)
					if neighbors != 2 {
						t.Errorf("Rail cell at (%d,%d) has %d neighbors, expected 2", x, y, neighbors)
					}
				}
			}
		}
	}
}

func TestGenerateRailLayer_RailOnBridge(t *testing.T) {
	width, height := 30, 30
	ground := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	railLayer := createEmptyLayer(width, height)

	// Create a platform with a bridge section
	for y := 5; y < 25; y++ {
		for x := 5; x < 25; x++ {
			if x >= 12 && x <= 17 && y >= 12 && y <= 17 {
				// Leave a gap that could be bridged
				bridge[y][x] = 1
			} else {
				ground[y][x] = 1
			}
		}
	}

	debug := GenerateRailLayer(railLayer, ground, bridge, width, height)

	t.Logf("Platforms found: %d", debug.PlatformsFound)

	// Verify any rail placed is on ground or bridge
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 1 {
				if ground[y][x] != 1 && bridge[y][x] != 1 {
					t.Errorf("Rail cell at (%d,%d) is not on ground or bridge", x, y)
				}
			}
		}
	}
}

func TestGenerateRailLayer_EdgeDistance(t *testing.T) {
	width, height := 30, 30
	ground := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)
	railLayer := createEmptyLayer(width, height)

	// Create a platform touching the edges
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ground[y][x] = 1
		}
	}

	debug := GenerateRailLayer(railLayer, ground, bridge, width, height)

	// Verify rail respects edge distance
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if railLayer[y][x] == 1 {
				if x < minRailEdgeDistance || x >= width-minRailEdgeDistance ||
					y < minRailEdgeDistance || y >= height-minRailEdgeDistance {
					t.Errorf("Rail cell at (%d,%d) is too close to edge (min distance: %d)", x, y, minRailEdgeDistance)
				}
			}
		}
	}

	t.Logf("Rail loops: %d", len(debug.RailLoops))
}

func TestFindRailPlatforms(t *testing.T) {
	width, height := 30, 30
	ground := createEmptyLayer(width, height)
	bridge := createEmptyLayer(width, height)

	// Create a platform
	for y := 5; y < 20; y++ {
		for x := 5; x < 20; x++ {
			ground[y][x] = 1
		}
	}

	platforms := findRailPlatforms(ground, bridge, width, height)

	if len(platforms) == 0 {
		t.Errorf("Expected to find at least one platform")
	}

	for _, p := range platforms {
		t.Logf("Found platform at (%d,%d) size %dx%d", p.X, p.Y, p.Width, p.Height)

		// Verify platform respects edge distance
		if p.X < minRailEdgeDistance || p.Y < minRailEdgeDistance {
			t.Errorf("Platform at (%d,%d) too close to top-left edge", p.X, p.Y)
		}
		if p.X+p.Width > width-minRailEdgeDistance || p.Y+p.Height > height-minRailEdgeDistance {
			t.Errorf("Platform at (%d,%d) size %dx%d extends too close to bottom-right edge", p.X, p.Y, p.Width, p.Height)
		}
	}
}

func TestGetRailIndentCells(t *testing.T) {
	width, height := 20, 20
	railLayer := createEmptyLayer(width, height)

	// Create a simple rail rectangle
	// Top edge
	for x := 5; x <= 15; x++ {
		railLayer[5][x] = 1
	}
	// Bottom edge
	for x := 5; x <= 15; x++ {
		railLayer[15][x] = 1
	}
	// Left edge
	for y := 5; y <= 15; y++ {
		railLayer[y][5] = 1
	}
	// Right edge
	for y := 5; y <= 15; y++ {
		railLayer[y][15] = 1
	}

	indentCells := GetRailIndentCells(railLayer, width, height)

	t.Logf("Found %d indent cells", len(indentCells))

	// The interior of the rectangle should be identified as indent cells
	for _, cell := range indentCells {
		if cell.X <= 5 || cell.X >= 15 || cell.Y <= 5 || cell.Y >= 15 {
			t.Errorf("Indent cell (%d,%d) should be inside the rail rectangle", cell.X, cell.Y)
		}
	}
}
