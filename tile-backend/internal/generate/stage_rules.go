package generate

import (
	"fmt"
	"math/rand"
	"tile-backend/internal/model"
)

// StageConfig defines enemy count ranges and constraints for a stage type
type StageConfig struct {
	StageType        string
	AllowedRoomTypes []string // empty = all allowed
	DoorRestrictions *DoorRestriction
	ChaserRange      [2]int // [min, max]
	ZonerRange       [2]int
	DPSRange         [2]int
	MobAirRange      [2]int
	StaticRange      [2]int // [0,0] means use request value
	BossArena        bool   // requires 6x6 clear center area
	PlacementRule    string // placement rule identifier
}

// DoorRestriction defines door open constraints for a stage
type DoorRestriction struct {
	ForbidCornerPair bool           // forbid 2-door diagonal combos (left+top, left+bottom, etc.)
	OnlyCornerPair   bool           // only allow 2-door diagonal combos or 1-door
	MaxDoors         int            // 0 = no limit
	AllowedDoors     []DoorPosition // if non-empty, only these doors are allowed
}

// StagePlacementHints tells generators how to place enemies for a specific stage
type StagePlacementHints struct {
	// DPS placement
	DPSYRange [2]int // restrict DPS y-position to [min,max], [0,0] = no restriction

	// Chaser placement
	ChaserSymmetric bool // place chasers symmetrically (left-right mirror)
	ChaserCenterY   bool // prefer center Y region

	// Zoner placement
	ZonerCentral bool // place zoner as close to room center as possible

	// Grouping
	GroupCount int              // 0 = no grouping, just use default placement
	Groups     []PlacementGroup // if GroupCount > 0, defines how enemies are split per group
}

// PlacementGroup defines enemy allocation for one spatial group
type PlacementGroup struct {
	Region      GroupRegion // which part of the room
	DPSCount    int
	ChaserCount int
	ZonerCount  int
	MobAirCount int
}

// GroupRegion defines a spatial region of the room
type GroupRegion int

const (
	RegionFull        GroupRegion = iota // entire room
	RegionTop                            // top half
	RegionBottom                         // bottom half
	RegionLeft                           // left half
	RegionRight                          // right half
	RegionTopLeft                        // top-left quadrant
	RegionTopRight                       // top-right quadrant
	RegionBottomLeft                     // bottom-left quadrant
	RegionBottomRight                    // bottom-right quadrant
)

// GetRegionBounds returns the y and x bounds [minY, maxY, minX, maxX] for a region
func GetRegionBounds(region GroupRegion, width, height int) (minY, maxY, minX, maxX int) {
	midY := height / 2
	midX := width / 2
	switch region {
	case RegionTop:
		return 0, midY, 0, width
	case RegionBottom:
		return midY, height, 0, width
	case RegionLeft:
		return 0, height, 0, midX
	case RegionRight:
		return 0, height, midX, width
	case RegionTopLeft:
		return 0, midY, 0, midX
	case RegionTopRight:
		return 0, midY, midX, width
	case RegionBottomLeft:
		return midY, height, 0, midX
	case RegionBottomRight:
		return midY, height, midX, width
	default: // RegionFull
		return 0, height, 0, width
	}
}

// StageValidationResult contains validation result and adjusted counts
type StageValidationResult struct {
	Valid          bool
	ChaserCount    int
	ZonerCount     int
	DPSCount       int
	MobAirCount    int
	StaticCount    int
	BossArena      *BossArenaInfo
	PlacementHints *StagePlacementHints
}

// BossArenaInfo describes the boss arena location
type BossArenaInfo struct {
	X, Y          int
	Width, Height int
}

// StageConfigJSON is the JSON-friendly version for the frontend API
type StageConfigJSON struct {
	StageType        string   `json:"stageType"`
	AllowedRoomTypes []string `json:"allowedRoomTypes"`
	ChaserRange      [2]int   `json:"chaserRange"`
	ZonerRange       [2]int   `json:"zonerRange"`
	DPSRange         [2]int   `json:"dpsRange"`
	MobAirRange      [2]int   `json:"mobAirRange"`
	BossArena        bool     `json:"bossArena"`
}

