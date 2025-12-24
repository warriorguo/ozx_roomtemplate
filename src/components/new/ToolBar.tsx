import { useState } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import { SaveLoadPanel } from './SaveLoadPanel';

interface NewTemplateDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (width: number, height: number) => void;
}

const NewTemplateDialog: React.FC<NewTemplateDialogProps> = ({ isOpen, onClose, onConfirm }) => {
  const [width, setWidth] = useState<string>('20');
  const [height, setHeight] = useState<string>('12');
  const [error, setError] = useState<string>('');

  if (!isOpen) return null;

  const handleConfirm = () => {
    const w = parseInt(width);
    const h = parseInt(height);

    if (isNaN(w) || isNaN(h)) {
      setError('Width and height must be valid numbers');
      return;
    }

    if (w < 1 || h < 1) {
      setError('Width and height must be at least 1');
      return;
    }

    if (w > 200 || h > 200) {
      setError('Width and height must be at most 200');
      return;
    }

    setError('');
    onConfirm(w, h);
    onClose();
  };

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
      zIndex: 1000
    }}>
      <div style={{
        backgroundColor: 'white',
        padding: '20px',
        borderRadius: '8px',
        minWidth: '300px',
        boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
      }}>
        <h3 style={{ margin: '0 0 20px 0' }}>Create New Template</h3>
        
        <div style={{ marginBottom: '15px' }}>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>
            Width:
          </label>
          <input
            type="number"
            value={width}
            onChange={(e) => setWidth(e.target.value)}
            min="1"
            max="200"
            style={{
              width: '100%',
              padding: '8px',
              border: '1px solid #ccc',
              borderRadius: '4px'
            }}
          />
        </div>

        <div style={{ marginBottom: '15px' }}>
          <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>
            Height:
          </label>
          <input
            type="number"
            value={height}
            onChange={(e) => setHeight(e.target.value)}
            min="1"
            max="200"
            style={{
              width: '100%',
              padding: '8px',
              border: '1px solid #ccc',
              borderRadius: '4px'
            }}
          />
        </div>

        {error && (
          <div style={{
            color: 'red',
            fontSize: '14px',
            marginBottom: '15px'
          }}>
            {error}
          </div>
        )}

        <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
          <button
            onClick={onClose}
            style={{
              padding: '8px 16px',
              border: '1px solid #ccc',
              backgroundColor: 'white',
              borderRadius: '4px',
              cursor: 'pointer'
            }}
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            style={{
              padding: '8px 16px',
              backgroundColor: '#4CAF50',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer'
            }}
          >
            Create
          </button>
        </div>
      </div>
    </div>
  );
};

export const ToolBar: React.FC = () => {
  const {
    template,
    uiState,
    apiState,
    createNewTemplate,
    toggleErrorDisplay,
    validateTemplateWithBackend,
  } = useNewTemplateStore();

  const [showNewDialog, setShowNewDialog] = useState(false);
  const [showSaveLoadPanel, setShowSaveLoadPanel] = useState(false);
  
  const validationResult = uiState.validationResult;
  const canExport = validationResult?.isValid ?? true; // 默认允许导出

  const handleNewTemplate = (width: number, height: number) => {
    createNewTemplate(width, height);
  };

  const exportTemplate = () => {
    // 允许导出即使有验证错误的模板，但给出警告
    if (validationResult && !validationResult.isValid) {
      const proceed = confirm(
        `Template has ${validationResult.errors.length} validation error(s). ` +
        'Do you want to export anyway?'
      );
      if (!proceed) return;
    }

    const jsonString = JSON.stringify(template, null, 2);
    const blob = new Blob([jsonString], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = 'template.json';
    link.click();
    
    URL.revokeObjectURL(url);
  };

  const copyToClipboard = async () => {
    // 允许复制即使有验证错误的模板，但给出警告
    if (validationResult && !validationResult.isValid) {
      const proceed = confirm(
        `Template has ${validationResult.errors.length} validation error(s). ` +
        'Do you want to copy anyway?'
      );
      if (!proceed) return;
    }

    try {
      const jsonString = JSON.stringify(template, null, 2);
      await navigator.clipboard.writeText(jsonString);
      alert('Template JSON copied to clipboard!');
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      alert('Failed to copy to clipboard. Please try again.');
    }
  };

  return (
    <div style={{
      padding: '20px',
      backgroundColor: '#fff',
      borderRadius: '8px',
      boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
      marginBottom: '20px'
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '20px',
        flexWrap: 'wrap'
      }}>
        {/* Template Actions */}
        <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
          <button
            onClick={() => setShowNewDialog(true)}
            style={{
              padding: '8px 16px',
              backgroundColor: '#4CAF50',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold'
            }}
          >
            📄 New
          </button>

          <button
            onClick={() => setShowSaveLoadPanel(true)}
            style={{
              padding: '8px 16px',
              backgroundColor: '#FF9800',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold'
            }}
          >
            💾 Save/Load
          </button>
          
          <button
            onClick={exportTemplate}
            disabled={!canExport}
            style={{
              padding: '8px 16px',
              backgroundColor: canExport ? '#2196F3' : '#ccc',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: canExport ? 'pointer' : 'not-allowed',
              fontWeight: 'bold'
            }}
          >
            📤 Export JSON
          </button>

          <button
            onClick={copyToClipboard}
            disabled={!canExport}
            style={{
              padding: '8px 16px',
              backgroundColor: canExport ? '#17A2B8' : '#ccc',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: canExport ? 'pointer' : 'not-allowed',
              fontWeight: 'bold'
            }}
          >
            📋 Copy JSON
          </button>

          <button
            onClick={() => validateTemplateWithBackend(true)}
            disabled={apiState.isLoading}
            style={{
              padding: '8px 16px',
              backgroundColor: '#9C27B0',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: apiState.isLoading ? 'not-allowed' : 'pointer',
              fontWeight: 'bold',
              opacity: apiState.isLoading ? 0.6 : 1,
            }}
          >
            🔍 {apiState.isLoading ? 'Validating...' : 'Validate'}
          </button>
        </div>

        {/* Status Display */}
        <div style={{ display: 'flex', gap: '15px', alignItems: 'center', marginLeft: 'auto' }}>
          <div style={{
            padding: '5px 10px',
            backgroundColor: '#e3f2fd',
            borderRadius: '4px',
            fontSize: '14px'
          }}>
            {template.width} × {template.height}
          </div>
          
          <div style={{
            padding: '5px 10px',
            backgroundColor: '#e8f5e8',
            borderRadius: '4px',
            fontSize: '14px'
          }}>
            Multi-layer editing
          </div>

          <div style={{
            padding: '5px 10px',
            backgroundColor: canExport ? '#e8f5e8' : '#ffebee',
            color: canExport ? '#2e7d32' : '#c62828',
            borderRadius: '4px',
            fontSize: '14px',
            fontWeight: 'bold'
          }}>
            {canExport ? '✓ Valid' : '✗ Invalid'}
            {validationResult && ` (${validationResult.errors.length} errors)`}
          </div>

          <label style={{ display: 'flex', alignItems: 'center', gap: '5px', fontSize: '14px' }}>
            <input
              type="checkbox"
              checked={uiState.showErrors}
              onChange={toggleErrorDisplay}
            />
            Show Errors
          </label>
        </div>
      </div>

      <NewTemplateDialog
        isOpen={showNewDialog}
        onClose={() => setShowNewDialog(false)}
        onConfirm={handleNewTemplate}
      />

      <SaveLoadPanel
        isOpen={showSaveLoadPanel}
        onClose={() => setShowSaveLoadPanel(false)}
      />
    </div>
  );
};