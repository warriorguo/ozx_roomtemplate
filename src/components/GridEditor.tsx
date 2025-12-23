import { useTemplateStore } from '../store/templateStore';
import { useEffect } from 'react';
import type { LayerType, GroundCell, StaticCell, MonsterCell } from '../types/template';

interface CellProps {
  x: number;
  y: number;
  groundValue: GroundCell;
  staticValue: StaticCell;
  monsterValue: MonsterCell;
  isVisible: { ground: boolean; static: boolean; monster: boolean };
  activeLayer: LayerType;
  isDragging: boolean;
  onCellClick: (x: number, y: number, layer: LayerType) => void;
  onCellHover: (x: number, y: number) => void;
  onCellLeave: () => void;
  onCellMouseDown: (x: number, y: number) => void;
  onCellMouseEnter: (x: number, y: number) => void;
}

const Cell: React.FC<CellProps> = ({
  x,
  y,
  groundValue,
  staticValue,
  monsterValue,
  isVisible,
  activeLayer,
  isDragging,
  onCellClick,
  onCellHover,
  onCellLeave,
  onCellMouseDown,
  onCellMouseEnter,
}) => {
  const getCellStyle = (): React.CSSProperties => {
    const baseStyle: React.CSSProperties = {
      width: '24px',
      height: '24px',
      border: '1px solid #ccc',
      position: 'relative',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '10px',
      fontWeight: 'bold',
    };

    let backgroundColor = '#666';
    if (isVisible.ground && groundValue === 1) {
      backgroundColor = '#90EE90';
    } else if (isVisible.ground && groundValue === 0) {
      backgroundColor = '#666';
    }

    const isGroundAvailable = groundValue === 1;
    const canEditStatic = isGroundAvailable;
    
    if (activeLayer === 'static' && !canEditStatic) {
      baseStyle.cursor = 'not-allowed';
      baseStyle.opacity = 0.5;
    }

    // Add drag cursor for ground layer
    if (activeLayer === 'ground') {
      baseStyle.cursor = isDragging ? 'grabbing' : 'grab';
    }

    // Add visual feedback during dragging
    if (isDragging && activeLayer === 'ground') {
      baseStyle.boxShadow = '0 0 0 1px #007bff';
      baseStyle.transform = 'scale(0.95)';
    }

    return {
      ...baseStyle,
      backgroundColor,
    };
  };

  const renderLayerContent = () => {
    const content = [];

    if (isVisible.static && staticValue === 1) {
      content.push(
        <div
          key="static"
          style={{
            position: 'absolute',
            top: 0,
            right: 0,
            width: '8px',
            height: '8px',
            backgroundColor: '#FFA500',
            borderRadius: '2px',
            fontSize: '6px',
            color: 'white',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          S
        </div>
      );
    }

    if (isVisible.monster && monsterValue > 0) {
      const isFlying = monsterValue === 2;
      const isOnNonGround = groundValue === 0 && isFlying;
      
      content.push(
        <div
          key="monster"
          style={{
            position: 'absolute',
            width: '12px',
            height: '12px',
            borderRadius: '50%',
            backgroundColor: isFlying ? (isOnNonGround ? '#000080' : '#87CEEB') : '#FFB6C1',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            fontSize: '8px',
            color: 'white',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {isFlying ? 'F' : 'L'}
        </div>
      );
    }

    return content;
  };

  const handleClick = () => {
    // If dragging is in progress, don't handle click
    if (isDragging) return;
    
    const canEdit = activeLayer !== 'static' || groundValue === 1;
    if (canEdit) {
      onCellClick(x, y, activeLayer);
    }
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    // Only handle left mouse button
    if (e.button !== 0) return;
    
    // Only start drag for ground layer
    if (activeLayer === 'ground') {
      onCellMouseDown(x, y);
    }
  };

  const handleMouseEnter = () => {
    onCellHover(x, y);
    
    // Handle drag operation
    if (isDragging && activeLayer === 'ground') {
      onCellMouseEnter(x, y);
    }
  };

  return (
    <div
      style={getCellStyle()}
      onClick={handleClick}
      onMouseDown={handleMouseDown}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={onCellLeave}
      title={`(${x}, ${y}) - Ground: ${groundValue}, Static: ${staticValue}, Monster: ${monsterValue}`}
    >
      {renderLayerContent()}
    </div>
  );
};

export const GridEditor: React.FC = () => {
  const { 
    template, 
    uiState, 
    toggleCell, 
    setHoveredCell, 
    clearHoveredCell,
    startDrag,
    dragToCell,
    endDrag
  } = useTemplateStore();

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

  const gridStyle: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns: `repeat(${template.width}, 24px)`,
    gap: '1px',
    padding: '20px',
    backgroundColor: '#f0f0f0',
    border: '2px solid #333',
    borderRadius: '4px',
    maxWidth: '800px',
    maxHeight: '600px',
    overflow: 'auto',
  };

  return (
    <div style={gridStyle}>
      {Array.from({ length: template.height }, (_, y) =>
        Array.from({ length: template.width }, (_, x) => (
          <Cell
            key={`${x}-${y}`}
            x={x}
            y={y}
            groundValue={template.ground[y][x]}
            staticValue={template.static[y][x]}
            monsterValue={template.monster[y][x]}
            isVisible={uiState.visible}
            activeLayer={uiState.activeLayer}
            isDragging={uiState.dragState.isDragging}
            onCellClick={toggleCell}
            onCellHover={setHoveredCell}
            onCellLeave={clearHoveredCell}
            onCellMouseDown={startDrag}
            onCellMouseEnter={dragToCell}
          />
        ))
      )}
    </div>
  );
};