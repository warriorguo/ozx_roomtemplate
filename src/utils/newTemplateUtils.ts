import type {
  Template,
  LayerType,
  CellValue,
  Grid,
  ValidationResult,
  ValidationError,
  LayerValidation,
  DoorStates,
  DoorSide
} from '../types/newTemplate';
import { calculateAllTileProperties } from './tilePropertiesCalculator';

const DOOR_SIDES = ['top', 'right', 'bottom', 'left'] as const;

/**
 * Door sides live in two spaces and the editor has to speak both (ORT-111).
 *
 * **Data space** is the stored grid: `doors{}`, `doorOverrides{}`, the
 * `openDoors` bitmask, the derived filename, everything the backend and the OZX
 * importer read. `template.ground[0]` is the data top edge.
 *
 * **Visual space** is what the canvas draws. ORT-76 renders the grid rotated 90°
 * CCW — the same transpose + Y-flip OZX applies — so the room on screen matches
 * the room in game:
 *
 * | data   | visual |
 * |--------|--------|
 * | top    | left   |
 * | right  | top    |
 * | bottom | right  |
 * | left   | bottom |
 *
 * Before ORT-111 only the *canvas* was rotated; every label stayed in data
 * space, so the opening drawn at the visual top was reported and saved as
 * "Right" and the door controls were unusable without doing the rotation in
 * your head.
 *
 * The fix is display-only: labels and controls are visual, storage stays data.
 * These two functions are the **only** place the rotation is written down —
 * `LayerEditor.getDoorBorderSide` calls `dataToVisual` for exactly that reason.
 * Anything crossing the view boundary goes through here rather than open-coding
 * a mapping that can drift.
 */
const DATA_TO_VISUAL: Record<DoorSide, DoorSide> = {
  top: 'left',
  right: 'top',
  bottom: 'right',
  left: 'bottom',
};

const VISUAL_TO_DATA: Record<DoorSide, DoorSide> = {
  left: 'top',
  top: 'right',
  right: 'bottom',
  bottom: 'left',
};

/** Stored side → the canvas edge it is drawn on. */
export function dataToVisual(side: DoorSide): DoorSide {
  return DATA_TO_VISUAL[side];
}

/** Canvas edge the user pointed at → the stored side it means. */
export function visualToData(side: DoorSide): DoorSide {
  return VISUAL_TO_DATA[side];
}

/**
 * Visual sides in reading order, for iterating door UI. Use this instead of
 * DOOR_SIDES wherever the list is rendered to the user.
 */
export const VISUAL_DOOR_SIDES: readonly DoorSide[] = ['top', 'right', 'bottom', 'left'];

/**
 * 根据 ground 连通性检测每个门是否“物理上”开通：
 * 门对应的两个中间格子在 ground 层都为 1 则视为连通。
 */
export function detectDoorConnectivity(template: Template): DoorStates {
  const { width, height, ground } = template;

  const midWidth = Math.floor(width / 2);
  const midHeight = Math.floor(height / 2);

  const topOpen =
    ground[0]?.[midWidth - 1] === 1 &&
    ground[0]?.[midWidth] === 1 ? 1 : 0;

  const bottomOpen =
    ground[height - 1]?.[midWidth - 1] === 1 &&
    ground[height - 1]?.[midWidth] === 1 ? 1 : 0;

  const leftOpen =
    ground[midHeight - 1]?.[0] === 1 &&
    ground[midHeight]?.[0] === 1 ? 1 : 0;

  const rightOpen =
    ground[midHeight - 1]?.[width - 1] === 1 &&
    ground[midHeight]?.[width - 1] === 1 ? 1 : 0;

  return {
    top: topOpen as 0 | 1,
    right: rightOpen as 0 | 1,
    bottom: bottomOpen as 0 | 1,
    left: leftOpen as 0 | 1,
  };
}

/** The sides the user has explicitly marked open (the override whitelist). */
function selectedOpenSides(template: Template): DoorSide[] {
  const ov = template.doorOverrides;
  return DOOR_SIDES.filter((s) => ov?.[s] === 1);
}

/**
 * 计算门的最终开通状态（写入 meta 与文件名 openDoors 的依据）。
 *
 * 规则（whitelist 模型）：
 * - 如果用户做了显式选择（doorOverrides 中有任意一侧为 1），则开门集合 = 所选集合，
 *   未选中的门一律视为关闭（即使 ground 连通）。
 * - 如果用户没有选择，则回退到 ground 连通性自动检测。
 *
 * 注意：选中但实际未连通的门属于无效选择，由 getInvalidDoorSelections 检出，
 * 并在保存时报错拦截（见 saveTemplate）。
 */
