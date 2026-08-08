package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectPlacementShortfalls covers the reporting rules directly: only
// layers that ran and fell short are listed, in pipeline order.
func TestCollectPlacementShortfalls(t *testing.T) {
	t.Run("everything met returns nil", func(t *testing.T) {
		got := collectPlacementShortfalls(
			&StaticDebugInfo{TargetCount: 3, PlacedCount: 3},
			&EnemyLayerDebugInfo{TargetCount: 2, PlacedCount: 2},
			&EnemyLayerDebugInfo{TargetCount: 8, PlacedCount: 8},
			&EnemyLayerDebugInfo{TargetCount: 8, PlacedCount: 8},
			&MobAirDebugInfo{TargetCount: 12, PlacedCount: 12},
		)
		assert.Nil(t, got, "no shortfall should leave the field absent")
	})

	t.Run("skipped layers are not shortfalls", func(t *testing.T) {
		got := collectPlacementShortfalls(
			&StaticDebugInfo{Skipped: true, SkipReason: "staticCount is 0"},
			&EnemyLayerDebugInfo{Skipped: true},
			&EnemyLayerDebugInfo{Skipped: true},
			&EnemyLayerDebugInfo{Skipped: true},
			&MobAirDebugInfo{Skipped: true},
		)
		assert.Nil(t, got, "a count of 0 was honoured exactly")
	})

	t.Run("over-placement is not a shortfall", func(t *testing.T) {
		// Grouped placement can round a per-group allocation up.
		got := collectPlacementShortfalls(nil,
			nil, nil, nil,
			&MobAirDebugInfo{TargetCount: 12, PlacedCount: 14},
		)
		assert.Nil(t, got)
	})

	t.Run("reports each short layer in pipeline order", func(t *testing.T) {
		got := collectPlacementShortfalls(
			&StaticDebugInfo{TargetCount: 3, PlacedCount: 3},
			&EnemyLayerDebugInfo{TargetCount: 3, PlacedCount: 1}, // zoner
			&EnemyLayerDebugInfo{TargetCount: 8, PlacedCount: 8}, // chaser, fine
			&EnemyLayerDebugInfo{TargetCount: 8, PlacedCount: 5}, // dps
			&MobAirDebugInfo{TargetCount: 12, PlacedCount: 11},
		)
		require.Len(t, got, 3)
		assert.Equal(t, []string{"zoner", "dps", "mobAir"},
			[]string{got[0].Layer, got[1].Layer, got[2].Layer})

		assert.Equal(t, 3, got[0].Requested)
		assert.Equal(t, 1, got[0].Placed)
		assert.Contains(t, got[0].Message, "placed 1 of 3 spawns")
		assert.Contains(t, got[2].Message, "placed 11 of 12 spawns")
	})

	t.Run("static is reported in blocks", func(t *testing.T) {
		got := collectPlacementShortfalls(
			&StaticDebugInfo{TargetCount: 9, PlacedCount: 2},
			nil, nil, nil, nil)
		require.Len(t, got, 1)
		assert.Contains(t, got[0].Message, "placed 2 of 9 blocks")
	})

	t.Run("nil debug info is skipped", func(t *testing.T) {
		assert.Nil(t, collectPlacementShortfalls(nil, nil, nil, nil, nil))
	})
}

// TestGenerateReportsShortfallInSmallRooms is the end-to-end half: a room too
// small for its stage must come back with warnings rather than silently light.
// This is the failure mode ORT-102's minimum room size used to hide by refusing
// the room outright, and ORT-109 removed that refusal.
func TestGenerateReportsShortfallInSmallRooms(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	sawShortfall := false
	for trial := 0; trial < 60; trial++ {
		resp, err := GenerateFullRoom(FullRoomGenerateRequest{
			Width: 16, Height: 8, Doors: doors,
			StageType: "peak", RoomCategory: "normal",
		})
		require.NoErrorf(t, err, "trial %d: 16x8 peak must still generate (ORT-109)", trial)

		for _, w := range resp.Warnings {
			sawShortfall = true
			assert.Lessf(t, w.Placed, w.Requested,
				"trial %d: %s reported as short but placed >= requested", trial, w.Layer)
			assert.NotEmptyf(t, w.Message, "trial %d: %s warning has no message", trial, w.Layer)
		}
	}

	assert.True(t, sawShortfall,
		"16x8 peak asks for more than the room can hold, so at least one of 60 rooms "+
			"should report a shortfall; if this stops happening the counts or the room "+
			"got big enough that the warning path is no longer exercised here")
}

// TestGenerateNoShortfallAtReferenceSize is the other side: a room with ample
// space must not report anything, so the field stays a signal rather than noise.
func TestGenerateNoShortfallAtReferenceSize(t *testing.T) {
	doors := []DoorPosition{DoorTop, DoorBottom, DoorLeft, DoorRight}

	short := 0
	const trials = 40
	for trial := 0; trial < trials; trial++ {
		resp, err := GenerateFullRoom(FullRoomGenerateRequest{
			Width: 24, Height: 16, Doors: doors,
			StageType: "teaching", RoomCategory: "normal",
		})
		require.NoErrorf(t, err, "trial %d", trial)
		if len(resp.Warnings) > 0 {
			short++
		}
	}

	// Allow a small tail: placement is randomised and a pathological ground
	// shape can still run a layer out of sites.
	assert.LessOrEqualf(t, short, trials/5,
		"teaching at 24x16 reported a shortfall in %d/%d rooms — the warning should be "+
			"rare at a comfortable size", short, trials)
}
