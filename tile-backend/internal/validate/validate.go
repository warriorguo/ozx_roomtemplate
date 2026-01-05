package validate

import (
	"fmt"
	"tile-backend/internal/model"
)

// ValidateTemplate performs comprehensive validation of a template payload
func ValidateTemplate(payload *model.TemplatePayload, strictValidation bool) *model.ValidationResult {
	result := &model.ValidationResult{
		Valid:  true,
		Errors: []model.ValidationError{},
	}

	// Basic structure validation
	if err := validateBasicStructure(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, model.ValidationError{
			Layer:  "meta",
			X:      0,
			Y:      0,
			Reason: err.Error(),
		})
		return result
	}

	// Validate dimensions
	if err := validateDimensions(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, model.ValidationError{
			Layer:  "meta",
			X:      0,
			Y:      0,
			Reason: err.Error(),
		})
		return result
	}

	// Validate each layer structure and values
	errors := validateLayers(payload)
	result.Errors = append(result.Errors, errors...)

	// If strict validation is enabled, perform logical validation
	if strictValidation {
		logicalErrors := validateLogicalRules(payload)
		result.Errors = append(result.Errors, logicalErrors...)
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// validateBasicStructure checks if all required fields are present
func validateBasicStructure(payload *model.TemplatePayload) error {
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}

	if payload.Ground == nil {
		return fmt.Errorf("ground layer is missing")
	}
	if payload.Static == nil {
		return fmt.Errorf("static layer is missing")
	}
	if payload.Turret == nil {
		return fmt.Errorf("turret layer is missing")
	}
	if payload.MobGround == nil {
		return fmt.Errorf("mobGround layer is missing")
	}
	if payload.MobAir == nil {
		return fmt.Errorf("mobAir layer is missing")
	}

	return nil
}

// validateDimensions checks if width and height are within valid ranges
func validateDimensions(payload *model.TemplatePayload) error {
	width := payload.Meta.Width
	height := payload.Meta.Height

	if width < 4 || height < 4 {
		return fmt.Errorf("width and height must be at least 4, got %dx%d", width, height)
	}

	if width > 200 || height > 200 {
		return fmt.Errorf("width and height must be at most 200, got %dx%d", width, height)
	}

	return nil
}

// validateLayers checks each layer's structure and value ranges
func validateLayers(payload *model.TemplatePayload) []model.ValidationError {
	var errors []model.ValidationError
	width := payload.Meta.Width
	height := payload.Meta.Height

	// Validate each layer
	layers := map[string]model.Layer{
		"ground":    payload.Ground,
		"static":    payload.Static,
		"turret":    payload.Turret,
		"mobGround": payload.MobGround,
		"mobAir":    payload.MobAir,
	}
	
	// Add softEdge layer if present (optional for backward compatibility)
	if payload.SoftEdge != nil {
		layers["softEdge"] = payload.SoftEdge
	}
	
	// Add bridge layer if present (optional for backward compatibility)
	if payload.Bridge != nil {
		layers["bridge"] = payload.Bridge
	}

	for layerName, layer := range layers {
		layerErrors := validateSingleLayer(layerName, layer, width, height)
		errors = append(errors, layerErrors...)
	}

	return errors
}

// validateSingleLayer validates a single layer's structure and values
func validateSingleLayer(layerName string, layer model.Layer, expectedWidth, expectedHeight int) []model.ValidationError {
	var errors []model.ValidationError

	// Check if layer has correct height
	if len(layer) != expectedHeight {
		errors = append(errors, model.ValidationError{
			Layer:  layerName,
			X:      0,
			Y:      0,
			Reason: fmt.Sprintf("layer has %d rows, expected %d", len(layer), expectedHeight),
		})
		return errors
	}

	// Check each row
	for y, row := range layer {
		// Check row width
		if len(row) != expectedWidth {
			errors = append(errors, model.ValidationError{
				Layer:  layerName,
				X:      0,
				Y:      y,
				Reason: fmt.Sprintf("row %d has %d columns, expected %d", y, len(row), expectedWidth),
			})
			continue
		}

		// Check cell values
		for x, value := range row {
			if value != 0 && value != 1 {
				errors = append(errors, model.ValidationError{
					Layer:  layerName,
					X:      x,
					Y:      y,
					Reason: fmt.Sprintf("invalid value %d, must be 0 or 1", value),
				})
			}
		}
	}

	return errors
}

