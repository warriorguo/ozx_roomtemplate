import { useTemplateStore } from '../store/templateStore';

export const InfoPanel: React.FC = () => {
  const { template, uiState } = useTemplateStore();
  const { hoveredCell } = uiState;

  const rules = {
    ground: [
      "Click to toggle: 0 ↔ 1",
      "Drag to paint/erase: Hold and drag mouse",
      "1 = walkable floor (light green)",
      "0 = non-walkable (gray)",
      "Changing to 0 clears static and land monsters"
    ],
    static: [
      "Click to toggle: 0 ↔ 1",
      "1 = can place items (orange corner 'S')",
      "0 = cannot place items", 
      "Only editable on ground tiles (ground = 1)"
    ],
    monster: [
      "Click to cycle spawn points:",
      "On ground: 0 → 1 (land) → 2 (flying) → 0",
      "Off ground: 0 → 2 (flying) → 0",
      "Land monsters (pink): ground only",
      "Flying monsters (blue): anywhere"
    ]
  };

  return (
    <div style={{
      width: '300px',
      padding: '20px',
      backgroundColor: '#f9f9f9',
      borderRadius: '8px',
      border: '1px solid #ddd'
    }}>
      <h3 style={{ margin: '0 0 15px 0' }}>Template Info</h3>
      
      <div style={{ marginBottom: '15px', fontSize: '14px' }}>
        <div><strong>Dimensions:</strong> {template.width} × {template.height}</div>
        {hoveredCell && (
          <div style={{ marginTop: '5px' }}>
            <strong>Hovered Cell:</strong> ({hoveredCell.x}, {hoveredCell.y})
            <div style={{ fontSize: '12px', marginTop: '2px' }}>
              Ground: {template.ground[hoveredCell.y][hoveredCell.x]}, 
              Static: {template.static[hoveredCell.y][hoveredCell.x]}, 
              Monster: {template.monster[hoveredCell.y][hoveredCell.x]}
            </div>
          </div>
        )}
      </div>

      <div style={{ marginBottom: '20px' }}>
        <h4 style={{ margin: '0 0 10px 0', color: '#333' }}>
          {uiState.activeLayer.charAt(0).toUpperCase() + uiState.activeLayer.slice(1)} Layer Rules:
        </h4>
        <ul style={{ 
          margin: 0, 
          paddingLeft: '20px', 
          fontSize: '12px',
          lineHeight: '1.4'
        }}>
          {rules[uiState.activeLayer].map((rule, index) => (
            <li key={index} style={{ marginBottom: '3px' }}>{rule}</li>
          ))}
        </ul>
      </div>

      <div style={{ fontSize: '12px', lineHeight: '1.4' }}>
        <h4 style={{ margin: '0 0 8px 0', color: '#333' }}>Legend:</h4>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <div style={{ width: '16px', height: '16px', backgroundColor: '#90EE90', border: '1px solid #ccc' }}></div>
            <span>Ground (walkable)</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <div style={{ width: '16px', height: '16px', backgroundColor: '#666', border: '1px solid #ccc' }}></div>
            <span>Non-ground</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <div style={{ width: '16px', height: '16px', backgroundColor: '#FFA500', border: '1px solid #ccc' }}>
              <div style={{ fontSize: '8px', color: 'white', textAlign: 'center' }}>S</div>
            </div>
            <span>Static item area</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <div style={{ width: '16px', height: '16px', backgroundColor: '#FFB6C1', borderRadius: '50%', border: '1px solid #ccc' }}></div>
            <span>Land monster spawn</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <div style={{ width: '16px', height: '16px', backgroundColor: '#87CEEB', borderRadius: '50%', border: '1px solid #ccc' }}></div>
            <span>Flying monster spawn</span>
          </div>
        </div>
      </div>
    </div>
  );
};