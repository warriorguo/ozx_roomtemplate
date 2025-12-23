// New 5-layer template system
export type CellValue = 0 | 1;
export type Grid<T> = T[][];

export interface Template {
  version: 1;
  width: number;
  height: number;
  ground: Grid<CellValue>;
  static: Grid<CellValue>;
  turret: Grid<CellValue>;
  mobGround: Grid<CellValue>;
  mobAir: Grid<CellValue>;
}

export type LayerType = "ground" | "static" | "turret" | "mobGround" | "mobAir";

export interface DragState {
  isDragging: boolean;
  activeLayer: LayerType;
  dragMode: 'set' | 'clear' | null;
  lastProcessedCell: { x: number; y: number } | null;
}

export interface LayerValidation {
  ground: boolean[][];
  static: boolean[][];
  turret: boolean[][];
  mobGround: boolean[][];
  mobAir: boolean[][];
}

export interface ValidationError {
  layer: LayerType;
  x: number;
  y: number;
  reason: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
  layerValidation: LayerValidation;
}

export interface UIState {
  activeLayer: LayerType;
  dragState: DragState;
  hoveredCell: { x: number; y: number } | null;
  layerVisibility: Record<LayerType, boolean>;
  validationResult: ValidationResult | null;
  showErrors: boolean;
}

// Ground auto-generation types
export interface RoomSpec {
  width: number;
  height: number;
  roomType: "rectangular" | "cross" | "custom";
  wallThickness: number;
  doorPositions: Array<{ x: number; y: number; direction: "north" | "south" | "east" | "west" }>;
}

export interface GenerationResult {
  ground: Grid<CellValue>;
  warnings: string[];
}