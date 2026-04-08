import { create } from 'zustand';
import { templateApi } from '../services/api';
import type { ProjectSummary, CreateProjectRequest, ProjectStats, AutoFillResult } from '../types/project';

interface ProjectStore {
  // State
  projects: ProjectSummary[];
  total: number;
  loading: boolean;
  error: string | null;
  selectedProjectId: string | null;
  stats: ProjectStats | null;
  statsLoading: boolean;
  autoFillResult: AutoFillResult | null;
  autoFillLoading: boolean;

  // Actions
  fetchProjects: () => Promise<void>;
  createProject: (req: CreateProjectRequest) => Promise<void>;
  updateProject: (id: string, req: CreateProjectRequest) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
  selectProject: (id: string | null) => void;
  fetchStats: (id: string) => Promise<void>;
  autoFill: (id: string) => Promise<void>;
  clearError: () => void;
}

export const useProjectStore = create<ProjectStore>((set, get) => ({
  projects: [],
  total: 0,
  loading: false,
  error: null,
  selectedProjectId: null,
  stats: null,
  statsLoading: false,
  autoFillResult: null,
  autoFillLoading: false,

  fetchProjects: async () => {
    set({ loading: true, error: null });
    try {
      const resp = await templateApi.listProjects({ limit: 100 });
      set({ projects: resp.items || [], total: resp.total, loading: false });
    } catch (e: unknown) {
      set({ loading: false, error: e instanceof Error ? e.message : 'Failed to fetch projects' });
    }
  },

  createProject: async (req) => {
    set({ loading: true, error: null });
    try {
      await templateApi.createProject(req);
      await get().fetchProjects();
    } catch (e: unknown) {
      set({ loading: false, error: e instanceof Error ? e.message : 'Failed to create project' });
    }
  },

  updateProject: async (id, req) => {
    set({ loading: true, error: null });
    try {
      await templateApi.updateProject(id, req);
      await get().fetchProjects();
      if (get().selectedProjectId === id) {
        await get().fetchStats(id);
      }
    } catch (e: unknown) {
      set({ loading: false, error: e instanceof Error ? e.message : 'Failed to update project' });
    }
  },

  deleteProject: async (id) => {
    set({ loading: true, error: null });
    try {
      await templateApi.deleteProject(id);
      if (get().selectedProjectId === id) {
        set({ selectedProjectId: null, stats: null });
      }
      await get().fetchProjects();
    } catch (e: unknown) {
      set({ loading: false, error: e instanceof Error ? e.message : 'Failed to delete project' });
    }
  },

  selectProject: (id) => {
    set({ selectedProjectId: id, stats: null, autoFillResult: null });
    if (id) {
      get().fetchStats(id);
    }
  },

  fetchStats: async (id) => {
    set({ statsLoading: true });
    try {
      const stats = await templateApi.getProjectStats(id);
      set({ stats, statsLoading: false });
    } catch (e: unknown) {
      set({ statsLoading: false, error: e instanceof Error ? e.message : 'Failed to fetch stats' });
    }
  },

  autoFill: async (id) => {
    set({ autoFillLoading: true, autoFillResult: null, error: null });
    try {
      const result = await templateApi.autoFillProject(id);
      set({ autoFillResult: result, autoFillLoading: false });
      // Refresh stats and project list after auto-fill
      await get().fetchStats(id);
      await get().fetchProjects();
    } catch (e: unknown) {
      set({ autoFillLoading: false, error: e instanceof Error ? e.message : 'Auto-fill failed' });
    }
  },

  clearError: () => set({ error: null }),
}));