export function calculateDoorStates(template: Template): DoorStates {
  const connectivity = detectDoorConnectivity(template);
  const selected = selectedOpenSides(template);
  if (selected.length === 0) {
    return connectivity;
  }
  return {
    top: selected.includes('top') ? 1 : 0,
    right: selected.includes('right') ? 1 : 0,
    bottom: selected.includes('bottom') ? 1 : 0,
    left: selected.includes('left') ? 1 : 0,
  };
}

/**
 * Returns the sides the user explicitly marked open that the ground layer does
 * NOT actually connect. A non-empty result means the door selection is invalid
 * and the template must not be saved.
 */
export function getInvalidDoorSelections(template: Template): DoorSide[] {
  const connectivity = detectDoorConnectivity(template);
  return selectedOpenSides(template).filter((s) => connectivity[s] !== 1);
}

export function createEmptyTemplate(width: number, height: number): Template {
  if (width <= 0 || height <= 0 || width > 200 || height > 200) {
    throw new Error('Width and height must be between 1 and 200');
  }

  const createLayer = (): Grid<CellValue> =>
    Array(height).fill(null).map(() => Array(width).fill(0));

  const template: Template = {
    version: 1,
    width,
    height,
    ground: createLayer(),
    softEdge: createLayer(),
    bridge: createLayer(),
    pipeline: createLayer(),
    rail: createLayer(),
    static: createLayer(),
    chaser: createLayer(),
    zoner: createLayer(),
    dps: createLayer(),
    mainPath: createLayer(),
    mobAir: createLayer(),
    doors: { top: 0, right: 0, bottom: 0, left: 0 },
    doorOverrides: {},
    stageType: 'teaching',
    roomType: 'full',
    roomCategory: 'normal',
    tileProperties: Array(height).fill(null).map(() => Array(width).fill(null)),
  };

  // 计算初始门状态
  template.doors = calculateDoorStates(template);

  // 计算初始 tile properties
  template.tileProperties = calculateAllTileProperties(template);

  return template;
}

export function setCellValue(
  template: Template,
  layer: LayerType,
  x: number,
  y: number,
  value: CellValue
): Template {
  if (x < 0 || x >= template.width || y < 0 || y >= template.height) {
    return template;
  }

  if (template[layer][y][x] === value) {
    return template;
  }

  const newTemplate = JSON.parse(JSON.stringify(template)) as Template;
  newTemplate[layer][y][x] = value;

  // 如果修改的是 ground 层，重新计算门状态
  if (layer === 'ground') {
    newTemplate.doors = calculateDoorStates(newTemplate);
  }

  // 重新计算 tile properties (任何层修改都需要重新计算)
  newTemplate.tileProperties = calculateAllTileProperties(newTemplate);

  return newTemplate;
}

// Rule-based validation functions
export function validateCellRules(
  template: Template,
  x: number,
  y: number,
  softEdgeSupport?: boolean[][]
): Record<LayerType, boolean> {
  const ground = template.ground[y][x];
  const bridge = template.bridge[y][x];
  const pipeline = template.pipeline[y][x];
  const rail = template.rail[y][x];
  const static_ = template.static[y][x];
  const chaser = template.chaser[y][x];
  const zoner = template.zoner[y][x];
  const dps = template.dps[y][x];
  // const mobAir = template.mobAir[y][x]; // Not used in validation

  return {
    ground: true, // Ground has no constraints
    softEdge: validateSoftEdgeCell(template, x, y, softEdgeSupport ?? computeSoftEdgeSupport(template)),
    bridge: validateBridgeCell(template, x, y),
    // Pipeline: must be on ground, cannot be on bridge
    pipeline: pipeline === 0 || (ground === 1 && bridge === 0),
    // Rail: must be on ground or bridge; segments cannot branch/intersect (max 2 rail neighbors). Endpoints are allowed.
    rail: validateRailCell(template, x, y),
    // Static: can't be on bridge, can't conflict with pipeline or rail
    static: static_ === 0 || ((ground === 1 || bridge === 1) && bridge === 0 && pipeline === 0 && rail === 0),
    // Chaser: requires ground=1, cannot be on static/bridge/pipeline/rail/zoner
    chaser: chaser === 0 || (ground === 1 && static_ === 0 && bridge === 0 && pipeline === 0 && rail === 0 && zoner === 0),
    // Zoner: requires ground=1, cannot be on static/bridge/pipeline/rail/chaser
    zoner: zoner === 0 || (ground === 1 && static_ === 0 && bridge === 0 && pipeline === 0 && rail === 0 && chaser === 0),
    // DPS: requires ground=1, cannot be on static/bridge/pipeline/rail/zoner (ORT-119)
    dps: dps === 0 || (ground === 1 && static_ === 0 && bridge === 0 && pipeline === 0 && rail === 0 && zoner === 0),
    mainPath: true, // MainPath is read-only, no constraints
    mobAir: true, // MobAir has no constraints
  };
}

