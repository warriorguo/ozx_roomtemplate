package generate

import (
	"fmt"
	"math/rand"
)

// SoftEdgePlacement represents a potential soft edge placement
type SoftEdgePlacement struct {
	StartX, StartY int
	Width, Height  int // Either (1, N) or (N, 1) where N >= 3
}

const softEdgeMinLength = 3 // Minimum length (N > 2, so N >= 3)

// generateSoftEdgeLayerWithDebug generates the soft edge layer with debug info
// Soft edges are 1×N or N×1 strips (N > 2) placed in ground concave areas
func generateSoftEdgeLayerWithDebug(softEdgeLayer, ground [][]int, doorPositions map[DoorPosition]Point, width, height, targetCount int) *SoftEdgeDebugInfo {
	debug := &SoftEdgeDebugInfo{
		TargetCount: targetCount,
		PlacedCount: 0,
		Placements:  []PlaceInfo{},
		Misses:      []MissInfo{},
	}

	if targetCount <= 0 {
		debug.Skipped = true
		debug.SkipReason = "targetCount is 0"
		return debug
	}

	// Find all valid soft edge placements (concave areas)
	placements := findValidSoftEdgePlacements(ground, softEdgeLayer, doorPositions, width, height)
	if len(placements) == 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "no valid concave areas found in ground layer",
		})
		return debug
	}

	// Shuffle placements for variety
	for i := len(placements) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		placements[i], placements[j] = placements[j], placements[i]
	}

	// Place soft edges until target count reached or placements exhausted
	remaining := targetCount
	overlapCount := 0
	for _, placement := range placements {
		if remaining <= 0 {
			break
		}

		// Verify placement is still valid (not overlapping with already placed)
		if !canPlaceSoftEdge(placement, softEdgeLayer, width, height) {
			overlapCount++
			continue
		}

		// Place the soft edge
		placeSoftEdge(softEdgeLayer, placement)
		remaining--
		debug.PlacedCount++

		// Record placement
		debug.Placements = append(debug.Placements, PlaceInfo{
			Position: fmt.Sprintf("(%d,%d)", placement.StartX, placement.StartY),
			Size:     fmt.Sprintf("%dx%d", placement.Width, placement.Height),
			Reason:   "ground concave area",
		})
	}

	// Record miss info
	if overlapCount > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: "overlapping with already placed soft edge",
			Count:  overlapCount,
		})
	}
	if remaining > 0 {
		debug.Misses = append(debug.Misses, MissInfo{
			Reason: fmt.Sprintf("only %d valid placements available, needed %d more", len(placements), remaining),
		})
	}

	return debug
}

// findValidSoftEdgePlacements finds all valid positions for soft edge placement
// A valid soft edge is a 1×N or N×1 strip (N >= 3) in a ground concave area
func findValidSoftEdgePlacements(ground, softEdgeLayer [][]int, doorPositions map[DoorPosition]Point, width, height int) []SoftEdgePlacement {
	var placements []SoftEdgePlacement

	// Find horizontal soft edges (1×N, height=1, width=N)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Try to find a horizontal concave area starting at (x, y)
			if placement := findHorizontalConcave(ground, softEdgeLayer, doorPositions, x, y, width, height); placement != nil {
				placements = append(placements, *placement)
			}
		}
	}

	// Find vertical soft edges (N×1, height=N, width=1)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Try to find a vertical concave area starting at (x, y)
			if placement := findVerticalConcave(ground, softEdgeLayer, doorPositions, x, y, width, height); placement != nil {
				placements = append(placements, *placement)
			}
		}
	}

	return placements
}

