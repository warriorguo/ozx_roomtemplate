import { useEffect, useState, useCallback, useMemo } from 'react';
import {
  templateApi,
  type TemplateSummary,
  ApiError,
} from '../../services/api';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import { LazyThumbnail } from './LazyThumbnail';

/**
 * Persistent left sidebar listing every template in the configured OZX
 * folder. Replaces the modal Load dialog for everyday switching.
 *
 * - Search filters by name via the existing `?name_like=` backend param.
 * - Rows pull thumbnails through <LazyThumbnail/> only as they scroll into
 *   view, so opening the app against a 261-template project stays cheap.
 * - The list refetches whenever `apiState.lastSaved` changes so a freshly
 *   saved template appears immediately and jumps to the top (server sorts
 *   by mtime desc).
 */
export const TemplateSidebar: React.FC = () => {
  const apiState = useNewTemplateStore((s) => s.apiState);
  const loadTemplateFromBackend = useNewTemplateStore((s) => s.loadTemplateFromBackend);

  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [items, setItems] = useState<TemplateSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Debounce the search input so each keystroke doesn't fire a request.
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search.trim()), 200);
    return () => clearTimeout(t);
  }, [search]);

  const fetchList = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await templateApi.listTemplates({
        limit: 1000,
        name_like: debouncedSearch || undefined,
      });
      setItems(resp.items);
      setTotal(resp.total);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Failed to load templates');
    } finally {
      setLoading(false);
    }
  }, [debouncedSearch]);

  useEffect(() => {
    fetchList();
  }, [fetchList]);

  // Refetch whenever a save completes — the just-saved template will hop to
  // the top once its mtime is fresh.
  useEffect(() => {
    if (apiState.lastSaved?.id) fetchList();
    // We deliberately don't put fetchList in deps here; it would create a
    // refresh loop when the user types into the search box.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiState.lastSaved?.savedAt]);

  const currentId = apiState.lastSaved?.id;

  const handleSelect = useCallback(
    (id: string) => {
      if (id === currentId) return;
      loadTemplateFromBackend(id).catch((err) => {
        console.warn('failed to load template', id, err);
      });
    },
    [currentId, loadTemplateFromBackend]
  );

  const headerStyle: React.CSSProperties = {
    padding: '12px 14px 8px',
    borderBottom: '1px solid #e0e0e0',
    backgroundColor: '#fafafa',
  };

  const rows = useMemo(
    () =>
      items.map((item) => {
        const isCurrent = item.id === currentId;
        const meta = [
          item.room_type ?? '?',
          item.stage_type ?? '?',
          `doors=${item.open_doors ?? '?'}`,
        ].join(' · ');
        return (
          <div
            key={item.id}
            onClick={() => handleSelect(item.id)}
            style={{
              display: 'flex',
              gap: 10,
              padding: '8px 12px',
              borderBottom: '1px solid #f0f0f0',
              cursor: 'pointer',
              backgroundColor: isCurrent ? '#E3F2FD' : 'transparent',
              transition: 'background-color 120ms',
            }}
            onMouseEnter={(e) => {
              if (!isCurrent) e.currentTarget.style.backgroundColor = '#F5F5F5';
            }}
            onMouseLeave={(e) => {
              if (!isCurrent) e.currentTarget.style.backgroundColor = 'transparent';
            }}
            title={item.name}
          >
            <LazyThumbnail templateId={item.id} alt={item.name} size={56} />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div
                style={{
                  fontWeight: isCurrent ? 600 : 500,
                  fontSize: 13,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  color: '#222',
                }}
              >
                {item.name}
              </div>
              <div
                style={{
                  marginTop: 2,
                  fontSize: 11,
                  color: '#777',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {meta}
              </div>
              <div style={{ fontSize: 11, color: '#aaa' }}>
                {item.width}×{item.height}
              </div>
            </div>
          </div>
        );
      }),
    [items, currentId, handleSelect]
  );

  return (
    <aside
      style={{
        flex: '0 0 25vw',
        minWidth: 280,
        maxWidth: 440,
        height: '100vh',
        borderRight: '1px solid #e0e0e0',
        backgroundColor: '#fff',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      <div style={headerStyle}>
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
          <strong style={{ fontSize: 14 }}>Templates</strong>
          <span style={{ fontSize: 11, color: '#888' }}>
            {loading ? 'loading…' : `${total} total`}
          </span>
        </div>
        <input
          type="search"
          placeholder="Search name…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{
            marginTop: 8,
            width: '100%',
            boxSizing: 'border-box',
            padding: '6px 8px',
            fontSize: 12,
            border: '1px solid #ccc',
            borderRadius: 4,
          }}
        />
      </div>

      {error && (
        <div style={{ padding: '12px 14px', color: '#C62828', fontSize: 12 }}>
          {error}
        </div>
      )}

      <div style={{ flex: 1, overflowY: 'auto' }}>
        {!loading && !error && items.length === 0 && (
          <div style={{ padding: '24px 14px', color: '#999', fontSize: 12, textAlign: 'center' }}>
            No templates match.
          </div>
        )}
        {rows}
      </div>
    </aside>
  );
};
