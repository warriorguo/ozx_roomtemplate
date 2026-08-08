package generate

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MobAir centre-seeded even distribution (ORT-104)
// ============================================================================

// generateMobAirRoom builds a room of the given stage and returns a function
// that lays a fresh mobAir layer over that same fixed room.
func generateMobAirRoom(t *testing.T, stage string, w, h int) func(target int) [][]int {
	t.Helper()
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}
	resp, err := GenerateFullRoom(FullRoomGenerateRequest{
		Width: w, Height: h, Doors: doors, StaticCount: 8,
		StageType: stage, RoomCategory: "normal",
	})
	require.NoError(t, err)

	p := resp.Payload
	doorPositions := getDoorCenterPositions(w, h, doors)
	return func(target int) [][]int {
		layer := createEmptyLayer(w, h)
		GenerateMobAirLayerNew(layer, p.Ground, p.SoftEdge, p.Bridge, p.Static,
			p.Zoner, p.Chaser, p.DPS, doorPositions, w, h, target)
		return layer
	}
}

func mobAirPoints(layer [][]int, width, height int) []Point {
	var pts []Point
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if layer[y][x] == 1 {
				pts = append(pts, Point{X: x, Y: y})
			}
		}
	}
	return pts
}

// TestGenerateMobAirLayer_Deterministic is the ORT-104 regression guard against
// the strategy coin flip: the same room must always produce the same layer.
func TestGenerateMobAirLayer_Deterministic(t *testing.T) {
	const w, h = 20, 12
	place := generateMobAirRoom(t, "peak", w, h)

	first := place(18)
	for run := 1; run < 10; run++ {
		again := place(18)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				require.Equalf(t, first[y][x], again[y][x],
					"run %d differs from run 0 at (%d,%d) — placement is not deterministic", run, x, y)
			}
		}
	}
}

// TestGenerateMobAirLayer_SeedsRoomCentre pins the acceptance criterion that the
// valid position closest to the room centre is always occupied.
func TestGenerateMobAirLayer_SeedsRoomCentre(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	for _, tc := range []struct {
		stage  string
		w, h   int
		target int
	}{
		{"teaching", 20, 12, 6},
		{"peak", 20, 12, 18},
		{"pressure", 18, 10, 8},
	} {
		t.Run(tc.stage, func(t *testing.T) {
			for trial := 0; trial < 25; trial++ {
				resp, err := GenerateFullRoom(FullRoomGenerateRequest{
					Width: tc.w, Height: tc.h, Doors: doors, StaticCount: 8,
					StageType: tc.stage, RoomCategory: "normal",
				})
				require.NoErrorf(t, err, "trial %d", trial)
				p := resp.Payload
				doorPositions := getDoorCenterPositions(tc.w, tc.h, doors)

				layer := createEmptyLayer(tc.w, tc.h)
				GenerateMobAirLayerNew(layer, p.Ground, p.SoftEdge, p.Bridge, p.Static,
					p.Zoner, p.Chaser, p.DPS, doorPositions, tc.w, tc.h, tc.target)

				// Smallest distance to the centre any valid cell could have, in
				// the empty room before anything was placed.
				center := Point{X: tc.w / 2, Y: tc.h / 2}
				empty := createEmptyLayer(tc.w, tc.h)
				bestPossible := -1
				for y := 0; y < tc.h; y++ {
					for x := 0; x < tc.w; x++ {
						pos := Point{X: x, Y: y}
						if !isValidMobAirPositionNew(pos, p.Ground, p.SoftEdge, p.Bridge, p.Static,
							p.Zoner, p.Chaser, p.DPS, empty, doorPositions, tc.w, tc.h) {
							continue
						}
						if d := manhattanDistance(pos, center); bestPossible < 0 || d < bestPossible {
							bestPossible = d
						}
					}
				}
				if bestPossible < 0 {
					continue // no valid cell at all; nothing to assert
				}

				placedBest := -1
				for _, pt := range mobAirPoints(layer, tc.w, tc.h) {
					if d := manhattanDistance(pt, center); placedBest < 0 || d < placedBest {
						placedBest = d
					}
				}
				assert.Equalf(t, bestPossible, placedBest,
					"trial %d: closest mobAir sits %d from the room centre but a valid cell exists at %d",
					trial, placedBest, bestPossible)
			}
		})
	}
}

