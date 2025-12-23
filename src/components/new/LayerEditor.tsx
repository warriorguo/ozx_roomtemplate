import { useEffect } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import type { LayerType, CellValue } from '../../types/newTemplate';

interface LayerEditorProps {
  layer: LayerType;
  title: string;
  color: string;
}

interface CellProps {
  x: number;
  y: number;
  value: CellValue;
  isValid: boolean;
  isActive: boolean;
  isVisible: boolean;
  showErrors: boolean;
  layer: LayerType;
  onCellClick: (x: number, y: number) => void;
  onCellMouseDown: (x: number, y: number) => void;
  onCellMouseEnter: (x: number, y: number) => void;
  onCellMouseLeave: () => void;
}

const Cell: React.FC<CellProps> = ({
  x,
  y,
  value,
  isValid,
  isActive,
  isVisible,
  showErrors,
  layer,
  onCellClick,
  onCellMouseDown,
  onCellMouseEnter,
  onCellMouseLeave,
}) => {
  const getCellStyle = (): React.CSSProperties => {
    const baseStyle: React.CSSProperties = {
      width: '20px',
      height: '20px',
      border: '1px solid #ddd',
      cursor: isActive ? 'pointer' : 'default',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '10px',
      fontWeight: 'bold',
      opacity: isVisible ? 1 : 0.3,
      position: 'relative',
    };

    // Background color based on value and layer
    if (value === 1 && isVisible) {
      switch (layer) {
        case 'ground':
          baseStyle.backgroundColor = '#90EE90'; // Light green
          break;
        case 'static':
          baseStyle.backgroundColor = '#FFA500'; // Orange
          break;
        case 'turret':
          baseStyle.backgroundColor = '#FF6B6B'; // Red
          break;
        case 'mobGround':
          baseStyle.backgroundColor = '#FFB6C1'; // Pink
          break;
        case 'mobAir':
          baseStyle.backgroundColor = '#87CEEB'; // Sky blue
          break;
      }
    } else if (value === 0) {
      baseStyle.backgroundColor = '#f5f5f5'; // Light gray
    }

    // Error styling - red border for invalid cells with value=1
    if (value === 1 && !isValid && showErrors && isVisible) {
      baseStyle.border = '2px solid #ff0000';
      baseStyle.boxShadow = '0 0 3px rgba(255, 0, 0, 0.5)';
    }

    // Active layer highlighting
    if (isActive) {
      baseStyle.border = '2px solid #007bff';
    }

    return baseStyle;
  };

  const handleClick = () => {
    if (!isActive) return;
    onCellClick(x, y);
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    if (!isActive || e.button !== 0) return;
    onCellMouseDown(x, y);
  };

  const handleMouseEnter = () => {
    if (!isActive) return;
    onCellMouseEnter(x, y);
  };

  return (
    <div
      style={getCellStyle()}
      onClick={handleClick}
      onMouseDown={handleMouseDown}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={onCellMouseLeave}
      title={`(${x}, ${y}) - ${layer}: ${value}${!isValid && value === 1 ? ' [INVALID]' : ''}`}
    >
      {value === 1 && isVisible ? '●' : ''}
    </div>
  );
};

export const LayerEditor: React.FC<LayerEditorProps> = ({ layer, title, color }) => {
  const {
    template,
    uiState,
    toggleCell,
    setActiveLayer,
    toggleLayerVisibility,
    setHoveredCell,
    clearHoveredCell,
    startDrag,
    dragToCell,
    endDrag,
  } = useNewTemplateStore();

  const isActive = uiState.activeLayer === layer;
  const isVisible = uiState.layerVisibility[layer];
  const validationResult = uiState.validationResult;

  // Handle global mouse up to end drag
  useEffect(() => {
    const handleMouseUp = () => {
      if (uiState.dragState.isDragging) {
        endDrag();
      }
    };

    document.addEventListener('mouseup', handleMouseUp);
    return () => document.removeEventListener('mouseup', handleMouseUp);
  }, [uiState.dragState.isDragging, endDrag]);

  const handleCellClick = (x: number, y: number) => {
    if (uiState.dragState.isDragging) return;
    toggleCell(layer, x, y);
  };

  const handleCellMouseDown = (x: number, y: number) => {
    startDrag(layer, x, y);
  };

  const handleCellMouseEnter = (x: number, y: number) => {
    setHoveredCell(x, y);
    if (uiState.dragState.isDragging && isActive) {
      dragToCell(layer, x, y);
    }
  };

  const handleSetActive = () => {
    setActiveLayer(layer);
  };

  const handleToggleVisibility = () => {
    toggleLayerVisibility(layer);
  };

  const gridStyle: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns: `repeat(${template.width}, 20px)`,
    gap: '1px',
    backgroundColor: '#f0f0f0',
    padding: '10px',
    borderRadius: '4px',
    maxWidth: '600px',
    maxHeight: '400px',
    overflow: 'auto',
  };

  const headerStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    padding: '10px',
    backgroundColor: isActive ? color : '#f8f9fa',
    border: isActive ? `2px solid ${color}` : '1px solid #dee2e6',
    borderRadius: '4px',
    marginBottom: '5px',
  };

  return (
    <div style={{ marginBottom: '20px' }}>
      <div style={headerStyle}>
        <button
          onClick={handleSetActive}
          style={{
            padding: '6px 12px',
            backgroundColor: isActive ? color : '#fff',
            color: isActive ? '#fff' : '#000',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontWeight: isActive ? 'bold' : 'normal',
          }}
        >
          {title}
        </button>
        
        <label style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
          <input
            type="checkbox"
            checked={isVisible}
            onChange={handleToggleVisibility}
          />
          👁️ Show
        </label>
        
        <span style={{ 
          fontSize: '12px', 
          color: '#666',
          marginLeft: 'auto' 
        }}>
          {isActive ? 'ACTIVE' : ''}
        </span>
      </div>

      <div style={gridStyle}>
        {Array.from({ length: template.height }, (_, y) =>
          Array.from({ length: template.width }, (_, x) => {
            const cellValue = template[layer][y][x];
            const cellValid = validationResult?.layerValidation[layer]?.[y]?.[x] ?? true;

            return (
              <Cell
                key={`${layer}-${x}-${y}`}
                x={x}
                y={y}
                value={cellValue}
                isValid={cellValid}
                isActive={isActive}
                isVisible={isVisible}
                showErrors={uiState.showErrors}
                layer={layer}
                onCellClick={handleCellClick}
                onCellMouseDown={handleCellMouseDown}
                onCellMouseEnter={handleCellMouseEnter}
                onCellMouseLeave={clearHoveredCell}
              />
            );
          })
        )}
      </div>
    </div>
  );
};