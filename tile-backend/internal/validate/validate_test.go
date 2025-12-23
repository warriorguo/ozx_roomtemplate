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
				Turret: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobGround: [][]int{
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
				Ground:    [][]int{{1}, {1}},
				Static:    [][]int{{0}, {1}},
				Turret:    [][]int{{0}, {0}},
				MobGround: [][]int{{0}, {0}},
				MobAir:    [][]int{{0}, {1}},
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
				Ground:    [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
				Static:    [][]int{{0, 1, 1, 0}}, // Wrong height
				Turret:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobGround: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobAir:    [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
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
				Ground:    [][]int{{1, 2, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}}, // Invalid value '2'
				Static:    [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Turret:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobGround: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobAir:    [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
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
				Ground:    [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
				Static:    [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Turret:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobGround: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				MobAir:    [][]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 1, 0}},
				Meta: model.TemplateMeta{
					Name:    "", // Empty name
					Version: 0, // Zero version
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
				Turret: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobGround: [][]int{
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
				Turret: [][]int{
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobGround: [][]int{
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
				Turret: [][]int{
					{1, 0, 0, 0}, // Turret at (0,0): ground=1, static=0
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobGround: [][]int{
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
				Turret: [][]int{
					{1, 0, 0, 0}, // Turret at (0,0): ground=0
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobGround: [][]int{
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
				Turret: [][]int{
					{1, 0, 0, 0}, // Turret=1 at (0,0) where ground=0
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
				},
				MobGround: [][]int{
					{1, 0, 0, 0}, // MobGround=1 at (0,0) where ground=0
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