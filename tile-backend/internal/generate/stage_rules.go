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
	GroupCount       [2]int // how many groups to split into
	BossArena        bool   // requires 6x6 clear center area
}

// DoorRestriction defines door open constraints for a stage
type DoorRestriction struct {
	ForbidCornerPair bool // forbid 2-door diagonal combos (left+top, left+bottom, etc.)
	OnlyCornerPair   bool // only allow 2-door diagonal combos or 1-door
	MaxDoors         int  // 0 = no limit
}

// StageValidationResult contains validation result and adjusted counts
type StageValidationResult struct {
	Valid       bool
	Error       string
	ChaserCount int
	ZonerCount  int
	DPSCount    int
	MobAirCount int
	StaticCount int
	BossArena   *BossArenaInfo // non-nil if boss arena found
}

// BossArenaInfo describes the boss arena location
type BossArenaInfo struct {
	X, Y          int
	Width, Height int
}

var stageConfigs = map[string]StageConfig{
	model.StageTeaching: {
		StageType:   model.StageTeaching,
		DPSRange:    [2]int{2, 3},
		ChaserRange: [2]int{0, 0},
		ZonerRange:  [2]int{0, 0},
		MobAirRange: [2]int{0, 0},
	},
	model.StageBuilding: {
		StageType:   model.StageBuilding,
		DPSRange:    [2]int{2, 3},
		ChaserRange: [2]int{2, 3},
		ZonerRange:  [2]int{0, 0},
		MobAirRange: [2]int{0, 0},
	},
	model.StagePressure: {
		StageType:        model.StagePressure,
		AllowedRoomTypes: []string{"full", "platform"}, // not bridge
		DPSRange:         [2]int{4, 6},
		ChaserRange:      [2]int{6, 8},
		ZonerRange:       [2]int{1, 1},
		MobAirRange:      [2]int{2, 4},
		GroupCount:       [2]int{2, 2},
	},
	model.StagePeak: {
		StageType:        model.StagePeak,
		AllowedRoomTypes: []string{"full"}, // only full
		DoorRestrictions: &DoorRestriction{ForbidCornerPair: true},
		DPSRange:         [2]int{6, 12},
		ChaserRange:      [2]int{6, 8},
		ZonerRange:       [2]int{2, 3},
		MobAirRange:      [2]int{2, 4},
		GroupCount:       [2]int{2, 4},
	},
	model.StageRelease: {
		StageType:   model.StageRelease,
		DPSRange:    [2]int{0, 3},
		ChaserRange: [2]int{0, 0},
		ZonerRange:  [2]int{0, 0},
		MobAirRange: [2]int{0, 0},
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

// ValidateAndApplyStage validates stage constraints and returns adjusted enemy counts.
// If stageType is empty, passes through the original request counts.
func ValidateAndApplyStage(stageType, roomType string, doors []DoorPosition, ground [][]int, width, height int) (*StageValidationResult, error) {
	if stageType == "" {
		// No stage type — no constraints
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
	result := &StageValidationResult{
		Valid:       true,
		ChaserCount: randRange(cfg.ChaserRange[0], cfg.ChaserRange[1]),
		ZonerCount:  randRange(cfg.ZonerRange[0], cfg.ZonerRange[1]),
		DPSCount:    randRange(cfg.DPSRange[0], cfg.DPSRange[1]),
		MobAirCount: randRange(cfg.MobAirRange[0], cfg.MobAirRange[1]),
	}

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

// validateDoorRestrictions checks door configuration against stage restrictions
func validateDoorRestrictions(r *DoorRestriction, doors []DoorPosition) error {
	doorCount := len(doors)

	if r.MaxDoors > 0 && doorCount > r.MaxDoors {
		return fmt.Errorf("max %d doors allowed, got %d", r.MaxDoors, doorCount)
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
// Corner pairs: top+left, top+right, bottom+left, bottom+right
func isCornerDoorPair(doors []DoorPosition) bool {
	if len(doors) != 2 {
		return false
	}
	set := map[DoorPosition]bool{doors[0]: true, doors[1]: true}

	cornerPairs := [][2]DoorPosition{
		{DoorTop, DoorLeft},
		{DoorTop, DoorRight},
		{DoorBottom, DoorLeft},
		{DoorBottom, DoorRight},
	}

	for _, pair := range cornerPairs {
		if set[pair[0]] && set[pair[1]] {
			return true
		}
	}
	return false
}

// findBossArena finds a 6x6 clear area in the center of the room
// The area must be at least 3 cells from all edges
func findBossArena(ground [][]int, width, height int) *BossArenaInfo {
	const arenaSize = 6
	const minEdgeDist = 3

	// Search from center outward
	centerX, centerY := width/2, height/2

	bestDist := width + height
	var best *BossArenaInfo

	for y := minEdgeDist; y+arenaSize <= height-minEdgeDist; y++ {
		for x := minEdgeDist; x+arenaSize <= width-minEdgeDist; x++ {
			// Check if all cells are ground
			allGround := true
			for dy := 0; dy < arenaSize && allGround; dy++ {
				for dx := 0; dx < arenaSize && allGround; dx++ {
					if ground[y+dy][x+dx] != 1 {
						allGround = false
					}
				}
			}
			if allGround {
				// Distance from center
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
