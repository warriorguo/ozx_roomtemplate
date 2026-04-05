import React, { useState, useEffect, useCallback } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import { templateApi, type BackendListResponse, type ListTemplatesParams, type TemplateSummary, ApiError } from '../../services/api';
import { formatTemplateInfo, generateDefaultTemplateName } from '../../services/templateConverter';

interface SaveLoadPanelProps {
  isOpen: boolean;
  onClose: () => void;
  mode: 'save' | 'load';
}

type RoomTypeFilter = 'all' | 'full' | 'bridge' | 'platform';
type DoorFilter = 'any' | 'connected' | 'disconnected';

interface Filters {
  roomType: RoomTypeFilter;
  stageType: string;
  doors: {
    top: DoorFilter;
    right: DoorFilter;
    bottom: DoorFilter;
    left: DoorFilter;
  };
}

const initialFilters: Filters = {
  roomType: 'all',
  stageType: 'all',
  doors: {
    top: 'any',
    right: 'any',
    bottom: 'any',
    left: 'any',
  },
};

const roomTypeLabels: Record<string, string> = {
  full: 'Full Room',
  bridge: 'Bridge',
  platform: 'Platform',
};

const stageTypeLabels: Record<string, string> = {
  teaching: 'Teaching',
  building: 'Building',
  pressure: 'Pressure',
  peak: 'Peak',
  release: 'Release',
  boss: 'Boss',
};

// Three-state checkbox component: null (any) -> true (yes) -> false (no) -> null (any)
const TriStateCheckbox: React.FC<{
  value: boolean | null;
  onChange: (value: boolean | null) => void;
  label: string;
}> = ({ value, onChange, label }) => {
  const handleClick = () => {
    if (value === null) onChange(true);
    else if (value === true) onChange(false);
    else onChange(null);
  };

  return (
    <div
      onClick={handleClick}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '6px',
        cursor: 'pointer',
        padding: '4px 8px',
        borderRadius: '4px',
        backgroundColor: value === null ? 'transparent' : value ? '#d4edda' : '#f8d7da',
        border: '1px solid',
        borderColor: value === null ? '#ddd' : value ? '#28a745' : '#dc3545',
        transition: 'all 0.15s',
        userSelect: 'none',
      }}
    >
      <span style={{
        width: '16px',
        height: '16px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: '12px',
        fontWeight: 'bold',
        color: value === null ? '#999' : value ? '#28a745' : '#dc3545',
      }}>
        {value === null ? '-' : value ? '✓' : '✗'}
      </span>
      <span style={{ fontSize: '12px', color: '#333' }}>{label}</span>
    </div>
  );
};