// validateLogicalRules checks the logical constraints between layers
func validateLogicalRules(payload *model.TemplatePayload) []model.ValidationError {
	var errors []model.ValidationError
	width := payload.Meta.Width
	height := payload.Meta.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ground := payload.Ground[y][x]
			var softEdge int = 0
			if payload.SoftEdge != nil && len(payload.SoftEdge) > y && len(payload.SoftEdge[y]) > x {
				softEdge = payload.SoftEdge[y][x]
			}
			var bridge int = 0
			if payload.Bridge != nil && len(payload.Bridge) > y && len(payload.Bridge[y]) > x {
				bridge = payload.Bridge[y][x]
			}
			static := payload.Static[y][x]
			turret := payload.Turret[y][x]
			mobGround := payload.MobGround[y][x]

			// SoftEdge validation rules
			if softEdge == 1 {
				// Rule: softEdge==1 => ground==0 (softEdge cannot overlap with ground)
				if ground == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "softEdge",
						X:      x,
						Y:      y,
						Reason: "soft edge cannot overlap with ground",
					})
				}

				// Rule: softEdge must be adjacent to at least one ground tile
				if !isAdjacentToGround(payload, x, y, width, height) {
					errors = append(errors, model.ValidationError{
						Layer:  "softEdge",
						X:      x,
						Y:      y,
						Reason: "soft edge must be adjacent to ground",
					})
				}
			}

			// Bridge validation rules
			if bridge == 1 {
				// Rule: bridge==1 => ground==0 (bridge can only be placed on unwalkable ground)
				if ground == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "bridge",
						X:      x,
						Y:      y,
						Reason: "bridge cannot be placed on walkable ground",
					})
				}

				// Rule: bridge must connect walkable areas
				if !bridgeConnectsWalkableAreas(payload, x, y, width, height) {
					errors = append(errors, model.ValidationError{
						Layer:  "bridge",
						X:      x,
						Y:      y,
						Reason: "bridge must connect walkable areas",
					})
				}
			}

			// Rule: static==1 => (ground==1 || bridge==1) && bridge==0
			if static == 1 {
				if ground == 0 && bridge == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "static",
						X:      x,
						Y:      y,
						Reason: "static items require walkable ground or bridge",
					})
				}
				if bridge == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "static",
						X:      x,
						Y:      y,
						Reason: "static items cannot be placed on bridge",
					})
				}
			}

			// Rule: turret==1 => (ground==1 || bridge==1) && static==0 && bridge==0
			if turret == 1 {
				if ground == 0 && bridge == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "turret",
						X:      x,
						Y:      y,
						Reason: "turrets require walkable ground or bridge",
					})
				}
				if bridge == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "turret",
						X:      x,
						Y:      y,
						Reason: "turrets cannot be placed on bridge",
					})
				}
				if static == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "turret",
						X:      x,
						Y:      y,
						Reason: "turrets cannot be placed on static items",
					})
				}
			}

			// Rule: mobGround==1 => (ground==1 || bridge==1) && static==0 && turret==0 && bridge==0
			if mobGround == 1 {
				if ground == 0 && bridge == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "mobGround",
						X:      x,
						Y:      y,
						Reason: "ground mobs require walkable ground or bridge",
					})
				}
				if bridge == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "mobGround",
						X:      x,
						Y:      y,
						Reason: "ground mobs cannot be placed on bridge",
					})
				}
				if static == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "mobGround",
						X:      x,
						Y:      y,
						Reason: "ground mobs cannot be placed on static items",
					})
				}
				if turret == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "mobGround",
						X:      x,
						Y:      y,
						Reason: "ground mobs cannot be placed on turrets",
					})
				}
			}

			// Note: mobAir has no constraints, can be placed anywhere
		}
	}

	return errors
}

// bridgeConnectsWalkableAreas checks if a bridge tile connects walkable areas
func bridgeConnectsWalkableAreas(payload *model.TemplatePayload, x, y, width, height int) bool {
	// Check all four directions (horizontal and vertical)
	directions := []struct{ dx, dy int }{
		{-1, 0}, {1, 0},  // left, right
		{0, -1}, {0, 1},  // up, down
	}

	for _, dir := range directions {
		x1, y1 := x+dir.dx, y+dir.dy
		x2, y2 := x-dir.dx, y-dir.dy

		// Check if this direction has walkable areas on both sides
		side1Walkable := isWalkable(payload, x1, y1, width, height)
		side2Walkable := isWalkable(payload, x2, y2, width, height)

		if side1Walkable && side2Walkable {
			return true // Bridge connects walkable areas
		}
	}

	return false // Bridge doesn't connect walkable areas
}

// isWalkable checks if a position is walkable (ground=1 or bridge=1)
func isWalkable(payload *model.TemplatePayload, x, y, width, height int) bool {
	if x < 0 || x >= width || y < 0 || y >= height {
		return false
	}

	ground := payload.Ground[y][x]
	var bridge int = 0
	if payload.Bridge != nil && len(payload.Bridge) > y && len(payload.Bridge[y]) > x {
		bridge = payload.Bridge[y][x]
	}

	return ground == 1 || bridge == 1
}

// isAdjacentToGround checks if a position is adjacent to at least one ground tile
func isAdjacentToGround(payload *model.TemplatePayload, x, y, width, height int) bool {
	// Check all four directions
	directions := []struct{ dx, dy int }{
		{-1, 0}, {1, 0},  // left, right
		{0, -1}, {0, 1},  // up, down
	}

	for _, dir := range directions {
		nx, ny := x+dir.dx, y+dir.dy
		
		// Check bounds
		if nx >= 0 && nx < width && ny >= 0 && ny < height {
			if payload.Ground[ny][nx] == 1 {
				return true // Found adjacent ground tile
			}
		}
	}

	return false // No adjacent ground tile found
}