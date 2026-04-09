import React, { useEffect, useState } from 'react';
import { useProjectStore } from '../../store/projectStore';
import { GalleryMode } from './GalleryMode';
import type { CreateProjectRequest, ProjectSummary, DimensionStat } from '../../types/project';
import { DOOR_BITMASK_LABELS } from '../../types/project';

const STAGE_TYPES = ['start', 'teaching', 'building', 'pressure', 'peak', 'release', 'boss'] as const;
const SHAPE_TYPES = ['full', 'bridge', 'platform'] as const;

// All valid door bitmasks (at least 1 door open)
const ALL_DOOR_MASKS = Array.from({ length: 15 }, (_, i) => i + 1);

const emptyForm = (): CreateProjectRequest => ({
  name: '',
  total_rooms: 100,
  shape_pct_full: 65,
  shape_pct_bridge: 10,
  shape_pct_platform: 25,
  door_distribution: {
    '3': 3, '5': 10, '6': 3, '7': 3,
    '9': 3, '10': 22, '11': 3,
    '12': 3, '13': 3, '14': 3,
    '15': 44,
  },
  stage_pct_start: 0,
  stage_pct_teaching: 10,
  stage_pct_building: 20,
  stage_pct_pressure: 33,
  stage_pct_peak: 13,
  stage_pct_release: 17,
  stage_pct_boss: 7,
});

const projectToForm = (p: ProjectSummary): CreateProjectRequest => ({
  name: p.name,
  total_rooms: p.total_rooms,
  shape_pct_full: p.shape_pct_full,
  shape_pct_bridge: p.shape_pct_bridge,
  shape_pct_platform: p.shape_pct_platform,
  door_distribution: { ...p.door_distribution },
  stage_pct_start: p.stage_pct_start,
  stage_pct_teaching: p.stage_pct_teaching,
  stage_pct_building: p.stage_pct_building,
  stage_pct_pressure: p.stage_pct_pressure,
  stage_pct_peak: p.stage_pct_peak,
  stage_pct_release: p.stage_pct_release,
  stage_pct_boss: p.stage_pct_boss,
});

interface Props {
  onBack: () => void;
}

export const ProjectsPage: React.FC<Props> = ({ onBack }) => {
  const store = useProjectStore();
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<CreateProjectRequest>(emptyForm());
  const [formErrors, setFormErrors] = useState<string[]>([]);

  useEffect(() => {
    store.fetchProjects();
  }, []);

  const validateForm = (): string[] => {
    const errors: string[] = [];
    if (!form.name.trim()) errors.push('Name is required');
    if (form.total_rooms <= 0) errors.push('Total rooms must be positive');
    const shapeSum = form.shape_pct_full + form.shape_pct_bridge + form.shape_pct_platform;
    if (shapeSum !== 100) errors.push(`Shape percentages must sum to 100 (got ${shapeSum})`);
    const stageSum = form.stage_pct_start + form.stage_pct_teaching + form.stage_pct_building +
      form.stage_pct_pressure + form.stage_pct_peak + form.stage_pct_release + form.stage_pct_boss;
    if (stageSum !== 100) errors.push(`Stage percentages must sum to 100 (got ${stageSum})`);
    const doorSum = Object.values(form.door_distribution).reduce((a, b) => a + b, 0);
    if (doorSum !== form.total_rooms) errors.push(`Door distribution must sum to ${form.total_rooms} (got ${doorSum})`);
    return errors;
  };

  const handleSubmit = async () => {
    const errors = validateForm();
    if (errors.length > 0) {
      setFormErrors(errors);
      return;
    }
    setFormErrors([]);
    if (editingId) {
      await store.updateProject(editingId, form);
    } else {
      await store.createProject(form);
    }
    if (!store.error) {
      setShowForm(false);
      setEditingId(null);
    }
  };

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyForm());
    setFormErrors([]);
    setShowForm(true);
  };

  const openEdit = (p: ProjectSummary) => {
    setEditingId(p.id);
    setForm(projectToForm(p));
    setFormErrors([]);
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (confirm('Delete this project? Templates will be unlinked but not deleted.')) {
      await store.deleteProject(id);
    }
  };

  const updateFormField = (field: keyof CreateProjectRequest, value: number | string) => {
    setForm(prev => ({ ...prev, [field]: value }));
  };

  const updateDoorDist = (mask: string, value: number) => {
    setForm(prev => {
      const dd = { ...prev.door_distribution };
      if (value <= 0) {
        delete dd[mask];
      } else {
        dd[mask] = value;
      }
      return { ...prev, door_distribution: dd };
    });
  };

  const selectedProject = store.projects.find(p => p.id === store.selectedProjectId);

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto', padding: 20, fontFamily: '-apple-system, BlinkMacSystemFont, sans-serif' }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <button onClick={onBack} style={btnStyle('#6c757d')}>Back to Editor</button>
          <h1 style={{ margin: 0, fontSize: 24 }}>Projects</h1>
        </div>
        <button onClick={openCreate} style={btnStyle('#28a745')}>+ New Project</button>
      </div>

      {store.error && (
        <div style={errorStyle}>
          {store.error}
          <button onClick={store.clearError} style={{ marginLeft: 10, background: 'none', border: 'none', cursor: 'pointer', fontSize: 16 }}>x</button>
        </div>
      )}

      <div style={{ display: 'flex', gap: 20 }}>
        {/* Project List */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <ProjectList
            projects={store.projects}
            loading={store.loading}
            selectedId={store.selectedProjectId}
            onSelect={store.selectProject}
            onEdit={openEdit}
            onDelete={handleDelete}
          />
        </div>

        {/* Detail / Stats Panel */}
        <div style={{ width: 420, flexShrink: 0 }}>
          {selectedProject && store.stats && (
            <StatsPanel
              project={selectedProject}
              stats={store.stats}
              statsLoading={store.statsLoading}
              autoFillLoading={store.autoFillLoading}
              autoFillResult={store.autoFillResult}
              onAutoFill={() => store.autoFill(selectedProject.id)}
              onGallery={() => store.openGallery(selectedProject.id)}
            />
          )}
          {store.selectedProjectId && store.statsLoading && (
            <div style={cardStyle}>Loading stats...</div>
          )}
        </div>
      </div>

      {/* Create/Edit Modal */}
      {showForm && (
        <ProjectFormModal
          form={form}
          editingId={editingId}
          formErrors={formErrors}
          loading={store.loading}
          onFieldChange={updateFormField}
          onDoorChange={updateDoorDist}
          onSubmit={handleSubmit}
          onClose={() => { setShowForm(false); setEditingId(null); }}
        />
      )}

      {/* Gallery Mode Overlay */}
      {store.galleryActive && <GalleryMode />}
    </div>
  );
};

