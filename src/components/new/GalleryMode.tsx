import React from 'react';
import { useProjectStore } from '../../store/projectStore';
import type { BackendTemplate } from '../../services/api';
import { DOOR_BITMASK_LABELS } from '../../types/project';

const CELL_SIZE = 30;

const layerColors: Record<string, { color: string; label: string }> = {
  mobAir:   { color: '#87CEEB', label: 'MobAir' },
  dps:      { color: '#FF4500', label: 'DPS' },
  zoner:    { color: '#FFD700', label: 'Zoner' },
  chaser:   { color: '#4169E1', label: 'Chaser' },
  mainPath: { color: '#00CED1', label: 'Path' },
  static:   { color: '#FFA500', label: 'Static' },
  rail:     { color: '#8B4513', label: 'Rail' },
  pipeline: { color: '#9932CC', label: 'Pipe' },
  bridge:   { color: '#9966CC', label: 'Bridge' },
  ground:   { color: '#90EE90', label: 'Ground' },
};

const getCellColor = (template: BackendTemplate, y: number, x: number): string => {
  const p = template.payload;
  if (p.mobAir?.[y]?.[x] === 1) return layerColors.mobAir.color;
  if (p.dps?.[y]?.[x] === 1) return layerColors.dps.color;
  if (p.zoner?.[y]?.[x] === 1) return layerColors.zoner.color;
  if (p.chaser?.[y]?.[x] === 1) return layerColors.chaser.color;
  if (p.mainPath?.[y]?.[x] === 1) return layerColors.mainPath.color;
  if (p.static?.[y]?.[x] === 1) return layerColors.static.color;
  if (p.rail?.[y]?.[x] === 1) return layerColors.rail.color;
  if (p.pipeline?.[y]?.[x] === 1) return layerColors.pipeline.color;
  if (p.bridge?.[y]?.[x] === 1) return layerColors.bridge.color;
  if (p.ground?.[y]?.[x] === 1) return layerColors.ground.color;
  return '#1a1a2e';
};

const CompositeGrid: React.FC<{ template: BackendTemplate }> = ({ template }) => {
  const rows = template.height;
  const cols = template.width;

  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: `repeat(${cols}, ${CELL_SIZE}px)`,
      gap: 0,
      borderRadius: 6,
      overflow: 'hidden',
      boxShadow: '0 0 40px rgba(0,0,0,0.5), inset 0 0 0 1px rgba(255,255,255,0.08)',
    }}>
      {Array.from({ length: rows }, (_, y) =>
        Array.from({ length: cols }, (_, x) => (
          <div
            key={`${y}-${x}`}
            style={{
              width: CELL_SIZE,
              height: CELL_SIZE,
              backgroundColor: getCellColor(template, y, x),
              borderRight: '1px solid rgba(0,0,0,0.08)',
              borderBottom: '1px solid rgba(0,0,0,0.08)',
            }}
          />
        ))
      )}
    </div>
  );
};

const InfoCard: React.FC<{ label: string; value: string; accent?: string }> = ({ label, value, accent }) => (
  <div style={{
    padding: '6px 12px',
    backgroundColor: 'rgba(255,255,255,0.06)',
    borderRadius: 6,
    border: '1px solid rgba(255,255,255,0.1)',
    textAlign: 'center',
    minWidth: 60,
  }}>
    <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.45)', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 2 }}>{label}</div>
    <div style={{ fontSize: 15, fontWeight: 700, color: accent || '#fff' }}>{value}</div>
  </div>
);

const EnemyBadge: React.FC<{ label: string; count: number; color: string }> = ({ label, count, color }) => (
  <div style={{
    display: 'flex',
    alignItems: 'center',
    gap: 5,
    padding: '3px 8px',
    backgroundColor: 'rgba(255,255,255,0.06)',
    borderRadius: 4,
    border: `1px solid ${color}40`,
  }}>
    <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: color, display: 'inline-block' }} />
    <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.7)' }}>{label}</span>
    <span style={{ fontSize: 13, fontWeight: 700, color }}>{count}</span>
  </div>
);

