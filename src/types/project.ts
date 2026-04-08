export interface Project {
  id: string;
  name: string;
  total_rooms: number;
  shape_pct_full: number;
  shape_pct_bridge: number;
  shape_pct_platform: number;
  door_distribution: Record<string, number>;
  stage_pct_start: number;
  stage_pct_teaching: number;
  stage_pct_building: number;
  stage_pct_pressure: number;
  stage_pct_peak: number;
  stage_pct_release: number;
  stage_pct_boss: number;
  created_at: string;
  updated_at: string;
}

export interface ProjectSummary extends Project {
  template_count: number;
}

export interface CreateProjectRequest {
  name: string;
  total_rooms: number;
  shape_pct_full: number;
  shape_pct_bridge: number;
  shape_pct_platform: number;
  door_distribution: Record<string, number>;
  stage_pct_start: number;
  stage_pct_teaching: number;
  stage_pct_building: number;
  stage_pct_pressure: number;
  stage_pct_peak: number;
  stage_pct_release: number;
  stage_pct_boss: number;
}

export interface ProjectListResponse {
  total: number;
  items: ProjectSummary[];
}

export interface DimensionStat {
  required: number;
  current: number;
  deficit: number;
}

export interface ProjectStats {
  total_rooms: number;
  template_count: number;
  shape: Record<string, DimensionStat>;
  door: Record<string, DimensionStat>;
  stage: Record<string, DimensionStat>;
}

export interface AutoFillItem {
  shape: string;
  door_mask: number;
  stage_type: string;
  template_id?: string;
  error?: string;
}

export interface AutoFillResult {
  total_generated: number;
  total_failed: number;
  items: AutoFillItem[];
}

// Door bitmask labels: Top=1, Right=2, Bottom=4, Left=8
export const DOOR_BITMASK_LABELS: Record<number, string> = {
  1: 'T',
  2: 'R',
  3: 'T+R',
  4: 'B',
  5: 'T+B',
  6: 'R+B',
  7: 'T+R+B',
  8: 'L',
  9: 'T+L',
  10: 'R+L',
  11: 'T+R+L',
  12: 'B+L',
  13: 'T+B+L',
  14: 'R+B+L',
  15: 'T+R+B+L',
};
