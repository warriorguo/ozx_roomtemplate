import { useMemo } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import type { CellValue } from '../../types/newTemplate';
import { computeThreatHeatmap, heatmapScoreToColor } from '../../utils/heatmapUtils';

interface CompositeCellProps {
  x: number;
  y: number;
  groundValue: CellValue;
  bridgeValue: CellValue;
  pipelineValue: CellValue;
  railValue: CellValue;
  staticValue: CellValue;
  chaserValue: CellValue;
  zonerValue: CellValue;
  dpsValue: CellValue;
  mainPathValue: CellValue;
  mobAirValue: CellValue;
  isValid: boolean;
  showErrors: boolean;
  heatmapScore: number;
  showHeatmap: boolean;
  onCellMouseEnter: (x: number, y: number) => void;
  onCellMouseLeave: () => void;
}

const CompositeCell: React.FC<CompositeCellProps> = ({
  x,
  y,
  groundValue,
  bridgeValue,
  pipelineValue,
  railValue,
  staticValue,
  chaserValue,
  zonerValue,
  dpsValue,
  mainPathValue,
  mobAirValue,
  isValid,
  showErrors,
  heatmapScore,
  showHeatmap,
  onCellMouseEnter,
  onCellMouseLeave,
}) => {
  const getCellStyle = (): React.CSSProperties => {
    const baseStyle: React.CSSProperties = {
      width: '30px',
      height: '30px',
      border: '1px solid #ddd',
      cursor: 'default',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '15px',
      fontWeight: 'bold',
      position: 'relative',
      userSelect: 'none',
      WebkitUserSelect: 'none',
      MozUserSelect: 'none',
      msUserSelect: 'none',
    };

    // 优先级：mobAir > dps > zoner > chaser > mainPath > static > rail > pipeline > bridge > ground
    if (mobAirValue === 1) {
      baseStyle.backgroundColor = '#87CEEB'; // Sky blue (mobAir)
    } else if (dpsValue === 1) {
      baseStyle.backgroundColor = '#FF4500'; // Orange-red (dps)
    } else if (zonerValue === 1) {
      baseStyle.backgroundColor = '#FFD700'; // Yellow (zoner)
    } else if (chaserValue === 1) {
      baseStyle.backgroundColor = '#4169E1'; // Blue (chaser)
    } else if (mainPathValue === 1) {
      baseStyle.backgroundColor = '#00CED1'; // Dark cyan (mainPath)
    } else if (staticValue === 1) {
      baseStyle.backgroundColor = '#FFA500'; // Orange (static)
    } else if (railValue === 1) {
      baseStyle.backgroundColor = '#8B4513'; // Brown (rail)
    } else if (pipelineValue === 1) {
      baseStyle.backgroundColor = '#9932CC'; // Purple (pipeline)
    } else if (bridgeValue === 1) {
      baseStyle.backgroundColor = '#9966CC'; // Light purple (bridge)
    } else if (groundValue === 1) {
      baseStyle.backgroundColor = '#90EE90'; // Light green (ground)
    } else {
      baseStyle.backgroundColor = '#ffffff'; // White (all 0)
      baseStyle.border = '1px solid #ddd';
    }

    // 如果验证失败，标红
    if (!isValid && showErrors) {
      baseStyle.border = '2px solid #ff0000';
      baseStyle.boxShadow = '0 0 3px rgba(255, 0, 0, 0.5)';
    }

    return baseStyle;
  };

  const handleMouseEnter = () => {
    onCellMouseEnter(x, y);
  };

  return (
    <div
      style={getCellStyle()}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={onCellMouseLeave}
      title={`(${x}, ${y}) - Ground:${groundValue} Bridge:${bridgeValue} Pipeline:${pipelineValue} Rail:${railValue} Static:${staticValue} Chaser:${chaserValue} Zoner:${zonerValue} DPS:${dpsValue} MainPath:${mainPathValue} MobAir:${mobAirValue}${!isValid ? ' [INVALID]' : ''}${showHeatmap ? ` Threat:${(heatmapScore * 100).toFixed(0)}%` : ''}`}
    >
      {showHeatmap && heatmapScore >= 0.01 && (
        <div style={{
          position: 'absolute',
          inset: 0,
          backgroundColor: heatmapScoreToColor(heatmapScore),
          pointerEvents: 'none',
        }} />
      )}
    </div>
  );
};

