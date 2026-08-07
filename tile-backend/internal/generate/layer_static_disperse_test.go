package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ORT-99: static placement is stage-driven. teaching/building/release keep the
// alternating centre-outward scatter; pressure/peak seed from the edge and then
// disperse by farthest-point selection.

// TestStaticDisperseHintPerStage pins which stages opt into dispersion.
// teaching/building/release (and start/boss) must stay on the default path so
// their output is unchanged.
func TestStaticDisperseHintPerStage(t *testing.T) {
	tests := []struct {
		stage        string
		width        int
		height       int
		wantDisperse bool
	}{
		{"start", 16, 8, false},
		{"teaching", 16, 8, false},
		{"building", 16, 8, false},
		{"release", 16, 8, false},
		{"pressure", 18, 10, true},
		{"peak", 20, 12, true},
	}

	for _, tt := range tests {
		t.Run(tt.stage, func(t *testing.T) {
			ground := makeFullGround(tt.width, tt.height)
			result, err := ValidateAndApplyStage(tt.stage, "full", nil, ground, tt.width, tt.height)
			require.NoError(t, err)
			require.NotNil(t, result.PlacementHints)
			assert.Equal(t, tt.wantDisperse, result.PlacementHints.StaticDisperse)
		})
	}
}

// staticBlocks returns the top-left corner of every placed 2x2 block.
func staticBlocks(layer [][]int, width, height int) []Point {
	var blocks []Point
	seen := make([][]bool, height)
	for y := range seen {
		seen[y] = make([]bool, width)
	}
	for y := 0; y < height-1; y++ {
		for x := 0; x < width-1; x++ {
			if seen[y][x] || layer[y][x] != 1 {
				continue
			}
			if layer[y][x+1] == 1 && layer[y+1][x] == 1 && layer[y+1][x+1] == 1 {
				blocks = append(blocks, Point{X: x, Y: y})
				seen[y][x], seen[y][x+1] = true, true
				seen[y+1][x], seen[y+1][x+1] = true, true
			}
		}
	}
	return blocks
}

// meanEdgeDistance is the average distance of each block from the nearest wall.
// Lower means cover sits closer to the perimeter.
func meanEdgeDistance(blocks []Point, width, height int) float64 {
	if len(blocks) == 0 {
		return 0
	}
	total := 0
	for _, b := range blocks {
		total += distanceFromEdge(b, width, height)
	}
	return float64(total) / float64(len(blocks))
}

// meanNearestNeighbour is the average distance from each block to its closest
// sibling. Higher means the blocks are spread further apart.
func meanNearestNeighbour(blocks []Point) float64 {
	if len(blocks) < 2 {
		return 0
	}
	total := 0
	for i, b := range blocks {
		others := append(append([]Point{}, blocks[:i]...), blocks[i+1:]...)
		total += distanceToNearest(b, others)
	}
	return float64(total) / float64(len(blocks))
}

// runStaticStrategy places targetCount blocks over a plain room and returns the
// resulting block corners, holding everything but the strategy constant.
func runStaticStrategy(t *testing.T, disperse bool, width, height, targetCount int) []Point {
	t.Helper()

	ground := makeFullGround(width, height)
	empty := makeEmptyLayer(width, height)
	staticLayer := makeEmptyLayer(width, height)
	doorPositions := getDoorCenterPositions(width, height, []DoorPosition{DoorTop, DoorBottom})

	var hints *StagePlacementHints
	if disperse {
		hints = &StagePlacementHints{StaticDisperse: true}
	}

	generateStaticLayerWithDebugAndRail(
		staticLayer, ground, empty, empty, empty,
		doorPositions, width, height, targetCount, hints,
	)
	return staticBlocks(staticLayer, width, height)
}

func makeEmptyLayer(width, height int) [][]int {
	l := make([][]int, height)
	for y := range l {
		l[y] = make([]int, width)
	}
	return l
}

// TestStaticDisperse_FavoursPerimeterAndSpacing is the ORT-99 acceptance test.
// Both arms use the same room, doors and block count, so the only variable is
// the strategy. Averaged over many trials, dispersion must put cover closer to
// the walls and further apart than the centre-first default.
func TestStaticDisperse_FavoursPerimeterAndSpacing(t *testing.T) {
	const (
		width  = 24
		height = 16
		blocks = 6
		trials = 60
	)

	var centreEdge, disperseEdge float64
	var centreSpacing, disperseSpacing float64
	centreTrials, disperseTrials := 0, 0

	for i := 0; i < trials; i++ {
		c := runStaticStrategy(t, false, width, height, blocks)
		if len(c) >= 2 {
			centreEdge += meanEdgeDistance(c, width, height)
			centreSpacing += meanNearestNeighbour(c)
			centreTrials++
		}

		d := runStaticStrategy(t, true, width, height, blocks)
		if len(d) >= 2 {
			disperseEdge += meanEdgeDistance(d, width, height)
			disperseSpacing += meanNearestNeighbour(d)
			disperseTrials++
		}
	}

	require.Greater(t, centreTrials, trials/2, "too few usable centre-first trials")
	require.Greater(t, disperseTrials, trials/2, "too few usable disperse trials")

	centreEdge /= float64(centreTrials)
	disperseEdge /= float64(disperseTrials)
	centreSpacing /= float64(centreTrials)
	disperseSpacing /= float64(disperseTrials)

	t.Logf("mean edge distance:     centre-first=%.2f disperse=%.2f", centreEdge, disperseEdge)
	t.Logf("mean nearest-neighbour: centre-first=%.2f disperse=%.2f", centreSpacing, disperseSpacing)

	assert.Less(t, disperseEdge, centreEdge,
		"dispersed statics should sit closer to the perimeter")
	assert.Greater(t, disperseSpacing, centreSpacing,
		"dispersed statics should be spaced further apart")
}

// TestStaticDisperse_PreservesInvariants guards the ORT-99 acceptance clause
// that the new strategy breaks none of the existing static rules.
func TestStaticDisperse_PreservesInvariants(t *testing.T) {
	for _, stage := range []string{"pressure", "peak"} {
		t.Run(stage, func(t *testing.T) {
			for trial := 0; trial < 25; trial++ {
				resp, err := GenerateFullRoom(FullRoomGenerateRequest{
					Width:     24,
					Height:    16,
					Doors:     []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight},
					StageType: stage,
				})
				require.NoError(t, err)

				static := resp.Payload.Static
				ground := resp.Payload.Ground

				for y := 0; y < 16; y++ {
					for x := 0; x < 24; x++ {
						if static[y][x] != 1 {
							continue
						}
						require.Equal(t, 1, ground[y][x],
							"trial %d: static at (%d,%d) not on ground", trial, x, y)
					}
				}

				// 8-directional non-contact between distinct blocks.
				blocks := staticBlocks(static, 24, 16)
				for i := 0; i < len(blocks); i++ {
					for j := i + 1; j < len(blocks); j++ {
						a, b := blocks[i], blocks[j]
						if abs(a.X-b.X) <= staticSize && abs(a.Y-b.Y) <= staticSize {
							t.Fatalf("trial %d: blocks (%d,%d) and (%d,%d) touch",
								trial, a.X, a.Y, b.X, b.Y)
						}
					}
				}
			}
		})
	}
}