// --- Sub-components ---

const ProjectList: React.FC<{
  projects: ProjectSummary[];
  loading: boolean;
  selectedId: string | null;
  onSelect: (id: string) => void;
  onEdit: (p: ProjectSummary) => void;
  onDelete: (id: string) => void;
}> = ({ projects, loading, selectedId, onSelect, onEdit, onDelete }) => (
  <div style={cardStyle}>
    <h3 style={{ margin: '0 0 12px' }}>All Projects</h3>
    {loading && <div>Loading...</div>}
    {!loading && projects.length === 0 && <div style={{ color: '#666' }}>No projects yet. Create one to get started.</div>}
    {projects.map(p => (
      <div
        key={p.id}
        onClick={() => onSelect(p.id)}
        style={{
          padding: '10px 12px',
          border: `2px solid ${selectedId === p.id ? '#007bff' : '#dee2e6'}`,
          borderRadius: 6,
          marginBottom: 8,
          cursor: 'pointer',
          backgroundColor: selectedId === p.id ? '#e7f1ff' : 'white',
          transition: 'all 0.15s',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: 600, fontSize: 15 }}>{p.name || '(unnamed)'}</div>
            <div style={{ fontSize: 12, color: '#666', marginTop: 2 }}>
              {p.template_count} / {p.total_rooms} rooms
              <span style={{ margin: '0 6px' }}>|</span>
              F:{p.shape_pct_full}% B:{p.shape_pct_bridge}% P:{p.shape_pct_platform}%
            </div>
          </div>
          <div style={{ display: 'flex', gap: 6 }} onClick={e => e.stopPropagation()}>
            <button onClick={() => onEdit(p)} style={smallBtnStyle('#007bff')}>Edit</button>
            <button onClick={() => onDelete(p.id)} style={smallBtnStyle('#dc3545')}>Del</button>
          </div>
        </div>
      </div>
    ))}
  </div>
);

