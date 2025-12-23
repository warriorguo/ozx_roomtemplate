import { ToolBar } from './ToolBar';
import { LayerEditor } from './LayerEditor';
import { GroundGenerator } from './GroundGenerator';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import type { LayerType } from '../../types/newTemplate';

const layerConfigs: Array<{
  layer: LayerType;
  title: string;
  color: string;
  description: string;
}> = [
  {
    layer: 'ground',
    title: 'Ground (地面)',
    color: '#90EE90',
    description: 'Walkable areas - foundation for all other layers'
  },
  {
    layer: 'static',
    title: 'Static (静态物品)',
    color: '#FFA500',
    description: 'Static objects placement areas (requires ground=1)'
  },
  {
    layer: 'turret',
    title: 'Turret (炮塔)',
    color: '#4169E1',
    description: 'Turret placement (requires ground=1, static=0)'
  },
  {
    layer: 'mobGround',
    title: 'Mob Ground (地面怪)',
    color: '#FFD700',
    description: 'Ground mob spawns (requires ground=1, static=0, turret=0)'
  },
  {
    layer: 'mobAir',
    title: 'Mob Air (飞行怪)',
    color: '#87CEEB',
    description: 'Air mob spawns (no constraints)'
  },
];

export const TileTemplateApp: React.FC = () => {
  const { uiState, template, apiState } = useNewTemplateStore();

  const ErrorSummary: React.FC = () => {
    const { validationResult } = uiState;
    
    if (!validationResult || validationResult.isValid) {
      return null;
    }

    const errorsByLayer = validationResult.errors.reduce((acc, error) => {
      if (!acc[error.layer]) acc[error.layer] = [];
      acc[error.layer].push(error);
      return acc;
    }, {} as Record<LayerType, typeof validationResult.errors>);

    return (
      <div style={{
        padding: '15px',
        backgroundColor: '#fff3cd',
        border: '1px solid #ffeaa7',
        borderRadius: '4px',
        marginBottom: '20px'
      }}>
        <h4 style={{ margin: '0 0 10px 0', color: '#856404' }}>
          ⚠️ Validation Errors ({validationResult.errors.length} total)
        </h4>
        
        {Object.entries(errorsByLayer).map(([layer, errors]) => (
          <div key={layer} style={{ marginBottom: '10px' }}>
            <strong style={{ color: '#721c24' }}>
              {layerConfigs.find(c => c.layer === layer)?.title}: {errors.length} errors
            </strong>
            <ul style={{ margin: '5px 0', paddingLeft: '20px', fontSize: '12px' }}>
              {errors.slice(0, 5).map((error, index) => (
                <li key={index} style={{ color: '#721c24' }}>
                  ({error.x}, {error.y}): {error.reason}
                </li>
              ))}
              {errors.length > 5 && (
                <li style={{ color: '#6c757d' }}>
                  ... and {errors.length - 5} more
                </li>
              )}
            </ul>
          </div>
        ))}
      </div>
    );
  };

  return (
    <div style={{
      minHeight: '100vh',
      backgroundColor: '#f8f9fa',
      padding: '20px'
    }}>
      <div style={{
        maxWidth: '1200px',
        margin: '0 auto'
      }}>
        <ToolBar />
        
        {/* API Status */}
        {apiState.error && (
          <div style={{
            padding: '15px',
            backgroundColor: '#f8d7da',
            border: '1px solid #f5c6cb',
            borderRadius: '4px',
            marginBottom: '20px',
            color: '#721c24',
          }}>
            <strong>❌ API Error:</strong> {apiState.error}
          </div>
        )}

        {apiState.lastSaved && (
          <div style={{
            padding: '15px',
            backgroundColor: '#d4edda',
            border: '1px solid #c3e6cb',
            borderRadius: '4px',
            marginBottom: '20px',
            color: '#155724',
          }}>
            <strong>✅ Last Saved:</strong> "{apiState.lastSaved.name}" (ID: {apiState.lastSaved.id})
          </div>
        )}

        {uiState.showErrors && <ErrorSummary />}
        
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'auto 300px',
          gap: '20px',
          alignItems: 'start'
        }}>
          {/* Main editing area */}
          <div>
            {layerConfigs.map(({ layer, title, color, description }) => (
              <div key={layer}>
                {layer === 'ground' && <GroundGenerator />}
                
                <div style={{
                  backgroundColor: 'white',
                  border: '1px solid #dee2e6',
                  borderRadius: '8px',
                  padding: '15px',
                  marginBottom: '20px',
                  boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
                }}>
                  <div style={{
                    marginBottom: '10px',
                    paddingBottom: '10px',
                    borderBottom: '1px solid #eee'
                  }}>
                    <h3 style={{
                      margin: '0 0 5px 0',
                      color: color,
                      fontSize: '18px'
                    }}>
                      {title}
                    </h3>
                    <p style={{
                      margin: 0,
                      fontSize: '12px',
                      color: '#6c757d'
                    }}>
                      {description}
                    </p>
                  </div>
                  
                  <LayerEditor
                    layer={layer}
                    title={title}
                    color={color}
                  />
                </div>
              </div>
            ))}
          </div>

          {/* Info sidebar */}
          <div style={{
            backgroundColor: 'white',
            border: '1px solid #dee2e6',
            borderRadius: '8px',
            padding: '15px',
            height: 'fit-content',
            position: 'sticky',
            top: '20px'
          }}>
            <h3 style={{ margin: '0 0 15px 0', fontSize: '16px' }}>
              📊 Template Info
            </h3>
            
            <div style={{ fontSize: '14px', lineHeight: '1.5' }}>
              <div style={{ marginBottom: '10px' }}>
                <strong>Dimensions:</strong> {template.width} × {template.height}
              </div>

              {/* API Status */}
              <div style={{ marginBottom: '15px' }}>
                <strong>Backend Status:</strong><br/>
                <div style={{ fontSize: '12px', marginTop: '5px' }}>
                  {apiState.isLoading && (
                    <div style={{ color: '#007bff' }}>🔄 Loading...</div>
                  )}
                  {apiState.lastSaved ? (
                    <div style={{ color: '#28a745' }}>
                      ✅ Saved: "{apiState.lastSaved.name}"<br/>
                      <span style={{ color: '#666' }}>ID: {apiState.lastSaved.id}</span>
                    </div>
                  ) : (
                    <div style={{ color: '#6c757d' }}>💾 Not saved</div>
                  )}
                  {apiState.error && (
                    <div style={{ color: '#dc3545', marginTop: '5px' }}>
                      ❌ {apiState.error}
                    </div>
                  )}
                </div>
              </div>
              
              {uiState.hoveredCell && (
                <div style={{ marginBottom: '15px' }}>
                  <strong>Hovered Cell:</strong> ({uiState.hoveredCell.x}, {uiState.hoveredCell.y})
                  <div style={{ fontSize: '12px', marginTop: '5px' }}>
                    Ground: {template.ground[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    Static: {template.static[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    Turret: {template.turret[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    MobGround: {template.mobGround[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    MobAir: {template.mobAir[uiState.hoveredCell.y][uiState.hoveredCell.x]}
                  </div>
                </div>
              )}

              <div style={{ marginBottom: '15px' }}>
                <strong>Edit Mode:</strong> Click any cell to toggle (0 ↔ 1)<br/>
                <strong>All layers:</strong> Directly editable
              </div>

              <div style={{ 
                padding: '10px',
                backgroundColor: '#f8f9fa',
                borderRadius: '4px',
                fontSize: '12px'
              }}>
                <h4 style={{ margin: '0 0 8px 0', fontSize: '13px' }}>
                  🎨 Usage Tips:
                </h4>
                <ul style={{ margin: 0, paddingLeft: '15px' }}>
                  <li>All layers are always editable</li>
                  <li>Click cells to toggle between 0 and 1</li>
                  <li>Drag to paint/erase multiple cells</li>
                  <li>Red borders indicate rule violations</li>
                  <li>Must fix all errors before export</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};