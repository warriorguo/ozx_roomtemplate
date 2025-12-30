/**
 * Utility functions to convert between frontend and backend template formats
 */

import type { Template as FrontendTemplate } from '../types/newTemplate';
import type {
  BackendTemplate,
  BackendTemplatePayload,
  BackendCreateRequest
} from './api';
import { calculateDoorStates } from '../utils/newTemplateUtils';
import { calculateAllTileProperties } from '../utils/tilePropertiesCalculator';

/**
 * Convert frontend template to backend create request format
 */
export function frontendToBackendCreateRequest(
  template: FrontendTemplate,
  name: string,
  thumbnail?: string
): BackendCreateRequest {
  return {
    name,
    payload: {
      ground: template.ground,
      static: template.static,
      turret: template.turret,
      mobGround: template.mobGround,
      mobAir: template.mobAir,
      doors: template.doors,
      attributes: template.attributes,
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
    ground: backendTemplate.payload.ground,
    static: backendTemplate.payload.static,
    turret: backendTemplate.payload.turret,
    mobGround: backendTemplate.payload.mobGround,
    mobAir: backendTemplate.payload.mobAir,
    doors: backendTemplate.payload.doors || { top: 0, right: 0, bottom: 0, left: 0 },
    attributes: backendTemplate.payload.attributes || {
      boss: false,
      elite: false,
      mob: false,
      treasure: false,
      teleport: false,
      story: false,
    },
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
  return {
    ground: template.ground,
    static: template.static,
    turret: template.turret,
    mobGround: template.mobGround,
    mobAir: template.mobAir,
    doors: template.doors,
    attributes: template.attributes,
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