package model

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// DoorDistribution maps bitmask values (0-15) as string keys to room counts.
type DoorDistribution map[string]int

// Project represents a complete project record
type Project struct {
	ID               uuid.UUID        `json:"id"`
	Name             string           `json:"name"`
	TotalRooms       int              `json:"total_rooms"`
	ShapePctFull     int              `json:"shape_pct_full"`
	ShapePctBridge   int              `json:"shape_pct_bridge"`
	ShapePctPlatform int              `json:"shape_pct_platform"`
	DoorDistribution DoorDistribution `json:"door_distribution"`
	StagePctStart    int              `json:"stage_pct_start"`
	StagePctTeaching int              `json:"stage_pct_teaching"`
	StagePctBuilding int              `json:"stage_pct_building"`
	StagePctPressure int              `json:"stage_pct_pressure"`
	StagePctPeak     int              `json:"stage_pct_peak"`
	StagePctRelease  int              `json:"stage_pct_release"`
	StagePctBoss     int              `json:"stage_pct_boss"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// ProjectSummary is used for list responses, includes computed template_count
type ProjectSummary struct {
	ID               uuid.UUID        `json:"id"`
	Name             string           `json:"name"`
	TotalRooms       int              `json:"total_rooms"`
	ShapePctFull     int              `json:"shape_pct_full"`
	ShapePctBridge   int              `json:"shape_pct_bridge"`
	ShapePctPlatform int              `json:"shape_pct_platform"`
	DoorDistribution DoorDistribution `json:"door_distribution"`
	StagePctStart    int              `json:"stage_pct_start"`
	StagePctTeaching int              `json:"stage_pct_teaching"`
	StagePctBuilding int              `json:"stage_pct_building"`
	StagePctPressure int              `json:"stage_pct_pressure"`
	StagePctPeak     int              `json:"stage_pct_peak"`
	StagePctRelease  int              `json:"stage_pct_release"`
	StagePctBoss     int              `json:"stage_pct_boss"`
	TemplateCount    int              `json:"template_count"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name             string           `json:"name"`
	TotalRooms       int              `json:"total_rooms"`
	ShapePctFull     int              `json:"shape_pct_full"`
	ShapePctBridge   int              `json:"shape_pct_bridge"`
	ShapePctPlatform int              `json:"shape_pct_platform"`
	DoorDistribution DoorDistribution `json:"door_distribution"`
	StagePctStart    int              `json:"stage_pct_start"`
	StagePctTeaching int              `json:"stage_pct_teaching"`
	StagePctBuilding int              `json:"stage_pct_building"`
	StagePctPressure int              `json:"stage_pct_pressure"`
	StagePctPeak     int              `json:"stage_pct_peak"`
	StagePctRelease  int              `json:"stage_pct_release"`
	StagePctBoss     int              `json:"stage_pct_boss"`
}

// UpdateProjectRequest is the same as CreateProjectRequest (full replace)
type UpdateProjectRequest = CreateProjectRequest

// ListProjectsQueryParams represents query parameters for listing projects
type ListProjectsQueryParams struct {
	Limit    int
	Offset   int
	NameLike string
}

// ListProjectsResponse represents the response for listing projects
type ListProjectsResponse struct {
	Total int              `json:"total"`
	Items []ProjectSummary `json:"items"`
}

// ValidateProjectRequest validates a CreateProjectRequest and returns a map of field->error.
// An empty map means the request is valid.
func ValidateProjectRequest(req *CreateProjectRequest) map[string]string {
	errors := make(map[string]string)

	if req.Name == "" {
		errors["name"] = "name is required"
	}
	if req.TotalRooms <= 0 {
		errors["total_rooms"] = "total_rooms must be positive"
	}

	// Shape percentages must be non-negative and sum to 100
	if req.ShapePctFull < 0 || req.ShapePctBridge < 0 || req.ShapePctPlatform < 0 {
		errors["shape_pct"] = "shape percentages must be non-negative"
	}
	shapeSum := req.ShapePctFull + req.ShapePctBridge + req.ShapePctPlatform
	if shapeSum != 100 {
		errors["shape_pct_sum"] = fmt.Sprintf("shape percentages must sum to 100, got %d", shapeSum)
	}

	// Stage percentages must sum to 100
	stagePcts := []int{req.StagePctStart, req.StagePctTeaching, req.StagePctBuilding,
		req.StagePctPressure, req.StagePctPeak, req.StagePctRelease, req.StagePctBoss}
	stageSum := 0
	for _, v := range stagePcts {
		if v < 0 {
			errors["stage_pct"] = "stage percentages must be non-negative"
			break
		}
		stageSum += v
	}
	if stageSum != 100 {
		errors["stage_pct_sum"] = fmt.Sprintf("stage percentages must sum to 100, got %d", stageSum)
	}

	// Door distribution: keys must be "0"-"15", values non-negative, sum equals total_rooms
	doorSum := 0
	for k, v := range req.DoorDistribution {
		intKey, err := strconv.Atoi(k)
		if err != nil || intKey < 0 || intKey > 15 {
			errors["door_distribution"] = fmt.Sprintf("invalid door bitmask key: %q (must be 0-15)", k)
			break
		}
		if v < 0 {
			errors["door_distribution"] = "door counts must be non-negative"
			break
		}
		doorSum += v
	}
	if _, hasErr := errors["door_distribution"]; !hasErr && req.TotalRooms > 0 && doorSum != req.TotalRooms {
		errors["door_distribution_sum"] = fmt.Sprintf("door distribution counts must sum to total_rooms (%d), got %d", req.TotalRooms, doorSum)
	}

	return errors
}
