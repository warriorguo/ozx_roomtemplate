import type { Template, TileProperties, Grid, CellValue } from '../types/newTemplate';

/**
 * Calculate Manhattan distance between two points
 */
function manhattanDistance(x1: number, y1: number, x2: number, y2: number): number {
  return Math.abs(x1 - x2) + Math.abs(y1 - y2);
}

/**
 * BFS to find shortest path distance from (startX, startY) to any target cell
 * Returns distance or null if no path exists
 */
function bfsDistance(
  ground: Grid<CellValue>,
  width: number,
  height: number,
  startX: number,
  startY: number,
  isTarget: (x: number, y: number) => boolean
): number | null {
  // BFS queue: [x, y, distance]
  const queue: Array<[number, number, number]> = [[startX, startY, 0]];
  const visited = new Set<string>();
  visited.add(`${startX},${startY}`);

  while (queue.length > 0) {
    const [x, y, dist] = queue.shift()!;

    // Check if this is a target
    if (isTarget(x, y)) {
      return dist;
    }

    // Explore 4 directions (up, down, left, right)
    const directions = [
      [0, -1], // up
      [0, 1],  // down
      [-1, 0], // left
      [1, 0],  // right
    ];

    for (const [dx, dy] of directions) {
      const nx = x + dx;
      const ny = y + dy;
      const key = `${nx},${ny}`;

      // Check bounds
      if (nx < 0 || nx >= width || ny < 0 || ny >= height) {
        continue;
      }

      // Check if already visited
      if (visited.has(key)) {
        continue;
      }

      // Check if walkable
      if (ground[ny][nx] !== 1) {
        continue;
      }

      visited.add(key);
      queue.push([nx, ny, dist + 1]);
    }
  }

  return null; // No path found
}

/**
 * Find nearest tile with value=1 in a layer
 * Returns Manhattan distance or null if no such tile exists
 */
function findNearestTile(
  layer: Grid<CellValue>,
  width: number,
  height: number,
  fromX: number,
  fromY: number
): number | null {
  let minDist: number | null = null;

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (layer[y][x] === 1) {
        const dist = manhattanDistance(fromX, fromY, x, y);
        if (minDist === null || dist < minDist) {
          minDist = dist;
        }
      }
    }
  }

  return minDist;
}

/**
 * Calculate distance to edge (unwalkable cell or room boundary)
 * Returns the minimum distance to any ground=0 cell or room edge
 */
function calculateDistanceToEdge(
  ground: Grid<CellValue>,
  width: number,
  height: number,
  fromX: number,
  fromY: number
): number {
  let minDist = Infinity;

  // Check distance to room edges
  minDist = Math.min(
    minDist,
    fromY,              // top edge
    height - 1 - fromY, // bottom edge
    fromX,              // left edge
    width - 1 - fromX   // right edge
  );

  // Check distance to all unwalkable cells (ground=0)
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (ground[y][x] === 0) {
        const dist = manhattanDistance(fromX, fromY, x, y);
        minDist = Math.min(minDist, dist);
      }
    }
  }

  return minDist;
}

/**
 * Get door positions based on template dimensions
 */
function getDoorPositions(width: number, height: number) {
  const midWidth = Math.floor(width / 2);
  const midHeight = Math.floor(height / 2);

  return {
    top: [
      { x: midWidth - 1, y: 0 },
      { x: midWidth, y: 0 },
    ],
    bottom: [
      { x: midWidth - 1, y: height - 1 },
      { x: midWidth, y: height - 1 },
    ],
    left: [
      { x: 0, y: midHeight - 1 },
      { x: 0, y: midHeight },
    ],
    right: [
      { x: width - 1, y: midHeight - 1 },
      { x: width - 1, y: midHeight },
    ],
  };
}

/**
 * Calculate properties for a single tile at position (x, y)
 */
export function calculateTileProperties(
  template: Template,
  x: number,
  y: number
): TileProperties {
  const { width, height, ground, static: staticLayer, turret, doors } = template;

  // Check if tile is walkable (ground layer)
  const walkable = ground[y][x] === 1;

  // Calculate distances to walls (Manhattan distance)
  const distToTopWall = y;
  const distToBottomWall = height - 1 - y;
  const distToLeftWall = x;
  const distToRightWall = width - 1 - x;

  // Calculate distance to center
  const centerX = (width - 1) / 2;
  const centerY = (height - 1) / 2;
  const distToCenter = manhattanDistance(x, y, centerX, centerY);

  // Get door positions
  const doorPositions = getDoorPositions(width, height);

  // Calculate BFS distances to doors (only if door is open and tile is walkable)
  let distToTopDoor: number | null = null;
  let distToBottomDoor: number | null = null;
  let distToLeftDoor: number | null = null;
  let distToRightDoor: number | null = null;

  if (walkable) {
    // Top door
    if (doors.top === 1) {
      distToTopDoor = bfsDistance(
        ground,
        width,
        height,
        x,
        y,
        (tx, ty) => doorPositions.top.some(p => p.x === tx && p.y === ty)
      );
    }

    // Bottom door
    if (doors.bottom === 1) {
      distToBottomDoor = bfsDistance(
        ground,
        width,
        height,
        x,
        y,
        (tx, ty) => doorPositions.bottom.some(p => p.x === tx && p.y === ty)
      );
    }

    // Left door
    if (doors.left === 1) {
      distToLeftDoor = bfsDistance(
        ground,
        width,
        height,
        x,
        y,
        (tx, ty) => doorPositions.left.some(p => p.x === tx && p.y === ty)
      );
    }

    // Right door
    if (doors.right === 1) {
      distToRightDoor = bfsDistance(
        ground,
        width,
        height,
        x,
        y,
        (tx, ty) => doorPositions.right.some(p => p.x === tx && p.y === ty)
      );
    }
  }

  // Calculate distance to nearest static tile
  const distToNearStatic = findNearestTile(staticLayer, width, height, x, y);

  // Calculate distance to nearest turret tile
  const distToNearTurret = findNearestTile(turret, width, height, x, y);

  // Calculate distance to edge (unwalkable cell or room boundary)
  const distToEdge = calculateDistanceToEdge(ground, width, height, x, y);

  return {
    walkable,
    distToTopWall,
    distToBottomWall,
    distToLeftWall,
    distToRightWall,
    distToCenter,
    distToEdge,
    distToTopDoor,
    distToBottomDoor,
    distToLeftDoor,
    distToRightDoor,
    distToNearStatic,
    distToNearTurret,
  };
}

/**
 * Calculate properties for all tiles in the template
 * Returns a 2D grid where each cell has properties (or null if not set)
 */
export function calculateAllTileProperties(template: Template): Grid<TileProperties | null> {
  const { width, height, static: staticLayer, turret, mobGround, mobAir } = template;

  const properties: Grid<TileProperties | null> = [];

  for (let y = 0; y < height; y++) {
    properties[y] = [];
    for (let x = 0; x < width; x++) {
      // Only calculate properties for non-ground layers that are set (value=1)
      const hasFeature =
        staticLayer[y][x] === 1 ||
        turret[y][x] === 1 ||
        mobGround[y][x] === 1 ||
        mobAir[y][x] === 1;

      if (hasFeature) {
        properties[y][x] = calculateTileProperties(template, x, y);
      } else {
        properties[y][x] = null;
      }
    }
  }

  return properties;
}
