import type { Template, Grid, GroundCell, StaticCell, MonsterCell, ValidationResult, ValidationError } from '../types/template';

export function createEmptyTemplate(width: number, height: number): Template {
  if (width <= 0 || height <= 0 || width > 200 || height > 200) {
    throw new Error('Width and height must be between 1 and 200');
  }

  const ground: Grid<GroundCell> = Array(height).fill(null).map(() => Array(width).fill(0));
  const staticLayer: Grid<StaticCell> = Array(height).fill(null).map(() => Array(width).fill(0));
  const monster: Grid<MonsterCell> = Array(height).fill(null).map(() => Array(width).fill(0));

  return {
    version: 1,
    width,
    height,
    ground,
    static: staticLayer,
    monster,
  };
}

export function toggleGround(template: Template, x: number, y: number): Template {
  if (x < 0 || x >= template.width || y < 0 || y >= template.height) {
    return template;
  }

  const newTemplate = JSON.parse(JSON.stringify(template)) as Template;
  const currentValue = newTemplate.ground[y][x];
  newTemplate.ground[y][x] = currentValue === 0 ? 1 : 0;

  if (newTemplate.ground[y][x] === 0) {
    newTemplate.static[y][x] = 0;
    if (newTemplate.monster[y][x] === 1) {
      newTemplate.monster[y][x] = 0;
    }
  }

  return newTemplate;
}

export function toggleStatic(template: Template, x: number, y: number): Template {
  if (x < 0 || x >= template.width || y < 0 || y >= template.height) {
    return template;
  }

  if (template.ground[y][x] === 0) {
    return template;
  }

  const newTemplate = JSON.parse(JSON.stringify(template)) as Template;
  const currentValue = newTemplate.static[y][x];
  newTemplate.static[y][x] = currentValue === 0 ? 1 : 0;

  return newTemplate;
}

export function toggleMonster(template: Template, x: number, y: number): Template {
  if (x < 0 || x >= template.width || y < 0 || y >= template.height) {
    return template;
  }

  const newTemplate = JSON.parse(JSON.stringify(template)) as Template;
  const currentValue = newTemplate.monster[y][x];
  const isGround = template.ground[y][x] === 1;

  if (isGround) {
    switch (currentValue) {
      case 0:
        newTemplate.monster[y][x] = 1;
        break;
      case 1:
        newTemplate.monster[y][x] = 2;
        break;
      case 2:
        newTemplate.monster[y][x] = 0;
        break;
    }
  } else {
    switch (currentValue) {
      case 0:
        newTemplate.monster[y][x] = 2;
        break;
      case 2:
        newTemplate.monster[y][x] = 0;
        break;
      default:
        newTemplate.monster[y][x] = 0;
    }
  }

  return newTemplate;
}

export function setGroundValue(template: Template, x: number, y: number, value: GroundCell): Template {
  if (x < 0 || x >= template.width || y < 0 || y >= template.height) {
    return template;
  }

  if (template.ground[y][x] === value) {
    return template;
  }

  const newTemplate = JSON.parse(JSON.stringify(template)) as Template;
  newTemplate.ground[y][x] = value;

  if (value === 0) {
    newTemplate.static[y][x] = 0;
    if (newTemplate.monster[y][x] === 1) {
      newTemplate.monster[y][x] = 0;
    }
  }

  return newTemplate;
}

export function validateTemplate(template: any): ValidationResult {
  const errors: ValidationError[] = [];

  if (!template || typeof template !== 'object') {
    return {
      isValid: false,
      errors: [{ layer: 'ground', x: 0, y: 0, message: 'Invalid template format' }],
      corrected: null,
    };
  }

  const { version, width, height, ground, static: staticLayer, monster } = template;

  if (version !== 1) {
    errors.push({ layer: 'ground', x: 0, y: 0, message: 'Unsupported version' });
  }

  if (!Number.isInteger(width) || !Number.isInteger(height) || width <= 0 || height <= 0 || width > 200 || height > 200) {
    errors.push({ layer: 'ground', x: 0, y: 0, message: 'Invalid width or height' });
  }

  if (!Array.isArray(ground) || ground.length !== height) {
    errors.push({ layer: 'ground', x: 0, y: 0, message: 'Invalid ground layer dimensions' });
  }

  if (!Array.isArray(staticLayer) || staticLayer.length !== height) {
    errors.push({ layer: 'static', x: 0, y: 0, message: 'Invalid static layer dimensions' });
  }

  if (!Array.isArray(monster) || monster.length !== height) {
    errors.push({ layer: 'monster', x: 0, y: 0, message: 'Invalid monster layer dimensions' });
  }

  if (errors.length > 0) {
    return { isValid: false, errors, corrected: null };
  }

  for (let y = 0; y < height; y++) {
    if (!Array.isArray(ground[y]) || ground[y].length !== width) {
      errors.push({ layer: 'ground', x: 0, y, message: `Invalid ground row ${y} length` });
      continue;
    }
    if (!Array.isArray(staticLayer[y]) || staticLayer[y].length !== width) {
      errors.push({ layer: 'static', x: 0, y, message: `Invalid static row ${y} length` });
      continue;
    }
    if (!Array.isArray(monster[y]) || monster[y].length !== width) {
      errors.push({ layer: 'monster', x: 0, y, message: `Invalid monster row ${y} length` });
      continue;
    }

    for (let x = 0; x < width; x++) {
      if (![0, 1].includes(ground[y][x])) {
        errors.push({ layer: 'ground', x, y, message: `Invalid ground value at (${x}, ${y})` });
      }
      if (![0, 1].includes(staticLayer[y][x])) {
        errors.push({ layer: 'static', x, y, message: `Invalid static value at (${x}, ${y})` });
      }
      if (![0, 1, 2].includes(monster[y][x])) {
        errors.push({ layer: 'monster', x, y, message: `Invalid monster value at (${x}, ${y})` });
      }
    }
  }

  return { isValid: errors.length === 0, errors, corrected: null };
}

export function sanitizeTemplate(template: any): Template {
  const validation = validateTemplate(template);
  
  if (!validation.isValid && validation.errors.some(e => e.message.includes('Invalid template format') || e.message.includes('dimensions'))) {
    throw new Error('Template cannot be sanitized due to structural issues');
  }

  const { width, height } = template;
  const corrected = createEmptyTemplate(width, height);

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (template.ground?.[y]?.[x] === 1) {
        corrected.ground[y][x] = 1;
      }

      if (template.static?.[y]?.[x] === 1 && corrected.ground[y][x] === 1) {
        corrected.static[y][x] = 1;
      }

      const monsterValue = template.monster?.[y]?.[x];
      if (monsterValue === 1 && corrected.ground[y][x] === 1) {
        corrected.monster[y][x] = 1;
      } else if (monsterValue === 2) {
        corrected.monster[y][x] = 2;
      }
    }
  }

  return corrected;
}