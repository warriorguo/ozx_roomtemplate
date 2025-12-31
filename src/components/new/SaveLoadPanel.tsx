import React, { useState, useEffect } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import { templateApi, type BackendListResponse, ApiError } from '../../services/api';
import { formatTemplateInfo, generateDefaultTemplateName } from '../../services/templateConverter';

interface SaveLoadPanelProps {
  isOpen: boolean;
  onClose: () => void;
  mode: 'save' | 'load';
}

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

  // Load template list when opening in load mode
  useEffect(() => {
    if (isOpen && mode === 'load') {
      loadTemplateList();
    }
  }, [isOpen, mode]);

  const loadTemplateList = async () => {
    setListLoading(true);
    setListError(null);
    
    try {
      const params = searchTerm ? { name_like: searchTerm } : undefined;
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
  };

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
    // Confirm deletion
    const confirmed = confirm(`Are you sure you want to delete "${templateName}"?\n\nThis action cannot be undone.`);
    if (!confirmed) return;

    try {
      await deleteTemplateFromBackend(templateId);
      // Reload the template list after successful deletion
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
        width: '600px',
        maxWidth: '90vw',
        maxHeight: '80vh',
        overflow: 'hidden',
        boxShadow: '0 10px 30px rgba(0,0,0,0.3)',
      }}>
        {/* Header */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '20px',
          borderBottom: '1px solid #eee',
        }}>
          <h2 style={{ margin: 0, fontSize: '20px', color: '#333' }}>
            {mode === 'save' ? '💾 Save Template' : '📁 Load Template'}
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
            ✕
          </button>
        </div>

        {/* Content */}
        <div style={{
          padding: '20px',
          maxHeight: '60vh',
          overflowY: 'auto',
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
                ✕
              </button>
            </div>
          )}

          {/* Success message */}
          {apiState.lastSaved && (
            <div style={{
              padding: '10px',
              backgroundColor: '#d4edda',
              border: '1px solid #c3e6cb',
              borderRadius: '4px',
              marginBottom: '20px',
              color: '#155724',
            }}>
              ✅ Template "{apiState.lastSaved.name}" saved successfully!
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
                Size: {template.width} × {template.height}<br/>
                Layers: Ground, Static, Turret, MobGround, MobAir
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
              
              {/* Search */}
              <form onSubmit={handleSearch} style={{ marginBottom: '20px' }}>
                <div style={{ display: 'flex', gap: '10px' }}>
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Search templates..."
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
                    🔍 Search
                  </button>
                </div>
              </form>

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

                  {templateList.items.length === 0 ? (
                    <div style={{
                      textAlign: 'center',
                      padding: '40px',
                      color: '#666',
                    }}>
                      No templates found.
                    </div>
                  ) : (
                    <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
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
                                  width: '60px',
                                  height: '60px',
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
                                  width: '60px',
                                  height: '60px',
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
                                      {info.info} • {info.size} • Created {info.created}
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