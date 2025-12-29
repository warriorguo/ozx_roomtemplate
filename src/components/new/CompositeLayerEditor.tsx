import { useNewTemplateStore } from '../../store/newTemplateStore';
import type { CellValue } from '../../types/newTemplate';

interface CompositeCellProps {
  x: number;
  y: number;
  groundValue: CellValue;
  staticValue: CellValue;
  turretValue: CellValue;
  mobGroundValue: CellValue;
  mobAirValue: CellValue;
  isValid: boolean;
  showErrors: boolean;
  onCellMouseEnter: (x: number, y: number) => void;
  onCellMouseLeave: () => void;
}

const CompositeCell: React.FC<CompositeCellProps> = ({
  x,
  y,
  groundValue,
  staticValue,
  turretValue,
  mobGroundValue,
  mobAirValue,
  isValid,
  showErrors,
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

    // 优先级：mobAir > mobGround > turret > static > ground
    if (mobAirValue === 1) {
      baseStyle.backgroundColor = '#87CEEB'; // Sky blue (mobAir)
    } else if (mobGroundValue === 1) {
      baseStyle.backgroundColor = '#FFD700'; // Yellow (mobGround)
    } else if (turretValue === 1) {
      baseStyle.backgroundColor = '#4169E1'; // Blue (turret)
    } else if (staticValue === 1) {
      baseStyle.backgroundColor = '#FFA500'; // Orange (static)
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
      title={`(${x}, ${y}) - Ground:${groundValue} Static:${staticValue} Turret:${turretValue} MobGround:${mobGroundValue} MobAir:${mobAirValue}${!isValid ? ' [INVALID]' : ''}`}
    >
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
  } = useNewTemplateStore();

  const validationResult = uiState.validationResult;
  const showCompositeView = uiState.showCompositeView;

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
              const staticValue = template.static[y][x];
              const turretValue = template.turret[y][x];
              const mobGroundValue = template.mobGround[y][x];
              const mobAirValue = template.mobAir[y][x];

              // 检查该位置所有层是否都有效
              const allLayersValid = (
                (validationResult?.layerValidation.ground?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.static?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.turret?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.mobGround?.[y]?.[x] ?? true) &&
                (validationResult?.layerValidation.mobAir?.[y]?.[x] ?? true)
              );

              return (
                <CompositeCell
                  key={`composite-${x}-${y}`}
                  x={x}
                  y={y}
                  groundValue={groundValue}
                  staticValue={staticValue}
                  turretValue={turretValue}
                  mobGroundValue={mobGroundValue}
                  mobAirValue={mobAirValue}
                  isValid={allLayersValid}
                  showErrors={uiState.showErrors}
                  onCellMouseEnter={setHoveredCell}
                  onCellMouseLeave={clearHoveredCell}
                />
              );
            })
          )}
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
            <span style={{ padding: '2px 8px', backgroundColor: '#FFD700', borderRadius: '3px', color: '#000' }}>地面怪 (MobGround)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#4169E1', borderRadius: '3px', color: '#fff' }}>炮塔 (Turret)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#FFA500', borderRadius: '3px', color: '#fff' }}>静态物品 (Static)</span>
            <span style={{ padding: '2px 8px', backgroundColor: '#90EE90', borderRadius: '3px', color: '#000' }}>地面 (Ground)</span>
          </div>
        </div>
      )}
    </div>
  );
};