export const GalleryMode: React.FC<{ onEdit?: (templateId: string) => void }> = ({ onEdit }) => {
  const {
    galleryTemplates, galleryIndex, galleryTotal,
    galleryLoading, galleryNext, galleryDelete, closeGallery,
  } = useProjectStore();

  if (galleryLoading) {
    return (
      <div style={overlayStyle}>
        <div style={{ color: 'rgba(255,255,255,0.7)', fontSize: 18, fontWeight: 300 }}>Loading gallery...</div>
      </div>
    );
  }

  const current = galleryTemplates[galleryIndex];
  if (!current) {
    return (
      <div style={overlayStyle}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ color: 'rgba(255,255,255,0.6)', fontSize: 18, marginBottom: 16 }}>No templates to review.</div>
          <button onClick={closeGallery} style={actionBtn('#6c757d')}>Close</button>
        </div>
      </div>
    );
  }

  const doors = current.payload.doors;
  const doorLabels: string[] = [];
  if (doors?.top) doorLabels.push('T');
  if (doors?.right) doorLabels.push('R');
  if (doors?.bottom) doorLabels.push('B');
  if (doors?.left) doorLabels.push('L');
  const doorDisplay = (current.open_doors != null ? DOOR_BITMASK_LABELS[current.open_doors] : null) || doorLabels.join('+') || '-';

  const progress = galleryTotal > 0 ? ((galleryIndex + 1) / galleryTotal) * 100 : 0;

  return (
    <div style={overlayStyle}>
      {/* Main panel */}
      <div style={{
        backgroundColor: '#16213e',
        borderRadius: 16,
        padding: '20px 24px',
        boxShadow: '0 20px 60px rgba(0,0,0,0.6), 0 0 0 1px rgba(255,255,255,0.06)',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 14,
        maxWidth: '95vw',
        maxHeight: '95vh',
        overflow: 'auto',
      }}>
        {/* Progress bar */}
        <div style={{ width: '100%', display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{ flex: 1, height: 4, backgroundColor: 'rgba(255,255,255,0.1)', borderRadius: 2, overflow: 'hidden' }}>
            <div style={{ height: '100%', width: `${progress}%`, backgroundColor: '#6f42c1', borderRadius: 2, transition: 'width 0.3s ease' }} />
          </div>
          <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.4)', whiteSpace: 'nowrap' }}>{galleryIndex + 1} / {galleryTotal}</span>
        </div>

        {/* Info cards row */}
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', justifyContent: 'center' }}>
          <InfoCard label="Shape" value={current.room_type || '-'} accent="#a78bfa" />
          <InfoCard label="Stage" value={current.stage_type || '-'} accent="#60a5fa" />
          <InfoCard label="Doors" value={doorDisplay} accent="#fb923c" />
          <InfoCard label="Size" value={`${current.width}x${current.height}`} />
          <InfoCard label="Views" value={String(current.view_count ?? 0)} accent="rgba(255,255,255,0.5)" />
        </div>

        {/* Enemy counts */}
        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', justifyContent: 'center' }}>
          {current.static_count != null && <EnemyBadge label="Static" count={current.static_count} color="#FFA500" />}
          {current.chaser_count != null && <EnemyBadge label="Chaser" count={current.chaser_count} color="#4169E1" />}
          {current.zoner_count != null && <EnemyBadge label="Zoner" count={current.zoner_count} color="#FFD700" />}
          {current.dps_count != null && <EnemyBadge label="DPS" count={current.dps_count} color="#FF4500" />}
          {current.mobair_count != null && <EnemyBadge label="MobAir" count={current.mobair_count} color="#87CEEB" />}
        </div>

        {/* Grid */}
        <CompositeGrid template={current} />

        {/* Legend */}
        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap', justifyContent: 'center' }}>
          {Object.entries(layerColors).map(([key, { color, label }]) => (
            <span key={key} style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, color: 'rgba(255,255,255,0.45)' }}>
              <span style={{ width: 10, height: 10, backgroundColor: color, borderRadius: 2, display: 'inline-block' }} />
              {label}
            </span>
          ))}
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', gap: 12, marginTop: 4 }}>
          <button onClick={closeGallery} style={actionBtn('#4b5563')}>
            Esc
          </button>
          <button
            onClick={() => { if (confirm('Delete this template?')) galleryDelete(); }}
            style={actionBtn('#dc2626')}
          >
            Delete
          </button>
          {onEdit && (
            <button onClick={() => onEdit(current.id)} style={actionBtn('#0d6efd')}>
              Edit
            </button>
          )}
          <button onClick={galleryNext} style={{
            ...actionBtn('#7c3aed'),
            minWidth: 140,
          }}>
            Next
          </button>
        </div>
      </div>
    </div>
  );
};

const overlayStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0, left: 0, right: 0, bottom: 0,
  backgroundColor: 'rgba(0,0,0,0.9)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  zIndex: 2000,
  backdropFilter: 'blur(8px)',
};

const actionBtn = (bg: string): React.CSSProperties => ({
  padding: '10px 24px',
  backgroundColor: bg,
  color: 'white',
  border: 'none',
  borderRadius: 8,
  cursor: 'pointer',
  fontSize: 15,
  fontWeight: 600,
  letterSpacing: 0.3,
  transition: 'opacity 0.15s, transform 0.1s',
  boxShadow: `0 2px 8px ${bg}60`,
});
