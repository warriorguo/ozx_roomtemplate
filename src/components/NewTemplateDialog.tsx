import { useState } from 'react';

interface NewTemplateDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (width: number, height: number) => void;
}

export const NewTemplateDialog: React.FC<NewTemplateDialogProps> = ({ isOpen, onClose, onConfirm }) => {
  const [width, setWidth] = useState<string>('15');
  const [height, setHeight] = useState<string>('11');
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