import { create } from 'zustand';
import type { Template, LayerType, UIState } from '../types/template';
import { createEmptyTemplate, toggleGround, toggleStatic, toggleMonster, setGroundValue } from '../utils/templateUtils';

interface TemplateStore {
  template: Template;
  uiState: UIState;
  
  createNewTemplate: (width: number, height: number) => void;
  loadTemplate: (template: Template) => void;
  toggleCell: (x: number, y: number, layer: LayerType) => void;
  setActiveLayer: (layer: LayerType) => void;
  toggleLayerVisibility: (layer: LayerType) => void;
  setHoveredCell: (x: number, y: number) => void;
  clearHoveredCell: () => void;
  
  // Drag operations (for ground layer only)
  startDrag: (x: number, y: number) => void;
  dragToCell: (x: number, y: number) => void;
  endDrag: () => void;
}

export const useTemplateStore = create<TemplateStore>((set, get) => ({
  template: createEmptyTemplate(15, 11),
  uiState: {
    activeLayer: 'ground',
    visible: {
      ground: true,
      static: true,
      monster: true,
    },
    hoveredCell: null,
    dragState: {
      isDragging: false,
      dragMode: null,
      lastProcessedCell: null,
    },
  },

  createNewTemplate: (width: number, height: number) => {
    set({
      template: createEmptyTemplate(width, height),
    });
  },

  loadTemplate: (template: Template) => {
    set({ template });
  },

  toggleCell: (x: number, y: number, layer: LayerType) => {
    const { template } = get();
    let newTemplate: Template;

    switch (layer) {
      case 'ground':
        newTemplate = toggleGround(template, x, y);
        break;
      case 'static':
        newTemplate = toggleStatic(template, x, y);
        break;
      case 'monster':
        newTemplate = toggleMonster(template, x, y);
        break;
      default:
        return;
    }

    set({ template: newTemplate });
  },

  setActiveLayer: (layer: LayerType) => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        activeLayer: layer,
      },
    }));
  },

  toggleLayerVisibility: (layer: LayerType) => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        visible: {
          ...state.uiState.visible,
          [layer]: !state.uiState.visible[layer],
        },
      },
    }));
  },

  setHoveredCell: (x: number, y: number) => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        hoveredCell: { x, y },
      },
    }));
  },

  clearHoveredCell: () => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        hoveredCell: null,
      },
    }));
  },

  startDrag: (x: number, y: number) => {
    const { template, uiState } = get();
    
    // Only allow dragging on ground layer
    if (uiState.activeLayer !== 'ground') return;
    
    const currentValue = template.ground[y]?.[x];
    if (currentValue === undefined) return;
    
    const dragMode = currentValue === 0 ? 'set' : 'clear';
    const newTemplate = setGroundValue(template, x, y, dragMode === 'set' ? 1 : 0);
    
    set((state) => ({
      template: newTemplate,
      uiState: {
        ...state.uiState,
        dragState: {
          isDragging: true,
          dragMode,
          lastProcessedCell: { x, y },
        },
      },
    }));
  },

  dragToCell: (x: number, y: number) => {
    const { template, uiState } = get();
    
    if (!uiState.dragState.isDragging || !uiState.dragState.dragMode) return;
    
    // Check if this cell was already processed
    const lastCell = uiState.dragState.lastProcessedCell;
    if (lastCell && lastCell.x === x && lastCell.y === y) return;
    
    const targetValue = uiState.dragState.dragMode === 'set' ? 1 : 0;
    const newTemplate = setGroundValue(template, x, y, targetValue);
    
    set((state) => ({
      template: newTemplate,
      uiState: {
        ...state.uiState,
        dragState: {
          ...state.uiState.dragState,
          lastProcessedCell: { x, y },
        },
      },
    }));
  },

  endDrag: () => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        dragState: {
          isDragging: false,
          dragMode: null,
          lastProcessedCell: null,
        },
      },
    }));
  },
}));