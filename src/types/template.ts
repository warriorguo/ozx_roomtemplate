export type GroundCell = 0 | 1;
export type StaticCell = 0 | 1;
export type MonsterCell = 0 | 1 | 2;

export type Grid<T> = T[][];

export interface Template {
  version: 1;
  width: number;
  height: number;
  ground: Grid<GroundCell>;
  static: Grid<StaticCell>;
  monster: Grid<MonsterCell>;
}

export type LayerType = "ground" | "static" | "monster";

export interface DragState {
  isDragging: boolean;
  dragMode: 'set' | 'clear' | null;
  lastProcessedCell: { x: number; y: number } | null;
}

export interface UIState {
  activeLayer: LayerType;
  visible: {
    ground: boolean;
    static: boolean;
    monster: boolean;
  };
  hoveredCell: { x: number; y: number } | null;
  dragState: DragState;
}

export interface ValidationError {
  layer: LayerType;
  x: number;
  y: number;
  message: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
  corrected: Template | null;
}