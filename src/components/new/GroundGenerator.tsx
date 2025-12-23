import { useState } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import type { RoomSpec } from '../../types/newTemplate';

export const GroundGenerator: React.FC = () => {
  const { template, generateGround } = useNewTemplateStore();
  
  const [roomType, setRoomType] = useState<RoomSpec['roomType']>('rectangular');
  const [wallThickness, setWallThickness] = useState<number>(1);
  const [doors, setDoors] = useState<RoomSpec['doorPositions']>([]);
  const [showGenerator, setShowGenerator] = useState<boolean>(false);

  const handleGenerate = () => {
    const spec: RoomSpec = {
      width: template.width,
      height: template.height,
      roomType,
      wallThickness,
      doorPositions: doors,
    };

    generateGround(spec);
  };

  const addDoor = () => {
    setDoors([...doors, { x: 0, y: 0, direction: 'north' }]);
  };

  const updateDoor = (index: number, updates: Partial<RoomSpec['doorPositions'][0]>) => {
    const newDoors = [...doors];
    newDoors[index] = { ...newDoors[index], ...updates };
    setDoors(newDoors);
  };

  const removeDoor = (index: number) => {
    setDoors(doors.filter((_, i) => i !== index));
  };

  if (!showGenerator) {
    return (
      <div style={{ 
        padding: '10px', 
        backgroundColor: '#f8f9fa', 
        borderRadius: '4px',
        marginBottom: '10px'
      }}>
        <button
          onClick={() => setShowGenerator(true)}
          style={{
            padding: '6px 12px',
            backgroundColor: '#28a745',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '12px'
          }}
        >
          🏗️ Auto-Generate Ground
        </button>
      </div>
    );
  }

  return (
    <div style={{ 
      padding: '15px', 
      backgroundColor: '#f8f9fa', 
      border: '1px solid #dee2e6',
      borderRadius: '4px',
      marginBottom: '10px'
    }}>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: '15px'
      }}>
        <h4 style={{ margin: 0 }}>🏗️ Ground Auto-Generator</h4>
        <button
          onClick={() => setShowGenerator(false)}
          style={{
            padding: '4px 8px',
            backgroundColor: '#6c757d',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '12px'
          }}
        >
          ✕
        </button>
      </div>

      <div style={{ 
        display: 'grid', 
        gridTemplateColumns: '1fr 1fr', 
        gap: '15px',
        marginBottom: '15px'
      }}>
        {/* Room Type */}
        <div>
          <label style={{ 
            display: 'block', 
            fontWeight: 'bold', 
            marginBottom: '5px',
            fontSize: '12px'
          }}>
            Room Type:
          </label>
          <select
            value={roomType}
            onChange={(e) => setRoomType(e.target.value as RoomSpec['roomType'])}
            style={{
              width: '100%',
              padding: '6px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontSize: '12px'
            }}
          >
            <option value="rectangular">Rectangular</option>
            <option value="cross">Cross Shape</option>
            <option value="custom">Custom (empty)</option>
          </select>
        </div>

        {/* Wall Thickness */}
        <div>
          <label style={{ 
            display: 'block', 
            fontWeight: 'bold', 
            marginBottom: '5px',
            fontSize: '12px'
          }}>
            Wall Thickness:
          </label>
          <input
            type="number"
            value={wallThickness}
            onChange={(e) => setWallThickness(parseInt(e.target.value) || 1)}
            min="1"
            max="5"
            style={{
              width: '100%',
              padding: '6px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontSize: '12px'
            }}
          />
        </div>
      </div>

      {/* Doors */}
      <div style={{ marginBottom: '15px' }}>
        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center',
          marginBottom: '10px'
        }}>
          <label style={{ 
            fontWeight: 'bold',
            fontSize: '12px'
          }}>
            Doors:
          </label>
          <button
            onClick={addDoor}
            style={{
              padding: '4px 8px',
              backgroundColor: '#17a2b8',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: '10px'
            }}
          >
            + Add Door
          </button>
        </div>

        {doors.map((door, index) => (
          <div key={index} style={{
            display: 'grid',
            gridTemplateColumns: '60px 60px 80px 30px',
            gap: '5px',
            alignItems: 'center',
            marginBottom: '5px',
            fontSize: '12px'
          }}>
            <input
              type="number"
              placeholder="X"
              value={door.x}
              onChange={(e) => updateDoor(index, { x: parseInt(e.target.value) || 0 })}
              min="0"
              max={template.width - 1}
              style={{
                padding: '4px',
                border: '1px solid #ccc',
                borderRadius: '4px',
                fontSize: '10px'
              }}
            />
            <input
              type="number"
              placeholder="Y"
              value={door.y}
              onChange={(e) => updateDoor(index, { y: parseInt(e.target.value) || 0 })}
              min="0"
              max={template.height - 1}
              style={{
                padding: '4px',
                border: '1px solid #ccc',
                borderRadius: '4px',
                fontSize: '10px'
              }}
            />
            <select
              value={door.direction}
              onChange={(e) => updateDoor(index, { direction: e.target.value as any })}
              style={{
                padding: '4px',
                border: '1px solid #ccc',
                borderRadius: '4px',
                fontSize: '10px'
              }}
            >
              <option value="north">North</option>
              <option value="south">South</option>
              <option value="east">East</option>
              <option value="west">West</option>
            </select>
            <button
              onClick={() => removeDoor(index)}
              style={{
                padding: '4px',
                backgroundColor: '#dc3545',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '10px'
              }}
            >
              ✕
            </button>
          </div>
        ))}
      </div>

      {/* Generate Button */}
      <div style={{ textAlign: 'center' }}>
        <button
          onClick={handleGenerate}
          style={{
            padding: '8px 20px',
            backgroundColor: '#28a745',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontWeight: 'bold',
            fontSize: '14px'
          }}
        >
          🔄 Generate Ground Layer
        </button>
      </div>

      <div style={{ 
        marginTop: '10px', 
        fontSize: '11px', 
        color: '#6c757d',
        textAlign: 'center'
      }}>
        ⚠️ This will overwrite the current ground layer
      </div>
    </div>
  );
};