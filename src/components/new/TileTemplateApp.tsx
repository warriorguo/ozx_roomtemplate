import { ToolBar } from './ToolBar';
import { LayerEditor } from './LayerEditor';
import { CompositeLayerEditor } from './CompositeLayerEditor';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import type { LayerType } from '../../types/newTemplate';
import { ROOM_TYPES } from '../../types/newTemplate';

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
    layer: 'softEdge',
    title: 'Soft Edge (软边缘)',
    color: '#808080',
    description: 'Soft edge tiles - must be adjacent to ground but not overlap'
  },
  {
    layer: 'bridge',
    title: 'Bridge (桥梁)',
    color: '#9966CC',
    description: 'Bridge tiles that span unwalkable areas to connect walkable areas'
  },
  {
    layer: 'static',
    title: 'Static (静态物品)',
    color: '#FFA500',
    description: 'Static objects placement areas (requires walkable ground, not on bridge)'
  },
  {
    layer: 'turret',
    title: 'Turret (炮塔)',
    color: '#4169E1',
    description: 'Turret placement (requires walkable ground, not on bridge/static)'
  },
  {
    layer: 'mobGround',
    title: 'Mob Ground (地面怪)',
    color: '#FFD700',
    description: 'Ground mob spawns (requires walkable ground, not on bridge/static/turret)'
  },
  {
    layer: 'mobAir',
    title: 'Mob Air (飞行怪)',
    color: '#87CEEB',
    description: 'Air mob spawns (no constraints)'
  },
];