// findHorizontalConcave finds a horizontal concave area (1×N) starting at (x, y)
// A horizontal concave is a void notch: void cells with ground on one horizontal edge (top or bottom)
// and ground cells on both ends (left and right), forming a U-shaped depression
func findHorizontalConcave(ground, softEdgeLayer [][]int, doorPositions map[DoorPosition]Point, startX, startY, width, height int) *SoftEdgePlacement {
	// Check if starting position is valid
	if startX >= width || startY >= height {
		return nil
	}

	// Starting position must be VOID (this is a void notch)
	if ground[startY][startX] != 0 {
		return nil
	}

	// Must have ground immediately to the left (this is the start of the notch)
	if startX == 0 || ground[startY][startX-1] != 1 {
		return nil
	}

	// Check door distance for starting position
	if !isFarEnoughFromDoors(startX, startY, doorPositions, softEdgeMinDoorDistance) {
		return nil
	}

	// Determine if this is a top-notch (ground below) or bottom-notch (ground above)
	hasGroundAbove := startY > 0 && ground[startY-1][startX] == 1
	hasGroundBelow := startY < height-1 && ground[startY+1][startX] == 1

	// Must have ground on exactly one horizontal side (forming a U-shape)
	if !hasGroundAbove && !hasGroundBelow {
		return nil // Not a concave notch - no ground on either horizontal side
	}
	if hasGroundAbove && hasGroundBelow {
		return nil // This is a tunnel, not a notch
	}

	// Find the length of this horizontal notch
	length := 1
	for x := startX + 1; x < width; x++ {
		// Must be void to continue the notch
		if ground[startY][x] != 0 {
			break
		}

		// Must maintain the same concave property
		gAbove := startY > 0 && ground[startY-1][x] == 1
		gBelow := startY < height-1 && ground[startY+1][x] == 1

		if hasGroundAbove && !gAbove {
			break // Ground above ended
		}
		if hasGroundBelow && !gBelow {
			break // Ground below ended
		}

		// Check door distance
		if !isFarEnoughFromDoors(x, startY, doorPositions, softEdgeMinDoorDistance) {
			break
		}

		length++
	}

	// Check if there's ground on the right side (closing the notch)
	endX := startX + length
	if endX >= width || ground[startY][endX] != 1 {
		return nil // Notch is open on the right, not a proper concave
	}

	// Must be at least 3 cells long
	if length < softEdgeMinLength {
		return nil
	}

	return &SoftEdgePlacement{
		StartX: startX,
		StartY: startY,
		Width:  length,
		Height: 1,
	}
}

// findVerticalConcave finds a vertical concave area (N×1) starting at (x, y)
// A vertical concave is a void notch: void cells with ground on one vertical edge (left or right)
// and ground cells on both ends (top and bottom), forming a U-shaped depression
func findVerticalConcave(ground, softEdgeLayer [][]int, doorPositions map[DoorPosition]Point, startX, startY, width, height int) *SoftEdgePlacement {
	// Check if starting position is valid
	if startX >= width || startY >= height {
		return nil
	}

	// Starting position must be VOID (this is a void notch)
	if ground[startY][startX] != 0 {
		return nil
	}

	// Must have ground immediately above (this is the start of the notch)
	if startY == 0 || ground[startY-1][startX] != 1 {
		return nil
	}

	// Check door distance for starting position
	if !isFarEnoughFromDoors(startX, startY, doorPositions, softEdgeMinDoorDistance) {
		return nil
	}

	// Determine if this is a left-notch (ground to the right) or right-notch (ground to the left)
	hasGroundLeft := startX > 0 && ground[startY][startX-1] == 1
	hasGroundRight := startX < width-1 && ground[startY][startX+1] == 1

	// Must have ground on exactly one vertical side (forming a U-shape)
	if !hasGroundLeft && !hasGroundRight {
		return nil // Not a concave notch - no ground on either vertical side
	}
	if hasGroundLeft && hasGroundRight {
		return nil // This is a tunnel, not a notch
	}

	// Find the length of this vertical notch
	length := 1
	for y := startY + 1; y < height; y++ {
		// Must be void to continue the notch
		if ground[y][startX] != 0 {
			break
		}

		// Must maintain the same concave property
		gLeft := startX > 0 && ground[y][startX-1] == 1
		gRight := startX < width-1 && ground[y][startX+1] == 1

		if hasGroundLeft && !gLeft {
			break // Ground on left ended
		}
		if hasGroundRight && !gRight {
			break // Ground on right ended
		}

		// Check door distance
		if !isFarEnoughFromDoors(startX, y, doorPositions, softEdgeMinDoorDistance) {
			break
		}

		length++
	}

	// Check if there's ground on the bottom side (closing the notch)
	endY := startY + length
	if endY >= height || ground[endY][startX] != 1 {
		return nil // Notch is open on the bottom, not a proper concave
	}

	// Must be at least 3 cells long
	if length < softEdgeMinLength {
		return nil
	}

	return &SoftEdgePlacement{
		StartX: startX,
		StartY: startY,
		Width:  1,
		Height: length,
	}
}

// placeSoftEdge places a soft edge on the layer
func placeSoftEdge(softEdgeLayer [][]int, placement SoftEdgePlacement) {
	for dy := 0; dy < placement.Height; dy++ {
		for dx := 0; dx < placement.Width; dx++ {
			x := placement.StartX + dx
			y := placement.StartY + dy
			softEdgeLayer[y][x] = 1
		}
	}
}
