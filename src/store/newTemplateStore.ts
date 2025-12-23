import { create } from 'zustand';
import type { 
  Template, 
  LayerType, 
  UIState, 
  RoomSpec
} from '../types/newTemplate';
import { 
  createEmptyTemplate, 
  setCellValue, 
  validateTemplate, 
  generateGroundFromSpec
} from '../utils/newTemplateUtils';
import { 
  templateApi, 
  type BackendTemplate,
  type BackendCreateResponse,
  type BackendListResponse,
  ApiError
} from '../services/api';
import { 
  frontendToBackendCreateRequest,
  backendToFrontendTemplate,
  frontendToBackendPayload,
  validateTemplateName,
  generateDefaultTemplateName
} from '../services/templateConverter';

// API state
export interface ApiState {
  isLoading: boolean;
  error: string | null;
  lastSaved?: {
    id: string;
    name: string;
    savedAt: string;
  };
}

interface NewTemplateStore {
  template: Template;
  uiState: UIState;
  apiState: ApiState;
  
  // Template management
  createNewTemplate: (width: number, height: number) => void;
  loadTemplate: (template: Template) => void;
  
  // Cell editing
  setCellValue: (layer: LayerType, x: number, y: number, value: 0 | 1) => void;
  toggleCell: (layer: LayerType, x: number, y: number) => void;
  
  // Layer management
  setActiveLayer: (layer: LayerType) => void;
  
  // Drag operations
  startDrag: (layer: LayerType, x: number, y: number) => void;
  dragToCell: (layer: LayerType, x: number, y: number) => void;
  endDrag: () => void;
  
  // UI state
  setHoveredCell: (x: number, y: number) => void;
  clearHoveredCell: () => void;
  toggleLayerVisibility: (layer: LayerType) => void;
  toggleErrorDisplay: () => void;
  
  // Validation
  validateTemplate: () => void;
  
  // Ground generation
  generateGround: (spec: RoomSpec) => void;
  
  // API operations
  saveTemplate: (name: string) => Promise<void>;
  loadTemplateFromBackend: (id: string) => Promise<void>;
  validateTemplateWithBackend: (strict?: boolean) => Promise<void>;
  clearApiError: () => void;
}

