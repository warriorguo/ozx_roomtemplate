import { create } from 'zustand';
import type {
  Template,
  LayerType,
  UIState,
  RoomAttribute,
  RoomType
} from '../types/newTemplate';
import {
  createEmptyTemplate,
  setCellValue,
  validateTemplate,
  calculateDoorStates
} from '../utils/newTemplateUtils';
import { calculateAllTileProperties } from '../utils/tilePropertiesCalculator';
import {
  templateApi,
  ApiError
} from '../services/api';
import {
  frontendToBackendCreateRequest,
  backendToFrontendTemplate,
  frontendToBackendPayload,
  validateTemplateName
} from '../services/templateConverter';
import { generateDetailedThumbnail } from '../utils/thumbnailGenerator';

// API state
export interface ApiState {
  isLoading: boolean;
  error: string | null;
  lastSaved?: {
    id: string;
    name: string;
    savedAt: string;
    thumbnail?: string; // Base64 encoded PNG
  };
}

interface NewTemplateStore {
  template: Template;
  uiState: UIState;
  apiState: ApiState;
  
  // Template management
  createNewTemplate: (width: number, height: number) => void;
  loadTemplate: (template: Template) => void;
  loadTemplateFromJSON: (jsonData: any) => Promise<void>;
  
  // Cell editing
  setCellValue: (layer: LayerType, x: number, y: number, value: 0 | 1) => void;
  toggleCell: (layer: LayerType, x: number, y: number) => void;
  
  // Brush functionality
  setBrushSize: (width: number, height: number) => void;
  applyBrush: (layer: LayerType, x: number, y: number) => void;
  applyBrushWithTargetValue: (layer: LayerType, x: number, y: number, targetValue: 0 | 1) => void;
  setBrushPreview: (layer: LayerType | null, x: number, y: number, visible: boolean) => void;
  
  // Ground special functionality
  invertGroundLayer: () => void;

  // Room attributes
  toggleRoomAttribute: (attribute: RoomAttribute) => void;
  setRoomType: (roomType: RoomType) => void;

  // Drag operations
  startDrag: (layer: LayerType, x: number, y: number) => void;
  dragToCell: (layer: LayerType, x: number, y: number) => void;
  endDrag: () => void;
  
  // UI state
  setHoveredCell: (x: number, y: number) => void;
  clearHoveredCell: () => void;
  toggleLayerVisibility: (layer: LayerType) => void;
  toggleErrorDisplay: () => void;
  toggleCompositeView: () => void;
  toggleAcceptPaste: () => void;
  
  // Validation
  validateTemplate: () => void;
  
  // API operations
  saveTemplate: (name: string) => Promise<void>;
  loadTemplateFromBackend: (id: string) => Promise<void>;
  deleteTemplateFromBackend: (id: string) => Promise<void>;
  validateTemplateWithBackend: (strict?: boolean) => Promise<void>;
  clearApiError: () => void;
}

