import { create } from 'zustand';
import { templateApi } from '../services/api';
import type { BackendTemplate } from '../services/api';
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

  // Gallery state
  galleryTemplates: BackendTemplate[];
  galleryIndex: number;
  galleryTotal: number;
  galleryLoading: boolean;
  galleryActive: boolean;

  // Actions
  fetchProjects: () => Promise<void>;
  createProject: (req: CreateProjectRequest) => Promise<void>;
  updateProject: (id: string, req: CreateProjectRequest) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
  selectProject: (id: string | null) => void;
  fetchStats: (id: string) => Promise<void>;
  autoFill: (id: string) => Promise<void>;
  clearError: () => void;

  // Gallery actions
  openGallery: (projectId: string) => Promise<void>;
  closeGallery: () => void;
  galleryNext: () => Promise<void>;
  galleryDelete: () => Promise<void>;
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

  galleryTemplates: [],
  galleryIndex: 0,
  galleryTotal: 0,
  galleryLoading: false,
  galleryActive: false,

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
      await get().fetchStats(id);
      await get().fetchProjects();
    } catch (e: unknown) {
      set({ autoFillLoading: false, error: e instanceof Error ? e.message : 'Auto-fill failed' });
    }
  },

  clearError: () => set({ error: null }),

  // Gallery actions
  openGallery: async (projectId) => {
    set({ galleryLoading: true, galleryActive: true, galleryIndex: 0, error: null });
    try {
      const resp = await templateApi.listProjectTemplates(projectId, 500);
      set({
        galleryTemplates: resp.items || [],
        galleryTotal: resp.total,
        galleryIndex: 0,
        galleryLoading: false,
      });
    } catch (e: unknown) {
      set({ galleryLoading: false, galleryActive: false, error: e instanceof Error ? e.message : 'Failed to load gallery' });
    }
  },

  closeGallery: () => {
    set({ galleryActive: false, galleryTemplates: [], galleryIndex: 0 });
    // Refresh stats after gallery session
    const pid = get().selectedProjectId;
    if (pid) {
      get().fetchStats(pid);
      get().fetchProjects();
    }
  },

  galleryNext: async () => {
    const { galleryTemplates, galleryIndex } = get();
    const current = galleryTemplates[galleryIndex];
    if (!current) return;

    // Increment view count
    try {
      await templateApi.incrementViewCount(current.id);
    } catch {
      // non-critical, continue
    }

    if (galleryIndex + 1 >= galleryTemplates.length) {
      // End of gallery
      get().closeGallery();
    } else {
      set({ galleryIndex: galleryIndex + 1 });
    }
  },

  galleryDelete: async () => {
    const { galleryTemplates, galleryIndex } = get();
    const current = galleryTemplates[galleryIndex];
    if (!current) return;

    try {
      await templateApi.deleteTemplate(current.id);
      const newTemplates = [...galleryTemplates];
      newTemplates.splice(galleryIndex, 1);

      if (newTemplates.length === 0) {
        get().closeGallery();
      } else {
        const newIndex = galleryIndex >= newTemplates.length ? newTemplates.length - 1 : galleryIndex;
        set({ galleryTemplates: newTemplates, galleryIndex: newIndex, galleryTotal: get().galleryTotal - 1 });
      }
    } catch (e: unknown) {
      set({ error: e instanceof Error ? e.message : 'Failed to delete template' });
    }
  },
}));
