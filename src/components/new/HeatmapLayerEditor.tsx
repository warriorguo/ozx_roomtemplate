import { useMemo } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import { computeThreatHeatmap } from '../../utils/heatmapUtils';

/** Standalone color for the heatmap layer (more opaque than the overlay version). */
function scoreToStandaloneColor(score: number): string {
  if (score < 0.01) return '#f0f0f0'; // near-zero → light gray (empty)

  type Stop = [number, [number, number, number]];
  const stops: Stop[] = [
    [0.0,  [220, 240, 255]],
    [0.25, [100, 210, 255]],
    [0.5,  [80,  210, 80]],
    [0.75, [255, 165, 0]],
    [1.0,  [220, 30,  30]],
  ];

  let lower: Stop = stops[0];
  let upper: Stop = stops[stops.length - 1];
  for (let i = 0; i < stops.length - 1; i++) {
    if (score >= stops[i][0] && score <= stops[i + 1][0]) {
      lower = stops[i];
      upper = stops[i + 1];
      break;
    }
  }

  const range = upper[0] - lower[0];
  const t = range === 0 ? 0 : (score - lower[0]) / range;
  const r = Math.round(lower[1][0] + t * (upper[1][0] - lower[1][0]));
  const g = Math.round(lower[1][1] + t * (upper[1][1] - lower[1][1]));
  const b = Math.round(lower[1][2] + t * (upper[1][2] - lower[1][2]));
  return `rgb(${r},${g},${b})`;
}

export const HeatmapLayerEditor: React.FC = () => {
  const { template, uiState, setHoveredCell, clearHoveredCell, toggleHeatmap } = useNewTemplateStore();
  const showHeatmap = uiState.showHeatmap;

  const heatmap = useMemo(
    () => computeThreatHeatmap(template),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [template.dps, template.zoner, template.width, template.height]
  );

  const headerStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    padding: '10px',
    backgroundColor: '#FF450015',
    border: '1px solid #FF4500',
    borderRadius: '4px',
    marginBottom: '5px',
  };

  const gridStyle: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns: `repeat(${template.width}, 30px)`,
    gap: '1px',
    backgroundColor: '#e0e0e0',
    padding: '10px',
    borderRadius: '4px',
    maxWidth: '900px',
    maxHeight: '600px',
    overflow: 'auto',
    userSelect: 'none',
  };

  return (
    <div style={{ marginBottom: '20px' }}>
      <div style={headerStyle}>
        <div style={{
          padding: '6px 12px',
          backgroundColor: '#FF4500',
          color: '#fff',
          borderRadius: '4px',
          fontWeight: 'bold',
          fontSize: '14px',
        }}>
          🔥 Threat Heatmap (威胁热力图)
        </div>

        <label style={{ display: 'flex', alignItems: 'center', gap: '5px', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={showHeatmap}
            onChange={toggleHeatmap}
          />
          👁️ Show
        </label>

        <div style={{ marginLeft: 'auto', fontSize: '12px', color: '#666', fontStyle: 'italic' }}>
          DPS (×1.0) + Zoner (×0.8) · 影响半径 5 格 · 只读
        </div>
      </div>

      {showHeatmap && (
        <>
          <div style={gridStyle} onMouseLeave={clearHoveredCell}>
            {Array.from({ length: template.height }, (_, y) =>
              Array.from({ length: template.width }, (_, x) => {
                const score = heatmap[y][x];
                const pct = Math.round(score * 100);
                return (
                  <div
                    key={`heatmap-${x}-${y}`}
                    style={{
                      width: '30px',
                      height: '30px',
                      backgroundColor: scoreToStandaloneColor(score),
                      border: '1px solid rgba(0,0,0,0.08)',
                      cursor: 'default',
                    }}
                    title={`(${x}, ${y}) Threat: ${pct}%`}
                    onMouseEnter={() => setHoveredCell(x, y)}
                    onMouseLeave={clearHoveredCell}
                  />
                );
              })
            )}
          </div>

          {/* Legend */}
          <div style={{
            marginTop: '8px',
            padding: '8px 10px',
            backgroundColor: '#f8f9fa',
            borderRadius: '4px',
            fontSize: '12px',
            color: '#666',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}>
            <span>低威胁</span>
            <div style={{
              width: '180px',
              height: '14px',
              borderRadius: '3px',
              background: 'linear-gradient(to right, #dceeff, #64d2ff, #50d250, #ffa500, #dc1e1e)',
              border: '1px solid #ccc',
            }} />
            <span>高威胁</span>
            <span style={{ marginLeft: '12px', color: '#999' }}>灰色 = 无威胁</span>
          </div>
        </>
      )}
    </div>
  );
};