const ORTHOGONAL_DIRECTIONS = [
  { dx: -1, dy: 0 }, // left
  { dx: 1, dy: 0 },  // right
  { dx: 0, dy: -1 }, // up
  { dx: 0, dy: 1 },  // down
];

// Check whether a cell touches at least one ground tile orthogonally
function isAdjacentToGround(template: Template, x: number, y: number): boolean {
  return ORTHOGONAL_DIRECTIONS.some(dir => {
    const nx = x + dir.dx;
    const ny = y + dir.dy;
    if (nx < 0 || nx >= template.width || ny < 0 || ny >= template.height) return false;
    return template.ground[ny][nx] === 1;
  });
}

// Compute, for every cell, whether its soft edge is anchored to the ground (ORT-116).
// Support is a least fixpoint over the softEdge layer:
//
//   base       — the cell is orthogonally adjacent to a ground tile
//   right+down — the cells to the visual right and below are both supported
//   left+up    — the cells to the visual left and above are both supported
//
// The propagation rules are stated in VISUAL space, the frame the canvas draws
// and the rule was specified in. The grid renders rotated 90° CCW (ORT-111), so
// in the data space this function walks:
//
//   visual right+down -> data (x, y+1) and (x-1, y)
//   visual left+up    -> data (x, y-1) and (x+1, y)
//
// Getting this backwards inverts the rule into its mirror image (ORT-117).
//
// The two propagation rules borrow support only from cells that are themselves
// supported, so a soft edge patch floating in the void with no ground anchor
// anywhere stays unsupported however large it is. Cells with softEdge === 0 are
// never supported and never lend support.
//
// Mirrors computeSoftEdgeSupport in tile-backend/internal/validate/validate.go.
export function computeSoftEdgeSupport(template: Template): boolean[][] {
  const supported: boolean[][] = [];
  for (let y = 0; y < template.height; y++) {
    supported[y] = new Array<boolean>(template.width).fill(false);
  }

  const queue: Array<{ x: number; y: number }> = [];

  // Seed: every soft edge cell that touches ground directly
  for (let y = 0; y < template.height; y++) {
    for (let x = 0; x < template.width; x++) {
      if (template.softEdge[y][x] !== 1) continue;
      if (isAdjacentToGround(template, x, y)) {
        supported[y][x] = true;
        queue.push({ x, y });
      }
    }
  }

  const isSupported = (x: number, y: number): boolean => {
    if (x < 0 || x >= template.width || y < 0 || y >= template.height) return false;
    return supported[y][x];
  };

  // The offsets are the data-space spelling of "visual right and below" /
  // "visual left and above" — see the rotation note above.
  const canBorrow = (x: number, y: number): boolean => {
    if (template.softEdge[y][x] !== 1) return false;
    return (isSupported(x, y + 1) && isSupported(x - 1, y))
      || (isSupported(x, y - 1) && isSupported(x + 1, y));
  };

  // Relax: a newly supported cell can only unlock the four neighbours that name
  // it in one of the two rules
  while (queue.length > 0) {
    const cell = queue.pop()!;
    // Both rules name all four orthogonal neighbours between them, so the
    // candidate set is the same either way; only canBorrow encodes which pair counts.
    const neighbours = [
      { x: cell.x - 1, y: cell.y }, { x: cell.x, y: cell.y - 1 },
      { x: cell.x + 1, y: cell.y }, { x: cell.x, y: cell.y + 1 },
    ];
    for (const n of neighbours) {
      if (n.x < 0 || n.x >= template.width || n.y < 0 || n.y >= template.height) continue;
      if (supported[n.y][n.x] || !canBorrow(n.x, n.y)) continue;
      supported[n.y][n.x] = true;
      queue.push(n);
    }
  }

  return supported;
}

