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
	if payload.Chaser == nil {
		return fmt.Errorf("chaser layer is missing")
	}
	if payload.Zoner == nil {
		return fmt.Errorf("zoner layer is missing")
	}
	if payload.DPS == nil {
		return fmt.Errorf("dps layer is missing")
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
		"ground": payload.Ground,
		"static": payload.Static,
		"chaser": payload.Chaser,
		"zoner":  payload.Zoner,
		"dps":    payload.DPS,
		"mobAir": payload.MobAir,
	}

	// Add softEdge layer if present (optional for backward compatibility)
	if payload.SoftEdge != nil {
		layers["softEdge"] = payload.SoftEdge
	}

	// Add bridge layer if present (optional for backward compatibility)
	if payload.Bridge != nil {
		layers["bridge"] = payload.Bridge
	}

	// Add pipeline layer if present (optional for backward compatibility)
	if payload.Pipeline != nil {
		layers["pipeline"] = payload.Pipeline
	}

	// Add rail layer if present (optional for backward compatibility)
	if payload.Rail != nil {
		layers["rail"] = payload.Rail
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

	// Soft edge support is a property of the whole layer, not of a single cell
	// (ORT-116), so it is computed once here rather than inside the loop.
	softEdgeSupport := computeSoftEdgeSupport(payload, width, height)

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
			var pipeline int = 0
			if payload.Pipeline != nil && len(payload.Pipeline) > y && len(payload.Pipeline[y]) > x {
				pipeline = payload.Pipeline[y][x]
			}
			var rail int = 0
			if payload.Rail != nil && len(payload.Rail) > y && len(payload.Rail[y]) > x {
				rail = payload.Rail[y][x]
			}
			static := payload.Static[y][x]
			var chaser int = 0
			if payload.Chaser != nil && len(payload.Chaser) > y && len(payload.Chaser[y]) > x {
				chaser = payload.Chaser[y][x]
			}
			var zoner int = 0
			if payload.Zoner != nil && len(payload.Zoner) > y && len(payload.Zoner[y]) > x {
				zoner = payload.Zoner[y][x]
			}
			var dps int = 0
			if payload.DPS != nil && len(payload.DPS) > y && len(payload.DPS[y]) > x {
				dps = payload.DPS[y][x]
			}

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

				// Rule: softEdge must be anchored to ground — either directly
				// adjacent to it, or supported through neighbouring soft edges
				// that are themselves anchored (ORT-116).
				if !softEdgeSupport[y][x] {
					errors = append(errors, model.ValidationError{
						Layer:  "softEdge",
						X:      x,
						Y:      y,
						Reason: "soft edge has no ground anchor",
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

			// Pipeline validation rules
			if pipeline == 1 {
				// Rule: pipeline==1 => ground==1 (pipeline must be on ground)
				if ground == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "pipeline",
						X:      x,
						Y:      y,
						Reason: "pipeline must be placed on ground",
					})
				}
				// Rule: pipeline cannot be on bridge
				if bridge == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "pipeline",
						X:      x,
						Y:      y,
						Reason: "pipeline cannot be placed on bridge",
					})
				}
			}

			// Rail validation rules
			if rail == 1 {
				// Rule: rail==1 => ground==1 || bridge==1 (rail must be on ground or bridge)
				if ground == 0 && bridge == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "rail",
						X:      x,
						Y:      y,
						Reason: "rail must be placed on ground or bridge",
					})
				}

				// Rule: rail segments cannot branch or intersect (max 2 neighbors per
				// cell). Closed-loop requirement was dropped — endpoints are valid.
				neighborCount := countRailNeighbors(payload, x, y, width, height)
				if neighborCount > 2 {
					errors = append(errors, model.ValidationError{
						Layer:  "rail",
						X:      x,
						Y:      y,
						Reason: fmt.Sprintf("rail segments cannot intersect (has %d neighbors, max 2)", neighborCount),
					})
				}
			}

			// Rule: static==1 => (ground==1 || bridge==1) && bridge==0 && pipeline==0 && rail==0
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
				if pipeline == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "static",
						X:      x,
						Y:      y,
						Reason: "static items cannot be placed on pipeline",
					})
				}
				if rail == 1 {
					errors = append(errors, model.ValidationError{
						Layer:  "static",
						X:      x,
						Y:      y,
						Reason: "static items cannot be placed on rail",
					})
				}
			}

			// Rule: chaser==1 => ground==1
			if chaser == 1 {
				if ground == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "chaser",
						X:      x,
						Y:      y,
						Reason: "chasers require walkable ground",
					})
				}
			}

			// Rule: zoner==1 => ground==1
			if zoner == 1 {
				if ground == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "zoner",
						X:      x,
						Y:      y,
						Reason: "zoners require walkable ground",
					})
				}
			}

			// Rule: dps==1 => ground==1
			if dps == 1 {
				if ground == 0 {
					errors = append(errors, model.ValidationError{
						Layer:  "dps",
						X:      x,
						Y:      y,
						Reason: "dps require walkable ground",
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
		{-1, 0}, {1, 0}, // left, right
		{0, -1}, {0, 1}, // up, down
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

// countRailNeighbors counts adjacent rail cells for a given position
func countRailNeighbors(payload *model.TemplatePayload, x, y, width, height int) int {
	if payload.Rail == nil {
		return 0
	}

	directions := []struct{ dx, dy int }{
		{-1, 0}, {1, 0}, // left, right
		{0, -1}, {0, 1}, // up, down
	}

	count := 0
	for _, dir := range directions {
		nx, ny := x+dir.dx, y+dir.dy

		if nx >= 0 && nx < width && ny >= 0 && ny < height {
			if len(payload.Rail) > ny && len(payload.Rail[ny]) > nx {
				if payload.Rail[ny][nx] == 1 {
					count++
				}
			}
		}
	}

	return count
}

// softEdgeCell reads the softEdge layer defensively — it is optional, so a
// short or absent grid reads as 0.
func softEdgeCell(payload *model.TemplatePayload, x, y int) int {
	if payload.SoftEdge == nil || len(payload.SoftEdge) <= y || len(payload.SoftEdge[y]) <= x {
		return 0
	}
	return payload.SoftEdge[y][x]
}

// isAdjacentToGround checks if a position is adjacent to at least one ground tile
func isAdjacentToGround(payload *model.TemplatePayload, x, y, width, height int) bool {
	// Check all four directions
	directions := []struct{ dx, dy int }{
		{-1, 0}, {1, 0}, // left, right
		{0, -1}, {0, 1}, // up, down
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

// computeSoftEdgeSupport returns, for every cell, whether its soft edge is
// anchored to the ground (ORT-116). Support is a least fixpoint over the
// softEdge layer:
//
//	base       — the cell is 4-adjacent to a ground tile
//	right+down — the cells to the visual right and below are both supported
//	left+up    — the cells to the visual left and above are both supported
//
// The two propagation rules are stated in VISUAL space, the frame the editor
// draws and the rule was specified in. The canvas renders the grid rotated 90°
// CCW (ORT-111), so in the data space this function actually walks:
//
//	visual right+down → data (x, y+1) and (x-1, y)
//	visual left+up    → data (x, y-1) and (x+1, y)
//
// Getting this backwards inverts the rule into its mirror image, which fails
// exactly the L-shaped corner the rule exists to accept (ORT-117).
//
// The two propagation rules borrow support only from cells that are themselves
// supported, so a soft edge patch floating in the void with no ground anchor
// anywhere stays unsupported however large it is. Cells with softEdge == 0 are
// never supported and never lend support.
//
// Seeded with the base cells and relaxed with a worklist, so it is O(cells).
func computeSoftEdgeSupport(payload *model.TemplatePayload, width, height int) [][]bool {
	supported := make([][]bool, height)
	for y := range supported {
		supported[y] = make([]bool, width)
	}

	type point struct{ x, y int }
	var queue []point

	// Seed: every soft edge cell that touches ground directly.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if softEdgeCell(payload, x, y) != 1 {
				continue
			}
			if isAdjacentToGround(payload, x, y, width, height) {
				supported[y][x] = true
				queue = append(queue, point{x, y})
			}
		}
	}

	// isSupported reports support with an out-of-bounds guard, so a cell on the
	// room border cannot borrow support from outside the grid.
	isSupported := func(x, y int) bool {
		if x < 0 || x >= width || y < 0 || y >= height {
			return false
		}
		return supported[y][x]
	}

	// canBorrow applies the two propagation rules to a single cell. The offsets
	// are the data-space spelling of "visual right and below" / "visual left and
	// above" — see the rotation note above.
	canBorrow := func(x, y int) bool {
		if softEdgeCell(payload, x, y) != 1 {
			return false
		}
		return (isSupported(x, y+1) && isSupported(x-1, y)) ||
			(isSupported(x, y-1) && isSupported(x+1, y))
	}

	// Relax: a newly supported cell can only unlock the four neighbours that
	// name it in one of the two rules.
	for len(queue) > 0 {
		p := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		// Both rules name all four orthogonal neighbours between them, so the
		// candidate set is the same either way; only canBorrow encodes which
		// pair counts.
		neighbours := []point{
			{p.x - 1, p.y}, {p.x, p.y - 1},
			{p.x + 1, p.y}, {p.x, p.y + 1},
		}
		for _, n := range neighbours {
			if n.x < 0 || n.x >= width || n.y < 0 || n.y >= height {
				continue
			}
			if supported[n.y][n.x] || !canBorrow(n.x, n.y) {
				continue
			}
			supported[n.y][n.x] = true
			queue = append(queue, n)
		}
	}

	return supported
}