// TestGenerateMobAirLayer_EvenlyDispersed measures that placements spread evenly
// rather than clustering, using the Clark-Evans nearest-neighbour index: the
// observed mean nearest-neighbour distance over the mean expected from a
// uniform-random set of the same size in the same area. R < 1 is clustered,
// R = 1 is indistinguishable from random, R > 1 is dispersed.
//
// A raw variance-of-spacing check would not do: a tight cluster has uniformly
// small spacings and so scores well on variance alone. The density-driven
// placement this replaced measured R = 0.95 at teaching — statistically
// indistinguishable from scattering the mobs at random.
func TestGenerateMobAirLayer_EvenlyDispersed(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	// teaching is the discriminating case: few mobs in a large room, so the
	// spacing constraint alone does not dictate the layout.
	const w, h, target, trials = 20, 12, 6, 40
	sumR := 0.0

	for trial := 0; trial < trials; trial++ {
		resp, err := GenerateFullRoom(FullRoomGenerateRequest{
			Width: w, Height: h, Doors: doors, StaticCount: 8,
			StageType: "teaching", RoomCategory: "normal",
		})
		require.NoErrorf(t, err, "trial %d", trial)
		p := resp.Payload
		doorPositions := getDoorCenterPositions(w, h, doors)

		layer := createEmptyLayer(w, h)
		GenerateMobAirLayerNew(layer, p.Ground, p.SoftEdge, p.Bridge, p.Static,
			p.Zoner, p.Chaser, p.DPS, doorPositions, w, h, target)

		pts := mobAirPoints(layer, w, h)
		if len(pts) < 2 {
			continue
		}

		// Area available to the layer, counted as actually-placeable cells.
		empty := createEmptyLayer(w, h)
		area := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if isValidMobAirPositionNew(Point{X: x, Y: y}, p.Ground, p.SoftEdge, p.Bridge, p.Static,
					p.Zoner, p.Chaser, p.DPS, empty, doorPositions, w, h) {
					area++
				}
			}
		}
		if area == 0 {
			continue
		}

		meanNN := 0.0
		for i, a := range pts {
			best := math.Inf(1)
			for j, b := range pts {
				if i == j {
					continue
				}
				if d := math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y)); d < best {
					best = d
				}
			}
			meanNN += best
		}
		meanNN /= float64(len(pts))

		expected := 0.5 * math.Sqrt(float64(area)/float64(len(pts)))
		sumR += meanNN / expected
	}

	avgR := sumR / trials
	t.Logf("Clark-Evans R = %.2f over %d trials (1.0 = random, >1 = dispersed)", avgR, trials)
	assert.Greaterf(t, avgR, 1.25,
		"mobAir placements are not meaningfully dispersed: Clark-Evans R = %.2f", avgR)
}

// TestGenerateMobAirLayer_ConstraintsHold checks the placement rules ORT-104
// left untouched.
func TestGenerateMobAirLayer_ConstraintsHold(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	for _, stage := range []string{"teaching", "building", "pressure", "peak", "release"} {
		t.Run(stage, func(t *testing.T) {
			cfg := GetStageConfig(stage)
			require.NotNil(t, cfg)
			w, h := cfg.MinWidth, cfg.MinHeight
			if w == 0 || h == 0 {
				w, h = 20, 12
			}

			for trial := 0; trial < 20; trial++ {
				resp, err := GenerateFullRoom(FullRoomGenerateRequest{
					Width: w, Height: h, Doors: doors, StaticCount: 8,
					StageType: stage, RoomCategory: "normal",
				})
				require.NoErrorf(t, err, "trial %d", trial)
				p := resp.Payload
				doorPositions := getDoorCenterPositions(w, h, doors)

				for _, pt := range mobAirPoints(p.MobAir, w, h) {
					assert.GreaterOrEqualf(t, pt.X, mobAirMinEdgeDistance, "trial %d: mobAir %v too close to left edge", trial, pt)
					assert.GreaterOrEqualf(t, pt.Y, mobAirMinEdgeDistance, "trial %d: mobAir %v too close to top edge", trial, pt)
					assert.Lessf(t, pt.X, w-mobAirMinEdgeDistance, "trial %d: mobAir %v too close to right edge", trial, pt)
					assert.Lessf(t, pt.Y, h-mobAirMinEdgeDistance, "trial %d: mobAir %v too close to bottom edge", trial, pt)

					for _, door := range doorPositions {
						assert.GreaterOrEqualf(t, manhattanDistance(pt, door), mobAirMinDoorDistance,
							"trial %d: mobAir %v within %d of door %v", trial, pt, mobAirMinDoorDistance, door)
					}

					assert.Equalf(t, 0, p.Static[pt.Y][pt.X], "trial %d: mobAir %v on static", trial, pt)
					assert.Equalf(t, 0, p.Zoner[pt.Y][pt.X], "trial %d: mobAir %v on zoner", trial, pt)
					assert.Equalf(t, 0, p.Chaser[pt.Y][pt.X], "trial %d: mobAir %v on chaser", trial, pt)
					assert.Equalf(t, 0, p.DPS[pt.Y][pt.X], "trial %d: mobAir %v on dps", trial, pt)
				}

				assert.Falsef(t, hasSameLayerAdjacency(p.MobAir, w, h),
					"trial %d: %s produced 8-directionally adjacent mobAir", trial, stage)
			}
		})
	}
}