export const CompositeLayerEditor: React.FC = () => {
  const {
    template,
    uiState,
    setHoveredCell,
    clearHoveredCell,
    toggleCompositeView,
    toggleHeatmap,
  } = useNewTemplateStore();

  const validationResult = uiState.validationResult;
  const showCompositeView = uiState.showCompositeView;
  const showHeatmap = uiState.showHeatmap;

  const heatmap = useMemo(
    () => (showHeatmap ? computeThreatHeatmap(template) : null),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [showHeatmap, template.dps, template.zoner, template.width, template.height]
  );

  const handleToggleVisibility = () => {
    toggleCompositeView();
  };

  const gridStyle: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns: `repeat(${template.width}, 30px)`,
    gap: '1px',
    backgroundColor: '#f0f0f0',
    padding: '10px',
    borderRadius: '4px',
    maxWidth: '900px',
    maxHeight: '600px',
    overflow: 'auto',
    userSelect: 'none',
    WebkitUserSelect: 'none',
    MozUserSelect: 'none',
    msUserSelect: 'none',
  };

  const headerStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    padding: '10px',
    backgroundColor: '#9C27B020',
    border: '1px solid #9C27B0',
    borderRadius: '4px',
    marginBottom: '5px',
  };

  return (
    <div style={{ marginBottom: '20px' }}>
      <div style={headerStyle}>
        <div
          style={{
            padding: '6px 12px',
            backgroundColor: '#9C27B0',
            color: '#fff',
            border: 'none',
            borderRadius: '4px',
            fontWeight: 'bold',
            fontSize: '14px',
          }}
        >
          🗂️ Composite Layer (总图层)
        </div>

        <label style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
          <input
            type="checkbox"
            checked={showCompositeView}
            onChange={handleToggleVisibility}
          />
          👁️ Show
        </label>

        <label style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
          <input
            type="checkbox"
            checked={showHeatmap}
            onChange={toggleHeatmap}
          />
          🔥 Threat Heatmap
        </label>

        <div style={{
          marginLeft: 'auto',
          fontSize: '12px',
          color: '#666',
          fontStyle: 'italic'
        }}>
          Read-only view showing all layers combined
        </div>
      </div>

      {showCompositeView && (
        <div
          style={gridStyle}
          onMouseLeave={clearHoveredCell}
        >
          {Array.from({ length: template.height }, (_, y) =>
            Array.from({ length: template.width }, (_, x) => {
              const groundValue = template.ground[y][x];
              const bridgeValue = template.bridge[y][x];
              const pipelineValue = template.pipeline[y][x];
              const railValue = template.rail[y][x];
              const staticValue = template.static[y][x];
              const chaserValue = template.chaser[y][x];
              const zonerValue = template.zoner[y][x];
              const dpsValue = template.dps[y][x];
              const mainPathValue = template.mainPath[y][x];
              const mobAirValue = template.mobAir[y][x];

              // 检查该位置所有层是否都有效
              const allLayersValid = (
                (validationResult?.layerValidation.ground?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.bridge?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.pipeline?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.rail?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.static?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.chaser?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.zoner?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.dps?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.mainPath?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.mobAir?.[y]?.[x] ?? true)
              );

              return (
                <CompositeCell
                  key={`composite-${x}-${y}`}
                  x={x}
                  y={y}
                  groundValue={groundValue}
                  bridgeValue={bridgeValue}
                  pipelineValue={pipelineValue}
                  railValue={railValue}
                  staticValue={staticValue}
                  chaserValue={chaserValue}
                  zonerValue={zonerValue}
                  dpsValue={dpsValue}
                  mainPathValue={mainPathValue}
                  mobAirValue={mobAirValue}
                  isValid={allLayersValid}
                  showErrors={uiState.showErrors}
                  heatmapScore={heatmap ? heatmap[y][x] : 0}
                  showHeatmap={showHeatmap}
                  onCellMouseEnter={setHoveredCell}
                  onCellMouseLeave={clearHoveredCell}
                />
              );
            })
          )}
        </div>
      )}

      {showCompositeView && showHeatmap && (
        <div style={{
          marginTop: '10px',
          padding: '10px',
          backgroundColor: '#f8f9fa',
          borderRadius: '4px',
          fontSize: '12px',
          color: '#666',
        }}>
          <strong>🔥 Threat Heatmap:</strong> DPS (权重 1.0) + Zoner (权重 0.8) 影响半径 5 格
          <div style={{ marginTop: '6px', display: 'flex', alignItems: 'center', gap: '4px' }}>
            <span>低威胁</span>
            <div style={{
              width: '160px',
              height: '14px',
              borderRadius: '3px',
              background: 'linear-gradient(to right, rgba(0,0,255,0.25), rgba(0,200,255,0.35), rgba(0,220,80,0.45), rgba(255,165,0,0.55), rgba(255,0,0,0.70))',
              border: '1px solid #ccc',
            }} />
            <span>高威胁</span>
          </div>
        </div>
      )}

      {showCompositeView && (
        <div style={{
          marginTop: '10px',
          padding: '10px',
          backgroundColor: '#f8f9fa',
          borderRadius: '4px',
          fontSize: '12px',
          color: '#666'
        }}>
          <strong>图层优先级 (从高到低):</strong>
          <div style={{ marginTop: '5px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
            <span style={{ padding: '2px 8px', backgroundColor: '#87CEEB', borderRadius: '3px', color: '#000' }}>飞行怪 (MobAir)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#FF4500', borderRadius: '3px', color: '#fff' }}>DPS</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#FFD700', borderRadius: '3px', color: '#000' }}>区域怪 (Zoner)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#4169E1', borderRadius: '3px', color: '#fff' }}>追踪怪 (Chaser)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#00CED1', borderRadius: '3px', color: '#000' }}>主路径 (MainPath)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#FFA500', borderRadius: '3px', color: '#fff' }}>静态物品 (Static)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#8B4513', borderRadius: '3px', color: '#fff' }}>轨道 (Rail)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#9932CC', borderRadius: '3px', color: '#fff' }}>管道 (Pipeline)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#9966CC', borderRadius: '3px', color: '#fff' }}>桥梁 (Bridge)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#90EE90', borderRadius: '3px', color: '#000' }}>地面 (Ground)</span>
          </div>
        </div>
      )}
    </div>
  );
};