// Validate soft edge placement: must be anchored to ground but not overlap with ground
function validateSoftEdgeCell(
  template: Template,
  x: number,
  y: number,
  softEdgeSupport: boolean[][]
): boolean {
  const softEdge = template.softEdge[y][x];
  if (softEdge === 0) return true; // Empty soft edge cells are always valid

  // Soft edge cannot overlap with ground
  if (template.ground[y][x] === 1) return false;

  return softEdgeSupport[y][x];
}

// Count adjacent rail cells for a given position
function countRailNeighbors(template: Template, x: number, y: number): number {
  const directions = [
    { dx: -1, dy: 0 }, // left
    { dx: 1, dy: 0 },  // right
    { dx: 0, dy: -1 }, // up
    { dx: 0, dy: 1 }   // down
  ];

  let count = 0;
  for (const dir of directions) {
    const nx = x + dir.dx;
    const ny = y + dir.dy;

    if (nx >= 0 && nx < template.width && ny >= 0 && ny < template.height) {
      if (template.rail[ny][nx] === 1) {
        count++;
      }
    }
  }
  return count;
}

// Validate rail cell: must be on ground or bridge, and cannot branch/intersect (max 2 rail neighbors). Endpoints (0 or 1 neighbor) are valid.
function validateRailCell(template: Template, x: number, y: number): boolean {
  const rail = template.rail[y][x];
  if (rail === 0) return true; // Empty rail cells are always valid

  const ground = template.ground[y][x];
  const bridge = template.bridge[y][x];

  // Rail must be on ground or bridge
  if (ground === 0 && bridge === 0) return false;

  const neighborCount = countRailNeighbors(template, x, y);
  return neighborCount <= 2;
}

// Validate bridge placement: bridge should span unwalkable areas (ground=0) to connect walkable areas
function validateBridgeCell(template: Template, x: number, y: number): boolean {
  const bridge = template.bridge[y][x];
  if (bridge === 0) return true; // Empty bridge cells are always valid
  
  const ground = template.ground[y][x];
  
  // Bridge can only be placed on unwalkable ground
  if (ground === 1) return false;
  
  // Check if bridge connects walkable areas in any direction
  const directions = [
    { dx: -1, dy: 0 }, // left
    { dx: 1, dy: 0 },  // right  
    { dx: 0, dy: -1 }, // up
    { dx: 0, dy: 1 }   // down
  ];
  
  for (const dir of directions) {
    const x1 = x + dir.dx;
    const y1 = y + dir.dy;
    const x2 = x - dir.dx;
    const y2 = y - dir.dy;
    
    // Check if this direction has walkable areas on both sides
    const side1Walkable = isWalkable(template, x1, y1);
    const side2Walkable = isWalkable(template, x2, y2);
    
    if (side1Walkable && side2Walkable) {
      return true; // Bridge connects walkable areas
    }
  }
  
  return false; // Bridge doesn't connect walkable areas
}

function isWalkable(template: Template, x: number, y: number): boolean {
  if (x < 0 || x >= template.width || y < 0 || y >= template.height) {
    return false;
  }
  return template.ground[y][x] === 1 || template.bridge[y][x] === 1;
}

