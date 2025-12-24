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
    if (isVisible) {
      if (value === 1) {
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
      } else {
        baseStyle.backgroundColor = '#ffffff'; // White for 0 values
        baseStyle.border = '1px solid #ddd'; // Light border for empty cells
      }
    } else {
      baseStyle.backgroundColor = '#f5f5f5'; // Light gray when layer is hidden
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
    setBrushSize,
    applyBrush,
    invertGroundLayer,
    setBrushPreview,
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
    
    // 使用笔刷模式
    if (uiState.brushSize.width > 1 || uiState.brushSize.height > 1) {
      applyBrush(layer, x, y);
    } else {
      toggleCell(layer, x, y);
    }
  };

  const handleCellMouseDown = (x: number, y: number) => {
    // 启动拖拽（单格模式和笔刷模式都支持）
    startDrag(layer, x, y);
  };

  const handleCellMouseEnter = (x: number, y: number) => {
    setHoveredCell(x, y);
    
    // 显示笔刷预览（仅在非单格笔刷时）
    if (uiState.brushSize.width > 1 || uiState.brushSize.height > 1) {
      setBrushPreview(layer, x, y, true);
    }
    
    if (uiState.dragState.isDragging && uiState.dragState.dragLayer === layer) {
      dragToCell(layer, x, y);
    }
  };


  const handleToggleVisibility = () => {
    toggleLayerVisibility(layer);
  };

  // 预定义的笔刷尺寸
  const brushSizes = [
    { width: 1, height: 1, label: '1×1' },
    { width: 2, height: 2, label: '2×2' },
    { width: 3, height: 2, label: '3×2' },
    { width: 2, height: 3, label: '2×3' },
    { width: 3, height: 3, label: '3×3' },
    { width: 3, height: 4, label: '3×4' },
    { width: 4, height: 3, label: '4×3' },
    { width: 4, height: 4, label: '4×4' },
  ];

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

      <div style={{ display: 'flex', gap: '15px', alignItems: 'flex-start' }}>
        {/* 画布区域 */}
        <div 
          style={{...gridStyle, position: 'relative'}}
          onMouseLeave={() => {
            setBrushPreview(null, 0, 0, false);
            clearHoveredCell();
          }}
        >
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

        {/* 笔刷预览 */}
        {uiState.brushPreview.visible && 
         uiState.brushPreview.layer === layer && 
         (uiState.brushSize.width > 1 || uiState.brushSize.height > 1) && (
          (() => {
            const { x: centerX, y: centerY } = uiState.brushPreview;
            const { width: brushWidth, height: brushHeight } = uiState.brushSize;
            
            // 计算笔刷的起始位置（以中心点为基准）
            const startX = Math.max(0, centerX - Math.floor(brushWidth / 2));
            const startY = Math.max(0, centerY - Math.floor(brushHeight / 2));
            const endX = Math.min(template.width, startX + brushWidth);
            const endY = Math.min(template.height, startY + brushHeight);
            
            const actualWidth = endX - startX;
            const actualHeight = endY - startY;
            
            return (
              <div
                style={{
                  position: 'absolute',
                  pointerEvents: 'none',
                  left: `${10 + startX * 31}px`,
                  top: `${10 + startY * 31}px`,
                  width: `${actualWidth * 31 - 1}px`,
                  height: `${actualHeight * 31 - 1}px`,
                  border: `2px solid ${color}`,
                  borderRadius: '3px',
                  backgroundColor: `${color}20`,
                  zIndex: 10,
                }}
              />
            );
          })()
        )}
        </div>

        {/* 控制面板 */}
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          gap: '8px',
          minWidth: '180px',
        }}>
          {/* 笔刷选择器 */}
          <div style={{
            backgroundColor: '#f8f9fa',
            border: '1px solid #dee2e6',
            borderRadius: '6px',
            padding: '10px',
          }}>
            <h4 style={{
              margin: '0 0 8px 0',
              fontSize: '13px',
              fontWeight: 'bold',
              color: '#333',
            }}>
              🖌️ Brush Size
            </h4>
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(2, 1fr)',
              gap: '4px',
            }}>
              {brushSizes.map(size => {
                const isActive = uiState.brushSize.width === size.width && 
                                uiState.brushSize.height === size.height;
                return (
                  <button
                    key={size.label}
                    onClick={() => setBrushSize(size.width, size.height)}
                    style={{
                      padding: '9px 12px',
                      fontSize: '17px',
                      border: '1px solid #ccc',
                      borderRadius: '6px',
                      backgroundColor: isActive ? color : '#fff',
                      color: isActive ? '#fff' : '#333',
                      cursor: 'pointer',
                      fontWeight: isActive ? 'bold' : 'normal',
                      transition: 'all 0.2s ease',
                    }}
                  >
                    {size.label}
                  </button>
                );
              })}
            </div>
          </div>

          {/* Ground特殊功能 */}
          {layer === 'ground' && (
            <div style={{
              backgroundColor: '#fff3cd',
              border: '1px solid #ffeaa7',
              borderRadius: '6px',
              padding: '10px',
            }}>
              <h4 style={{
                margin: '0 0 8px 0',
                fontSize: '13px',
                fontWeight: 'bold',
                color: '#856404',
              }}>
                ⚡ Special Actions
              </h4>
              <button
                onClick={invertGroundLayer}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  fontSize: '12px',
                  backgroundColor: '#FF5722',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                  fontWeight: 'bold',
                }}
              >
                🔄 Invert All
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};