export const SaveLoadPanel: React.FC<SaveLoadPanelProps> = ({ isOpen, onClose, mode }) => {
  const {
    template,
    apiState,
    saveTemplate,
    loadTemplateFromBackend,
    deleteTemplateFromBackend,
    clearApiError
  } = useNewTemplateStore();

  const [templateName, setTemplateName] = useState('');
  const [templateList, setTemplateList] = useState<BackendListResponse | null>(null);
  const [listLoading, setListLoading] = useState(false);
  const [listError, setListError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState<Filters>(initialFilters);
  const [showFilters, setShowFilters] = useState(true);

  // Initialize template name when component opens
  useEffect(() => {
    if (isOpen && !templateName) {
      if (apiState.lastSaved?.name) {
        setTemplateName(apiState.lastSaved.name);
      } else {
        setTemplateName(generateDefaultTemplateName());
      }
    }
  }, [isOpen, apiState.lastSaved?.name, templateName]);

  const buildFilterParams = useCallback((): ListTemplatesParams => {
    const params: ListTemplatesParams = {};

    if (searchTerm) params.name_like = searchTerm;
    if (filters.roomType !== 'all') params.room_type = filters.roomType;
    if (filters.stageType !== 'all') params.stage_type = filters.stageType;

    // Door connectivity
    if (filters.doors.top === 'connected') params.top_door_connected = true;
    if (filters.doors.top === 'disconnected') params.top_door_connected = false;
    if (filters.doors.right === 'connected') params.right_door_connected = true;
    if (filters.doors.right === 'disconnected') params.right_door_connected = false;
    if (filters.doors.bottom === 'connected') params.bottom_door_connected = true;
    if (filters.doors.bottom === 'disconnected') params.bottom_door_connected = false;
    if (filters.doors.left === 'connected') params.left_door_connected = true;
    if (filters.doors.left === 'disconnected') params.left_door_connected = false;

    return params;
  }, [searchTerm, filters]);

  const loadTemplateList = useCallback(async () => {
    setListLoading(true);
    setListError(null);

    try {
      const params = buildFilterParams();
      const response = await templateApi.listTemplates(params);
      setTemplateList(response);
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to load template list';
      setListError(errorMessage);
    } finally {
      setListLoading(false);
    }
  }, [buildFilterParams]);

  // Load template list when opening in load mode or filters change
  useEffect(() => {
    if (isOpen && mode === 'load') {
      loadTemplateList();
    }
  }, [isOpen, mode, loadTemplateList]);

  const handleSave = async () => {
    if (!templateName.trim()) {
      return;
    }

    try {
      await saveTemplate(templateName.trim());
      if (!apiState.error) {
        onClose();
      }
    } catch (error) {
      // Error is handled by the store
    }
  };

  const handleLoad = async (templateId: string) => {
    try {
      await loadTemplateFromBackend(templateId);
      if (!apiState.error) {
        onClose();
      }
    } catch (error) {
      // Error is handled by the store
    }
  };

  const handleDelete = async (templateId: string, templateName: string) => {
    const confirmed = confirm(`Are you sure you want to delete "${templateName}"?\n\nThis action cannot be undone.`);
    if (!confirmed) return;

    try {
      await deleteTemplateFromBackend(templateId);
      if (!apiState.error) {
        loadTemplateList();
      }
    } catch (error) {
      // Error is handled by the store
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    loadTemplateList();
  };

  const resetFilters = () => {
    setFilters(initialFilters);
    setSearchTerm('');
  };

  const renderDoorInfo = (item: TemplateSummary) => {
    // Use open_doors bitmask (Top=1, Right=2, Bottom=4, Left=8) if available,
    // fall back to doors_connected for backward compat
    const bitmask = item.open_doors;
    const doors = item.doors_connected;
    if (bitmask == null && !doors) return null;

    const doorIcons = bitmask != null ? [
      { key: 'top', label: 'T', connected: (bitmask & 1) !== 0 },
      { key: 'right', label: 'R', connected: (bitmask & 2) !== 0 },
      { key: 'bottom', label: 'B', connected: (bitmask & 4) !== 0 },
      { key: 'left', label: 'L', connected: (bitmask & 8) !== 0 },
    ] : [
      { key: 'top', label: 'T', connected: doors!.top },
      { key: 'right', label: 'R', connected: doors!.right },
      { key: 'bottom', label: 'B', connected: doors!.bottom },
      { key: 'left', label: 'L', connected: doors!.left },
    ];

    return (
      <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
        {doorIcons.map(door => (
          <span
            key={door.key}
            style={{
              padding: '2px 6px',
              borderRadius: '3px',
              fontSize: '11px',
              fontWeight: 'bold',
              backgroundColor: door.connected ? '#28a745' : '#6c757d',
              color: 'white',
            }}
            title={`${door.key.charAt(0).toUpperCase() + door.key.slice(1)} door: ${door.connected ? 'Connected' : 'Disconnected'}`}
          >
            {door.label}
          </span>
        ))}
      </div>
    );
  };

  const renderStageType = (item: TemplateSummary) => {
    const stageType = item.stage_type;
    if (!stageType) return <span style={{ color: '#999', fontSize: '11px' }}>N/A</span>;

    return (
      <span
        style={{
          padding: '2px 6px',
          borderRadius: '3px',
          fontSize: '11px',
          backgroundColor: '#17a2b8',
          color: 'white',
        }}
      >
        {stageTypeLabels[stageType] || stageType}
      </span>
    );
  };

  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(0, 0, 0, 0.5)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
    }}>
      <div style={{
        backgroundColor: 'white',
        borderRadius: '8px',
        width: '900px',
        maxWidth: '95vw',
        maxHeight: '90vh',
        overflow: 'hidden',
        boxShadow: '0 10px 30px rgba(0,0,0,0.3)',
        display: 'flex',
        flexDirection: 'column',
      }}>
        {/* Header */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '20px',
          borderBottom: '1px solid #eee',
          flexShrink: 0,
        }}>
          <h2 style={{ margin: 0, fontSize: '20px', color: '#333' }}>
            {mode === 'save' ? 'Save Template' : 'Load Template'}
          </h2>

          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              fontSize: '20px',
              cursor: 'pointer',
              padding: '5px',
            }}
          >
            X
          </button>
        </div>

        {/* Content */}
        <div style={{
          padding: '20px',
          overflowY: 'auto',
          flex: 1,
        }}>
          {/* Error display */}
          {(apiState.error || listError) && (
            <div style={{
              padding: '10px',
              backgroundColor: '#f8d7da',
              border: '1px solid #f5c6cb',
              borderRadius: '4px',
              marginBottom: '20px',
              color: '#721c24',
            }}>
              {apiState.error || listError}
              <button
                onClick={() => {
                  clearApiError();
                  setListError(null);
                }}
                style={{
                  float: 'right',
                  background: 'none',
                  border: 'none',
                  color: '#721c24',
                  cursor: 'pointer',
                }}
              >
                X
              </button>
            </div>
          )}

          {/* Success message */}
          {apiState.lastSaved && mode === 'save' && (
            <div style={{
              padding: '10px',
              backgroundColor: '#d4edda',
              border: '1px solid #c3e6cb',
              borderRadius: '4px',
              marginBottom: '20px',
              color: '#155724',
            }}>
              Template "{apiState.lastSaved.name}" saved successfully!
            </div>
          )}

          {/* Save Mode */}
          {mode === 'save' && (
            <div>
              <div style={{ marginBottom: '20px' }}>
                <label style={{
                  display: 'block',
                  marginBottom: '5px',
                  fontWeight: 'bold',
                }}>
                  Template Name:
                </label>
                <input
                  type="text"
                  value={templateName}
                  onChange={(e) => setTemplateName(e.target.value)}
                  placeholder="Enter template name..."
                  style={{
                    width: '100%',
                    padding: '10px',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                    fontSize: '14px',
                  }}
                />
              </div>

              <div style={{
                padding: '15px',
                backgroundColor: '#f8f9fa',
                borderRadius: '4px',
                marginBottom: '20px',
                fontSize: '14px',
              }}>
                <strong>Template Info:</strong><br/>
                Size: {template.width} x {template.height}<br/>
                Layers: Ground, Static, Chaser, Zoner, DPS, MainPath, MobAir
              </div>

              <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
                <button
                  onClick={onClose}
                  style={{
                    padding: '10px 20px',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                    backgroundColor: 'white',
                    cursor: 'pointer',
                  }}
                >
                  Cancel
                </button>
                <button
                  onClick={handleSave}
                  disabled={!templateName.trim() || apiState.isLoading}
                  style={{
                    padding: '10px 20px',
                    border: 'none',
                    borderRadius: '4px',
                    backgroundColor: '#007bff',
                    color: 'white',
                    cursor: templateName.trim() && !apiState.isLoading ? 'pointer' : 'not-allowed',
                    opacity: templateName.trim() && !apiState.isLoading ? 1 : 0.6,
                  }}
                >
                  {apiState.isLoading ? 'Saving...' : 'Save Template'}
                </button>
              </div>
            </div>
          )}

          {/* Load Mode */}
          {mode === 'load' && (
            <div>
              {/* Search and Filter Toggle */}
              <div style={{ marginBottom: '15px' }}>
                <form onSubmit={handleSearch} style={{ display: 'flex', gap: '10px', marginBottom: '10px' }}>
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Search templates by name..."
                    style={{
                      flex: 1,
                      padding: '10px',
                      border: '1px solid #ddd',
                      borderRadius: '4px',
                      fontSize: '14px',
                    }}
                  />
                  <button
                    type="submit"
                    disabled={listLoading}
                    style={{
                      padding: '10px 20px',
                      border: 'none',
                      borderRadius: '4px',
                      backgroundColor: '#007bff',
                      color: 'white',
                      cursor: listLoading ? 'not-allowed' : 'pointer',
                      opacity: listLoading ? 0.6 : 1,
                    }}
                  >
                    Search
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowFilters(!showFilters)}
                    style={{
                      padding: '10px 15px',
                      border: '1px solid #ddd',
                      borderRadius: '4px',
                      backgroundColor: showFilters ? '#e9ecef' : 'white',
                      cursor: 'pointer',
                    }}
                  >
                    Filters {showFilters ? '[-]' : '[+]'}
                  </button>
                </form>

                {/* Filter Panel */}
                {showFilters && (
                  <div style={{
                    padding: '15px',
                    backgroundColor: '#f8f9fa',
                    borderRadius: '4px',
                    border: '1px solid #e9ecef',
                  }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '15px' }}>
                      <strong style={{ fontSize: '14px' }}>Filters</strong>
                      <button
                        onClick={resetFilters}
                        style={{
                          padding: '4px 10px',
                          border: '1px solid #ddd',
                          borderRadius: '4px',
                          backgroundColor: 'white',
                          cursor: 'pointer',
                          fontSize: '12px',
                        }}
                      >
                        Reset All
                      </button>
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '20px' }}>
                      {/* Room Type Filter */}
                      <div>
                        <div style={{ fontWeight: 'bold', fontSize: '13px', marginBottom: '8px', color: '#555' }}>Room Type</div>
                        <select
                          value={filters.roomType}
                          onChange={(e) => setFilters({ ...filters, roomType: e.target.value as RoomTypeFilter })}
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ddd',
                            borderRadius: '4px',
                            fontSize: '13px',
                          }}
                        >
                          <option value="all">All Types</option>
                          <option value="full">Full Room</option>
                          <option value="bridge">Bridge</option>
                          <option value="platform">Platform</option>
                        </select>
                      </div>

                      {/* Door Connectivity Filter */}
                      <div>
                        <div style={{ fontWeight: 'bold', fontSize: '13px', marginBottom: '8px', color: '#555' }}>Door Connectivity</div>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '6px' }}>
                          {(['top', 'right', 'bottom', 'left'] as const).map(door => {
                            const doorValue = filters.doors[door] === 'any' ? null : filters.doors[door] === 'connected';
                            return (
                              <TriStateCheckbox
                                key={door}
                                label={door.charAt(0).toUpperCase() + door.slice(1)}
                                value={doorValue}
                                onChange={(val) => setFilters({
                                  ...filters,
                                  doors: {
                                    ...filters.doors,
                                    [door]: val === null ? 'any' : val ? 'connected' : 'disconnected'
                                  }
                                })}
                              />
                            );
                          })}
                        </div>
                      </div>

                      {/* Stage Type Filter */}
                      <div>
                        <div style={{ fontWeight: 'bold', fontSize: '13px', marginBottom: '8px', color: '#555' }}>Stage Type</div>
                        <select
                          value={filters.stageType}
                          onChange={(e) => setFilters({ ...filters, stageType: e.target.value })}
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ddd',
                            borderRadius: '4px',
                            fontSize: '13px',
                          }}
                        >
                          <option value="all">All Types</option>
                          <option value="teaching">Teaching</option>
                          <option value="building">Building</option>
                          <option value="pressure">Pressure</option>
                          <option value="peak">Peak</option>
                          <option value="release">Release</option>
                          <option value="boss">Boss</option>
                        </select>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {/* Template List */}
              {listLoading && (
                <div style={{
                  textAlign: 'center',
                  padding: '40px',
                  color: '#666',
                }}>
                  Loading templates...
                </div>
              )}

              {templateList && !listLoading && (
                <div>
                  <div style={{
                    marginBottom: '15px',
                    color: '#666',
                    fontSize: '14px',
                  }}>
                    Found {templateList.total} template(s)
                  </div>

                  {!templateList.items || templateList.items.length === 0 ? (
                    <div style={{
                      textAlign: 'center',
                      padding: '40px',
                      color: '#666',
                    }}>
                      No templates found.
                    </div>
                  ) : (
                    <div style={{ maxHeight: '450px', overflowY: 'auto' }}>
                      {templateList.items.map((item) => {
                        const info = formatTemplateInfo(item as any);
                        return (
                          <div
                            key={item.id}
                            style={{
                              border: '1px solid #ddd',
                              borderRadius: '4px',
                              padding: '15px',
                              marginBottom: '10px',
                              cursor: 'pointer',
                              transition: 'background-color 0.2s',
                            }}
                            onMouseEnter={(e) => {
                              e.currentTarget.style.backgroundColor = '#f8f9fa';
                            }}
                            onMouseLeave={(e) => {
                              e.currentTarget.style.backgroundColor = 'white';
                            }}
                            onClick={() => handleLoad(item.id)}
                          >
                            <div style={{
                              display: 'flex',
                              gap: '15px',
                              alignItems: 'flex-start',
                            }}>
                              {/* Thumbnail */}
                              {item.thumbnail ? (
                                <div style={{
                                  flexShrink: 0,
                                  width: '80px',
                                  height: '80px',
                                  border: '1px solid #ddd',
                                  borderRadius: '4px',
                                  overflow: 'hidden',
                                  backgroundColor: '#f8f9fa',
                                }}>
                                  <img
                                    src={item.thumbnail}
                                    alt={`${info.displayName} thumbnail`}
                                    style={{
                                      width: '100%',
                                      height: '100%',
                                      objectFit: 'cover',
                                      backgroundColor: '#fff',
                                    }}
                                  />
                                </div>
                              ) : (
                                <div style={{
                                  flexShrink: 0,
                                  width: '80px',
                                  height: '80px',
                                  border: '1px solid #ddd',
                                  borderRadius: '4px',
                                  backgroundColor: '#f0f0f0',
                                  display: 'flex',
                                  alignItems: 'center',
                                  justifyContent: 'center',
                                  fontSize: '12px',
                                  color: '#666',
                                  textAlign: 'center',
                                }}>
                                  No<br/>Preview
                                </div>
                              )}

                              {/* Template Info */}
                              <div style={{ flex: 1, minWidth: 0 }}>
                                <div style={{
                                  display: 'flex',
                                  justifyContent: 'space-between',
                                  alignItems: 'flex-start',
                                  marginBottom: '8px',
                                }}>
                                  <div style={{ flex: 1, minWidth: 0 }}>
                                    <strong style={{
                                      fontSize: '16px',
                                      display: 'block',
                                      whiteSpace: 'nowrap',
                                      overflow: 'hidden',
                                      textOverflow: 'ellipsis',
                                    }}>
                                      {info.displayName}
                                    </strong>
                                    <div style={{
                                      fontSize: '12px',
                                      color: '#666',
                                      marginTop: '4px',
                                    }}>
                                      {item.width}x{item.height} | Created {info.created}
                                    </div>
                                  </div>
                                  <div style={{ display: 'flex', gap: '8px', marginLeft: '10px', flexShrink: 0 }}>
                                    <button
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        handleLoad(item.id);
                                      }}
                                      disabled={apiState.isLoading}
                                      style={{
                                        padding: '6px 12px',
                                        border: 'none',
                                        borderRadius: '4px',
                                        backgroundColor: '#28a745',
                                        color: 'white',
                                        cursor: apiState.isLoading ? 'not-allowed' : 'pointer',
                                        fontSize: '12px',
                                        opacity: apiState.isLoading ? 0.6 : 1,
                                      }}
                                    >
                                      {apiState.isLoading ? 'Loading...' : 'Load'}
                                    </button>
                                    <button
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        handleDelete(item.id, item.name);
                                      }}
                                      disabled={apiState.isLoading}
                                      style={{
                                        padding: '6px 12px',
                                        border: 'none',
                                        borderRadius: '4px',
                                        backgroundColor: '#dc3545',
                                        color: 'white',
                                        cursor: apiState.isLoading ? 'not-allowed' : 'pointer',
                                        fontSize: '12px',
                                        opacity: apiState.isLoading ? 0.6 : 1,
                                      }}
                                    >
                                      Delete
                                    </button>
                                  </div>
                                </div>

                                {/* Additional Info Row */}
                                <div style={{
                                  display: 'flex',
                                  gap: '20px',
                                  alignItems: 'center',
                                  flexWrap: 'wrap',
                                  marginTop: '8px',
                                }}>
                                  {/* Room Type */}
                                  <div style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
                                    <span style={{ fontSize: '12px', color: '#666' }}>Type:</span>
                                    <span style={{
                                      padding: '2px 8px',
                                      borderRadius: '3px',
                                      fontSize: '11px',
                                      backgroundColor: '#6f42c1',
                                      color: 'white',
                                    }}>
                                      {item.room_type ? roomTypeLabels[item.room_type] || item.room_type : 'N/A'}
                                    </span>
                                  </div>

                                  {/* Doors */}
                                  <div style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
                                    <span style={{ fontSize: '12px', color: '#666' }}>Doors:</span>
                                    {renderDoorInfo(item)}
                                  </div>

                                  {/* Stage Type */}
                                  <div style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
                                    <span style={{ fontSize: '12px', color: '#666' }}>Stage:</span>
                                    {renderStageType(item)}
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
