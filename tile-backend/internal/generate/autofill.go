package generate

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"tile-backend/internal/model"

	"github.com/google/uuid"
)

// workItem represents a single room to generate during auto-fill
type workItem struct {
	shape     string // "all", "bridge", "platform"
	doorMask  int    // bitmask 0-15
	stageType string
}

// stageShapeCompat returns whether a stage type is compatible with a room shape.
// Shape uses DB values: "full", "bridge", "platform".
func stageShapeCompat(stageType, dbShape string) bool {
	cfg := GetStageConfig(stageType)
	if cfg == nil {
		return true // unknown stage, allow anything
	}
	if len(cfg.AllowedRoomTypes) == 0 {
		return true // no restriction
	}
	for _, allowed := range cfg.AllowedRoomTypes {
		if allowed == dbShape {
			return true
		}
	}
	return false
}

// stageDoorCompat returns whether a door bitmask is compatible with a stage type.
func stageDoorCompat(stageType string, doorMask int) bool {
	cfg := GetStageConfig(stageType)
	if cfg == nil {
		return true
	}
	doors := bitmaskToDoors(doorMask)
	if cfg.DoorRestrictions == nil {
		return len(doors) >= 2
	}
	if err := validateDoorRestrictions(cfg.DoorRestrictions, doors); err != nil {
		return false
	}
	return true
}

// bitmaskToDoors converts a bitmask (Top=1, Right=2, Bottom=4, Left=8) to DoorPosition slice.
func bitmaskToDoors(mask int) []DoorPosition {
	var doors []DoorPosition
	if mask&1 != 0 {
		doors = append(doors, DoorTop)
	}
	if mask&2 != 0 {
		doors = append(doors, DoorRight)
	}
	if mask&4 != 0 {
		doors = append(doors, DoorBottom)
	}
	if mask&8 != 0 {
		doors = append(doors, DoorLeft)
	}
	return doors
}

// dbShapeToGenShape maps DB room_type to the roomShape value used in payloads.
// DB stores "full", payload uses "all".
func dbShapeToGenShape(dbShape string) string {
	if dbShape == "full" {
		return "all"
	}
	return dbShape
}

// TemplateCreator is the subset of store.TemplateStore needed by AutoFill.
type TemplateCreator interface {
	Create(ctx context.Context, template model.Template) (*model.Template, error)
}

// AutoFill generates rooms to fill project deficits and saves them.
func AutoFill(ctx context.Context, project *model.Project, stats *model.ProjectStats, templateStore TemplateCreator) (*model.AutoFillResult, error) {
	items := buildWorkItems(project, stats)

	result := &model.AutoFillResult{
		Items: make([]model.AutoFillItem, 0, len(items)),
	}

	for _, item := range items {
		ri := model.AutoFillItem{
			Shape:     item.shape,
			DoorMask:  item.doorMask,
			StageType: item.stageType,
		}

		payload, err := generateRoom(item)
		if err != nil {
			ri.Error = err.Error()
			result.TotalFailed++
			result.Items = append(result.Items, ri)
			continue
		}

		// Build template and save
		tmpl := model.Template{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("autofill-%s-%s-%d", item.shape, item.stageType, item.doorMask),
			Version:   1,
			Width:     payload.Meta.Width,
			Height:    payload.Meta.Height,
			Payload:   *payload,
			ProjectID: &project.ID,
		}

		saved, err := templateStore.Create(ctx, tmpl)
		if err != nil {
			ri.Error = fmt.Sprintf("save failed: %s", err.Error())
			result.TotalFailed++
			result.Items = append(result.Items, ri)
			continue
		}

		ri.TemplateID = &saved.ID
		result.TotalGenerated++
		result.Items = append(result.Items, ri)
	}

	return result, nil
}