export function validateTemplate(template: Template): ValidationResult {
  const errors: ValidationError[] = [];
  // Soft edge support is a property of the whole layer, not of a single cell
  // (ORT-116), so it is computed once here rather than inside the loop.
  const softEdgeSupport = computeSoftEdgeSupport(template);
  const layerValidation: LayerValidation = {
    ground: [],
    softEdge: [],
    bridge: [],
    pipeline: [],
    rail: [],
    static: [],
    chaser: [],
    zoner: [],
    dps: [],
    mainPath: [],
    mobAir: [],
  };

  // Initialize validation grids
  for (let y = 0; y < template.height; y++) {
    layerValidation.ground[y] = [];
    layerValidation.softEdge[y] = [];
    layerValidation.bridge[y] = [];
    layerValidation.pipeline[y] = [];
    layerValidation.rail[y] = [];
    layerValidation.static[y] = [];
    layerValidation.chaser[y] = [];
    layerValidation.zoner[y] = [];
    layerValidation.dps[y] = [];
    layerValidation.mainPath[y] = [];
    layerValidation.mobAir[y] = [];

    for (let x = 0; x < template.width; x++) {
      const cellValidation = validateCellRules(template, x, y, softEdgeSupport);

      // Store validation results
      layerValidation.ground[y][x] = cellValidation.ground;
      layerValidation.softEdge[y][x] = cellValidation.softEdge;
      layerValidation.bridge[y][x] = cellValidation.bridge;
      layerValidation.pipeline[y][x] = cellValidation.pipeline;
      layerValidation.rail[y][x] = cellValidation.rail;
      layerValidation.static[y][x] = cellValidation.static;
      layerValidation.chaser[y][x] = cellValidation.chaser;
      layerValidation.zoner[y][x] = cellValidation.zoner;
      layerValidation.dps[y][x] = cellValidation.dps;
      layerValidation.mainPath[y][x] = cellValidation.mainPath;
      layerValidation.mobAir[y][x] = cellValidation.mobAir;

      // Collect errors for cells that have value=1 but are invalid
      const layers: LayerType[] = ['softEdge', 'bridge', 'pipeline', 'rail', 'static', 'chaser', 'zoner', 'dps', 'mobAir'];

      layers.forEach(layer => {
        if (template[layer][y][x] === 1 && !cellValidation[layer]) {
          errors.push({
            layer,
            x,
            y,
            reason: getValidationErrorReason(layer, template, x, y),
          });
        }
      });
    }
  }

  return {
    isValid: errors.length === 0,
    errors,
    layerValidation,
  };
}

function getValidationErrorReason(
  layer: LayerType,
  template: Template,
  x: number,
  y: number
): string {
  const ground = template.ground[y][x];
  const bridge = template.bridge[y][x];
  const pipeline = template.pipeline[y][x];
  const rail = template.rail[y][x];
  const static_ = template.static[y][x];
  const chaser = template.chaser[y][x];
  const zoner = template.zoner[y][x];

  switch (layer) {
    case 'softEdge':
      if (ground === 1) return 'Soft edge cannot overlap with ground';
      return 'Soft edge has no ground anchor';
    case 'bridge':
      if (ground === 1) return 'Bridge cannot be placed on walkable ground';
      return 'Bridge must connect walkable areas';
    case 'pipeline':
      if (ground === 0) return 'Pipeline must be placed on ground';
      if (bridge === 1) return 'Pipeline cannot be placed on bridge';
      return 'Unknown error';
    case 'rail':
      if (ground === 0 && bridge === 0) return 'Rail must be placed on ground or bridge';
      const railNeighbors = countRailNeighbors(template, x, y);
      if (railNeighbors > 2) return `Rail segments cannot intersect (has ${railNeighbors} neighbors, max 2)`;
      return 'Unknown error';
    case 'static':
      if (ground === 0 && bridge === 0) return 'Static items require walkable ground or bridge';
      if (bridge === 1) return 'Static items cannot be placed on bridge';
      if (pipeline === 1) return 'Static items cannot be placed on pipeline';
      if (rail === 1) return 'Static items cannot be placed on rail';
      return 'Unknown error';
    case 'chaser':
      if (ground === 0) return 'Chasers require walkable ground';
      if (static_ === 1) return 'Chasers cannot be placed on static items';
      if (bridge === 1) return 'Chasers cannot be placed on bridge';
      if (pipeline === 1) return 'Chasers cannot be placed on pipeline';
      if (rail === 1) return 'Chasers cannot be placed on rail';
      if (zoner === 1) return 'Chasers cannot be placed on zoner';
      return 'Unknown error';
    case 'zoner':
      if (ground === 0) return 'Zoners require walkable ground';
      if (static_ === 1) return 'Zoners cannot be placed on static items';
      if (bridge === 1) return 'Zoners cannot be placed on bridge';
      if (pipeline === 1) return 'Zoners cannot be placed on pipeline';
      if (rail === 1) return 'Zoners cannot be placed on rail';
      if (chaser === 1) return 'Zoners cannot be placed on chaser';
      return 'Unknown error';
    case 'dps':
      if (ground === 0) return 'DPS requires walkable ground';
      if (static_ === 1) return 'DPS cannot be placed on static items';
      if (bridge === 1) return 'DPS cannot be placed on bridge';
      if (pipeline === 1) return 'DPS cannot be placed on pipeline';
      if (rail === 1) return 'DPS cannot be placed on rail';
      if (zoner === 1) return 'DPS cannot be placed on zoner';
      return 'Unknown error';
    default:
      return 'Unknown validation error';
  }
}

