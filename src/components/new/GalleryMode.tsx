import React from 'react';
import { useProjectStore } from '../../store/projectStore';
import type { BackendTemplate } from '../../services/api';
import { DOOR_BITMASK_LABELS } from '../../types/project';

const CELL_SIZE = 28;

const layerColors: Record<string, string> = {
  mobAir: '#87CEEB',
  dps: '#FF4500',
  zoner: '#FFD700',
  chaser: '#4169E1',
  mainPath: '#00CED1',
  static: '#FFA500',
  rail: '#8B4513',
  pipeline: '#9932CC',
  bridge: '#9966CC',
  ground: '#90EE90',
};

const getCellColor = (template: BackendTemplate, y: number, x: number): string => {
  const p = template.payload;
  if (p.mobAir?.[y]?.[x] === 1) return layerColors.mobAir;
  if (p.dps?.[y]?.[x] === 1) return layerColors.dps;
  if (p.zoner?.[y]?.[x] === 1) return layerColors.zoner;
  if (p.chaser?.[y]?.[x] === 1) return layerColors.chaser;
  if (p.mainPath?.[y]?.[x] === 1) return layerColors.mainPath;
  if (p.static?.[y]?.[x] === 1) return layerColors.static;
  if (p.rail?.[y]?.[x] === 1) return layerColors.rail;
  if (p.pipeline?.[y]?.[x] === 1) return layerColors.pipeline;
  if (p.bridge?.[y]?.[x] === 1) return layerColors.bridge;
  if (p.ground?.[y]?.[x] === 1) return layerColors.ground;
  return '#ffffff';
};

const CompositeGrid: React.FC<{ template: BackendTemplate }> = ({ template }) => {
  const rows = template.height;
  const cols = template.width;

  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: `repeat(${cols}, ${CELL_SIZE}px)`,
      gap: '1px',
      backgroundColor: '#e0e0e0',
      borderRadius: 4,
      overflow: 'auto',
      maxHeight: '70vh',
    }}>
      {Array.from({ length: rows }, (_, y) =>
        Array.from({ length: cols }, (_, x) => (
          <div
            key={`${y}-${x}`}
            style={{
              width: CELL_SIZE,
              height: CELL_SIZE,
              backgroundColor: getCellColor(template, y, x),
            }}
          />
        ))
      )}
    </div>
  );
};

const InfoPanel: React.FC<{ template: BackendTemplate; index: number; total: number }> = ({ template, index, total }) => {
  const doors = template.payload.doors;
  const openDoorLabels = [];
  if (doors?.top) openDoorLabels.push('T');
  if (doors?.right) openDoorLabels.push('R');
  if (doors?.bottom) openDoorLabels.push('B');
  if (doors?.left) openDoorLabels.push('L');

  const openDoorsFromBitmask = template.open_doors != null ? DOOR_BITMASK_LABELS[template.open_doors] : null;

  return (
    <div style={{
      display: 'flex',
      flexWrap: 'wrap',
      gap: '16px',
      padding: '12px 16px',
      backgroundColor: '#f8f9fa',
      borderRadius: 6,
      fontSize: 14,
      alignItems: 'center',
    }}>
      <Tag label="Index" value={`${index + 1} / ${total}`} />
      <Tag label="Name" value={template.name || '(unnamed)'} />
      <Tag label="Size" value={`${template.width}x${template.height}`} />
      {template.room_type && <Tag label="Shape" value={template.room_type} color="#6f42c1" />}
      {template.stage_type && <Tag label="Stage" value={template.stage_type} color="#0d6efd" />}
      <Tag label="Doors" value={openDoorsFromBitmask || openDoorLabels.join('+') || 'none'} color="#fd7e14" />
      {template.static_count != null && <Tag label="Static" value={String(template.static_count)} />}
      {template.chaser_count != null && <Tag label="Chaser" value={String(template.chaser_count)} />}
      {template.zoner_count != null && <Tag label="Zoner" value={String(template.zoner_count)} />}
      {template.dps_count != null && <Tag label="DPS" value={String(template.dps_count)} />}
      {template.mobair_count != null && <Tag label="MobAir" value={String(template.mobair_count)} />}
      <Tag label="Views" value={String(template.view_count ?? 0)} />
    </div>
  );
};

const Tag: React.FC<{ label: string; value: string; color?: string }> = ({ label, value, color }) => (
  <span>
    <span style={{ color: '#6c757d', fontSize: 12 }}>{label}: </span>
    <span style={{ fontWeight: 600, color: color || '#212529' }}>{value}</span>
  </span>
);

export const GalleryMode: React.FC = () => {
  const {
    galleryTemplates, galleryIndex, galleryTotal,
    galleryLoading, galleryNext, galleryDelete, closeGallery,
  } = useProjectStore();

  if (galleryLoading) {
    return (
      <div style={overlayStyle}>
        <div style={{ color: 'white', fontSize: 20 }}>Loading gallery...</div>
      </div>
    );
  }

  const current = galleryTemplates[galleryIndex];
  if (!current) {
    return (
      <div style={overlayStyle}>
        <div style={{ color: 'white', fontSize: 20, textAlign: 'center' }}>
          <div>No templates to review.</div>
          <button onClick={closeGallery} style={btnStyle('#6c757d', 14)}>Close</button>
        </div>
      </div>
    );
  }

  return (
    <div style={overlayStyle}>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 12,
        maxWidth: '95vw',
        maxHeight: '95vh',
      }}>
        {/* Info */}
        <InfoPanel template={current} index={galleryIndex} total={galleryTotal} />

        {/* Grid */}
        <CompositeGrid template={current} />

        {/* Legend */}
        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap', fontSize: 12 }}>
          {Object.entries(layerColors).map(([name, color]) => (
            <span key={name} style={{ display: 'flex', alignItems: 'center', gap: 3 }}>
              <span style={{ width: 12, height: 12, backgroundColor: color, borderRadius: 2, display: 'inline-block' }} />
              {name}
            </span>
          ))}
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', gap: 16 }}>
          <button onClick={closeGallery} style={btnStyle('#6c757d', 16)}>Close</button>
          <button
            onClick={() => { if (confirm('Delete this template?')) galleryDelete(); }}
            style={btnStyle('#dc3545', 16)}
          >
            Delete
          </button>
          <button onClick={galleryNext} style={btnStyle('#28a745', 16)}>
            Next ({galleryIndex + 1}/{galleryTemplates.length})
          </button>
        </div>
      </div>
    </div>
  );
};

const overlayStyle: React.CSSProperties = {
  position: 'fixed',
  top: 0, left: 0, right: 0, bottom: 0,
  backgroundColor: 'rgba(0,0,0,0.85)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  zIndex: 2000,
};

const btnStyle = (bg: string, size: number): React.CSSProperties => ({
  padding: `${size * 0.6}px ${size * 1.5}px`,
  backgroundColor: bg,
  color: 'white',
  border: 'none',
  borderRadius: 4,
  cursor: 'pointer',
  fontSize: size,
  fontWeight: 600,
});