export const useNewTemplateStore = create<NewTemplateStore>((set, get) => {
  const initialTemplate = createEmptyTemplate(20, 12);
  const initialValidation = validateTemplate(initialTemplate);
  
  return {
    template: initialTemplate,
    uiState: {
      dragState: {
        isDragging: false,
        dragLayer: null,
        dragMode: null,
        lastProcessedCell: null,
        justEndedDrag: false,
      },
      hoveredCell: null,
      layerVisibility: {
        ground: true,
        softEdge: true,
        bridge: true,
        static: true,
        turret: true,
        mobGround: true,
        mobAir: true,
      },
      validationResult: initialValidation,
      showErrors: true,
      brushSize: { width: 1, height: 1 },
      brushPreview: {
        layer: null,
        x: 0,
        y: 0,
        visible: false,
      },
      showCompositeView: true,  // 总图层默认打开
      acceptPaste: false,  // 默认关闭paste功能
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


  startDrag: (layer: LayerType, x: number, y: number) => {
    const { template, uiState } = get();
    
    const currentValue = template[layer][y]?.[x];
    if (currentValue === undefined) return;
    
    // 检查是否是笔刷模式
    const isBrushMode = uiState.brushSize.width > 1 || uiState.brushSize.height > 1;
    
    if (isBrushMode) {
      // 笔刷模式：应用笔刷并设置笔刷拖拽状态
      const centerValue = template[layer][y]?.[x];
      const targetValue = centerValue === 1 ? 0 : 1;
      
      get().applyBrush(layer, x, y);
      
      set((state) => ({
        uiState: {
          ...state.uiState,
          dragState: {
            isDragging: true,
            dragLayer: layer,
            dragMode: 'brush',
            lastProcessedCell: { x, y },
            brushTargetValue: targetValue,
            justEndedDrag: false, // 清除标志
          },
        },
      }));
    } else {
      // 单格模式：原有逻辑
      const dragMode = currentValue === 0 ? 'set' : 'clear';
      
      get().toggleCell(layer, x, y);
      
      set((state) => ({
        uiState: {
          ...state.uiState,
          dragState: {
            isDragging: true,
            dragLayer: layer,
            dragMode,
            lastProcessedCell: { x, y },
            justEndedDrag: false, // 清除标志
          },
        },
      }));
    }
  },

  dragToCell: (layer: LayerType, x: number, y: number) => {
    const { uiState, template } = get();
    
    if (!uiState.dragState.isDragging || uiState.dragState.dragLayer !== layer) return;
    
    // Check if this cell was already processed
    const lastCell = uiState.dragState.lastProcessedCell;
    if (lastCell && lastCell.x === x && lastCell.y === y) return;
    
    if (uiState.dragState.dragMode === 'brush') {
      // 笔刷模式：应用笔刷到新位置
      get().applyBrushWithTargetValue(layer, x, y, uiState.dragState.brushTargetValue!);
    } else {
      // 单格模式：原有逻辑
      const targetValue = uiState.dragState.dragMode === 'set' ? 1 : 0;
      const currentValue = template[layer][y]?.[x];
      
      if (currentValue !== undefined && currentValue !== targetValue) {
        get().setCellValue(layer, x, y, targetValue);
      }
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
          dragLayer: null,
          dragMode: null,
          lastProcessedCell: null,
          justEndedDrag: true, // 标记刚刚结束拖拽
        },
      },
    }));

    // 短暂延迟后清除 justEndedDrag 标志，避免阻止后续的正常点击
    setTimeout(() => {
      set((state) => ({
        uiState: {
          ...state.uiState,
          dragState: {
            ...state.uiState.dragState,
            justEndedDrag: false,
          },
        },
      }));
    }, 50); // 50ms 足够让 onClick 事件完成
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

  toggleCompositeView: () => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        showCompositeView: !state.uiState.showCompositeView,
      },
    }));
  },

  toggleAcceptPaste: () => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        acceptPaste: !state.uiState.acceptPaste,
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
      // Generate thumbnail
      const thumbnail = await generateDetailedThumbnail(template, 120);
      
      const request = frontendToBackendCreateRequest(template, name, thumbnail);
      const response = await templateApi.createTemplate(request);
      
      set((state) => ({
        apiState: {
          ...state.apiState,
          isLoading: false,
          lastSaved: {
            id: response.id,
            name: response.name,
            savedAt: response.created_at,
            thumbnail: thumbnail,
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
            thumbnail: backendTemplate.thumbnail,
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

  deleteTemplateFromBackend: async (id: string) => {
    set((state) => ({
      apiState: {
        ...state.apiState,
        isLoading: true,
        error: null,
      },
    }));

    try {
      await templateApi.deleteTemplate(id);
      
      set((state) => ({
        apiState: {
          ...state.apiState,
          isLoading: false,
        },
      }));
    } catch (error) {
      const errorMessage = error instanceof ApiError 
        ? error.message 
        : 'Failed to delete template';
      
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
          softEdge: Array(template.height).fill(null).map(() => Array(template.width).fill(true)),
          bridge: Array(template.height).fill(null).map(() => Array(template.width).fill(true)),
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

  // Brush functionality
  setBrushSize: (width: number, height: number) => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        brushSize: { width, height },
      },
    }));
  },

  applyBrush: (layer: LayerType, centerX: number, centerY: number) => {
    const { template, uiState } = get();
    const { brushSize } = uiState;
    
    // 计算笔刷的起始位置（以中心点为基准）
    const startX = Math.max(0, centerX - Math.floor(brushSize.width / 2));
    const startY = Math.max(0, centerY - Math.floor(brushSize.height / 2));
    const endX = Math.min(template.width, startX + brushSize.width);
    const endY = Math.min(template.height, startY + brushSize.height);
    
    // 判断笔刷模式：如果中心格子是1，则整个笔刷区域设为0，反之亦然
    const centerValue = template[layer][centerY]?.[centerX];
    if (centerValue === undefined) return;
    
    const targetValue = centerValue === 1 ? 0 : 1;
    
    let newTemplate = { ...template };
    const newLayer = newTemplate[layer].map(row => [...row]);
    
    // 应用笔刷
    for (let y = startY; y < endY; y++) {
      for (let x = startX; x < endX; x++) {
        if (y < template.height && x < template.width) {
          newLayer[y][x] = targetValue;
        }
      }
    }
    
    newTemplate[layer] = newLayer;

    // 如果修改的是 ground 层，重新计算门状态
    if (layer === 'ground') {
      newTemplate.doors = calculateDoorStates(newTemplate);
    }

    // 重新计算 tile properties
    newTemplate.tileProperties = calculateAllTileProperties(newTemplate);

    const validation = validateTemplate(newTemplate);

    set({
      template: newTemplate,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
  },

  applyBrushWithTargetValue: (layer: LayerType, centerX: number, centerY: number, targetValue: 0 | 1) => {
    const { template, uiState } = get();
    const { brushSize } = uiState;
    
    // 计算笔刷的起始位置（以中心点为基准）
    const startX = Math.max(0, centerX - Math.floor(brushSize.width / 2));
    const startY = Math.max(0, centerY - Math.floor(brushSize.height / 2));
    const endX = Math.min(template.width, startX + brushSize.width);
    const endY = Math.min(template.height, startY + brushSize.height);
    
    let newTemplate = { ...template };
    const newLayer = newTemplate[layer].map(row => [...row]);
    
    // 应用笔刷，使用指定的目标值
    for (let y = startY; y < endY; y++) {
      for (let x = startX; x < endX; x++) {
        if (y < template.height && x < template.width) {
          newLayer[y][x] = targetValue;
        }
      }
    }
    
    newTemplate[layer] = newLayer;

    // 如果修改的是 ground 层，重新计算门状态
    if (layer === 'ground') {
      newTemplate.doors = calculateDoorStates(newTemplate);
    }

    // 重新计算 tile properties
    newTemplate.tileProperties = calculateAllTileProperties(newTemplate);

    const validation = validateTemplate(newTemplate);

    set({
      template: newTemplate,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
  },

  setBrushPreview: (layer: LayerType | null, x: number, y: number, visible: boolean) => {
    set((state) => ({
      uiState: {
        ...state.uiState,
        brushPreview: {
          layer,
          x,
          y,
          visible,
        },
      },
    }));
  },

  // Ground special functionality
  invertGroundLayer: () => {
    const { template } = get();

    let newTemplate = { ...template };
    const newGround = template.ground.map(row =>
      row.map(cell => cell === 1 ? 0 : 1)
    );

    newTemplate.ground = newGround;

    // 重新计算门状态
    newTemplate.doors = calculateDoorStates(newTemplate);

    // 重新计算 tile properties
    newTemplate.tileProperties = calculateAllTileProperties(newTemplate);

    const validation = validateTemplate(newTemplate);

    set({
      template: newTemplate,
      uiState: {
        ...get().uiState,
        validationResult: validation,
      },
    });
  },

  // Room attributes
  toggleRoomAttribute: (attribute: RoomAttribute) => {
    const { template } = get();

    const newTemplate = { ...template };
    newTemplate.attributes = {
      ...template.attributes,
      [attribute]: !template.attributes[attribute],
    };

    set({
      template: newTemplate,
    });
  },

  setRoomType: (roomType: RoomType) => {
    const { template } = get();

    const newTemplate = { ...template };
    newTemplate.roomType = roomType;

    set({
      template: newTemplate,
    });
  },

  // Load template from JSON data (for paste functionality)
  loadTemplateFromJSON: async (jsonData: any): Promise<void> => {
    try {
      // Validate JSON structure
      if (!jsonData || typeof jsonData !== 'object') {
        throw new Error('Invalid JSON: Expected an object');
      }

      // Check if it has the expected structure and construct backend template format
      let backendTemplate;
      
      if (jsonData.payload && jsonData.name) {
        // Format from Copy JSON: { name: string, payload: { ground, static, ..., meta } }
        const payload = jsonData.payload;
        
        // Construct backend template format
        backendTemplate = {
          id: 'pasted-template',
          name: jsonData.name,
          width: payload.meta?.width || 20,
          height: payload.meta?.height || 12,
          payload: {
            ground: payload.ground,
            bridge: payload.bridge,
            static: payload.static,
            turret: payload.turret,
            mobGround: payload.mobGround,
            mobAir: payload.mobAir,
            doors: payload.doors,
            attributes: payload.attributes,
            roomType: payload.roomType,
            tileProperties: payload.tileProperties,
          },
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
      } else if (jsonData.meta && jsonData.ground && jsonData.static) {
        // Direct payload format: { ground, static, turret, ..., meta }
        backendTemplate = {
          id: 'pasted-template',
          name: jsonData.meta?.name || 'pasted-template',
          width: jsonData.meta?.width || 20,
          height: jsonData.meta?.height || 12,
          payload: jsonData,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
      } else if (jsonData.id && jsonData.payload) {
        // Backend template format: { id, name, payload: {...} }
        backendTemplate = jsonData;
      } else {
        throw new Error('Invalid JSON structure: Expected template data with required fields (ground, static, turret, mobGround, mobAir)');
      }

      // Validate that the payload has required fields
      const payload = backendTemplate.payload;
      if (!payload.ground || !payload.static || !payload.turret || !payload.mobGround || !payload.mobAir) {
        throw new Error('Invalid template: Missing required layer data (ground, bridge, static, turret, mobGround, mobAir)');
      }
      
      // Initialize bridge layer if not present (for backward compatibility)
      if (!payload.bridge) {
        payload.bridge = Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0));
      }

      // Convert to frontend template format
      const frontendTemplate = backendToFrontendTemplate(backendTemplate);
      
      // Recalculate doors and tile properties based on current ground layer
      frontendTemplate.doors = calculateDoorStates(frontendTemplate);
      frontendTemplate.tileProperties = calculateAllTileProperties(frontendTemplate);
      
      const validation = validateTemplate(frontendTemplate);
      
      set({
        template: frontendTemplate,
        uiState: {
          ...get().uiState,
          validationResult: validation,
        },
        apiState: {
          ...get().apiState,
          error: null,
        },
      });
    } catch (error) {
      const errorMessage = error instanceof Error 
        ? error.message 
        : 'Failed to load template from JSON';
      
      set((state) => ({
        apiState: {
          ...state.apiState,
          error: errorMessage,
        },
      }));
      
      throw error; // Re-throw to allow caller to handle
    }
  },
};});