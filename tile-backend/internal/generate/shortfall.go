package generate

import "fmt"

// layerShortfall is the shape every layer's debug info shares for this purpose:
// what was asked for, what was placed, and whether the layer ran at all.
type layerShortfall struct {
	layer     string
	skipped   bool
	requested int
	placed    int
}

// collectPlacementShortfalls reports every layer that placed fewer spawns than
// requested, in pipeline order. Returns nil when everything met its target, so
// the response field stays absent in the common case.
//
// A skipped layer is not a shortfall — a count of 0 was honoured exactly. Nor
// is placing *more* than requested, which grouped placement can do when a
// stage's per-group allocation rounds up.
func collectPlacementShortfalls(
	static *StaticDebugInfo,
	zoner, chaser, dps *EnemyLayerDebugInfo,
	mobAir *MobAirDebugInfo,
) []PlacementShortfall {
	layers := make([]layerShortfall, 0, 5)

	if static != nil {
		layers = append(layers, layerShortfall{"static", static.Skipped, static.TargetCount, static.PlacedCount})
	}
	for _, l := range []struct {
		name  string
		debug *EnemyLayerDebugInfo
	}{{"zoner", zoner}, {"chaser", chaser}, {"dps", dps}} {
		if l.debug != nil {
			layers = append(layers, layerShortfall{l.name, l.debug.Skipped, l.debug.TargetCount, l.debug.PlacedCount})
		}
	}
	if mobAir != nil {
		layers = append(layers, layerShortfall{"mobAir", mobAir.Skipped, mobAir.TargetCount, mobAir.PlacedCount})
	}

	var out []PlacementShortfall
	for _, l := range layers {
		if l.skipped || l.requested <= 0 || l.placed >= l.requested {
			continue
		}
		unit := "spawns"
		if l.layer == "static" {
			unit = "blocks"
		}
		out = append(out, PlacementShortfall{
			Layer:     l.layer,
			Requested: l.requested,
			Placed:    l.placed,
			Message: fmt.Sprintf("%s: placed %d of %d %s — the room ran out of legal sites under the spacing constraint",
				l.layer, l.placed, l.requested, unit),
		})
	}
	return out
}
