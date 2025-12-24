import type { 
  Template, 
  LayerType, 
  CellValue, 
  Grid, 
  ValidationResult, 
  ValidationError, 
  LayerValidation
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

