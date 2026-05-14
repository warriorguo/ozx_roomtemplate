import { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import {
  templateApi,
  type TemplateSummary,
  type LocalConfig,
  ApiError,
} from '../../services/api';
import { useNewTemplateStore } from '../../store/newTemplateStore';
import { LazyThumbnail } from './LazyThumbnail';
import { formatOpenDoors } from '../../services/templateConverter';

/**
 * Persistent left sidebar listing every template in the configured OZX
 * folder. Replaces the modal Load dialog for everyday switching.
 *
 * - Search is client-side substring match against the displayed label so
 *   what the user sees is what they search.
 * - Rows pull thumbnails through <LazyThumbnail/> only as they scroll into
 *   view, so opening the app against a 261-template project stays cheap.
 * - The list refetches whenever `apiState.lastSaved` changes so a freshly
 *   saved template appears immediately and jumps to the top (server sorts
 *   by mtime desc).
 */

// Source of truth for the sidebar row label. Falls back to `name` when
// `path` is absent (cloud backend mode).
function getDisplayLabel(item: TemplateSummary): string {
  return item.path
    ? item.path.split('/').pop()!.replace(/\.json$/, '')
    : item.name;
}
export const TemplateSidebar: React.FC = () => {
  const apiState = useNewTemplateStore((s) => s.apiState);
  const loadTemplateFromBackend = useNewTemplateStore((s) => s.loadTemplateFromBackend);

  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [items, setItems] = useState<TemplateSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [config, setConfig] = useState<LocalConfig | null>(null);

  // Pull the local-mode config once so we know whether OzxRoomView is wired up.
  useEffect(() => {
    let cancelled = false;
    templateApi.getLocalConfig().then((c) => {
      if (!cancelled) setConfig(c);
    }).catch(() => { /* cloud backend; nothing to do */ });
    return () => { cancelled = true; };
  }, []);

  // Debounce the search input so each keystroke doesn't fire a request.
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search.trim()), 200);
    return () => clearTimeout(t);
  }, [search]);

  const fetchList = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await templateApi.listTemplates({ limit: 1000 });
      setItems(resp.items);
      setTotal(resp.total);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Failed to load templates');
    } finally {
      setLoading(false);
    }
  }, []);

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

  const [deletingId, setDeletingId] = useState<string | null>(null);

  // Right-click context menu state. `null` means hidden. We render a single
  // shared menu (not one per row) and anchor it to the click coordinates.
  type MenuState = { x: number; y: number; item: TemplateSummary };
  const [menu, setMenu] = useState<MenuState | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);

  // Close the menu on any outside click, scroll, or Escape. We have to be
  // careful here: a capture-phase mousedown listener would close the menu
  // *before* the menu item's own click handler can run (mousedown fires
  // before click). So we use bubble phase and check whether the event target
  // is inside the menu DOM — if it is, the menu's own onClick will handle it.
  useEffect(() => {
    if (!menu) return;
    const onDown = (e: MouseEvent) => {
      const node = menuRef.current;
      if (node && e.target instanceof Node && node.contains(e.target)) return;
      setMenu(null);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenu(null);
    };
    const onScroll = () => setMenu(null);
    document.addEventListener('mousedown', onDown);
    document.addEventListener('scroll', onScroll, true);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('scroll', onScroll, true);
      document.removeEventListener('keydown', onKey);
    };
  }, [menu]);

  const handleContextMenu = useCallback((e: React.MouseEvent, item: TemplateSummary) => {
    e.preventDefault();
    setMenu({ x: e.clientX, y: e.clientY, item });
  }, []);

  /**
   * Launches an external macOS app with arguments via the Swift bridge.
   * Falls back to an alert in dev/browser mode (where there is no bridge)
   * so the user at least sees what would have happened.
   */
  const openWith = useCallback((appPath: string, args: string[]) => {
    const bridge = (window as any).webkit?.messageHandlers?.openWith;
    if (bridge && typeof bridge.postMessage === 'function') {
      try {
        bridge.postMessage({ app: appPath, args });
        return;
      } catch (err) {
        console.warn('openWith bridge failed', err);
      }
    }
    alert(`OzxRoomView launch is only available in the macOS app.\n\nWould have run:\n  ${appPath} ${args.join(' ')}`);
  }, []);

  const copyToClipboard = useCallback(async (text: string) => {
    // When running inside the Swift wrapper we have a native bridge to
    // NSPasteboard — both navigator.clipboard.writeText and execCommand are
    // unreliable in WKWebView. The bridge is registered as the "copy"
    // message handler in MainWindow.swift.
    const bridge = (window as any).webkit?.messageHandlers?.copy;
    if (bridge && typeof bridge.postMessage === 'function') {
      try {
        bridge.postMessage(text);
        return;
      } catch (err) {
        console.warn('native copy bridge failed, falling back', err);
      }
    }
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text);
        return;
      }
    } catch (err) {
      console.warn('clipboard.writeText failed, falling back to execCommand', err);
    }
    try {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.left = '-9999px';
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
    } catch (err) {
      console.warn('execCommand copy failed', err);
    }
  }, []);

  const handleDelete = useCallback(
    async (e: React.MouseEvent, item: TemplateSummary) => {
      // Don't let the click bubble up to the row, which would try to load
      // the template we're about to delete.
      e.stopPropagation();
      e.preventDefault();
      const ok = window.confirm(`Delete "${item.name}"?\n\nThis removes the .json from the OZX folder. Unity will regenerate its .meta on next import.`);
      if (!ok) return;
      setDeletingId(item.id);
      try {
        await templateApi.deleteTemplate(item.id);
        // Optimistic: drop from the local list immediately, then re-fetch
        // to reconcile with whatever else may have changed on disk.
        setItems((prev) => prev.filter((t) => t.id !== item.id));
        setTotal((t) => Math.max(0, t - 1));
        fetchList();
      } catch (err) {
        const msg = err instanceof ApiError ? err.message : String(err);
        alert(`Delete failed: ${msg}`);
      } finally {
        setDeletingId(null);
      }
    },
    [fetchList]
  );

  const headerStyle: React.CSSProperties = {
    padding: '12px 14px 8px',
    borderBottom: '1px solid #e0e0e0',
    backgroundColor: '#fafafa',
  };

  const filteredItems = useMemo(() => {
    const q = debouncedSearch.toLowerCase();
    if (!q) return items;
    return items.filter((item) => getDisplayLabel(item).toLowerCase().includes(q));
  }, [items, debouncedSearch]);

  const rows = useMemo(
    () =>
      filteredItems.map((item) => {
        const isCurrent = item.id === currentId;
        const isDeleting = item.id === deletingId;
        const label = getDisplayLabel(item);
        const meta = [
          item.room_type ?? '?',
          item.stage_type ?? '?',
          `doors=${formatOpenDoors(item.open_doors)}`,
        ].join(' · ');
        return (
          <div
            key={item.id}
            onClick={() => handleSelect(item.id)}
            onContextMenu={(e) => handleContextMenu(e, item)}
            className="ort-sidebar-row"
            style={{
              position: 'relative',
              display: 'flex',
              gap: 10,
              padding: '8px 12px',
              borderBottom: '1px solid #f0f0f0',
              cursor: 'pointer',
              backgroundColor: isCurrent ? '#E3F2FD' : 'transparent',
              transition: 'background-color 120ms',
              opacity: isDeleting ? 0.5 : 1,
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
            <div style={{ flex: 1, minWidth: 0, paddingRight: 24 }}>
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
                {label}
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

            <button
              type="button"
              aria-label={`Delete ${item.name}`}
              title="Delete this template"
              disabled={isDeleting}
              onClick={(e) => handleDelete(e, item)}
              className="ort-sidebar-delete"
              style={{
                position: 'absolute',
                top: 6,
                right: 6,
                width: 22,
                height: 22,
                lineHeight: '18px',
                padding: 0,
                fontSize: 14,
                fontWeight: 600,
                border: '1px solid transparent',
                borderRadius: 4,
                background: 'transparent',
                color: '#B71C1C',
                cursor: isDeleting ? 'wait' : 'pointer',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = '#FFEBEE';
                e.currentTarget.style.borderColor = '#FFCDD2';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'transparent';
                e.currentTarget.style.borderColor = 'transparent';
              }}
            >
              ×
            </button>
          </div>
        );
      }),
    [filteredItems, currentId, deletingId, handleSelect, handleDelete]
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
            {loading
              ? 'loading…'
              : debouncedSearch
                ? `${filteredItems.length} / ${total}`
                : `${total} total`}
          </span>
        </div>
        <input
          type="search"
          placeholder="Search filename…"
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
        {!loading && !error && filteredItems.length === 0 && (
          <div style={{ padding: '24px 14px', color: '#999', fontSize: 12, textAlign: 'center' }}>
            No templates match.
          </div>
        )}
        {rows}
      </div>

      {menu && (
        // Render through a portal-like fixed overlay so we don't get clipped
        // by the sidebar's overflow:auto.
        <div
          ref={menuRef}
          role="menu"
          style={{
            position: 'fixed',
            left: menu.x,
            top: menu.y,
            zIndex: 1000,
            minWidth: 200,
            background: '#fff',
            border: '1px solid #ccc',
            borderRadius: 4,
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            padding: '4px 0',
            fontSize: 13,
          }}
        >
          <button
            type="button"
            role="menuitem"
            disabled={!menu.item.path}
            title={menu.item.path || '(path unavailable)'}
            onClick={() => {
              if (menu.item.path) {
                void copyToClipboard(menu.item.path);
              }
              setMenu(null);
            }}
            style={menuItemStyle(!menu.item.path)}
          >
            Copy path
          </button>
          <button
            type="button"
            role="menuitem"
            disabled={!menu.item.path}
            onClick={() => {
              if (menu.item.path) {
                // The parent directory is more useful than the file itself
                // when pasting into Finder's "Go to Folder".
                const dir = menu.item.path.replace(/\/[^/]+$/, '');
                void copyToClipboard(dir);
              }
              setMenu(null);
            }}
            style={menuItemStyle(!menu.item.path)}
          >
            Copy folder
          </button>
          <button
            type="button"
            role="menuitem"
            onClick={() => {
              void copyToClipboard(menu.item.name);
              setMenu(null);
            }}
            style={menuItemStyle(false)}
          >
            Copy name
          </button>
          {(() => {
            const roomViewPath = config?.ozx_room_view_path?.trim() ?? '';
            const filePath = menu.item.path ?? '';
            const enabled = !!roomViewPath && !!filePath;
            return (
              <button
                type="button"
                role="menuitem"
                disabled={!enabled}
                title={
                  !roomViewPath
                    ? 'Set ozx_room_view_path in config.json to enable'
                    : !filePath
                      ? '(no on-disk path available)'
                      : `${roomViewPath} --room ${filePath}`
                }
                onClick={() => {
                  if (enabled) {
                    openWith(roomViewPath, ['--room', filePath]);
                  }
                  setMenu(null);
                }}
                style={menuItemStyle(!enabled)}
              >
                Open with OzxRoomView
              </button>
            );
          })()}
        </div>
      )}
    </aside>
  );
};

function menuItemStyle(disabled: boolean): React.CSSProperties {
  return {
    display: 'block',
    width: '100%',
    textAlign: 'left',
    padding: '6px 14px',
    border: 'none',
    background: 'transparent',
    cursor: disabled ? 'not-allowed' : 'pointer',
    color: disabled ? '#aaa' : '#222',
    fontSize: 13,
  };
}
