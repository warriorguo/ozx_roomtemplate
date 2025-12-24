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
  dragLayer: LayerType | null;
  dragMode: 'set' | 'clear' | 'brush' | null;
  lastProcessedCell: { x: number; y: number } | null;
  brushTargetValue?: 0 | 1; // 笔刷模式下的目标值
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

export interface BrushSize {
  width: number;
  height: number;
}

export interface UIState {
  dragState: DragState;
  hoveredCell: { x: number; y: number } | null;
  layerVisibility: Record<LayerType, boolean>;
  validationResult: ValidationResult | null;
  showErrors: boolean;
  brushSize: BrushSize;
  brushPreview: {
    layer: LayerType | null;
    x: number;
    y: number;
    visible: boolean;
  };
}

