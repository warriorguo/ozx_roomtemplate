import { useState, useEffect } from 'react';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import { SaveLoadPanel } from './SaveLoadPanel';
import { frontendToBackendPayload } from '../../services/templateConverter';

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
    createNewTemplate,
    toggleErrorDisplay,
    toggleAcceptPaste,
    loadTemplateFromJSON,
  } = useNewTemplateStore();

  const [showNewDialog, setShowNewDialog] = useState(false);
  const [showSavePanel, setShowSavePanel] = useState(false);
  const [showLoadPanel, setShowLoadPanel] = useState(false);
  
  const validationResult = uiState.validationResult;

  const handleNewTemplate = (width: number, height: number) => {
    createNewTemplate(width, height);
  };

  const copyToClipboard = async () => {
    try {
      // 使用后端格式导出，包含 payload 包装和所有字段（包括 roomType）
      const payload = frontendToBackendPayload(template, 'exported-template');
      const exportData = {
        name: payload.meta.name,
        payload: payload
      };

      const jsonString = JSON.stringify(exportData, null, 2);
      await navigator.clipboard.writeText(jsonString);
      alert('Template JSON copied to clipboard!');
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      alert('Failed to copy to clipboard. Please try again.');
    }
  };

  // Handle paste events
  const handlePaste = async (event: ClipboardEvent) => {
    if (!uiState.acceptPaste) {
      return;
    }

    try {
      const clipboardText = event.clipboardData?.getData('text');
      if (!clipboardText) {
        return;
      }

      // Parse JSON
      let jsonData;
      try {
        jsonData = JSON.parse(clipboardText);
      } catch (parseError) {
        alert('Invalid JSON format in clipboard');
        return;
      }

      // Load template from JSON
      await loadTemplateFromJSON(jsonData);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load template from clipboard';
      alert(`Error: ${errorMessage}`);
    }
  };

  // Add/remove paste event listener based on acceptPaste state
  useEffect(() => {
    if (uiState.acceptPaste) {
      document.addEventListener('paste', handlePaste);
    } else {
      document.removeEventListener('paste', handlePaste);
    }

    return () => {
      document.removeEventListener('paste', handlePaste);
    };
  }, [uiState.acceptPaste, loadTemplateFromJSON]);

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
            onClick={() => setShowSavePanel(true)}
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
            💾 Save
          </button>

          <button
            onClick={() => setShowLoadPanel(true)}
            style={{
              padding: '8px 16px',
              backgroundColor: '#2196F3',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold'
            }}
          >
            📁 Load
          </button>

          <button
            onClick={copyToClipboard}
            style={{
              padding: '8px 16px',
              backgroundColor: '#17A2B8',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold'
            }}
          >
            📋 Copy JSON
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
            backgroundColor: validationResult?.isValid !== false ? '#e8f5e8' : '#ffebee',
            color: validationResult?.isValid !== false ? '#2e7d32' : '#c62828',
            borderRadius: '4px',
            fontSize: '14px',
            fontWeight: 'bold'
          }}>
            {validationResult?.isValid !== false ? '✓ Valid' : '✗ Invalid'}
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

          <label style={{ display: 'flex', alignItems: 'center', gap: '5px', fontSize: '14px' }}>
            <input
              type="checkbox"
              checked={uiState.acceptPaste}
              onChange={toggleAcceptPaste}
            />
            Accept Paste
          </label>
        </div>
      </div>

      <NewTemplateDialog
        isOpen={showNewDialog}
        onClose={() => setShowNewDialog(false)}
        onConfirm={handleNewTemplate}
      />

      <SaveLoadPanel
        isOpen={showSavePanel}
        onClose={() => setShowSavePanel(false)}
        mode="save"
      />

      <SaveLoadPanel
        isOpen={showLoadPanel}
        onClose={() => setShowLoadPanel(false)}
        mode="load"
      />
    </div>
  );
};