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
  groundValue: CellValue;      // ground层的值
  softEdgeValue: CellValue;    // softEdge层的值
  bridgeValue: CellValue;      // bridge层的值
  pipelineValue: CellValue;    // pipeline层的值
  railValue: CellValue;        // rail层的值
  staticValue: CellValue;      // static层的值
  chaserValue: CellValue;      // chaser层的值
  zonerValue: CellValue;       // zoner层的值
  dpsValue: CellValue;         // dps层的值
  mainPathValue: CellValue;    // mainPath层的值
  mobAirValue: CellValue;      // mobAir层的值
  isCompositeView: boolean;    // 是否为总图层视图
  allLayersValid: boolean;     // 该位置所有层是否都有效
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
  groundValue,
  softEdgeValue: _softEdgeValue,
  bridgeValue,
  pipelineValue,
  railValue,
  staticValue,
  chaserValue,
  zonerValue,
  dpsValue,
  mainPathValue,
  mobAirValue,
  isCompositeView,
  allLayersValid,
  onCellClick,
  onCellMouseDown,
  onCellMouseEnter,
  onCellMouseLeave,
}) => {
  const { template } = useNewTemplateStore();

  // 判断当前格子是否是门位置，以及哪一侧需要标记浅棕色
  const getDoorBorderSide = (): 'top' | 'bottom' | 'left' | 'right' | null => {
    // 只在 ground 层显示门标记
    if (layer !== 'ground') return null;

    const width = template.width;
    const height = template.height;

    // 计算中间两格的位置
    const midWidth = Math.floor(width / 2);
    const midHeight = Math.floor(height / 2);

    // 顶部边缘的中间两格
    if (y === 0 && (x === midWidth - 1 || x === midWidth)) {
      return 'top';
    }

    // 底部边缘的中间两格
    if (y === height - 1 && (x === midWidth - 1 || x === midWidth)) {
      return 'bottom';
    }

    // 左侧边缘的中间两格
    if (x === 0 && (y === midHeight - 1 || y === midHeight)) {
      return 'left';
    }

    // 右侧边缘的中间两格
    if (x === width - 1 && (y === midHeight - 1 || y === midHeight)) {
      return 'right';
    }

    return null;
  };

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

    // 总图层视图：按优先级显示所有层的数据
    if (isCompositeView && isVisible) {
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

      // 如果有任何层验证失败，标红
      if (!allLayersValid && showErrors) {
        baseStyle.border = '2px solid #ff0000';
        baseStyle.boxShadow = '0 0 3px rgba(255, 0, 0, 0.5)';
      }

      return baseStyle;
    }

    // 普通视图：Background color based on value and layer
    if (isVisible) {
      if (value === 1) {
        // 当前层值为1时，显示该层的特征颜色
        switch (layer) {
          case 'ground':
            baseStyle.backgroundColor = '#90EE90'; // Light green
            break;
          case 'softEdge':
            baseStyle.backgroundColor = '#808080'; // Gray
            break;
          case 'bridge':
            baseStyle.backgroundColor = '#9966CC'; // Light purple
            break;
          case 'pipeline':
            baseStyle.backgroundColor = '#9932CC'; // Purple
            break;
          case 'rail':
            baseStyle.backgroundColor = '#8B4513'; // Brown
            break;
          case 'static':
            baseStyle.backgroundColor = '#FFA500'; // Orange
            break;
          case 'chaser':
            baseStyle.backgroundColor = '#4169E1'; // Blue
            break;
          case 'zoner':
            baseStyle.backgroundColor = '#FFD700'; // Yellow
            break;
          case 'dps':
            baseStyle.backgroundColor = '#FF4500'; // Orange-red
            break;
          case 'mainPath':
            baseStyle.backgroundColor = '#00CED1'; // Dark cyan
            break;
          case 'mobAir':
            baseStyle.backgroundColor = '#87CEEB'; // Sky blue
            break;
        }
      } else {
        // 当前层值为0时，根据依赖关系显示不同背景色
        switch (layer) {
          case 'softEdge':
            // softEdge层：ground=1 显示不可放置(浅红色)，ground=0 显示白色
            if (groundValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5'; // 浅红色 - 不能与ground重叠
              baseStyle.border = '1px solid #FFB3B3';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;
          case 'bridge':
            // bridge层：ground=0 显示可放置(白色)，ground=1 显示不可放置(浅红色)
            if (groundValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5'; // 浅红色 - 不能放置在walkable ground上
              baseStyle.border = '1px solid #FFB3B3';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'pipeline':
            // pipeline层：必须在ground上，不能在bridge上
            if (bridgeValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5'; // 浅红色 - 不能放置在bridge上
              baseStyle.border = '1px solid #FFB3B3';
            } else if (groundValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9'; // 浅绿色 - 可放置
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'rail':
            // rail层：可以在ground或bridge上
            if (groundValue === 1 || bridgeValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9'; // 浅绿色 - 可放置
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'static':
            // static层：bridge/pipeline/rail=1 显示浅红色，ground=1 显示浅绿色
            if (bridgeValue === 1) {
              baseStyle.backgroundColor = '#E5D3FF'; // 浅紫色 - 不能放置在bridge上
              baseStyle.border = '1px solid #D1B3FF';
            } else if (pipelineValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5'; // 浅红色 - 不能放置在pipeline上
              baseStyle.border = '1px solid #FFB3B3';
            } else if (railValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5'; // 浅红色 - 不能放置在rail上
              baseStyle.border = '1px solid #FFB3B3';
            } else if (groundValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9'; // 浅绿色
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'chaser':
            // chaser层：bridge/pipeline/rail=1 显示冲突色，static/zoner=1 显示浅橘色，ground=1 显示浅绿色
            if (bridgeValue === 1) {
              baseStyle.backgroundColor = '#E5D3FF';
              baseStyle.border = '1px solid #D1B3FF';
            } else if (pipelineValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5';
              baseStyle.border = '1px solid #FFB3B3';
            } else if (railValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5';
              baseStyle.border = '1px solid #FFB3B3';
            } else if (zonerValue === 1) {
              baseStyle.backgroundColor = '#FFFDE7';
              baseStyle.border = '1px solid #FFF9C4';
            } else if (staticValue === 1) {
              baseStyle.backgroundColor = '#FFE5CC';
              baseStyle.border = '1px solid #FFD4A3';
            } else if (groundValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9';
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'zoner':
            // zoner层：bridge/pipeline/rail=1 显示冲突色，chaser=1 显示浅蓝色，static=1 显示浅橘色，ground=1 显示浅绿色
            if (bridgeValue === 1) {
              baseStyle.backgroundColor = '#E5D3FF';
              baseStyle.border = '1px solid #D1B3FF';
            } else if (pipelineValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5';
              baseStyle.border = '1px solid #FFB3B3';
            } else if (railValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5';
              baseStyle.border = '1px solid #FFB3B3';
            } else if (chaserValue === 1) {
              baseStyle.backgroundColor = '#E3F2FD';
              baseStyle.border = '1px solid #BBDEFB';
            } else if (staticValue === 1) {
              baseStyle.backgroundColor = '#FFE5CC';
              baseStyle.border = '1px solid #FFD4A3';
            } else if (groundValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9';
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'dps':
            // dps层：bridge/pipeline/rail=1 显示冲突色，zoner=1 显示浅黄色，ground=1 显示浅绿色
            if (bridgeValue === 1) {
              baseStyle.backgroundColor = '#E5D3FF';
              baseStyle.border = '1px solid #D1B3FF';
            } else if (pipelineValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5';
              baseStyle.border = '1px solid #FFB3B3';
            } else if (railValue === 1) {
              baseStyle.backgroundColor = '#FFE5E5';
              baseStyle.border = '1px solid #FFB3B3';
            } else if (zonerValue === 1) {
              baseStyle.backgroundColor = '#FFFDE7';
              baseStyle.border = '1px solid #FFF9C4';
            } else if (groundValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9';
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'mainPath':
            // mainPath层：read-only, ground=1 显示浅绿色
            if (groundValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9';
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          case 'mobAir':
            // mobAir层：ground=1 或 bridge=1 显示浅绿色
            if (groundValue === 1 || bridgeValue === 1) {
              baseStyle.backgroundColor = '#E8F5E9'; // 浅绿色
              baseStyle.border = '1px solid #C8E6C9';
            } else {
              baseStyle.backgroundColor = '#ffffff';
              baseStyle.border = '1px solid #ddd';
            }
            break;

          default:
            // ground层和其他层
            baseStyle.backgroundColor = '#ffffff';
            baseStyle.border = '1px solid #ddd';
            break;
        }
      }
    } else {
      baseStyle.backgroundColor = '#f5f5f5'; // Light gray when layer is hidden
    }

    // Error styling - red border for invalid cells with value=1
    if (value === 1 && !isValid && showErrors && isVisible) {
      baseStyle.border = '2px solid #ff0000';
      baseStyle.boxShadow = '0 0 3px rgba(255, 0, 0, 0.5)';
    }

    // 门边框标记 - 浅棕色 (#D2B48C)
    const doorSide = getDoorBorderSide();
    if (doorSide && isVisible) {
      const doorBorderColor = '#D2B48C'; // 浅棕色 (tan)
      const doorBorderWidth = '3px';

      switch (doorSide) {
        case 'top':
          baseStyle.borderTop = `${doorBorderWidth} solid ${doorBorderColor}`;
          break;
        case 'bottom':
          baseStyle.borderBottom = `${doorBorderWidth} solid ${doorBorderColor}`;
          break;
        case 'left':
          baseStyle.borderLeft = `${doorBorderWidth} solid ${doorBorderColor}`;
          break;
        case 'right':
          baseStyle.borderRight = `${doorBorderWidth} solid ${doorBorderColor}`;
          break;
      }
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
      (document.body.style as any).webkitUserSelect = 'none';
    } else {
      document.body.style.userSelect = '';
      (document.body.style as any).webkitUserSelect = '';
    }

    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mouseup', handleMouseUp);
      // Cleanup: restore text selection
      document.body.style.userSelect = '';
      (document.body.style as any).webkitUserSelect = '';
    };
  }, [uiState.dragState.isDragging, endDrag]);

  const handleCellClick = (x: number, y: number) => {
    // 如果正在拖拽或刚刚结束拖拽，忽略点击事件（避免重复触发）
    if (uiState.dragState.isDragging || uiState.dragState.justEndedDrag) return;

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
            const groundValue = template.ground[y][x];          // 获取对应位置的ground值
            const softEdgeValue = template.softEdge[y][x];      // 获取对应位置的softEdge值
            const bridgeValue = template.bridge[y][x];          // 获取对应位置的bridge值
            const pipelineValue = template.pipeline[y][x];      // 获取对应位置的pipeline值
            const railValue = template.rail[y][x];              // 获取对应位置的rail值
            const staticValue = template.static[y][x];          // 获取对应位置的static值
            const chaserValue = template.chaser[y][x];          // 获取对应位置的chaser值
            const zonerValue = template.zoner[y][x];            // 获取对应位置的zoner值
            const dpsValue = template.dps[y][x];                // 获取对应位置的dps值
            const mainPathValue = template.mainPath[y][x];      // 获取对应位置的mainPath值
            const mobAirValue = template.mobAir[y][x];          // 获取对应位置的mobAir值

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
                groundValue={groundValue}
                softEdgeValue={softEdgeValue}
                bridgeValue={bridgeValue}
                pipelineValue={pipelineValue}
                railValue={railValue}
                staticValue={staticValue}
                chaserValue={chaserValue}
                zonerValue={zonerValue}
                dpsValue={dpsValue}
                mainPathValue={mainPathValue}
                mobAirValue={mobAirValue}
                isCompositeView={false}
                allLayersValid={true}
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