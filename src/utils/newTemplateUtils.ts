import type {
  Template,
  LayerType,
  CellValue,
  Grid,
  ValidationResult,
  ValidationError,
  LayerValidation,
  DoorStates
} from '../types/newTemplate';
import { calculateAllTileProperties } from './tilePropertiesCalculator';

/**
 * 计算门的开通状态
 * 如果门对应的两个格子在 ground 层都为 1，则该门视为开通
 */
export function calculateDoorStates(template: Template): DoorStates {
  const { width, height, ground } = template;

  // 计算中间位置
  const midWidth = Math.floor(width / 2);
  const midHeight = Math.floor(height / 2);

  // 检查顶部门（y=0，中间两格）
  const topOpen =
    ground[0]?.[midWidth - 1] === 1 &&
    ground[0]?.[midWidth] === 1 ? 1 : 0;

  // 检查底部门（y=height-1，中间两格）
  const bottomOpen =
    ground[height - 1]?.[midWidth - 1] === 1 &&
    ground[height - 1]?.[midWidth] === 1 ? 1 : 0;

  // 检查左侧门（x=0，中间两格）
  const leftOpen =
    ground[midHeight - 1]?.[0] === 1 &&
    ground[midHeight]?.[0] === 1 ? 1 : 0;

  // 检查右侧门（x=width-1，中间两格）
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
    turret: createLayer(),
    mobGround: createLayer(),
    mobAir: createLayer(),
    doors: { top: 0, right: 0, bottom: 0, left: 0 },
    attributes: {
      boss: false,
      elite: false,
      mob: false,
      treasure: false,
      teleport: false,
      story: false,
    },
    roomType: 'full',
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
  y: number
): Record<LayerType, boolean> {
  const ground = template.ground[y][x];
  const bridge = template.bridge[y][x];
  const pipeline = template.pipeline[y][x];
  const rail = template.rail[y][x];
  const static_ = template.static[y][x];
  const turret = template.turret[y][x];
  const mobGround = template.mobGround[y][x];
  // const mobAir = template.mobAir[y][x]; // Not used in validation

  return {
    ground: true, // Ground has no constraints
    softEdge: validateSoftEdgeCell(template, x, y),
    bridge: validateBridgeCell(template, x, y),
    // Pipeline: must be on ground, cannot be on bridge
    pipeline: pipeline === 0 || (ground === 1 && bridge === 0),
    // Rail: must be on ground or bridge
    rail: rail === 0 || (ground === 1 || bridge === 1),
    // Static: can't be on bridge, can't conflict with pipeline or rail
    static: static_ === 0 || ((ground === 1 || bridge === 1) && bridge === 0 && pipeline === 0 && rail === 0),
    // Turret: can't be on bridge, can't be on static, can't conflict with pipeline or rail
    turret: turret === 0 || ((ground === 1 || bridge === 1) && static_ === 0 && bridge === 0 && pipeline === 0 && rail === 0),
    // MobGround: can't be on bridge, can't be on static/turret, can't conflict with pipeline or rail
    mobGround: mobGround === 0 || ((ground === 1 || bridge === 1) && static_ === 0 && turret === 0 && bridge === 0 && pipeline === 0 && rail === 0),
    mobAir: true, // MobAir has no constraints
  };
}

// Validate soft edge placement: must be adjacent to ground but not overlap with ground
function validateSoftEdgeCell(template: Template, x: number, y: number): boolean {
  const softEdge = template.softEdge[y][x];
  if (softEdge === 0) return true; // Empty soft edge cells are always valid
  
  const ground = template.ground[y][x];
  
  // Soft edge cannot overlap with ground
  if (ground === 1) return false;
  
  // Check if soft edge is adjacent to at least one ground tile
  const directions = [
    { dx: -1, dy: 0 }, // left
    { dx: 1, dy: 0 },  // right
    { dx: 0, dy: -1 }, // up
    { dx: 0, dy: 1 }   // down
  ];
  
  for (const dir of directions) {
    const nx = x + dir.dx;
    const ny = y + dir.dy;
    
    // Check bounds
    if (nx >= 0 && nx < template.width && ny >= 0 && ny < template.height) {
      if (template.ground[ny][nx] === 1) {
        return true; // Found adjacent ground tile
      }
    }
  }
  
  return false; // No adjacent ground tile found
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
  const layerValidation: LayerValidation = {
    ground: [],
    softEdge: [],
    bridge: [],
    pipeline: [],
    rail: [],
    static: [],
    turret: [],
    mobGround: [],
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
    layerValidation.turret[y] = [];
    layerValidation.mobGround[y] = [];
    layerValidation.mobAir[y] = [];

    for (let x = 0; x < template.width; x++) {
      const cellValidation = validateCellRules(template, x, y);

      // Store validation results
      layerValidation.ground[y][x] = cellValidation.ground;
      layerValidation.softEdge[y][x] = cellValidation.softEdge;
      layerValidation.bridge[y][x] = cellValidation.bridge;
      layerValidation.pipeline[y][x] = cellValidation.pipeline;
      layerValidation.rail[y][x] = cellValidation.rail;
      layerValidation.static[y][x] = cellValidation.static;
      layerValidation.turret[y][x] = cellValidation.turret;
      layerValidation.mobGround[y][x] = cellValidation.mobGround;
      layerValidation.mobAir[y][x] = cellValidation.mobAir;

      // Collect errors for cells that have value=1 but are invalid
      const layers: LayerType[] = ['softEdge', 'bridge', 'pipeline', 'rail', 'static', 'turret', 'mobGround', 'mobAir'];

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
  const turret = template.turret[y][x];

  switch (layer) {
    case 'softEdge':
      if (ground === 1) return 'Soft edge cannot overlap with ground';
      return 'Soft edge must be adjacent to ground';
    case 'bridge':
      if (ground === 1) return 'Bridge cannot be placed on walkable ground';
      return 'Bridge must connect walkable areas';
    case 'pipeline':
      if (ground === 0) return 'Pipeline must be placed on ground';
      if (bridge === 1) return 'Pipeline cannot be placed on bridge';
      return 'Unknown error';
    case 'rail':
      if (ground === 0 && bridge === 0) return 'Rail must be placed on ground or bridge';
      return 'Unknown error';
    case 'static':
      if (ground === 0 && bridge === 0) return 'Static items require walkable ground or bridge';
      if (bridge === 1) return 'Static items cannot be placed on bridge';
      if (pipeline === 1) return 'Static items cannot be placed on pipeline';
      if (rail === 1) return 'Static items cannot be placed on rail';
      return 'Unknown error';
    case 'turret':
      if (ground === 0 && bridge === 0) return 'Turrets require walkable ground or bridge';
      if (bridge === 1) return 'Turrets cannot be placed on bridge';
      if (static_ === 1) return 'Turrets cannot be placed on static items';
      if (pipeline === 1) return 'Turrets cannot be placed on pipeline';
      if (rail === 1) return 'Turrets cannot be placed on rail';
      return 'Unknown error';
    case 'mobGround':
      if (ground === 0 && bridge === 0) return 'Ground mobs require walkable ground or bridge';
      if (bridge === 1) return 'Ground mobs cannot be placed on bridge';
      if (static_ === 1) return 'Ground mobs cannot be placed on static items';
      if (turret === 1) return 'Ground mobs cannot be placed on turrets';
      if (pipeline === 1) return 'Ground mobs cannot be placed on pipeline';
      if (rail === 1) return 'Ground mobs cannot be placed on rail';
      return 'Unknown error';
    default:
      return 'Unknown validation error';
  }
}

