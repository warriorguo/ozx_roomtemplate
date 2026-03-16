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

// Line segment structure for Pipeline and Rail
export interface Point {
  x: number;
  y: number;
}

export interface LineSegment {
  start: Point;
  end: Point;
}

export interface BackendTemplatePayload {
  ground: number[][];
  softEdge?: number[][]; // Optional for backward compatibility
  bridge?: number[][]; // Optional for backward compatibility
  pipeline?: number[][]; // Optional for backward compatibility
  pipelineLines?: LineSegment[]; // Line segments describing pipeline paths
  rail?: number[][]; // Optional for backward compatibility
  railLines?: LineSegment[]; // Line segments describing rail paths
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

export interface DoorsConnected {
  top: boolean;
  right: boolean;
  bottom: boolean;
  left: boolean;
}

export interface RoomAttributes {
  boss: boolean;
  elite: boolean;
  mob: boolean;
  treasure: boolean;
  teleport: boolean;
  story: boolean;
}

export interface TemplateSummary {
  id: string;
  name: string;
  version: number;
  width: number;
  height: number;
  thumbnail?: string;
  walkable_ratio?: number;
  room_type?: 'full' | 'bridge' | 'platform';
  room_attributes?: RoomAttributes;
  doors_connected?: DoorsConnected;
  static_count?: number;
  turret_count?: number;
  mobground_count?: number;
  mobair_count?: number;
  created_at: string;
  updated_at: string;
}

export interface BackendListResponse {
  total: number;
  items: TemplateSummary[];
}

export interface ListTemplatesParams {
  limit?: number;
  offset?: number;
  name_like?: string;
  room_type?: 'full' | 'bridge' | 'platform';
  // Attribute filters
  has_boss?: boolean;
  has_elite?: boolean;
  has_mob?: boolean;
  has_treasure?: boolean;
  has_teleport?: boolean;
  has_story?: boolean;
  // Door connectivity filters
  top_door_connected?: boolean;
  right_door_connected?: boolean;
  bottom_door_connected?: boolean;
  left_door_connected?: boolean;
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
// Use relative path by default for nginx proxy, can override with env var for local dev
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

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
  async listTemplates(params?: ListTemplatesParams): Promise<BackendListResponse> {
    const searchParams = new URLSearchParams();

    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.name_like) searchParams.set('name_like', params.name_like);
    if (params?.room_type) searchParams.set('room_type', params.room_type);

    // Attribute filters
    if (params?.has_boss !== undefined) searchParams.set('has_boss', params.has_boss.toString());
    if (params?.has_elite !== undefined) searchParams.set('has_elite', params.has_elite.toString());
    if (params?.has_mob !== undefined) searchParams.set('has_mob', params.has_mob.toString());
    if (params?.has_treasure !== undefined) searchParams.set('has_treasure', params.has_treasure.toString());
    if (params?.has_teleport !== undefined) searchParams.set('has_teleport', params.has_teleport.toString());
    if (params?.has_story !== undefined) searchParams.set('has_story', params.has_story.toString());

    // Door connectivity filters
    if (params?.top_door_connected !== undefined) searchParams.set('top_door_connected', params.top_door_connected.toString());
    if (params?.right_door_connected !== undefined) searchParams.set('right_door_connected', params.right_door_connected.toString());
    if (params?.bottom_door_connected !== undefined) searchParams.set('bottom_door_connected', params.bottom_door_connected.toString());
    if (params?.left_door_connected !== undefined) searchParams.set('left_door_connected', params.left_door_connected.toString());

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

