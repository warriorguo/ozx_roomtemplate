package generate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Zoner footprint tests (ORT-103)
//
// A zoner occupies a 2x2 block wherever a valid site exists and falls back to a
// single cell only when none does. The game collapses a connected block into
// one spawn, so a block is one enemy and the ORT-93 non-contact rule applies
// between blocks rather than between cells.
// ============================================================================

// zonerFixture builds an open room and everything GenerateZonerLayer needs to
// run against it.
type zonerFixture struct {
	width, height int
	ground        [][]int
	softEdge      [][]int
	bridge        [][]int
	staticLayer   [][]int
	doorPositions map[DoorPosition]Point
	mainPath      *MainPathData
}

func newZonerFixture(t *testing.T, ground [][]int, width, height int, doors []DoorPosition) *zonerFixture {
	t.Helper()
	doorPositions := getDoorCenterPositions(width, height, doors)
	bridge := createEmptyLayer(width, height)
	mainPath, _ := ComputeMainPath(ground, bridge, doorPositions, width, height)
	require.NotNil(t, mainPath, "main path should be computable for the fixture room")
	return &zonerFixture{
		width: width, height: height,
		ground:        ground,
		softEdge:      createEmptyLayer(width, height),
		bridge:        bridge,
		staticLayer:   createEmptyLayer(width, height),
		doorPositions: doorPositions,
		mainPath:      mainPath,
	}
}

func (f *zonerFixture) generate(targetCount int) ([][]int, *EnemyLayerDebugInfo) {
	zoner := createEmptyLayer(f.width, f.height)
	debug := GenerateZonerLayer(zoner, f.ground, f.softEdge, f.bridge, nil, f.staticLayer,
		f.doorPositions, f.mainPath, f.width, f.height, targetCount)
	return zoner, debug
}

// TestGenerateZonerLayer_PrefersBlocks pins the headline behaviour: in an open
// room every zoner is a 2x2 block, and the count is in spawns — six zoners are
// six blocks, i.e. 24 cells.
func TestGenerateZonerLayer_PrefersBlocks(t *testing.T) {
	const width, height, target = 24, 16, 6
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	for trial := 0; trial < 20; trial++ {
		f := newZonerFixture(t, makeFullGround(width, height), width, height, doors)
		zoner, debug := f.generate(target)

		require.Equalf(t, target, debug.PlacedCount,
			"trial %d: open %dx%d room should fit %d zoners, misses=%v",
			trial, width, height, target, debug.Misses)

		for _, p := range debug.Placements {
			assert.Equalf(t, "2x2", p.Size,
				"trial %d: zoner at %s fell back to %s in an open room", trial, p.Position, p.Size)
		}
		assert.Equalf(t, target, countZonerUnits(zoner), "trial %d: spawn count", trial)
		assert.Equalf(t, target*zonerSize*zonerSize, countCells(zoner),
			"trial %d: every spawn should occupy a full %dx%d footprint", trial, zonerSize, zonerSize)
	}
}

// TestGenerateZonerLayer_FallsBackTo1x1 covers the documented fallback: a
// one-cell-wide corridor has no 2x2 site anywhere, so every zoner must be a
// single cell rather than the layer coming back empty.
func TestGenerateZonerLayer_FallsBackTo1x1(t *testing.T) {
	const width, height = 24, 9
	const corridorY = 4

	// Ground is a single horizontal corridor, so no 2x2 block fits.
	ground := createEmptyLayer(width, height)
	for x := 0; x < width; x++ {
		ground[corridorY][x] = 1
	}

	f := newZonerFixture(t, ground, width, height, []DoorPosition{DoorLeft, DoorRight})
	zoner, debug := f.generate(3)

	require.Greaterf(t, debug.PlacedCount, 0,
		"corridor room placed no zoners at all, misses=%v", debug.Misses)
	for _, p := range debug.Placements {
		assert.Equalf(t, "1x1", p.Size, "zoner at %s claimed a 2x2 block in a 1-wide corridor", p.Position)
	}
	assert.Equal(t, debug.PlacedCount, countCells(zoner),
		"each fallback zoner should occupy exactly one cell")
	assert.False(t, hasInvalidZonerGroup(zoner, width, height),
		"fallback cells must still keep 8-directional distance from each other")
}