const StatsPanel: React.FC<{
  project: ProjectSummary;
  stats: import('../../types/project').ProjectStats;
  statsLoading: boolean;
  autoFillLoading: boolean;
  autoFillResult: import('../../types/project').AutoFillResult | null;
  onAutoFill: () => void;
  onGallery: () => void;
}> = ({ project, stats, statsLoading, autoFillLoading, autoFillResult, onAutoFill, onGallery }) => {
  const totalDeficit = Object.values(stats.stage).reduce((sum, s) => sum + s.deficit, 0);

  return (
    <div style={cardStyle}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <h3 style={{ margin: 0 }}>{project.name}</h3>
        <span style={{ fontSize: 13, color: '#666' }}>{stats.template_count} / {stats.total_rooms} rooms</span>
      </div>

      {statsLoading ? <div>Loading stats...</div> : (
        <>
          <StatSection title="Shape" data={stats.shape} labelFn={k => k.charAt(0).toUpperCase() + k.slice(1)} />
          <StatSection title="Stage" data={stats.stage} labelFn={k => k.charAt(0).toUpperCase() + k.slice(1)} />
          <StatSection title="Door Config" data={stats.door} labelFn={k => DOOR_BITMASK_LABELS[Number(k)] || k} />

          {totalDeficit > 0 && (
            <button
              onClick={onAutoFill}
              disabled={autoFillLoading}
              style={{
                ...btnStyle('#fd7e14'),
                width: '100%',
                marginTop: 12,
                opacity: autoFillLoading ? 0.6 : 1,
              }}
            >
              {autoFillLoading ? 'Generating...' : `Auto-Fill (${totalDeficit} rooms needed)`}
            </button>
          )}

          {stats.template_count > 0 && (
            <button
              onClick={onGallery}
              style={{ ...btnStyle('#6f42c1'), width: '100%', marginTop: 8 }}
            >
              Gallery ({stats.template_count} maps)
            </button>
          )}

          {autoFillResult && (
            <div style={{ marginTop: 10, padding: 10, backgroundColor: '#f8f9fa', borderRadius: 4, fontSize: 13 }}>
              <strong>Auto-Fill Result:</strong> {autoFillResult.total_generated} generated, {autoFillResult.total_failed} failed
              {autoFillResult.items.filter(i => i.error).map((item, idx) => (
                <div key={idx} style={{ color: '#dc3545', marginTop: 4 }}>
                  {item.shape}/{item.stage_type}: {item.error}
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
};

const StatSection: React.FC<{
  title: string;
  data: Record<string, DimensionStat>;
  labelFn: (key: string) => string;
}> = ({ title, data, labelFn }) => {
  const entries = Object.entries(data).sort((a, b) => b[1].deficit - a[1].deficit);
  if (entries.length === 0) return null;

  return (
    <div style={{ marginBottom: 12 }}>
      <div style={{ fontWeight: 600, fontSize: 13, marginBottom: 4, color: '#495057' }}>{title}</div>
      {entries.map(([key, stat]) => (
        <div key={key} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 3, fontSize: 13 }}>
          <span style={{ width: 70, flexShrink: 0 }}>{labelFn(key)}</span>
          <div style={{ flex: 1, height: 16, backgroundColor: '#e9ecef', borderRadius: 3, overflow: 'hidden', position: 'relative' }}>
            <div style={{
              height: '100%',
              width: stat.required > 0 ? `${Math.min(100, (stat.current / stat.required) * 100)}%` : '0%',
              backgroundColor: stat.deficit > 0 ? '#ffc107' : '#28a745',
              transition: 'width 0.3s',
            }} />
          </div>
          <span style={{ width: 70, textAlign: 'right', flexShrink: 0, color: stat.deficit > 0 ? '#dc3545' : '#28a745', fontWeight: stat.deficit > 0 ? 600 : 400 }}>
            {stat.current}/{stat.required}
          </span>
        </div>
      ))}
    </div>
  );
};

const ProjectFormModal: React.FC<{
  form: CreateProjectRequest;
  editingId: string | null;
  formErrors: string[];
  loading: boolean;
  onFieldChange: (field: keyof CreateProjectRequest, value: number | string) => void;
  onDoorChange: (mask: string, value: number) => void;
  onSubmit: () => void;
  onClose: () => void;
}> = ({ form, editingId, formErrors, loading, onFieldChange, onDoorChange, onSubmit, onClose }) => {
  const shapeSum = form.shape_pct_full + form.shape_pct_bridge + form.shape_pct_platform;
  const stageSum = form.stage_pct_start + form.stage_pct_teaching + form.stage_pct_building +
    form.stage_pct_pressure + form.stage_pct_peak + form.stage_pct_release + form.stage_pct_boss;
  const doorSum = Object.values(form.door_distribution).reduce((a, b) => a + b, 0);

  return (
    <div style={overlayStyle}>
      <div style={{ backgroundColor: 'white', borderRadius: 8, padding: 24, width: 600, maxHeight: '90vh', overflow: 'auto', boxShadow: '0 10px 30px rgba(0,0,0,0.3)' }}>
        <h2 style={{ margin: '0 0 16px' }}>{editingId ? 'Edit Project' : 'New Project'}</h2>

        {formErrors.length > 0 && (
          <div style={errorStyle}>{formErrors.map((e, i) => <div key={i}>{e}</div>)}</div>
        )}

        {/* Name & Total */}
        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>Name</label>
          <input
            value={form.name}
            onChange={e => onFieldChange('name', e.target.value)}
            style={inputStyle}
          />
        </div>
        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>Total Rooms</label>
          <input
            type="number"
            min={1}
            value={form.total_rooms}
            onChange={e => onFieldChange('total_rooms', parseInt(e.target.value) || 0)}
            style={inputStyle}
          />
        </div>

        {/* Shape Distribution */}
        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>
            Shape Distribution (%)
            <span style={{ fontWeight: 400, color: shapeSum === 100 ? '#28a745' : '#dc3545', marginLeft: 8 }}>
              Sum: {shapeSum}/100
            </span>
          </label>
          <div style={{ display: 'flex', gap: 8 }}>
            {SHAPE_TYPES.map(s => (
              <div key={s} style={{ flex: 1 }}>
                <div style={{ fontSize: 12, color: '#666', marginBottom: 2 }}>{s}</div>
                <input
                  type="number"
                  min={0}
                  max={100}
                  value={form[`shape_pct_${s}` as keyof CreateProjectRequest] as number}
                  onChange={e => onFieldChange(`shape_pct_${s}` as keyof CreateProjectRequest, parseInt(e.target.value) || 0)}
                  style={inputStyle}
                />
              </div>
            ))}
          </div>
        </div>

        {/* Stage Distribution */}
        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>
            Stage Distribution (%)
            <span style={{ fontWeight: 400, color: stageSum === 100 ? '#28a745' : '#dc3545', marginLeft: 8 }}>
              Sum: {stageSum}/100
            </span>
          </label>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 8 }}>
            {STAGE_TYPES.map(s => (
              <div key={s}>
                <div style={{ fontSize: 12, color: '#666', marginBottom: 2 }}>{s}</div>
                <input
                  type="number"
                  min={0}
                  max={100}
                  value={form[`stage_pct_${s}` as keyof CreateProjectRequest] as number}
                  onChange={e => onFieldChange(`stage_pct_${s}` as keyof CreateProjectRequest, parseInt(e.target.value) || 0)}
                  style={inputStyle}
                />
              </div>
            ))}
          </div>
        </div>

        {/* Door Distribution */}
        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>
            Door Distribution (count)
            <span style={{ fontWeight: 400, color: doorSum === form.total_rooms ? '#28a745' : '#dc3545', marginLeft: 8 }}>
              Sum: {doorSum}/{form.total_rooms}
            </span>
          </label>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 6 }}>
            {ALL_DOOR_MASKS.map(mask => (
              <div key={mask}>
                <div style={{ fontSize: 11, color: '#666', marginBottom: 2 }} title={`Bitmask: ${mask}`}>
                  {DOOR_BITMASK_LABELS[mask]}
                </div>
                <input
                  type="number"
                  min={0}
                  value={form.door_distribution[String(mask)] || 0}
                  onChange={e => onDoorChange(String(mask), parseInt(e.target.value) || 0)}
                  style={{ ...inputStyle, padding: '4px 6px', fontSize: 13 }}
                />
              </div>
            ))}
          </div>
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 20 }}>
          <button onClick={onClose} style={btnStyle('#6c757d')}>Cancel</button>
          <button onClick={onSubmit} disabled={loading} style={{ ...btnStyle('#007bff'), opacity: loading ? 0.6 : 1 }}>
            {loading ? 'Saving...' : (editingId ? 'Update' : 'Create')}
          </button>
        </div>
      </div>
    </div>
  );
};