  /**
   * Generate a bridge-type room
   */
  async generateBridge(request: BridgeGenerateRequest): Promise<BridgeGenerateResponse> {
    return this.makeRequest<BridgeGenerateResponse>('/generate/bridge', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  /**
   * Generate a platform-type room
   */
  async generatePlatform(request: PlatformGenerateRequest): Promise<PlatformGenerateResponse> {
    return this.makeRequest<PlatformGenerateResponse>('/generate/platform', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  /**
   * Generate a full-type room
   */
  async generateFullRoom(request: FullRoomGenerateRequest): Promise<FullRoomGenerateResponse> {
    return this.makeRequest<FullRoomGenerateResponse>('/generate/fullroom', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }
}

// Bridge generation types
export type DoorPosition = 'top' | 'right' | 'bottom' | 'left';

export interface BridgeGenerateRequest {
  width: number;
  height: number;
  doors: DoorPosition[];
  softEdgeCount?: number;
  railEnabled?: boolean;
  staticCount?: number;
  turretCount?: number;
  mobGroundCount?: number;
  mobAirCount?: number;
}

// Debug info types for bridge generation
export interface DoorConnectionInfo {
  from: string;
  to: string;
  pathType: string;
  brushSize: string;
}

export interface PlatformInfo {
  strategy: string;
  brushSize: string;
  points: string[];
  mirror: string;
}

export interface FloatingIslandInfo {
  position: string;
  size: string;
  fromArea: string;
  skipped?: boolean;
  skipReason?: string;
}

export interface GroundDebugInfo {
  doorConnections: DoorConnectionInfo[];
  platforms: PlatformInfo[];
  floatingIslands?: FloatingIslandInfo[];
}

export interface PlaceInfo {
  position: string;
  size: string;
  reason?: string;
}

export interface MissInfo {
  reason: string;
  count?: number;
}

export interface StaticDebugInfo {
  skipped: boolean;
  skipReason?: string;
  targetCount: number;
  placedCount: number;
  placements: PlaceInfo[];
  misses?: MissInfo[];
}

export interface TurretDebugInfo {
  skipped: boolean;
  skipReason?: string;
  targetCount: number;
  placedCount: number;
  placements: PlaceInfo[];
  misses?: MissInfo[];
}

export interface MobGroupInfo {
  groupIndex: number;
  strategy: string;
  targetCount: number;
  placedCount: number;
  placements: PlaceInfo[];
  misses?: MissInfo[];
}

export interface MobGroundDebugInfo {
  skipped: boolean;
  skipReason?: string;
  targetCount: number;
  placedCount: number;
  groups: MobGroupInfo[];
  misses?: MissInfo[];
}

export interface MobAirDebugInfo {
  skipped: boolean;
  skipReason?: string;
  targetCount: number;
  placedCount: number;
  strategy: string;
  placements: PlaceInfo[];
  misses?: MissInfo[];
}

export interface SoftEdgeDebugInfo {
  skipped: boolean;
  skipReason?: string;
  targetCount: number;
  placedCount: number;
  placements: PlaceInfo[];
  misses?: MissInfo[];
}

export interface BridgeConnection {
  from: string;
  to: string;
  position: string;
  size: string;
}

export interface BridgeLayerDebugInfo {
  skipped: boolean;
  skipReason?: string;
  islandsFound: number;
  bridgesPlaced: number;
  connections: BridgeConnection[];
  concaveGapBridges?: BridgeConnection[];
  misses?: MissInfo[];
}

export interface GenerateDebugInfo {
  ground?: GroundDebugInfo;
  softEdge?: SoftEdgeDebugInfo;
  bridgeLayer?: BridgeLayerDebugInfo;
  static?: StaticDebugInfo;
  turret?: TurretDebugInfo;
  mobGround?: MobGroundDebugInfo;
  mobAir?: MobAirDebugInfo;
}

export interface BridgeGenerateResponse {
  payload: BackendTemplatePayload;
  debugInfo?: GenerateDebugInfo;
}

// Platform generation types
export interface PlatformGenerateRequest {
  width: number;
  height: number;
  doors: DoorPosition[];
  softEdgeCount?: number;
  railEnabled?: boolean;
  staticCount?: number;
  turretCount?: number;
  mobGroundCount?: number;
  mobAirCount?: number;
}

export interface PlatformPlaceInfo {
  position: string;
  size: string;
  group?: string;
}

export interface EraserOpInfo {
  method: string;
  position: string;
  size: string;
  rolledBack?: boolean;
  reason?: string;
}

export interface PlatformGroundDebugInfo {
  strategy: string;
  platforms: PlatformPlaceInfo[];
  doorConnections: DoorConnectionInfo[];
  eraserOps?: EraserOpInfo[];
}

export interface PlatformDebugInfo {
  ground?: PlatformGroundDebugInfo;
  softEdge?: SoftEdgeDebugInfo;
  bridgeLayer?: BridgeLayerDebugInfo;
  static?: StaticDebugInfo;
  turret?: TurretDebugInfo;
  mobGround?: MobGroundDebugInfo;
  mobAir?: MobAirDebugInfo;
}

export interface PlatformGenerateResponse {
  payload: BackendTemplatePayload;
  debugInfo?: PlatformDebugInfo;
}

// Full room generation types
export interface FullRoomGenerateRequest {
  width: number;
  height: number;
  doors: DoorPosition[];
  softEdgeCount?: number;
  railEnabled?: boolean;
  staticCount?: number;
  turretCount?: number;
  mobGroundCount?: number;
  mobAirCount?: number;
}

export interface CornerEraseInfo {
  corner: string;
  position: string;
  size: string;
  rolledBack?: boolean;
  reason?: string;
}

export interface CornerEraseDebugInfo {
  skipped: boolean;
  skipReason?: string;
  brushType?: string;
  brushSize?: string;
  combo?: string;
  corners?: CornerEraseInfo[];
}

export interface CenterPitInfo {
  position: string;
  size: string;
  rolledBack?: boolean;
  reason?: string;
}

export interface CenterPitsDebugInfo {
  skipped: boolean;
  skipReason?: string;
  brushSize?: string;
  pitCount?: number;
  symmetry?: string;
  pits?: CenterPitInfo[];
}

export interface FullRoomGroundDebugInfo {
  cornerErase?: CornerEraseDebugInfo;
  centerPits?: CenterPitsDebugInfo;
}

export interface RailDebugInfo {
  skipped: boolean;
  skipReason?: string;
  platformsFound: number;
  railLoops: RailLoopInfo[];
  misses?: MissInfo[];
}

export interface RailLoopInfo {
  platform: string;
  boundingBox: string;
  perimeter: number;
  indents: IndentInfo[];
}

export interface IndentInfo {
  position: string;
  direction: string;
  size: number;
}

export interface FullRoomDebugInfo {
  ground?: FullRoomGroundDebugInfo;
  softEdge?: SoftEdgeDebugInfo;
  bridgeLayer?: BridgeLayerDebugInfo;
  rail?: RailDebugInfo;
  static?: StaticDebugInfo;
  turret?: TurretDebugInfo;
  mobGround?: MobGroundDebugInfo;
  mobAir?: MobAirDebugInfo;
}

export interface FullRoomGenerateResponse {
  payload: BackendTemplatePayload;
  debugInfo?: FullRoomDebugInfo;
}

// Create singleton instance
export const templateApi = new TemplateApiService();

// ApiError is already exported above