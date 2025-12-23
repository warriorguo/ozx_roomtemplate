import type { 
  Template, 
  LayerType, 
  CellValue, 
  Grid, 
  ValidationResult, 
  ValidationError, 
  LayerValidation,
  RoomSpec,
  GenerationResult
} from '../types/newTemplate';

export function createEmptyTemplate(width: number, height: number): Template {
  if (width <= 0 || height <= 0 || width > 200 || height > 200) {
    throw new Error('Width and height must be between 1 and 200');
  }

  const createLayer = (): Grid<CellValue> => 
    Array(height).fill(null).map(() => Array(width).fill(0));

  return {
    version: 1,
    width,
    height,
    ground: createLayer(),
    static: createLayer(),
    turret: createLayer(),
    mobGround: createLayer(),
    mobAir: createLayer(),
  };
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

  return newTemplate;
}

// Rule-based validation functions
export function validateCellRules(
  template: Template, 
  x: number, 
  y: number
): Record<LayerType, boolean> {
  const ground = template.ground[y][x];
  const static_ = template.static[y][x];
  const turret = template.turret[y][x];
  const mobGround = template.mobGround[y][x];
  // const mobAir = template.mobAir[y][x]; // Not used in validation

  return {
    ground: true, // Ground has no constraints
    static: static_ === 0 || ground === 1,
    turret: turret === 0 || (ground === 1 && static_ === 0),
    mobGround: mobGround === 0 || (ground === 1 && static_ === 0 && turret === 0),
    mobAir: true, // MobAir has no constraints
  };
}

export function validateTemplate(template: Template): ValidationResult {
  const errors: ValidationError[] = [];
  const layerValidation: LayerValidation = {
    ground: [],
    static: [],
    turret: [],
    mobGround: [],
    mobAir: [],
  };

  // Initialize validation grids
  for (let y = 0; y < template.height; y++) {
    layerValidation.ground[y] = [];
    layerValidation.static[y] = [];
    layerValidation.turret[y] = [];
    layerValidation.mobGround[y] = [];
    layerValidation.mobAir[y] = [];
    
    for (let x = 0; x < template.width; x++) {
      const cellValidation = validateCellRules(template, x, y);
      
      // Store validation results
      layerValidation.ground[y][x] = cellValidation.ground;
      layerValidation.static[y][x] = cellValidation.static;
      layerValidation.turret[y][x] = cellValidation.turret;
      layerValidation.mobGround[y][x] = cellValidation.mobGround;
      layerValidation.mobAir[y][x] = cellValidation.mobAir;

      // Collect errors for cells that have value=1 but are invalid
      const layers: LayerType[] = ['static', 'turret', 'mobGround', 'mobAir'];
      
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
  const static_ = template.static[y][x];
  const turret = template.turret[y][x];

  switch (layer) {
    case 'static':
      return ground === 0 ? 'Static items require walkable ground' : 'Unknown error';
    case 'turret':
      if (ground === 0) return 'Turrets require walkable ground';
      if (static_ === 1) return 'Turrets cannot be placed on static items';
      return 'Unknown error';
    case 'mobGround':
      if (ground === 0) return 'Ground mobs require walkable ground';
      if (static_ === 1) return 'Ground mobs cannot be placed on static items';
      if (turret === 1) return 'Ground mobs cannot be placed on turrets';
      return 'Unknown error';
    default:
      return 'Unknown validation error';
  }
}


// Ground auto-generation functions
export function generateGroundFromSpec(spec: RoomSpec): GenerationResult {
  const { width, height, roomType, wallThickness, doorPositions } = spec;
  const ground: Grid<CellValue> = Array(height).fill(null).map(() => Array(width).fill(0));
  const warnings: string[] = [];

  // Generate base room shape
  switch (roomType) {
    case 'rectangular':
      generateRectangularRoom(ground, width, height, wallThickness);
      break;
    case 'cross':
      generateCrossRoom(ground, width, height, wallThickness);
      break;
    case 'custom':
      // For custom rooms, start with empty ground
      break;
  }

  // Add doors
  for (const door of doorPositions) {
    if (isValidDoorPosition(door, width, height)) {
      placeDoor(ground, door, wallThickness);
    } else {
      warnings.push(`Invalid door position: (${door.x}, ${door.y})`);
    }
  }

  return { ground, warnings };
}

function generateRectangularRoom(
  ground: Grid<CellValue>, 
  width: number, 
  height: number, 
  wallThickness: number
): void {
  for (let y = wallThickness; y < height - wallThickness; y++) {
    for (let x = wallThickness; x < width - wallThickness; x++) {
      ground[y][x] = 1;
    }
  }
}

function generateCrossRoom(
  ground: Grid<CellValue>, 
  width: number, 
  height: number, 
  wallThickness: number
): void {
  const centerX = Math.floor(width / 2);
  const centerY = Math.floor(height / 2);
  const armWidth = Math.floor(Math.min(width, height) / 3);

  // Horizontal arm
  for (let y = centerY - Math.floor(armWidth / 2); y <= centerY + Math.floor(armWidth / 2); y++) {
    if (y >= 0 && y < height) {
      for (let x = wallThickness; x < width - wallThickness; x++) {
        ground[y][x] = 1;
      }
    }
  }

  // Vertical arm
  for (let x = centerX - Math.floor(armWidth / 2); x <= centerX + Math.floor(armWidth / 2); x++) {
    if (x >= 0 && x < width) {
      for (let y = wallThickness; y < height - wallThickness; y++) {
        ground[y][x] = 1;
      }
    }
  }
}

function isValidDoorPosition(
  door: { x: number; y: number; direction: string }, 
  width: number, 
  height: number
): boolean {
  const { x, y, direction } = door;
  
  if (x < 0 || x >= width || y < 0 || y >= height) return false;

  switch (direction) {
    case 'north':
      return y === 0;
    case 'south':
      return y === height - 1;
    case 'east':
      return x === width - 1;
    case 'west':
      return x === 0;
    default:
      return false;
  }
}

function placeDoor(
  ground: Grid<CellValue>,
  door: { x: number; y: number; direction: string },
  wallThickness: number
): void {
  const { x, y, direction } = door;

  switch (direction) {
    case 'north':
      for (let dy = 0; dy < wallThickness; dy++) {
        if (y + dy < ground.length) {
          ground[y + dy][x] = 1;
        }
      }
      break;
    case 'south':
      for (let dy = 0; dy < wallThickness; dy++) {
        if (y - dy >= 0) {
          ground[y - dy][x] = 1;
        }
      }
      break;
    case 'east':
      for (let dx = 0; dx < wallThickness; dx++) {
        if (x - dx >= 0) {
          ground[y][x - dx] = 1;
        }
      }
      break;
    case 'west':
      for (let dx = 0; dx < wallThickness; dx++) {
        if (x + dx < ground[0].length) {
          ground[y][x + dx] = 1;
        }
      }
      break;
  }
}