// --- Shared styles ---

const btnStyle = (bg: string): React.CSSProperties => ({
  padding: '8px 16px',
  backgroundColor: bg,
  color: 'white',
  border: 'none',
  borderRadius: 4,
  cursor: 'pointer',
  fontSize: 14,
  fontWeight: 500,
});

const smallBtnStyle = (bg: string): React.CSSProperties => ({
  padding: '4px 10px',
  backgroundColor: bg,
  color: 'white',
  border: 'none',
  borderRadius: 3,
  cursor: 'pointer',
  fontSize: 12,
});

const cardStyle: React.CSSProperties = {
  backgroundColor: 'white',
  border: '1px solid #dee2e6',
  borderRadius: 8,
  padding: 16,
  boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
};

const overlayStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0, left: 0, right: 0, bottom: 0,
  backgroundColor: 'rgba(0,0,0,0.5)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  zIndex: 1000,
};

const errorStyle: React.CSSProperties = {
  padding: 10,
  backgroundColor: '#f8d7da',
  border: '1px solid #f5c6cb',
  borderRadius: 4,
  color: '#721c24',
  marginBottom: 12,
};

const labelStyle: React.CSSProperties = {
  display: 'block',
  fontWeight: 600,
  fontSize: 14,
  marginBottom: 4,
};

const inputStyle: React.CSSProperties = {
  width: '100%',
  padding: '8px 10px',
  border: '1px solid #ddd',
  borderRadius: 4,
  fontSize: 14,
  boxSizing: 'border-box',
};
