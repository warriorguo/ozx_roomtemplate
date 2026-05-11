import { useEffect, useState, useCallback } from 'react';
import { templateApi, type LocalConfig, ApiError } from '../../services/api';

/**
 * Shows the active OZX project folder when running against the local-client
 * backend (ORT-65..ORT-69). Renders nothing when /api/v1/config is absent —
 * the cloud backend on main doesn't expose it.
 *
 * Lets the user switch projects in-place: PUT /api/v1/config hot-swaps the
 * filesystem store so subsequent template loads/saves hit the new directory
 * without restarting the binary.
 */
export const ProjectBanner: React.FC = () => {
  const [config, setConfig] = useState<LocalConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [draftRoot, setDraftRoot] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const cfg = await templateApi.getLocalConfig();
      setConfig(cfg);
    } catch (err) {
      console.error('Failed to fetch local config', err);
      setConfig(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  if (loading) return null;
  if (!config) return null;

  const handleStartEdit = () => {
    setDraftRoot(config.project_root);
    setError(null);
    setEditing(true);
  };

  const handleCancel = () => {
    setEditing(false);
    setError(null);
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      const updated = await templateApi.updateLocalConfig({ project_root: draftRoot.trim() });
      setConfig(updated);
      setEditing(false);
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      setError(msg);
    } finally {
      setSaving(false);
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text);
      } else {
        const ta = document.createElement('textarea');
        ta.value = text;
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
      }
    } catch (err) {
      console.error('Copy failed', err);
    }
  };

  return (
    <div
      style={{
        background: config.uses_fallback ? '#FFF7E6' : '#E6F7FF',
        border: `1px solid ${config.uses_fallback ? '#FFE0A0' : '#A0D8FF'}`,
        borderRadius: 6,
        padding: '8px 12px',
        marginBottom: 12,
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        flexWrap: 'wrap',
        fontSize: 13,
      }}
    >
      <span style={{ fontWeight: 600 }}>{config.uses_fallback ? '⚠ No OZX project' : '📂 Project'}</span>

      {!editing && (
        <>
          <code
            title={config.templates_dir}
            style={{
              background: 'rgba(255,255,255,0.6)',
              padding: '2px 6px',
              borderRadius: 4,
              fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
              maxWidth: 480,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {config.project_root || '(not set — using per-user fallback)'}
          </code>
          {config.project_root && (
            <button
              onClick={() => copyToClipboard(config.project_root)}
              style={btnStyleSecondary}
              title="Copy project root"
            >
              📋
            </button>
          )}
          <span style={{ color: '#666' }}>→</span>
          <code
            title="Where templates are read/written"
            style={{
              background: 'rgba(255,255,255,0.6)',
              padding: '2px 6px',
              borderRadius: 4,
              fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
              maxWidth: 480,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              color: '#555',
            }}
          >
            {config.templates_dir}
          </code>
          <button onClick={handleStartEdit} style={btnStylePrimary}>
            Switch project…
          </button>
          <span style={{ color: '#888', marginLeft: 'auto', fontSize: 11 }}>
            config: <code>{config.config_path}</code>
          </span>
        </>
      )}

      {editing && (
        <>
          <input
            value={draftRoot}
            onChange={(e) => setDraftRoot(e.target.value)}
            placeholder="/absolute/path/to/ozx_base"
            spellCheck={false}
            style={{
              flex: '1 1 320px',
              minWidth: 240,
              padding: '4px 8px',
              border: '1px solid #ccc',
              borderRadius: 4,
              fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
              fontSize: 13,
            }}
            autoFocus
          />
          <button onClick={handleSave} disabled={saving} style={btnStylePrimary}>
            {saving ? 'Saving…' : 'Save'}
          </button>
          <button onClick={handleCancel} disabled={saving} style={btnStyleSecondary}>
            Cancel
          </button>
          {error && (
            <span style={{ color: '#C62828', fontSize: 12, width: '100%' }}>{error}</span>
          )}
        </>
      )}
    </div>
  );
};

const btnStylePrimary: React.CSSProperties = {
  padding: '4px 10px',
  background: '#1976D2',
  color: 'white',
  border: 'none',
  borderRadius: 4,
  cursor: 'pointer',
  fontSize: 12,
  fontWeight: 600,
};

const btnStyleSecondary: React.CSSProperties = {
  padding: '4px 10px',
  background: 'white',
  color: '#333',
  border: '1px solid #ccc',
  borderRadius: 4,
  cursor: 'pointer',
  fontSize: 12,
};
