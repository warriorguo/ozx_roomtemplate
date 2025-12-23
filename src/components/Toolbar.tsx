import { useState } from 'react';
import { useTemplateStore } from '../store/templateStore';
import { ImportExport } from './ImportExport';
import { LayerControl } from './LayerControl';
import { NewTemplateDialog } from './NewTemplateDialog';

export const Toolbar: React.FC = () => {
  const { template, createNewTemplate } = useTemplateStore();
  const [showNewDialog, setShowNewDialog] = useState(false);

  const handleNewTemplate = (width: number, height: number) => {
    createNewTemplate(width, height);
  };

  return (
    <div style={{ 
      display: 'flex', 
      flexDirection: 'column', 
      gap: '10px', 
      padding: '20px',
      backgroundColor: '#fff',
      borderRadius: '8px',
      boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '15px',
        flexWrap: 'wrap'
      }}>
        <h2 style={{ margin: 0, color: '#333' }}>Room Template Editor</h2>
        
        <button
          onClick={() => setShowNewDialog(true)}
          style={{
            padding: '8px 16px',
            backgroundColor: '#4CAF50',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '14px',
            fontWeight: 'bold'
          }}
        >
          New Template
        </button>

        <div style={{
          padding: '5px 10px',
          backgroundColor: '#e3f2fd',
          borderRadius: '4px',
          fontSize: '14px',
          border: '1px solid #bbdefb'
        }}>
          Size: {template.width} × {template.height}
        </div>
      </div>

      <LayerControl />
      <ImportExport />

      <NewTemplateDialog
        isOpen={showNewDialog}
        onClose={() => setShowNewDialog(false)}
        onConfirm={handleNewTemplate}
      />
    </div>
  );
};