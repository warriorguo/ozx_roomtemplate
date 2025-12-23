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
			static := payload.Static[y][x]
			turret := payload.Turret[y][x]
			mobGround := payload.MobGround[y][x]

			// Rule: static==1 => ground==1
			if static == 1 && ground == 0 {
				errors = append(errors, model.ValidationError{
					Layer:  "static",
					X:      x,
					Y:      y,
					Reason: "static items require walkable ground",
				})
			}

			// Rule: turret==1 => ground==1 && static==0
			if turret == 1 {
				if ground == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "turret",
						X:      x,
						Y:      y,
						Reason: "turrets require walkable ground",
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

			// Rule: mobGround==1 => ground==1 && static==0 && turret==0
			if mobGround == 1 {
				if ground == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "mobGround",
						X:      x,
						Y:      y,
						Reason: "ground mobs require walkable ground",
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