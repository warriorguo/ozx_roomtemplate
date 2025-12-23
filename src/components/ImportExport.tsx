import { useState, useRef } from 'react';
import { useTemplateStore } from '../store/templateStore';
import { downloadTemplate, downloadSeparateLayers, copyToClipboard, importTemplate } from '../utils/fileUtils';

export const ImportExport: React.FC = () => {
  const { template, loadTemplate } = useTemplateStore();
  const [message, setMessage] = useState<string>('');
  const [isError, setIsError] = useState<boolean>(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const showMessage = (msg: string, error: boolean = false) => {
    setMessage(msg);
    setIsError(error);
    setTimeout(() => setMessage(''), 3000);
  };

  const handleExportSingle = () => {
    downloadTemplate(template);
    showMessage('Template exported successfully!');
  };

  const handleExportSeparate = () => {
    downloadSeparateLayers(template);
    showMessage('Layer files exported successfully!');
  };

  const handleCopyToClipboard = async () => {
    try {
      await copyToClipboard(template);
      showMessage('Template copied to clipboard!');
    } catch (error) {
      showMessage('Failed to copy to clipboard', true);
    }
  };

  const handleImport = () => {
    fileInputRef.current?.click();
  };

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    try {
      const { template: importedTemplate, warnings } = await importTemplate(file);
      loadTemplate(importedTemplate);
      
      let msg = 'Template imported successfully!';
      if (warnings.length > 0) {
        msg += ` (${warnings.join(' ')})`;
      }
      showMessage(msg);
    } catch (error) {
      showMessage(`Import failed: ${error}`, true);
    }

    event.target.value = '';
  };

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
      <span style={{ fontWeight: 'bold', marginRight: '10px' }}>File:</span>
      
      <button
        onClick={handleImport}
        style={{
          padding: '8px 16px',
          backgroundColor: '#4CAF50',
          color: 'white',
          border: 'none',
          borderRadius: '4px',
          cursor: 'pointer',
          fontSize: '14px'
        }}
      >
        Import JSON
      </button>

      <button
        onClick={handleExportSingle}
        style={{
          padding: '8px 16px',
          backgroundColor: '#2196F3',
          color: 'white',
          border: 'none',
          borderRadius: '4px',
          cursor: 'pointer',
          fontSize: '14px'
        }}
      >
        Export Single
      </button>

      <button
        onClick={handleExportSeparate}
        style={{
          padding: '8px 16px',
          backgroundColor: '#FF9800',
          color: 'white',
          border: 'none',
          borderRadius: '4px',
          cursor: 'pointer',
          fontSize: '14px'
        }}
      >
        Export Layers
      </button>

      <button
        onClick={handleCopyToClipboard}
        style={{
          padding: '8px 16px',
          backgroundColor: '#9C27B0',
          color: 'white',
          border: 'none',
          borderRadius: '4px',
          cursor: 'pointer',
          fontSize: '14px'
        }}
      >
        Copy JSON
      </button>

      <input
        ref={fileInputRef}
        type="file"
        accept=".json"
        onChange={handleFileSelect}
        style={{ display: 'none' }}
      />

      {message && (
        <div style={{
          padding: '5px 10px',
          borderRadius: '4px',
          backgroundColor: isError ? '#ffebee' : '#e8f5e8',
          color: isError ? '#c62828' : '#2e7d32',
          fontSize: '12px',
          marginLeft: '10px'
        }}>
          {message}
        </div>
      )}
    </div>
  );
};