// buildWorkItems creates a list of rooms to generate from project deficits.
// It cross-matches shape, door, and stage deficits, prioritizing combinations
// with the largest total deficit.
func buildWorkItems(project *model.Project, stats *model.ProjectStats) []workItem {
	// Collect dimensions with deficit > 0
	type dimDeficit struct {
		key     string
		deficit int
	}

	var shapeDeficits []dimDeficit
	for k, v := range stats.Shape {
		if v.Deficit > 0 {
			shapeDeficits = append(shapeDeficits, dimDeficit{k, v.Deficit})
		}
	}
	var doorDeficits []dimDeficit
	for k, v := range stats.Door {
		if v.Deficit > 0 {
			doorDeficits = append(doorDeficits, dimDeficit{k, v.Deficit})
		}
	}
	var stageDeficits []dimDeficit
	for k, v := range stats.Stage {
		if v.Deficit > 0 {
			stageDeficits = append(stageDeficits, dimDeficit{k, v.Deficit})
		}
	}

	// Sort each by deficit descending (prioritize largest gaps)
	sort.Slice(shapeDeficits, func(i, j int) bool { return shapeDeficits[i].deficit > shapeDeficits[j].deficit })
	sort.Slice(doorDeficits, func(i, j int) bool { return doorDeficits[i].deficit > doorDeficits[j].deficit })
	sort.Slice(stageDeficits, func(i, j int) bool { return stageDeficits[i].deficit > stageDeficits[j].deficit })

	// Track remaining deficits
	shapeRemaining := make(map[string]int)
	for _, d := range shapeDeficits {
		shapeRemaining[d.key] = d.deficit
	}
	doorRemaining := make(map[string]int)
	for _, d := range doorDeficits {
		doorRemaining[d.key] = d.deficit
	}
	stageRemaining := make(map[string]int)
	for _, d := range stageDeficits {
		stageRemaining[d.key] = d.deficit
	}

	var items []workItem

	// Greedy matching: for each stage (largest deficit first), find compatible shape and door
	for _, sd := range stageDeficits {
		stage := sd.key
		for stageRemaining[stage] > 0 {
			// Find best shape: largest remaining deficit that is compatible with this stage
			bestShape := ""
			bestShapeDef := 0
			for _, sh := range shapeDeficits {
				if shapeRemaining[sh.key] > 0 && stageShapeCompat(stage, sh.key) {
					if shapeRemaining[sh.key] > bestShapeDef {
						bestShape = sh.key
						bestShapeDef = shapeRemaining[sh.key]
					}
				}
			}
			if bestShape == "" {
				// No compatible shape with deficit, pick any compatible shape
				for _, sh := range []string{"full", "bridge", "platform"} {
					if stageShapeCompat(stage, sh) {
						bestShape = sh
						break
					}
				}
			}
			if bestShape == "" {
				break // no compatible shape at all
			}

			// Find best door: largest remaining deficit that is compatible with this stage
			bestDoor := ""
			bestDoorDef := 0
			for _, dd := range doorDeficits {
				mask, _ := strconv.Atoi(dd.key)
				if doorRemaining[dd.key] > 0 && stageDoorCompat(stage, mask) {
					if doorRemaining[dd.key] > bestDoorDef {
						bestDoor = dd.key
						bestDoorDef = doorRemaining[dd.key]
					}
				}
			}
			if bestDoor == "" {
				// No compatible door with deficit — pick any from project config that's compatible
				for k := range project.DoorDistribution {
					mask, _ := strconv.Atoi(k)
					if stageDoorCompat(stage, mask) {
						bestDoor = k
						break
					}
				}
			}
			if bestDoor == "" {
				break // no compatible door config
			}

			doorMask, _ := strconv.Atoi(bestDoor)
			items = append(items, workItem{
				shape:     bestShape,
				doorMask:  doorMask,
				stageType: stage,
			})

			stageRemaining[stage]--
			if shapeRemaining[bestShape] > 0 {
				shapeRemaining[bestShape]--
			}
			if doorRemaining[bestDoor] > 0 {
				doorRemaining[bestDoor]--
			}
		}
	}

	return items
}

// generateRoom calls the appropriate generator for a work item.
func generateRoom(item workItem) (*model.TemplatePayload, error) {
	doors := bitmaskToDoors(item.doorMask)
	if len(doors) == 0 {
		return nil, fmt.Errorf("door bitmask %d has no doors", item.doorMask)
	}

	// Default dimensions
	width, height := 20, 12

	switch item.shape {
	case "full":
		req := FullRoomGenerateRequest{
			Width:        width,
			Height:       height,
			Doors:        doors,
			StageType:    item.stageType,
			RoomCategory: "normal",
		}
		resp, err := GenerateFullRoom(req)
		if err != nil {
			return nil, err
		}
		return &resp.Payload, nil

	case "bridge":
		req := BridgeGenerateRequest{
			Width:        width,
			Height:       height,
			Doors:        doors,
			StageType:    item.stageType,
			RoomCategory: "normal",
		}
		resp, err := GenerateBridgeRoom(req)
		if err != nil {
			return nil, err
		}
		return &resp.Payload, nil

	case "platform":
		req := PlatformGenerateRequest{
			Width:        width,
			Height:       height,
			Doors:        doors,
			StageType:    item.stageType,
			RoomCategory: "normal",
		}
		resp, err := GeneratePlatformRoom(req)
		if err != nil {
			return nil, err
		}
		return &resp.Payload, nil

	default:
		return nil, fmt.Errorf("unknown shape: %s", item.shape)
	}
}