export const useNewTemplateStore = create<NewTemplateStore>((set, get) => ({
  template: createEmptyTemplate(20, 12),
  uiState: {
    activeLayer: 'ground',
    dragState: {
      isDragging: false,
      activeLayer: 'ground',
      dragMode: null,
      lastProcessedCell: null,
    },
    hoveredCell: null,
    layerVisibility: {
      ground: true,
      static: true,
      turret: true,
      mobGround: true,
      mobAir: true,
    },
    validationResult: null,
    showErrors: true,
  },
  apiState: {
    isLoading: false,
    error: null,
  },

  createNewTemplate: (width: number, height: number) => {
    const newTemplate = createEmptyTemplate(width, height);
    const validation = validateTemplate(newTemplate);
    
    set({
      template: newTemplate,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
  },

  loadTemplate: (template: Template) => {
    const validation = validateTemplate(template);
    
    set({
      template,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
  },

  setCellValue: (layer: LayerType, x: number, y: number, value: 0 | 1) => {
    const { template } = get();
    const newTemplate = setCellValue(template, layer, x, y, value);
    const validation = validateTemplate(newTemplate);
    
    set({
      template: newTemplate,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
  },

  toggleCell: (layer: LayerType, x: number, y: number) => {
    const { template } = get();
    const currentValue = template[layer][y]?.[x];
    if (currentValue === undefined) return;
    
    const newValue = currentValue === 0 ? 1 : 0;
    const newTemplate = setCellValue(template, layer, x, y, newValue);
    const validation = validateTemplate(newTemplate);
    
    set({
      template: newTemplate,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
  },

  setActiveLayer: (layer: LayerType) => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        activeLayer: layer,
      },
    }));
  },

  startDrag: (layer: LayerType, x: number, y: number) => {
    const { uiState, template } = get();
    
    if (uiState.activeLayer !== layer) return;
    
    const currentValue = template[layer][y]?.[x];
    if (currentValue === undefined) return;
    
    // Determine drag mode based on current cell value
    const dragMode = currentValue === 0 ? 'set' : 'clear';
    
    // Toggle the starting cell
    get().toggleCell(layer, x, y);
    
    set((state) => ({
      uiState: {
        ...state.uiState,
        dragState: {
          isDragging: true,
          activeLayer: layer,
          dragMode,
          lastProcessedCell: { x, y },
        },
      },
    }));
  },

  dragToCell: (layer: LayerType, x: number, y: number) => {
    const { uiState, template } = get();
    
    if (!uiState.dragState.isDragging || uiState.dragState.activeLayer !== layer) return;
    
    // Check if this cell was already processed
    const lastCell = uiState.dragState.lastProcessedCell;
    if (lastCell && lastCell.x === x && lastCell.y === y) return;
    
    // Apply drag mode to the cell
    const targetValue = uiState.dragState.dragMode === 'set' ? 1 : 0;
    const currentValue = template[layer][y]?.[x];
    
    // Only update if the value would actually change
    if (currentValue !== undefined && currentValue !== targetValue) {
      get().setCellValue(layer, x, y, targetValue);
    }
    
    set((state) => ({
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
          ...state.uiState.dragState,
          isDragging: false,
          dragMode: null,
          lastProcessedCell: null,
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

  toggleLayerVisibility: (layer: LayerType) => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        layerVisibility: {
          ...state.uiState.layerVisibility,
          [layer]: !state.uiState.layerVisibility[layer],
        },
      },
    }));
  },

  toggleErrorDisplay: () => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        showErrors: !state.uiState.showErrors,
      },
    }));
  },

  validateTemplate: () => {
    const { template } = get();
    const validation = validateTemplate(template);
    
    set((state) => ({
      uiState: {
        ...state.uiState,
        validationResult: validation,
      },
    }));
  },

  generateGround: (spec: RoomSpec) => {
    const { template } = get();
    const { ground, warnings } = generateGroundFromSpec(spec);
    
    const newTemplate = {
      ...template,
      ground,
    };
    
    const validation = validateTemplate(newTemplate);
    
    set({
      template: newTemplate,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
    
    // Show warnings if any
    if (warnings.length > 0) {
      console.warn('Ground generation warnings:', warnings);
    }
  },

  // API operations
  saveTemplate: async (name: string) => {
    const { template } = get();
    
    // Validate name
    const nameValidation = validateTemplateName(name);
    if (!nameValidation.valid) {
      set((state) => ({
        apiState: {
          ...state.apiState,
          error: nameValidation.error || 'Invalid template name',
        },
      }));
      return;
    }

    // Set loading state
    set((state) => ({
      apiState: {
        ...state.apiState,
        isLoading: true,
        error: null,
      },
    }));

    try {
      const request = frontendToBackendCreateRequest(template, name);
      const response = await templateApi.createTemplate(request);
      
      set((state) => ({
        apiState: {
          ...state.apiState,
          isLoading: false,
          lastSaved: {
            id: response.id,
            name: response.name,
            savedAt: response.created_at,
          },
        },
      }));
    } catch (error) {
      const errorMessage = error instanceof ApiError 
        ? error.message 
        : 'Failed to save template';
      
      set((state) => ({
        apiState: {
          ...state.apiState,
          isLoading: false,
          error: errorMessage,
        },
      }));
    }
  },

  loadTemplateFromBackend: async (id: string) => {
    set((state) => ({
      apiState: {
        ...state.apiState,
        isLoading: true,
        error: null,
      },
    }));

    try {
      const backendTemplate = await templateApi.getTemplate(id);
      const frontendTemplate = backendToFrontendTemplate(backendTemplate);
      const validation = validateTemplate(frontendTemplate);
      
      set({
        template: frontendTemplate,
        uiState: {
          ...get().uiState,
          validationResult: validation,
        },
        apiState: {
          isLoading: false,
          error: null,
          lastSaved: {
            id: backendTemplate.id,
            name: backendTemplate.name,
            savedAt: backendTemplate.updated_at,
          },
        },
      });
    } catch (error) {
      const errorMessage = error instanceof ApiError 
        ? error.message 
        : 'Failed to load template';
      
      set((state) => ({
        apiState: {
          ...state.apiState,
          isLoading: false,
          error: errorMessage,
        },
      }));
    }
  },

  validateTemplateWithBackend: async (strict = false) => {
    const { template } = get();
    
    set((state) => ({
      apiState: {
        ...state.apiState,
        isLoading: true,
        error: null,
      },
    }));

    try {
      const payload = frontendToBackendPayload(template, 'validation-template');
      const result = await templateApi.validateTemplate(payload, strict);
      
      // Convert backend validation result to frontend format
      const validationResult = {
        isValid: result.valid,
        errors: result.errors.map(error => ({
          layer: error.layer as LayerType,
          x: error.x,
          y: error.y,
          reason: error.reason,
        })),
        layerValidation: {
          ground: Array(template.height).fill(null).map(() => Array(template.width).fill(true)),
          static: Array(template.height).fill(null).map(() => Array(template.width).fill(true)),
          turret: Array(template.height).fill(null).map(() => Array(template.width).fill(true)),
          mobGround: Array(template.height).fill(null).map(() => Array(template.width).fill(true)),
          mobAir: Array(template.height).fill(null).map(() => Array(template.width).fill(true)),
        },
      };

      // Mark invalid cells in layer validation
      result.errors.forEach(error => {
        if (error.layer in validationResult.layerValidation) {
          validationResult.layerValidation[error.layer as LayerType][error.y][error.x] = false;
        }
      });
      
      set({
        uiState: {
          ...get().uiState,
          validationResult,
        },
        apiState: {
          ...get().apiState,
          isLoading: false,
        },
      });
    } catch (error) {
      const errorMessage = error instanceof ApiError 
        ? error.message 
        : 'Failed to validate template';
      
      set((state) => ({
        apiState: {
          ...state.apiState,
          isLoading: false,
          error: errorMessage,
        },
      }));
    }
  },

  clearApiError: () => {
    set((state) => ({
      apiState: {
        ...state.apiState,
        error: null,
      },
    }));
  },
}));