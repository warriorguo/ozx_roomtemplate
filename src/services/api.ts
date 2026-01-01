/**
 * API service for communicating with the tile template backend
 */

// Types matching the backend API
export interface BackendTemplate {
  id: string;
  name: string;
  version: number;
  width: number;
  height: number;
  payload: BackendTemplatePayload;
  thumbnail?: string; // Base64 encoded PNG
  created_at: string;
  updated_at: string;
}

export interface BackendTileProperties {
  walkable: boolean;
  distToTopWall: number;
  distToBottomWall: number;
  distToLeftWall: number;
  distToRightWall: number;
  distToCenter: number;
  distToEdge: number;
  distToTopDoor: number | null;
  distToBottomDoor: number | null;
  distToLeftDoor: number | null;
  distToRightDoor: number | null;
  distToNearStatic: number | null;
  distToNearTurret: number | null;
}

export interface BackendTemplatePayload {
  ground: number[][];
  bridge?: number[][]; // Optional for backward compatibility
  static: number[][];
  turret: number[][];
  mobGround: number[][];
  mobAir: number[][];
  doors?: {
    top: 0 | 1;
    right: 0 | 1;
    bottom: 0 | 1;
    left: 0 | 1;
  };
  attributes?: {
    boss: boolean;
    elite: boolean;
    mob: boolean;
    treasure: boolean;
    teleport: boolean;
    story: boolean;
  };
  roomType?: 'full' | 'bridge' | 'platform';
  tileProperties?: (BackendTileProperties | null)[][];
  meta: {
    name: string;
    version: number;
    width: number;
    height: number;
  };
}

export interface BackendCreateRequest {
  name: string;
  payload: BackendTemplatePayload;
  thumbnail?: string; // Base64 encoded PNG
}

export interface BackendCreateResponse {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface BackendListResponse {
  total: number;
  items: Array<{
    id: string;
    name: string;
    version: number;
    width: number;
    height: number;
    thumbnail?: string; // Base64 encoded PNG
    created_at: string;
    updated_at: string;
  }>;
}

export interface BackendValidationResult {
  valid: boolean;
  errors: Array<{
    layer: string;
    x: number;
    y: number;
    reason: string;
  }>;
}

export interface BackendErrorResponse {
  error: string;
  message: string;
  details?: Record<string, string>;
}

// API Configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8090/api/v1';

export class ApiError extends Error {
  public status: number;
  public details?: Record<string, string>;

  constructor(
    message: string,
    status: number,
    details?: Record<string, string>
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

/**
 * Main API service class
 */
export class TemplateApiService {
  private baseUrl: string;

  constructor(baseUrl = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  /**
   * Make HTTP request with error handling
   */
  private async makeRequest<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        let errorData: BackendErrorResponse;
        try {
          errorData = await response.json();
        } catch {
          throw new ApiError(
            `HTTP ${response.status}: ${response.statusText}`,
            response.status
          );
        }
        
        throw new ApiError(
          errorData.message || errorData.error,
          response.status,
          errorData.details
        );
      }

      // Handle empty responses (like for DELETE requests)
      const contentType = response.headers.get('Content-Type');
      if (!contentType?.includes('application/json')) {
        return {} as T;
      }

      return await response.json();
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      
      // Network or other errors
      throw new ApiError(
        `Network error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        0
      );
    }
  }

  /**
   * Create a new template
   */
  async createTemplate(request: BackendCreateRequest): Promise<BackendCreateResponse> {
    return this.makeRequest<BackendCreateResponse>('/templates', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  /**
   * List templates with optional filtering
   */
  async listTemplates(params?: {
    limit?: number;
    offset?: number;
    name_like?: string;
  }): Promise<BackendListResponse> {
    const searchParams = new URLSearchParams();
    
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.name_like) searchParams.set('name_like', params.name_like);
    
    const query = searchParams.toString();
    const endpoint = query ? `/templates?${query}` : '/templates';
    
    return this.makeRequest<BackendListResponse>(endpoint);
  }

  /**
   * Get a specific template by ID
   */
  async getTemplate(id: string): Promise<BackendTemplate> {
    return this.makeRequest<BackendTemplate>(`/templates/${id}`);
  }

  /**
   * Delete a specific template by ID
   */
  async deleteTemplate(id: string): Promise<void> {
    await this.makeRequest<void>(`/templates/${id}`, {
      method: 'DELETE',
    });
  }

  /**
   * Validate a template without saving it
   */
  async validateTemplate(
    payload: BackendTemplatePayload,
    strict = false
  ): Promise<BackendValidationResult> {
    const query = strict ? '?strict=true' : '';
    return this.makeRequest<BackendValidationResult>(`/templates/validate${query}`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  }

  /**
   * Check service health
   */
  async healthCheck(): Promise<{ status: string }> {
    return this.makeRequest<{ status: string }>('/health');
  }
}

// Create singleton instance
export const templateApi = new TemplateApiService();

// ApiError is already exported above