var stageConfigs = map[string]StageConfig{
	model.StageStart: {
		StageType:        model.StageStart,
		DoorRestrictions: &DoorRestriction{MaxDoors: 1, AllowedDoors: []DoorPosition{DoorLeft}},
		DPSRange:         [2]int{0, 0},
		ChaserRange:      [2]int{0, 0},
		ZonerRange:       [2]int{0, 0},
		MobAirRange:      [2]int{0, 0},
		PlacementRule:    "start",
	},
	model.StageTeaching: {
		StageType:     model.StageTeaching,
		DPSRange:      [2]int{2, 3},
		ChaserRange:   [2]int{0, 0},
		ZonerRange:    [2]int{0, 0},
		MobAirRange:   [2]int{0, 0},
		PlacementRule: "teaching",
	},
	model.StageBuilding: {
		StageType:     model.StageBuilding,
		DPSRange:      [2]int{2, 3},
		ChaserRange:   [2]int{2, 3},
		ZonerRange:    [2]int{0, 0},
		MobAirRange:   [2]int{0, 0},
		PlacementRule: "building",
	},
	model.StagePressure: {
		StageType:        model.StagePressure,
		AllowedRoomTypes: []string{"full", "platform"}, // not bridge
		DPSRange:         [2]int{4, 6},
		ChaserRange:      [2]int{6, 8},
		ZonerRange:       [2]int{1, 1},
		MobAirRange:      [2]int{2, 4},
		PlacementRule:    "pressure",
	},
	model.StagePeak: {
		StageType:        model.StagePeak,
		AllowedRoomTypes: []string{"full"}, // only full
		DoorRestrictions: &DoorRestriction{ForbidCornerPair: true},
		DPSRange:         [2]int{6, 12},
		ChaserRange:      [2]int{6, 8},
		ZonerRange:       [2]int{2, 3},
		MobAirRange:      [2]int{2, 4},
		PlacementRule:    "peak",
	},
	model.StageRelease: {
		StageType:     model.StageRelease,
		DPSRange:      [2]int{0, 2},
		ChaserRange:   [2]int{0, 0},
		ZonerRange:    [2]int{0, 0},
		MobAirRange:   [2]int{0, 0},
		PlacementRule: "teaching", // same as teaching
	},
	model.StageBoss: {
		StageType:        model.StageBoss,
		AllowedRoomTypes: []string{"full", "platform"}, // not bridge
		DoorRestrictions: &DoorRestriction{OnlyCornerPair: true, MaxDoors: 2},
		DPSRange:         [2]int{0, 0},
		ChaserRange:      [2]int{0, 0},
		ZonerRange:       [2]int{0, 0},
		MobAirRange:      [2]int{0, 0},
		BossArena:        true,
		PlacementRule:    "boss",
	},
}

// GetStageConfig returns the config for a stage type, or nil if not found
func GetStageConfig(stageType string) *StageConfig {
	cfg, ok := stageConfigs[stageType]
	if !ok {
		return nil
	}
	return &cfg
}

// GetAllStageConfigs returns all stage configs as JSON-friendly structs for the frontend
func GetAllStageConfigs() []StageConfigJSON {
	order := []string{
		model.StageTeaching, model.StageBuilding, model.StagePressure,
		model.StagePeak, model.StageRelease, model.StageBoss,
	}
	var result []StageConfigJSON
	for _, st := range order {
		cfg := stageConfigs[st]
		result = append(result, StageConfigJSON{
			StageType:        cfg.StageType,
			AllowedRoomTypes: cfg.AllowedRoomTypes,
			ChaserRange:      cfg.ChaserRange,
			ZonerRange:       cfg.ZonerRange,
			DPSRange:         cfg.DPSRange,
			MobAirRange:      cfg.MobAirRange,
			BossArena:        cfg.BossArena,
		})
	}
	return result
}