// TestZonerBlocksSatisfyPerCellConstraints checks that every cell of every
// placed block honours the constraints the single-cell path checks — the
// acceptance criterion that a block is not allowed to spill onto a static, a
// door approach, or off the main path's reach.
func TestZonerBlocksSatisfyPerCellConstraints(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	for _, stage := range []string{"teaching", "building", "pressure", "peak", "release"} {
		cfg := GetStageConfig(stage)
		require.NotNilf(t, cfg, "stage %q not found", stage)
		w, h := cfg.MinWidth, cfg.MinHeight
		if w == 0 || h == 0 {
			w, h = 20, 12
		}

		t.Run(fmt.Sprintf("%s_%dx%d", stage, w, h), func(t *testing.T) {
			for trial := 0; trial < 25; trial++ {
				resp, err := GenerateFullRoom(FullRoomGenerateRequest{
					Width: w, Height: h, Doors: doors, StaticCount: 8,
					StageType: stage, RoomCategory: "normal",
				})
				require.NoErrorf(t, err, "trial %d", trial)
				p := resp.Payload

				// Distinct zoners never touch. Because touching blocks merge into
				// one connected group, a group of any shape other than 2x2 or 1x1
				// is exactly the ORT-93 violation this guards.
				require.Falsef(t, hasInvalidZonerGroup(p.Zoner, w, h),
					"trial %d: %s produced a zoner group that is neither a 2x2 block nor a lone cell", trial, stage)

				forbidden := getDoorForbiddenCellsRadius(
					getDoorCenterPositions(w, h, doors), w, h, doorForbiddenRadius)

				for y := 0; y < h; y++ {
					for x := 0; x < w; x++ {
						if p.Zoner[y][x] != 1 {
							continue
						}
						assert.Equalf(t, 1, p.Ground[y][x], "trial %d: zoner (%d,%d) off ground", trial, x, y)
						assert.Equalf(t, 0, p.Static[y][x], "trial %d: zoner (%d,%d) on static", trial, x, y)
						assert.Equalf(t, 0, p.Chaser[y][x], "trial %d: zoner (%d,%d) on chaser", trial, x, y)
						if len(p.SoftEdge) > 0 {
							assert.Equalf(t, 0, p.SoftEdge[y][x], "trial %d: zoner (%d,%d) on softEdge", trial, x, y)
						}
						if len(p.Bridge) > 0 {
							assert.Equalf(t, 0, p.Bridge[y][x], "trial %d: zoner (%d,%d) on bridge", trial, x, y)
						}
						if len(p.Rail) > 0 {
							assert.Equalf(t, 0, p.Rail[y][x], "trial %d: zoner (%d,%d) on rail", trial, x, y)
						}
						assert.Falsef(t, forbidden[Point{X: x, Y: y}],
							"trial %d: zoner (%d,%d) inside a door forbidden zone", trial, x, y)

						// DirectDistance is Chebyshev distance to the nearest main
						// path cell, so measure it the same way here.
						assert.LessOrEqualf(t, chebyshevToMainPath(p.MainPath, x, y, w, h), zonerMaxPathDist,
							"trial %d: zoner (%d,%d) further than %d from the main path", trial, x, y, zonerMaxPathDist)
					}
				}
			}
		})
	}
}

// chebyshevToMainPath returns the Chebyshev distance from (x,y) to the nearest
// main path cell, or a large sentinel when the layer carries no path.
func chebyshevToMainPath(mainPath [][]int, x, y, width, height int) int {
	best := width + height
	for py := 0; py < height && py < len(mainPath); py++ {
		for px := 0; px < width && px < len(mainPath[py]); px++ {
			if mainPath[py][px] != 1 {
				continue
			}
			d := abs(px - x)
			if dy := abs(py - y); dy > d {
				d = dy
			}
			if d < best {
				best = d
			}
		}
	}
	return best
}
