import React, { useCallback, useEffect, useState } from 'react';
import { templateApi, ApiError, type TilemapTagRegistry } from '../../services/api';
import { useNewTemplateStore } from '../../store/newTemplateStore';

/**
 * Tilemap tags (ORT-158 / OZX-2152). OZX builds a template's TagMask from its `tags` and
 * throws on a name its registry (TilemapData/tilemap_tags.json) does not list, so this panel
 * offers the registry's names as checkboxes and nothing else — there is no free-text input.
 * A name the template already carries but the registry lacks is shown in red with a remove
 * button; the backend refuses to save it.
 *
 * The registry belongs to the OZX project and can change while the editor runs, so it is
 * re-read on mount, whenever a different template is loaded, and on demand.
 */
export const TemplateTagsPanel: React.FC = () => {
  const tags = useNewTemplateStore((s) => s.template.tags);
  const setTags = useNewTemplateStore((s) => s.setTags);
  const loadedId = useNewTemplateStore((s) => s.apiState.lastSaved?.id);

  const [registry, setRegistry] = useState<TilemapTagRegistry | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      setRegistry(await templateApi.getTilemapTags());
      setLoadError(null);
    } catch (e) {
      setLoadError(e instanceof ApiError ? e.message : 'Failed to load the tag registry');
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh, loadedId]);

  const current = tags ?? [];
  const registered = registry?.tags ?? [];
  const unknown = current.filter((t) => !registered.includes(t));

  // Unticking keeps the rest in their authored order; ticking appends.
  const toggle = (name: string, on: boolean) => {
    setTags(on ? [...current.filter((t) => t !== name), name] : current.filter((t) => t !== name));
  };

  return (
    <div style={{ marginBottom: '15px' }}>
      <strong>🏷️ Tags:</strong>
      <button
        onClick={refresh}
        title="Re-read tilemap_tags.json"
        style={{ marginLeft: '8px', fontSize: '11px', padding: '1px 6px', cursor: 'pointer' }}
      >
        ↻
      </button>
      <div style={{
        marginTop: '8px',
        padding: '10px',
        backgroundColor: '#f8f9fa',
        borderRadius: '6px',
        border: '1px solid #dee2e6',
        fontSize: '13px'
      }}>
        {loadError && <div style={{ color: '#dc3545' }}>{loadError}</div>}
        {registry?.error && (
          <div style={{ color: '#dc3545' }}>
            tilemap_tags.json is invalid — the game would refuse it: {registry.error}
          </div>
        )}
        {registry && !registry.found && (
          <div style={{ color: '#666' }}>
            This project has no <code>tilemap_tags.json</code>, so there are no tags to pick.
          </div>
        )}
        {registered.length > 0 && (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px 14px' }}>
            {registered.map((name) => (
              <label key={name} style={{ display: 'flex', alignItems: 'center', gap: '5px', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={current.includes(name)}
                  onChange={(e) => toggle(name, e.target.checked)}
                />
                {name}
              </label>
            ))}
          </div>
        )}
        {unknown.length > 0 && (
          <div style={{ marginTop: '8px', color: '#dc3545' }}>
            Not in the registry (the game throws on these; saving is refused):
            {unknown.map((name) => (
              <span key={name} style={{ marginLeft: '6px', whiteSpace: 'nowrap' }}>
                <code>{name}</code>
                <button
                  onClick={() => setTags(current.filter((t) => t !== name))}
                  title={`Remove ${name}`}
                  style={{ marginLeft: '2px', fontSize: '11px', padding: '0 4px', cursor: 'pointer' }}
                >
                  ×
                </button>
              </span>
            ))}
          </div>
        )}
        {registered.length > 0 && (
          <div style={{ marginTop: '6px', fontSize: '11px', color: '#666' }}>
            Rooms filter templates by these (require / exclude / any). Untagged is fine.
          </div>
        )}
      </div>
    </div>
  );
};
