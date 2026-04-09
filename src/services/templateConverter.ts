/**
 * Utility functions to convert between frontend and backend template formats
 */

import type { Template as FrontendTemplate, Grid, CellValue } from '../types/newTemplate';
import type {
  BackendTemplate,
  BackendTemplatePayload,
  BackendCreateRequest
} from './api';
import { calculateDoorStates } from '../utils/newTemplateUtils';
import { calculateAllTileProperties } from '../utils/tilePropertiesCalculator';
import { extractLineSegments, hasAnyCells } from '../utils/lineExtractor';

/**
 * Convert frontend template to backend create request format
 */
export function frontendToBackendCreateRequest(
  template: FrontendTemplate,
  name: string,
  thumbnail?: string,
  projectId?: string
): BackendCreateRequest {
  // Extract line segments for pipeline and rail
  const pipelineLines = hasAnyCells(template.pipeline)
    ? extractLineSegments(template.pipeline)
    : undefined;
  const railLines = hasAnyCells(template.rail)
    ? extractLineSegments(template.rail)
    : undefined;

  return {
    name,
    payload: {
      ground: template.ground,
      softEdge: template.softEdge,
      bridge: template.bridge,
      pipeline: template.pipeline,
      pipelineLines,
      rail: template.rail,
      railLines,
      static: template.static,
      chaser: template.chaser,
      zoner: template.zoner,
      dps: template.dps,
      mainPath: template.mainPath,
      mobAir: template.mobAir,
      doors: template.doors,
      stageType: template.stageType,
      roomType: template.roomType,
      tileProperties: template.tileProperties,
      meta: {
        name,
        version: template.version,
        width: template.width,
        height: template.height,
      },
    },
    thumbnail,
    ...(projectId ? { project_id: projectId } : {}),
  };
}

/**
 * Convert backend template to frontend format
 */
export function backendToFrontendTemplate(
  backendTemplate: BackendTemplate
): FrontendTemplate {
  const template: FrontendTemplate = {
    version: 1,
    width: backendTemplate.width,
    height: backendTemplate.height,
    ground: backendTemplate.payload.ground as Grid<CellValue>,
    softEdge: (backendTemplate.payload.softEdge ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    bridge: (backendTemplate.payload.bridge ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    pipeline: (backendTemplate.payload.pipeline ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    rail: (backendTemplate.payload.rail ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    static: backendTemplate.payload.static as Grid<CellValue>,
    chaser: (backendTemplate.payload.chaser ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    zoner: (backendTemplate.payload.zoner ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    dps: (backendTemplate.payload.dps ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    mainPath: (backendTemplate.payload.mainPath ||
      Array(backendTemplate.height).fill(null).map(() => Array(backendTemplate.width).fill(0))) as Grid<CellValue>,
    mobAir: backendTemplate.payload.mobAir as Grid<CellValue>,
    doors: backendTemplate.payload.doors || { top: 0, right: 0, bottom: 0, left: 0 },
    stageType: (backendTemplate.payload.stageType as any) || 'teaching',
    roomType: backendTemplate.payload.roomType || 'full',
    tileProperties: backendTemplate.payload.tileProperties ||
      Array(backendTemplate.height).fill(null).map(() =>
        Array(backendTemplate.width).fill(null)
      ),
  };

  // 如果后端没有 doors 字段，或者需要重新计算，则计算门状态
  if (!backendTemplate.payload.doors) {
    template.doors = calculateDoorStates(template);
  }

  // 如果后端没有 tileProperties 字段，重新计算
  if (!backendTemplate.payload.tileProperties) {
    template.tileProperties = calculateAllTileProperties(template);
  }

  return template;
}

/**
 * Convert frontend template to backend payload for validation
 */
export function frontendToBackendPayload(
  template: FrontendTemplate,
  name: string = 'validation-template'
): BackendTemplatePayload {
  // Extract line segments for pipeline and rail
  const pipelineLines = hasAnyCells(template.pipeline)
    ? extractLineSegments(template.pipeline)
    : undefined;
  const railLines = hasAnyCells(template.rail)
    ? extractLineSegments(template.rail)
    : undefined;

  return {
    ground: template.ground,
    softEdge: template.softEdge,
    bridge: template.bridge,
    pipeline: template.pipeline,
    pipelineLines,
    rail: template.rail,
    railLines,
    static: template.static,
    chaser: template.chaser,
    zoner: template.zoner,
    dps: template.dps,
    mainPath: template.mainPath,
    mobAir: template.mobAir,
    doors: template.doors,
    stageType: template.stageType,
    roomType: template.roomType,
    tileProperties: template.tileProperties,
    meta: {
      name,
      version: template.version,
      width: template.width,
      height: template.height,
    },
  };
}

/**
 * Validate template name for backend requirements
 */
export function validateTemplateName(name: string): { valid: boolean; error?: string } {
  if (!name.trim()) {
    return { valid: false, error: 'Template name cannot be empty' };
  }
  
  if (name.length < 3) {
    return { valid: false, error: 'Template name must be at least 3 characters long' };
  }
  
  if (name.length > 100) {
    return { valid: false, error: 'Template name must be less than 100 characters' };
  }
  
  // Allow alphanumeric, spaces, hyphens, underscores
  if (!/^[a-zA-Z0-9\s\-_]+$/.test(name)) {
    return { 
      valid: false, 
      error: 'Template name can only contain letters, numbers, spaces, hyphens, and underscores' 
    };
  }
  
  return { valid: true };
}

/**
 * Generate a default template name based on current timestamp
 */
export function generateDefaultTemplateName(): string {
  const now = new Date();
  const timestamp = now.toISOString().slice(0, 19).replace('T', '_').replace(/:/g, '-');
  return `room-template-${timestamp}`;
}

/**
 * Format template metadata for display
 */
export function formatTemplateInfo(template: BackendTemplate): {
  displayName: string;
  info: string;
  size: string;
  created: string;
} {
  return {
    displayName: template.name,
    info: `Version ${template.version}`,
    size: `${template.width}×${template.height}`,
    created: new Date(template.created_at).toLocaleDateString(),
  };
}