// ValidateAndApplyStage validates stage constraints and returns adjusted enemy counts + placement hints.
func ValidateAndApplyStage(stageType, roomType string, doors []DoorPosition, ground [][]int, width, height int) (*StageValidationResult, error) {
	if stageType == "" {
		return &StageValidationResult{Valid: true}, nil
	}

	cfg := GetStageConfig(stageType)
	if cfg == nil {
		return nil, fmt.Errorf("unknown stage type: %s", stageType)
	}

	// Validate room type
	if len(cfg.AllowedRoomTypes) > 0 {
		allowed := false
		for _, rt := range cfg.AllowedRoomTypes {
			if rt == roomType {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("stage %s does not allow room type %s (allowed: %v)", stageType, roomType, cfg.AllowedRoomTypes)
		}
	}

	// Validate door restrictions
	if cfg.DoorRestrictions != nil {
		if err := validateDoorRestrictions(cfg.DoorRestrictions, doors); err != nil {
			return nil, fmt.Errorf("stage %s door restriction: %w", stageType, err)
		}
	}

	// Generate random counts within ranges
	chaserCount := randRange(cfg.ChaserRange[0], cfg.ChaserRange[1])
	zonerCount := randRange(cfg.ZonerRange[0], cfg.ZonerRange[1])
	dpsCount := randRange(cfg.DPSRange[0], cfg.DPSRange[1])
	mobAirCount := randRange(cfg.MobAirRange[0], cfg.MobAirRange[1])

	result := &StageValidationResult{
		Valid:       true,
		ChaserCount: chaserCount,
		ZonerCount:  zonerCount,
		DPSCount:    dpsCount,
		MobAirCount: mobAirCount,
	}

	// Build placement hints based on stage
	result.PlacementHints = buildPlacementHints(cfg, chaserCount, zonerCount, dpsCount, mobAirCount, width, height)

	// Boss arena check
	if cfg.BossArena {
		arena := findBossArena(ground, width, height)
		if arena == nil {
			return nil, fmt.Errorf("stage %s requires a 6x6 clear area in center (distance > 3 from edges), not found", stageType)
		}
		result.BossArena = arena
	}

	return result, nil
}

// buildPlacementHints creates stage-specific placement hints
func buildPlacementHints(cfg *StageConfig, chaserCount, zonerCount, dpsCount, mobAirCount, width, height int) *StagePlacementHints {
	hints := &StagePlacementHints{}

	switch cfg.PlacementRule {
	case "teaching":
		// DPS only, restricted to y ∈ [5,7]
		hints.DPSYRange = [2]int{5, 7}

	case "building":
		// Chaser: symmetric left-right, center area
		hints.ChaserSymmetric = true
		hints.ChaserCenterY = true
		// DPS: default placement (no special constraint)

	case "pressure":
		// Zoner central, split into 2 groups
		hints.ZonerCentral = true
		hints.GroupCount = 2
		// Split enemies into 2 groups, pick top/bottom or left/right randomly
		groupDPS := splitCount(dpsCount, 2)
		groupChaser := splitCount(chaserCount, 2)
		groupZoner := splitCount(zonerCount, 2)
		groupMobAir := splitCount(mobAirCount, 2)
		regions := pickHalves()
		for i := 0; i < 2; i++ {
			hints.Groups = append(hints.Groups, PlacementGroup{
				Region:      regions[i],
				DPSCount:    groupDPS[i],
				ChaserCount: groupChaser[i],
				ZonerCount:  groupZoner[i],
				MobAirCount: groupMobAir[i],
			})
		}

	case "peak":
		// Split into 2-4 groups
		groupCount := randRange(2, 4)
		hints.GroupCount = groupCount

		groupDPS := splitCount(dpsCount, groupCount)
		groupChaser := splitCount(chaserCount, groupCount)
		groupZoner := splitCount(zonerCount, groupCount)
		groupMobAir := splitCount(mobAirCount, groupCount)

		var regions []GroupRegion
		if groupCount == 2 {
			regions = pickHalves()
		} else {
			regions = []GroupRegion{RegionTopLeft, RegionTopRight, RegionBottomLeft, RegionBottomRight}
		}

		for i := 0; i < groupCount && i < len(regions); i++ {
			hints.Groups = append(hints.Groups, PlacementGroup{
				Region:      regions[i],
				DPSCount:    groupDPS[i],
				ChaserCount: groupChaser[i],
				ZonerCount:  groupZoner[i],
				MobAirCount: groupMobAir[i],
			})
		}

	case "boss":
		// No enemies, boss arena only

	case "start":
		// No enemies, starting room
	}

	return hints
}

// pickHalves randomly returns top/bottom or left/right region pair
func pickHalves() []GroupRegion {
	if rand.Float64() < 0.5 {
		return []GroupRegion{RegionTop, RegionBottom}
	}
	return []GroupRegion{RegionLeft, RegionRight}
}

// splitCount splits total into n roughly equal parts
func splitCount(total, n int) []int {
	if n <= 0 {
		return nil
	}
	parts := make([]int, n)
	base := total / n
	remainder := total % n
	for i := 0; i < n; i++ {
		parts[i] = base
		if i < remainder {
			parts[i]++
		}
	}
	return parts
}

// validateDoorRestrictions checks door configuration against stage restrictions
func validateDoorRestrictions(r *DoorRestriction, doors []DoorPosition) error {
	doorCount := len(doors)

	if r.MaxDoors > 0 && doorCount > r.MaxDoors {
		return fmt.Errorf("max %d doors allowed, got %d", r.MaxDoors, doorCount)
	}

	if len(r.AllowedDoors) > 0 {
		allowed := make(map[DoorPosition]bool, len(r.AllowedDoors))
		for _, d := range r.AllowedDoors {
			allowed[d] = true
		}
		for _, d := range doors {
			if !allowed[d] {
				return fmt.Errorf("door %q is not allowed for this stage (allowed: %v)", d, r.AllowedDoors)
			}
		}
	}

	if doorCount == 2 {
		isCornerPair := isCornerDoorPair(doors)
		if r.ForbidCornerPair && isCornerPair {
			return fmt.Errorf("2-door diagonal/corner combos are forbidden for this stage")
		}
		if r.OnlyCornerPair && !isCornerPair {
			return fmt.Errorf("only diagonal/corner door pairs allowed for this stage")
		}
	}

	if r.OnlyCornerPair && doorCount > 2 {
		return fmt.Errorf("only 1 or 2 doors (corner pair) allowed for this stage")
	}

	return nil
}

// isCornerDoorPair checks if 2 doors form a diagonal/corner pair
func isCornerDoorPair(doors []DoorPosition) bool {
	if len(doors) != 2 {
		return false
	}
	set := map[DoorPosition]bool{doors[0]: true, doors[1]: true}
	cornerPairs := [][2]DoorPosition{
		{DoorTop, DoorLeft}, {DoorTop, DoorRight},
		{DoorBottom, DoorLeft}, {DoorBottom, DoorRight},
	}
	for _, pair := range cornerPairs {
		if set[pair[0]] && set[pair[1]] {
			return true
		}
	}
	return false
}

// findBossArena finds a 6x6 clear area in the center of the room
func findBossArena(ground [][]int, width, height int) *BossArenaInfo {
	const arenaSize = 6
	const minEdgeDist = 3

	centerX, centerY := width/2, height/2
	bestDist := width + height
	var best *BossArenaInfo

	for y := minEdgeDist; y+arenaSize <= height-minEdgeDist; y++ {
		for x := minEdgeDist; x+arenaSize <= width-minEdgeDist; x++ {
			allGround := true
			for dy := 0; dy < arenaSize && allGround; dy++ {
				for dx := 0; dx < arenaSize && allGround; dx++ {
					if ground[y+dy][x+dx] != 1 {
						allGround = false
					}
				}
			}
			if allGround {
				midX := x + arenaSize/2
				midY := y + arenaSize/2
				dist := abs(midX-centerX) + abs(midY-centerY)
				if dist < bestDist {
					bestDist = dist
					best = &BossArenaInfo{X: x, Y: y, Width: arenaSize, Height: arenaSize}
				}
			}
		}
	}
	return best
}

// randRange returns a random int in [min, max] inclusive
func randRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}