export const TileTemplateApp: React.FC = () => {
  const { uiState, template, apiState, toggleRoomAttribute, setRoomType } = useNewTemplateStore();

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

            {/* Composite Layer (总图层) */}
            <div style={{
              backgroundColor: 'white',
              border: '2px solid #9C27B0',
              borderRadius: '8px',
              padding: '15px',
              marginBottom: '20px',
              boxShadow: '0 2px 4px rgba(156, 39, 176, 0.15)'
            }}>
              <div style={{
                marginBottom: '10px',
                paddingBottom: '10px',
                borderBottom: '1px solid #eee'
              }}>
                <h3 style={{
                  margin: '0 0 5px 0',
                  color: '#9C27B0',
                  fontSize: '18px'
                }}>
                  🗂️ Composite Layer (总图层)
                </h3>
                <p style={{
                  margin: 0,
                  fontSize: '12px',
                  color: '#6c757d'
                }}>
                  Read-only view showing all layers combined with priority: MobAir &gt; MobGround &gt; Turret &gt; Static &gt; Bridge &gt; Ground
                </p>
              </div>

              <CompositeLayerEditor />
            </div>
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

              {/* Door States */}
              <div style={{ marginBottom: '15px' }}>
                <strong>🚪 Doors:</strong>
                <div style={{
                  marginTop: '8px',
                  padding: '10px',
                  backgroundColor: '#f8f9fa',
                  borderRadius: '6px',
                  border: '1px solid #dee2e6',
                  fontSize: '13px'
                }}>
                  <div style={{
                    display: 'grid',
                    gridTemplateColumns: '1fr 1fr',
                    gap: '8px'
                  }}>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '6px'
                    }}>
                      <span style={{
                        width: '12px',
                        height: '12px',
                        borderRadius: '50%',
                        backgroundColor: template.doors.top ? '#28a745' : '#dc3545',
                        display: 'inline-block'
                      }}></span>
                      <span>Top</span>
                    </div>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '6px'
                    }}>
                      <span style={{
                        width: '12px',
                        height: '12px',
                        borderRadius: '50%',
                        backgroundColor: template.doors.right ? '#28a745' : '#dc3545',
                        display: 'inline-block'
                      }}></span>
                      <span>Right</span>
                    </div>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '6px'
                    }}>
                      <span style={{
                        width: '12px',
                        height: '12px',
                        borderRadius: '50%',
                        backgroundColor: template.doors.bottom ? '#28a745' : '#dc3545',
                        display: 'inline-block'
                      }}></span>
                      <span>Bottom</span>
                    </div>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '6px'
                    }}>
                      <span style={{
                        width: '12px',
                        height: '12px',
                        borderRadius: '50%',
                        backgroundColor: template.doors.left ? '#28a745' : '#dc3545',
                        display: 'inline-block'
                      }}></span>
                      <span>Left</span>
                    </div>
                  </div>
                  <div style={{
                    marginTop: '8px',
                    paddingTop: '8px',
                    borderTop: '1px solid #dee2e6',
                    fontSize: '11px',
                    color: '#6c757d'
                  }}>
                    💡 Door opens when both middle cells = 1 in ground layer
                  </div>
                </div>
              </div>

              {/* Room Type */}
              <div style={{ marginBottom: '15px' }}>
                <strong>🏠 Room Type:</strong>
                <div style={{
                  marginTop: '8px',
                  padding: '10px',
                  backgroundColor: '#f8f9fa',
                  borderRadius: '6px',
                  border: '1px solid #dee2e6',
                  fontSize: '13px'
                }}>
                  <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '8px'
                  }}>
                    {ROOM_TYPES.map(({ type, label, description }) => (
                      <label
                        key={type}
                        style={{
                          display: 'flex',
                          alignItems: 'flex-start',
                          gap: '8px',
                          cursor: 'pointer',
                          padding: '6px',
                          borderRadius: '4px',
                          backgroundColor: template.roomType === type ? '#e7f3ff' : 'transparent',
                          border: template.roomType === type ? '1px solid #0066cc' : '1px solid transparent',
                          transition: 'all 0.2s',
                        }}
                        onMouseEnter={(e) => {
                          if (template.roomType !== type) {
                            e.currentTarget.style.backgroundColor = '#e9ecef';
                          }
                        }}
                        onMouseLeave={(e) => {
                          if (template.roomType !== type) {
                            e.currentTarget.style.backgroundColor = 'transparent';
                          }
                        }}
                      >
                        <input
                          type="radio"
                          name="roomType"
                          checked={template.roomType === type}
                          onChange={() => setRoomType(type)}
                          style={{
                            cursor: 'pointer',
                            marginTop: '2px',
                            accentColor: '#0066cc',
                          }}
                        />
                        <div style={{ flex: 1 }}>
                          <div style={{
                            fontWeight: template.roomType === type ? 'bold' : 'normal',
                            color: template.roomType === type ? '#0066cc' : '#212529',
                            marginBottom: '2px',
                          }}>
                            {label}
                          </div>
                          <div style={{
                            fontSize: '11px',
                            color: '#6c757d',
                            lineHeight: '1.3',
                          }}>
                            {description}
                          </div>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>
              </div>

              {/* Room Attributes */}
              <div style={{ marginBottom: '15px' }}>
                <strong>🏷️ Room Attributes:</strong>
                <div style={{
                  marginTop: '8px',
                  padding: '10px',
                  backgroundColor: '#f8f9fa',
                  borderRadius: '6px',
                  border: '1px solid #dee2e6',
                  fontSize: '13px'
                }}>
                  <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '6px'
                  }}>
                    {[
                      { key: 'boss' as const, label: '👹 Boss', color: '#dc3545' },
                      { key: 'elite' as const, label: '⚔️ Elite', color: '#fd7e14' },
                      { key: 'mob' as const, label: '🐛 Mob', color: '#6c757d' },
                      { key: 'treasure' as const, label: '💎 Treasure', color: '#ffc107' },
                      { key: 'teleport' as const, label: '🌀 Teleport', color: '#17a2b8' },
                      { key: 'story' as const, label: '📖 Story', color: '#6f42c1' },
                    ].map(({ key, label, color }) => (
                      <label
                        key={key}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '8px',
                          cursor: 'pointer',
                          padding: '4px',
                          borderRadius: '4px',
                          transition: 'background-color 0.2s',
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.backgroundColor = '#e9ecef';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.backgroundColor = 'transparent';
                        }}
                      >
                        <input
                          type="checkbox"
                          checked={template.attributes[key]}
                          onChange={() => toggleRoomAttribute(key)}
                          style={{
                            cursor: 'pointer',
                            accentColor: color,
                          }}
                        />
                        <span style={{
                          fontWeight: template.attributes[key] ? 'bold' : 'normal',
                          color: template.attributes[key] ? color : '#6c757d',
                        }}>
                          {label}
                        </span>
                      </label>
                    ))}
                  </div>
                </div>
              </div>

              {/* Thumbnail */}
              {apiState.lastSaved?.thumbnail && (
                <div style={{ marginBottom: '15px' }}>
                  <strong>Thumbnail:</strong><br/>
                  <div style={{ 
                    marginTop: '8px',
                    padding: '8px',
                    backgroundColor: '#f8f9fa',
                    borderRadius: '6px',
                    border: '1px solid #dee2e6',
                    textAlign: 'center'
                  }}>
                    <img
                      src={apiState.lastSaved.thumbnail}
                      alt="Template Thumbnail"
                      style={{
                        width: '120px',
                        height: '120px',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                        backgroundColor: '#fff',
                      }}
                    />
                    <div style={{ 
                      fontSize: '11px', 
                      color: '#666', 
                      marginTop: '4px' 
                    }}>
                      {apiState.lastSaved.name}
                    </div>
                  </div>
                </div>
              )}

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
                    <strong>Layer Values:</strong><br/>
                    Ground: {template.ground[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    Bridge: {template.bridge[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    Static: {template.static[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    Turret: {template.turret[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    MobGround: {template.mobGround[uiState.hoveredCell.y][uiState.hoveredCell.x]}<br/>
                    MobAir: {template.mobAir[uiState.hoveredCell.y][uiState.hoveredCell.x]}
                  </div>

                  {(() => {
                    const props = template.tileProperties[uiState.hoveredCell.y][uiState.hoveredCell.x];
                    if (!props) return null;

                    return (
                      <div style={{
                        marginTop: '10px',
                        padding: '8px',
                        backgroundColor: '#e3f2fd',
                        borderRadius: '4px',
                        border: '1px solid #90caf9',
                        fontSize: '11px',
                        lineHeight: '1.6'
                      }}>
                        <strong style={{ color: '#1976d2' }}>📊 Tile Properties:</strong><br/>
                        <div style={{ marginTop: '4px' }}>
                          <strong>Walkable:</strong> {props.walkable ? '✓ Yes' : '✗ No'}<br/>
                          <strong>Distance to Edge:</strong> {props.distToEdge}<br/>

                          <div style={{ marginTop: '4px' }}>
                            <strong>Wall Distances:</strong><br/>
                            &nbsp;&nbsp;Top: {props.distToTopWall} | Bottom: {props.distToBottomWall}<br/>
                            &nbsp;&nbsp;Left: {props.distToLeftWall} | Right: {props.distToRightWall}<br/>
                            &nbsp;&nbsp;Center: {props.distToCenter}
                          </div>

                          <div style={{ marginTop: '4px' }}>
                            <strong>Door Distances (BFS):</strong><br/>
                            &nbsp;&nbsp;Top: {props.distToTopDoor ?? '-'}<br/>
                            &nbsp;&nbsp;Bottom: {props.distToBottomDoor ?? '-'}<br/>
                            &nbsp;&nbsp;Left: {props.distToLeftDoor ?? '-'}<br/>
                            &nbsp;&nbsp;Right: {props.distToRightDoor ?? '-'}
                          </div>

                          <div style={{ marginTop: '4px' }}>
                            <strong>Feature Distances:</strong><br/>
                            &nbsp;&nbsp;Nearest Static: {props.distToNearStatic ?? '-'}<br/>
                            &nbsp;&nbsp;Nearest Turret: {props.distToNearTurret ?? '-'}
                          </div>
                        </div>
                      </div>
                    );
                  })()}
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