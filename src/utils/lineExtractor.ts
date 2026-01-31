/**
 * Utility to extract minimum line segments from a grid layer (Pipeline/Rail)
 * Lines are horizontal (1×N) or vertical (N×1) where N > 1
 */

import type { LineSegment, Point } from '../services/api';

interface Segment {
  start: Point;
  end: Point;
  cells: Set<string>; // Set of "x,y" strings for cells covered
  length: number;
  orientation: 'horizontal' | 'vertical';
}

/**
 * Convert cell coordinates to a string key
 */
function cellKey(x: number, y: number): string {
  return `${x},${y}`;
}

/**
 * Find all maximal horizontal segments in a row
 */
function findHorizontalSegments(grid: number[][], y: number): Segment[] {
  const segments: Segment[] = [];
  const width = grid[0]?.length || 0;
  let startX = -1;

  for (let x = 0; x <= width; x++) {
    const value = x < width ? grid[y][x] : 0;

    if (value === 1 && startX === -1) {
      startX = x;
    } else if (value === 0 && startX !== -1) {
      const length = x - startX;
      if (length >= 1) {
        const cells = new Set<string>();
        for (let cx = startX; cx < x; cx++) {
          cells.add(cellKey(cx, y));
        }
        segments.push({
          start: { x: startX, y },
          end: { x: x - 1, y },
          cells,
          length,
          orientation: 'horizontal',
        });
      }
      startX = -1;
    }
  }

  return segments;
}

/**
 * Find all maximal vertical segments in a column
 */
function findVerticalSegments(grid: number[][], x: number): Segment[] {
  const segments: Segment[] = [];
  const height = grid.length;
  let startY = -1;

  for (let y = 0; y <= height; y++) {
    const value = y < height ? grid[y][x] : 0;

    if (value === 1 && startY === -1) {
      startY = y;
    } else if (value === 0 && startY !== -1) {
      const length = y - startY;
      if (length >= 1) {
        const cells = new Set<string>();
        for (let cy = startY; cy < y; cy++) {
          cells.add(cellKey(x, cy));
        }
        segments.push({
          start: { x, y: startY },
          end: { x, y: y - 1 },
          cells,
          length,
          orientation: 'vertical',
        });
      }
      startY = -1;
    }
  }

  return segments;
}

/**
 * Extract minimum line segments from a grid using greedy set cover
 * @param grid 2D array of 0s and 1s
 * @returns Array of line segments covering all cells with minimum segments
 */
export function extractLineSegments(grid: number[][]): LineSegment[] {
  if (!grid || grid.length === 0 || !grid[0] || grid[0].length === 0) {
    return [];
  }

  const height = grid.length;
  const width = grid[0].length;

  // Collect all cells that need to be covered
  const allCells = new Set<string>();
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (grid[y][x] === 1) {
        allCells.add(cellKey(x, y));
      }
    }
  }

  if (allCells.size === 0) {
    return [];
  }

  // Find all possible horizontal and vertical segments
  const allSegments: Segment[] = [];

  // Horizontal segments
  for (let y = 0; y < height; y++) {
    allSegments.push(...findHorizontalSegments(grid, y));
  }

  // Vertical segments
  for (let x = 0; x < width; x++) {
    allSegments.push(...findVerticalSegments(grid, x));
  }

  // Greedy set cover: always pick the segment that covers the most uncovered cells
  const coveredCells = new Set<string>();
  const selectedSegments: Segment[] = [];

  while (coveredCells.size < allCells.size) {
    let bestSegment: Segment | null = null;
    let bestNewCoverage = 0;

    for (const segment of allSegments) {
      // Count how many new cells this segment would cover
      let newCoverage = 0;
      for (const cell of segment.cells) {
        if (!coveredCells.has(cell)) {
          newCoverage++;
        }
      }

      // Prefer longer segments when coverage is equal
      if (newCoverage > bestNewCoverage ||
          (newCoverage === bestNewCoverage && bestSegment && segment.length > bestSegment.length)) {
        bestNewCoverage = newCoverage;
        bestSegment = segment;
      }
    }

    if (bestSegment && bestNewCoverage > 0) {
      selectedSegments.push(bestSegment);
      for (const cell of bestSegment.cells) {
        coveredCells.add(cell);
      }
    } else {
      // No more segments can cover new cells, but we still have uncovered cells
      // This shouldn't happen if the grid has valid data
      break;
    }
  }

  // Convert to LineSegment format
  return selectedSegments.map(seg => ({
    start: seg.start,
    end: seg.end,
  }));
}

/**
 * Check if a grid layer has any cells set to 1
 */
export function hasAnyCells(grid: number[][]): boolean {
  if (!grid || grid.length === 0) return false;
  for (const row of grid) {
    for (const cell of row) {
      if (cell === 1) return true;
    }
  }
  return false;
}
