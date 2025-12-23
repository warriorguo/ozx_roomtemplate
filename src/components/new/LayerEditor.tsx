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
      width: '30px',
      height: '30px',
      border: '1px solid #ddd',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '15px',
      fontWeight: 'bold',
      opacity: isVisible ? 1 : 0.3,
      position: 'relative',
      userSelect: 'none', // 禁用文本选择
      WebkitUserSelect: 'none', // Safari
      MozUserSelect: 'none', // Firefox
      msUserSelect: 'none', // IE/Edge
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
          baseStyle.backgroundColor = '#4169E1'; // Blue
          break;
        case 'mobGround':
          baseStyle.backgroundColor = '#FFD700'; // Yellow
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


    return baseStyle;
  };

  const handleClick = () => {
    onCellClick(x, y);
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    if (e.button !== 0) return;
    e.preventDefault(); // 阻止默认行为
    onCellMouseDown(x, y);
  };

  const handleMouseEnter = () => {
    onCellMouseEnter(x, y);
  };

  return (
    <div
      style={getCellStyle()}
      onClick={handleClick}
      onMouseDown={handleMouseDown}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={onCellMouseLeave}
      onDragStart={(e) => e.preventDefault()} // 禁用拖拽
      onSelectStart={(e) => e.preventDefault()} // 禁用选择
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
    toggleLayerVisibility,
    setHoveredCell,
    clearHoveredCell,
    startDrag,
    dragToCell,
    endDrag,
  } = useNewTemplateStore();

  const isVisible = uiState.layerVisibility[layer];
  const validationResult = uiState.validationResult;

  // Handle global mouse up to end drag and prevent text selection during drag
  useEffect(() => {
    const handleMouseUp = () => {
      if (uiState.dragState.isDragging) {
        endDrag();
      }
    };

    // Disable text selection during drag
    if (uiState.dragState.isDragging) {
      document.body.style.userSelect = 'none';
      document.body.style.webkitUserSelect = 'none';
      document.body.style.mozUserSelect = 'none';
      document.body.style.msUserSelect = 'none';
    } else {
      document.body.style.userSelect = '';
      document.body.style.webkitUserSelect = '';
      document.body.style.mozUserSelect = '';
      document.body.style.msUserSelect = '';
    }

    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mouseup', handleMouseUp);
      // Cleanup: restore text selection
      document.body.style.userSelect = '';
      document.body.style.webkitUserSelect = '';
      document.body.style.mozUserSelect = '';
      document.body.style.msUserSelect = '';
    };
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
    if (uiState.dragState.isDragging && uiState.dragState.dragLayer === layer) {
      dragToCell(layer, x, y);
    }
  };


  const handleToggleVisibility = () => {
    toggleLayerVisibility(layer);
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
    userSelect: 'none', // 禁用文本选择
    WebkitUserSelect: 'none', // Safari
    MozUserSelect: 'none', // Firefox
    msUserSelect: 'none', // IE/Edge
  };

  const headerStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    padding: '10px',
    backgroundColor: color + '20', // Add transparency
    border: `1px solid ${color}`,
    borderRadius: '4px',
    marginBottom: '5px',
  };

  return (
    <div style={{ marginBottom: '20px' }}>
      <div style={headerStyle}>
        <div
          style={{
            padding: '6px 12px',
            backgroundColor: color,
            color: '#fff',
            border: 'none',
            borderRadius: '4px',
            fontWeight: 'bold',
            fontSize: '14px',
          }}
        >
          {title}
        </div>
        
        <label style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
          <input
            type="checkbox"
            checked={isVisible}
            onChange={handleToggleVisibility}
          />
          👁️ Show
        </label>
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