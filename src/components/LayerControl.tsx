import { useTemplateStore } from '../store/templateStore';
import type { LayerType } from '../types/template';

export const LayerControl: React.FC = () => {
  const { uiState, setActiveLayer, toggleLayerVisibility } = useTemplateStore();

  const layers: Array<{ key: LayerType; label: string; color: string }> = [
    { key: 'ground', label: 'Ground', color: '#90EE90' },
    { key: 'static', label: 'Static', color: '#FFA500' },
    { key: 'monster', label: 'Monster', color: '#FFB6C1' },
  ];

  return (
    <div style={{ 
      display: 'flex', 
      gap: '10px', 
      padding: '10px', 
      backgroundColor: '#f5f5f5', 
      borderRadius: '4px',
      alignItems: 'center',
      flexWrap: 'wrap'
    }}>
      <span style={{ fontWeight: 'bold', marginRight: '10px' }}>Layers:</span>
      
      {layers.map(({ key, label, color }) => (
        <div key={key} style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
          <button
            onClick={() => setActiveLayer(key)}
            style={{
              padding: '6px 12px',
              border: uiState.activeLayer === key ? '2px solid #333' : '1px solid #ccc',
              backgroundColor: uiState.activeLayer === key ? color : 'white',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: uiState.activeLayer === key ? 'bold' : 'normal',
              fontSize: '12px'
            }}
          >
            {label}
          </button>
          
          <label style={{ display: 'flex', alignItems: 'center', gap: '3px', fontSize: '12px' }}>
            <input
              type="checkbox"
              checked={uiState.visible[key]}
              onChange={() => toggleLayerVisibility(key)}
              style={{ margin: 0 }}
            />
            👁️
          </label>
        </div>
      ))}
      
      <div style={{ 
        marginLeft: '20px', 
        padding: '5px', 
        backgroundColor: 'white', 
        borderRadius: '4px', 
        fontSize: '12px',
        border: '1px solid #ccc'
      }}>
        Active: <strong>{uiState.activeLayer}</strong>
      </div>
    </div>
  );
};