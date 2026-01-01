// New 5-layer template system
export type CellValue = 0 | 1;
export type Grid<T> = T[][];

export interface DoorStates {
  top: 0 | 1;
  right: 0 | 1;
  bottom: 0 | 1;
  left: 0 | 1;
}

export type RoomAttribute = 'boss' | 'elite' | 'mob' | 'treasure' | 'teleport' | 'story';

export interface RoomAttributes {
  boss: boolean;
  elite: boolean;
  mob: boolean;
  treasure: boolean;
  teleport: boolean;
  story: boolean;
}

export type RoomType = 'full' | 'bridge' | 'platform';

export interface RoomTypeInfo {
  type: RoomType;
  label: string;
  description: string;
}

export const ROOM_TYPES: RoomTypeInfo[] = [
  { type: 'full', label: '全房间', description: '房间的ground tile几乎全部点开' },
  { type: 'bridge', label: '桥梁', description: '门到门连接是桥梁关系' },
  { type: 'platform', label: '平台', description: '有大块的平台和edge' },
];

export interface TileProperties {
  walkable: boolean;
  distToTopWall: number;
  distToBottomWall: number;
  distToLeftWall: number;
  distToRightWall: number;
  distToCenter: number;
  distToEdge: number;                // distance to unwalkable cell (ground=0) or room edge
  distToTopDoor: number | null;      // null if door not open
  distToBottomDoor: number | null;   // null if door not open
  distToLeftDoor: number | null;     // null if door not open
  distToRightDoor: number | null;    // null if door not open
  distToNearStatic: number | null;   // null if no static tiles
  distToNearTurret: number | null;   // null if no turret tiles
}

export interface Template {
  version: 1;
  width: number;
  height: number;
  ground: Grid<CellValue>;
  bridge: Grid<CellValue>;
  static: Grid<CellValue>;
  turret: Grid<CellValue>;
  mobGround: Grid<CellValue>;
  mobAir: Grid<CellValue>;
  doors: DoorStates;
  attributes: RoomAttributes;
  roomType: RoomType;
  tileProperties: Grid<TileProperties | null>; // null if cell is not set (value=0)
}

export type LayerType = "ground" | "bridge" | "static" | "turret" | "mobGround" | "mobAir";

export interface DragState {
  isDragging: boolean;
  dragLayer: LayerType | null;
  dragMode: 'set' | 'clear' | 'brush' | null;
  lastProcessedCell: { x: number; y: number } | null;
  brushTargetValue?: 0 | 1; // 笔刷模式下的目标值
  justEndedDrag: boolean; // 标记是否刚刚结束拖拽，防止onClick重复触发
}

export interface LayerValidation {
  ground: boolean[][];
  bridge: boolean[][];
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
  showCompositeView: boolean;  // 是否显示总图层视图（仅在mobAir层）
  acceptPaste: boolean;  // 是否接受paste操作
}

