package model

import (
	"encoding/json"
)

// ComputeTemplateStats calculates all computed fields for a template
func ComputeTemplateStats(template *Template) {
	// Calculate walkable ratio
	walkableRatio := CalculateWalkableRatio(template.Payload.Ground, template.Width, template.Height)
	template.WalkableRatio = &walkableRatio

	// Set room type if available in payload
	if template.Payload.RoomType != nil {
		template.RoomType = template.Payload.RoomType
	}

	// Set room attributes if available in payload
	if template.Payload.Attributes != nil {
		template.RoomAttributes = template.Payload.Attributes
	}

	// Calculate door connectivity
	if template.Payload.Doors != nil {
		doorsConnected := CalculateDoorsConnected(
			template.Payload.Ground,
			template.Payload.Doors,
			template.Width,
			template.Height,
		)
		template.DoorsConnected = &doorsConnected
	}

	// Count layer tiles
	staticCount := CountLayerTiles(template.Payload.Static)
	template.StaticCount = &staticCount

	turretCount := CountLayerTiles(template.Payload.Turret)
	template.TurretCount = &turretCount

	mobGroundCount := CountLayerTiles(template.Payload.MobGround)
	template.MobGroundCount = &mobGroundCount

	mobAirCount := CountLayerTiles(template.Payload.MobAir)
	template.MobAirCount = &mobAirCount
}

// CalculateWalkableRatio calculates the ratio of walkable tiles to total tiles
func CalculateWalkableRatio(ground Layer, width, height int) float64 {
	if width == 0 || height == 0 {
		return 0.0
	}

	walkableCount := 0
	totalTiles := width * height

	for y := 0; y < height && y < len(ground); y++ {
		for x := 0; x < width && x < len(ground[y]); x++ {
			if ground[y][x] == 1 {
				walkableCount++
			}
		}
	}

	return float64(walkableCount) / float64(totalTiles)
}

// CountLayerTiles counts the number of tiles with value 1 in a layer
func CountLayerTiles(layer Layer) int {
	count := 0
	for _, row := range layer {
		for _, cell := range row {
			if cell == 1 {
				count++
			}
		}
	}
	return count
}

// CalculateDoorsConnected checks if each door connects to walkable areas
func CalculateDoorsConnected(ground Layer, doors *DoorStates, width, height int) DoorsConnected {
	result := DoorsConnected{}

	if doors == nil || len(ground) == 0 {
		return result
	}

	midWidth := width / 2
	midHeight := height / 2

	// Check top door (y=0, middle two cells)
	if doors.Top == 1 && height > 0 {
		result.Top = isDoorConnected(ground, midWidth-1, 0, width, height) &&
			isDoorConnected(ground, midWidth, 0, width, height)
	}

	// Check bottom door (y=height-1, middle two cells)
	if doors.Bottom == 1 && height > 0 {
		result.Bottom = isDoorConnected(ground, midWidth-1, height-1, width, height) &&
			isDoorConnected(ground, midWidth, height-1, width, height)
	}

	// Check left door (x=0, middle two cells)
	if doors.Left == 1 && width > 0 {
		result.Left = isDoorConnected(ground, 0, midHeight-1, width, height) &&
			isDoorConnected(ground, 0, midHeight, width, height)
	}

	// Check right door (x=width-1, middle two cells)
	if doors.Right == 1 && width > 0 {
		result.Right = isDoorConnected(ground, width-1, midHeight-1, width, height) &&
			isDoorConnected(ground, width-1, midHeight, width, height)
	}

	return result
}

// isDoorConnected checks if a specific door cell is walkable
func isDoorConnected(ground Layer, x, y, width, height int) bool {
	if y < 0 || y >= height || x < 0 || x >= width {
		return false
	}
	if y >= len(ground) || x >= len(ground[y]) {
		return false
	}
	return ground[y][x] == 1
}

// SerializeRoomAttributes converts RoomAttributes to JSON for database storage
func SerializeRoomAttributes(attrs *RoomAttributes) ([]byte, error) {
	if attrs == nil {
		return nil, nil
	}
	return json.Marshal(attrs)
}

// DeserializeRoomAttributes converts JSON to RoomAttributes
func DeserializeRoomAttributes(data []byte) (*RoomAttributes, error) {
	if data == nil {
		return nil, nil
	}
	var attrs RoomAttributes
	if err := json.Unmarshal(data, &attrs); err != nil {
		return nil, err
	}
	return &attrs, nil
}

// SerializeDoorsConnected converts DoorsConnected to JSON for database storage
func SerializeDoorsConnected(doors *DoorsConnected) ([]byte, error) {
	if doors == nil {
		return nil, nil
	}
	return json.Marshal(doors)
}

// DeserializeDoorsConnected converts JSON to DoorsConnected
func DeserializeDoorsConnected(data []byte) (*DoorsConnected, error) {
	if data == nil {
		return nil, nil
	}
	var doors DoorsConnected
	if err := json.Unmarshal(data, &doors); err != nil {
		return nil, err
	}
	return &doors, nil
}
