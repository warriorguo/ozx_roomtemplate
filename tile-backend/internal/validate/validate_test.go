package validate

import (
	"testing"
	"tile-backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestValidateTemplate_BasicStructure_Fixed(t *testing.T) {
	tests := []struct {
		name     string
		payload  *model.TemplatePayload
		expected bool
		errors   int
	}{
		{
			name: "valid basic template",
			payload: &model.TemplatePayload{
				Ground: [][]int{
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
				},
				Static: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Chaser: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				Zoner: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				DPS: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobAir: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			expected: true,
			errors:   0,
		},
		{
			name: "invalid dimensions - too small",
			payload: &model.TemplatePayload{
				Ground: [][]int{{1}, {1}},
				Static: [][]int{{0}, {1}},
				Chaser: [][]int{{0}, {0}},
				Zoner:  [][]int{{0}, {0}},
				DPS:    [][]int{{0}, {0}},
				MobAir: [][]int{{0}, {1}},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   1,
					Height:  2,
				},
			},
			expected: false,
			errors:   1,
		},
		{
			name: "invalid dimensions - too large",
			payload: &model.TemplatePayload{
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   250,
					Height:  2,
				},
			},
			expected: false,
			errors:   1,
		},
		{
			name: "mismatched layer dimensions",
			payload: &model.TemplatePayload{
				Ground: [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
				Static: [][]int{{0, 1, 1, 0}}, // Wrong height
				Chaser: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				Zoner:  [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				DPS:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobAir: [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			expected: false,
			errors:   1,
		},
		{
			name: "invalid cell values",
			payload: &model.TemplatePayload{
				Ground: [][]int{{1, 2, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}}, // Invalid value '2'
				Static: [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Chaser: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				Zoner:  [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				DPS:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobAir: [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			expected: false,
			errors:   1,
		},
		{
			name: "empty name and version",
			payload: &model.TemplatePayload{
				Ground: [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
				Static: [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Chaser: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				Zoner:  [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				DPS:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobAir: [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Meta: model.TemplateMeta{
					Name:    "", // Empty name
					Version: 0,  // Zero version
					Width:   4,
					Height:  4,
				},
			},
			expected: true, // Basic structure validation doesn't check name/version
			errors:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateTemplate(tt.payload, false)
			assert.Equal(t, tt.expected, result.Valid)
			if tt.errors > 0 {
				assert.GreaterOrEqual(t, len(result.Errors), tt.errors)
			} else {
				assert.Equal(t, tt.errors, len(result.Errors))
			}
		})
	}
}

func TestValidateTemplate_LogicalRules_Fixed(t *testing.T) {
	tests := []struct {
		name     string
		payload  *model.TemplatePayload
		strict   bool
		expected bool
		errors   int
	}{
		{
			name: "static requires ground - valid",
			payload: &model.TemplatePayload{
				Ground: [][]int{
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
				},
				Static: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1}, // Static only where ground=1
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Chaser: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				Zoner: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				DPS: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobAir: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			strict:   true,
			expected: true,
			errors:   0,
		},
		{
			name: "static requires ground - invalid",
			payload: &model.TemplatePayload{
				Ground: [][]int{
					{0, 1, 1, 1}, // Ground=0 at (0,0)
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
				},
				Static: [][]int{
					{1, 1, 1, 0}, // Static=1 at (0,0) where ground=0
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Chaser: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				Zoner: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				DPS: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobAir: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			strict:   true,
			expected: false,
			errors:   1,
		},
		{
			name: "turret requires ground and no static - valid",
			payload: &model.TemplatePayload{
				Ground: [][]int{
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
				},
				Static: [][]int{
					{0, 0, 1, 0}, // No static at (0,0)
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Chaser: [][]int{
					{1, 0, 0, 0}, // Turret at (0,0): ground=1, static=0
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				Zoner: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				DPS: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobAir: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			strict:   true,
			expected: true,
			errors:   0,
		},
		{
			name: "turret requires ground - invalid",
			payload: &model.TemplatePayload{
				Ground: [][]int{
					{0, 1, 1, 1}, // Ground=0 at (0,0)
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
				},
				Static: [][]int{
					{0, 0, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Chaser: [][]int{
					{1, 0, 0, 0}, // Turret at (0,0): ground=0
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				Zoner: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				DPS: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobAir: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			strict:   true,
			expected: false,
			errors:   1,
		},
		{
			name: "non-strict mode ignores logical rules",
			payload: &model.TemplatePayload{
				Ground: [][]int{
					{0, 1, 1, 1}, // Ground=0 at (0,0)
					{1, 1, 1, 1},
					{1, 1, 1, 1},
					{1, 1, 1, 1},
				},
				Static: [][]int{
					{1, 1, 1, 0}, // Static=1 at (0,0) where ground=0
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Chaser: [][]int{
					{1, 0, 0, 0}, // Turret=1 at (0,0) where ground=0
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				Zoner: [][]int{
					{1, 0, 0, 0}, // Zoner=1 at (0,0) where ground=0
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				DPS: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobAir: [][]int{
					{0, 1, 1, 0},
					{1, 0, 0, 1},
					{1, 0, 0, 1},
					{0, 1, 1, 0},
				},
				Meta: model.TemplateMeta{
					Name:    "test",
					Version: 1,
					Width:   4,
					Height:  4,
				},
			},
			strict:   false, // Non-strict mode
			expected: true,
			errors:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateTemplate(tt.payload, tt.strict)
			assert.Equal(t, tt.expected, result.Valid, "Expected valid=%t for test %s", tt.expected, tt.name)
			if tt.errors > 0 {
				assert.GreaterOrEqual(t, len(result.Errors), tt.errors, "Expected at least %d errors for test %s, got %d", tt.errors, tt.name, len(result.Errors))
			} else {
				assert.Equal(t, tt.errors, len(result.Errors), "Expected exactly %d errors for test %s, got %d", tt.errors, tt.name, len(result.Errors))
			}

			// Check that error details are populated for failed validations
			if !result.Valid {
				for _, err := range result.Errors {
					assert.NotEmpty(t, err.Layer)
					assert.NotEmpty(t, err.Reason)
					assert.GreaterOrEqual(t, err.X, 0)
					assert.GreaterOrEqual(t, err.Y, 0)
				}
			}
		})
	}
}

// softEdgeTestPayload builds a minimal strict-validatable payload carrying only
// a ground and a softEdge layer; every other required layer is zeroed, so the
// only logical rules that can fire are the soft edge ones.
func softEdgeTestPayload(ground, softEdge [][]int) *model.TemplatePayload {
	height := len(ground)
	width := len(ground[0])

	zeros := func() [][]int {
		layer := make([][]int, height)
		for y := range layer {
			layer[y] = make([]int, width)
		}
		return layer
	}

	return &model.TemplatePayload{
		Ground:   ground,
		SoftEdge: softEdge,
		Static:   zeros(),
		Chaser:   zeros(),
		Zoner:    zeros(),
		DPS:      zeros(),
		MobAir:   zeros(),
		Meta: model.TemplateMeta{
			Name:    "softedge",
			Version: 1,
			Width:   width,
			Height:  height,
		},
	}
}

// softEdgeErrors returns only the softEdge errors of a strict validation run.
func softEdgeErrors(payload *model.TemplatePayload) []model.ValidationError {
	result := ValidateTemplate(payload, true)
	var out []model.ValidationError
	for _, e := range result.Errors {
		if e.Layer == "softEdge" {
			out = append(out, e)
		}
	}
	return out
}

// TestValidateTemplate_SoftEdgeSupport covers ORT-116: soft edge validity is a
// least fixpoint over the layer, not a per-cell adjacency test. The propagation
// pairs are stated in visual space, so the data-space fixtures below look
// rotated — see the note on computeSoftEdgeSupport.
func TestValidateTemplate_SoftEdgeSupport(t *testing.T) {
	tests := []struct {
		name       string
		ground     [][]int
		softEdge   [][]int
		wantErrors int
		wantReason string
	}{
		{
			// Anchored on its data left and bottom edges, so the far corner is
			// reached only through the below+left rule.
			name: "patch anchored below-left is fully supported",
			ground: [][]int{
				{1, 0, 0, 0},
				{1, 0, 0, 0},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			softEdge: [][]int{
				{0, 1, 1, 1},
				{0, 1, 1, 1},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			wantErrors: 0,
		},
		{
			// Mirror image: anchored on its data top and right edges, so the far
			// corner is reached only through the above+right rule.
			name: "patch anchored above-right is fully supported",
			ground: [][]int{
				{1, 1, 1, 1},
				{0, 0, 0, 1},
				{0, 0, 0, 1},
				{0, 0, 0, 1},
			},
			softEdge: [][]int{
				{0, 0, 0, 0},
				{1, 1, 1, 0},
				{1, 1, 1, 0},
				{1, 1, 1, 0},
			},
			wantErrors: 0,
		},
		{
			// No ground anywhere, so nothing seeds the fixpoint and the patch
			// cannot bootstrap itself.
			name: "patch with no ground anchor is unsupported",
			ground: [][]int{
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			softEdge: [][]int{
				{0, 0, 0, 0},
				{0, 1, 1, 0},
				{0, 1, 1, 0},
				{0, 0, 0, 0},
			},
			wantErrors: 4,
			wantReason: "soft edge has no ground anchor",
		},
		{
			// (0,1) is anchored, but (1,1) needs BOTH cells of a pair — its left
			// is supported and its remaining three neighbours are bare.
			name: "one supported neighbour is not enough",
			ground: [][]int{
				{1, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			softEdge: [][]int{
				{0, 0, 0, 0},
				{1, 1, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			wantErrors: 1,
			wantReason: "soft edge has no ground anchor",
		},
		{
			// A diagonal neighbour never lends support on its own.
			name: "diagonal contact alone is unsupported",
			ground: [][]int{
				{1, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			softEdge: [][]int{
				{0, 0, 0, 0},
				{1, 0, 0, 0},
				{0, 1, 0, 0},
				{0, 0, 0, 0},
			},
			wantErrors: 1,
			wantReason: "soft edge has no ground anchor",
		},
		{
			// The pair that the rule deliberately excludes: below and right.
			// (2,1) and (1,2) are both supported soft edge, and (1,1) touches no
			// ground, so only the excluded pair could rescue it. Accepting it
			// would make the rule symmetric in all four corners.
			name: "the excluded below-right pair stays unsupported",
			ground: [][]int{
				{0, 0, 0, 0},
				{0, 0, 0, 1},
				{0, 0, 0, 0},
				{0, 1, 0, 0},
			},
			softEdge: [][]int{
				{0, 0, 0, 0},
				{0, 1, 1, 0},
				{0, 1, 0, 0},
				{0, 0, 0, 0},
			},
			wantErrors: 1,
			wantReason: "soft edge has no ground anchor",
		},
		{
			// The overlap rule is untouched by the relaxation.
			name: "soft edge on ground still errors",
			ground: [][]int{
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			softEdge: [][]int{
				{0, 0, 0, 0},
				{0, 1, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			wantErrors: 1,
			wantReason: "soft edge cannot overlap with ground",
		},
		{
			// The pre-ORT-116 base case: a plain ground-adjacent rim.
			name: "directly adjacent rim is still supported",
			ground: [][]int{
				{0, 0, 0, 0},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			softEdge: [][]int{
				{1, 1, 1, 1},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := softEdgeErrors(softEdgeTestPayload(tt.ground, tt.softEdge))
			assert.Len(t, errs, tt.wantErrors)
			if tt.wantReason != "" {
				for _, e := range errs {
					assert.Equal(t, tt.wantReason, e.Reason)
				}
			}
		})
	}
}

// TestValidateTemplate_SoftEdgeCornerInVisualSpace is the ORT-117 regression
// guard: a 16x10 default room whose soft edge traces an L along two walls, the
// shape a user draws in the editor. The corner of the L touches no ground and
// is anchored only through the propagation rule.
//
// In visual space the corner's right and bottom neighbours are soft edge. The
// canvas renders rotated 90 degrees CCW, so in the data space stored here those
// are the corner's BELOW and LEFT neighbours — the pair the rule must accept.
// The original ORT-116 implementation checked the opposite diagonal and flagged
// this corner.
func TestValidateTemplate_SoftEdgeCornerInVisualSpace(t *testing.T) {
	const width, height = 16, 10

	ground := make([][]int, height)
	softEdge := make([][]int, height)
	for y := 0; y < height; y++ {
		ground[y] = make([]int, width)
		softEdge[y] = make([]int, width)
	}

	// The room floor: everything except the two walls the soft edge runs along.
	for y := 2; y < height; y++ {
		for x := 0; x < width-2; x++ {
			ground[y][x] = 1
		}
	}

	// The L: one arm down data column 14, the other along data row 1. They meet
	// at (14,1), which has no ground neighbour of its own.
	for y := 1; y < height; y++ {
		softEdge[y][width-2] = 1
	}
	for x := 0; x < width-1; x++ {
		softEdge[1][x] = 1
	}

	payload := softEdgeTestPayload(ground, softEdge)
	assert.Empty(t, softEdgeErrors(payload), "the L corner at (14,1) must be anchored")

	supported := computeSoftEdgeSupport(payload, width, height)
	assert.True(t, supported[1][width-2], "corner (14,1) should borrow support from below+left")
}

// TestComputeSoftEdgeSupport_LargePatch checks that support propagates all the
// way across a wide patch anchored on two edges, i.e. that the fixpoint is
// iterated rather than applied in a single sweep.
func TestComputeSoftEdgeSupport_LargePatch(t *testing.T) {
	const size = 8

	ground := make([][]int, size)
	softEdge := make([][]int, size)
	for y := 0; y < size; y++ {
		ground[y] = make([]int, size)
		softEdge[y] = make([]int, size)
		for x := 0; x < size; x++ {
			// Ground fills the bottom row and the left column only, so support
			// has to climb up and to the right through the below+left rule.
			if y == size-1 || x == 0 {
				ground[y][x] = 1
			} else {
				softEdge[y][x] = 1
			}
		}
	}

	payload := softEdgeTestPayload(ground, softEdge)
	supported := computeSoftEdgeSupport(payload, size, size)

	for y := 0; y < size-1; y++ {
		for x := 1; x < size; x++ {
			assert.True(t, supported[y][x], "cell (%d,%d) should be supported", x, y)
		}
	}
	assert.Empty(t, softEdgeErrors(payload))
}
