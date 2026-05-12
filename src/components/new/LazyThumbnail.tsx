import { useEffect, useRef, useState } from 'react';
import { templateApi } from '../../services/api';
import { backendToFrontendTemplate } from '../../services/templateConverter';
import { generateThumbnail } from '../../utils/thumbnailGenerator';

/**
 * Single-row lazy thumbnail for the Load panel.
 *
 * The OZX file format has no embedded thumbnail, so we generate one
 * client-side from the Ground layer (already supported by
 * `generateThumbnail`). To avoid rendering 261 templates at once when the
 * dialog opens, we wait until the row scrolls into view via
 * IntersectionObserver and then fetch the full template + draw it onto a
 * detached canvas.
 *
 * Two module-scope caches keep this cheap on re-renders:
 *   - `cache`    : id → data-URL of an already-rendered thumbnail.
 *   - `inflight` : id → Promise resolving to the data-URL, so two
 *                  components mounting for the same id never race.
 */

const cache = new Map<string, string>();
const inflight = new Map<string, Promise<string>>();

async function loadAndRender(id: string, size: number): Promise<string> {
  if (cache.has(id)) return cache.get(id)!;
  if (inflight.has(id)) return inflight.get(id)!;

  const promise = (async () => {
    const backend = await templateApi.getTemplate(id);
    const frontend = backendToFrontendTemplate(backend);
    const dataURL = await generateThumbnail(frontend, size);
    cache.set(id, dataURL);
    return dataURL;
  })();
  inflight.set(id, promise);
  try {
    return await promise;
  } finally {
    inflight.delete(id);
  }
}

interface Props {
  templateId: string;
  alt: string;
  size?: number;
}

export const LazyThumbnail: React.FC<Props> = ({ templateId, alt, size = 80 }) => {
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const [src, setSrc] = useState<string | null>(() => cache.get(templateId) ?? null);
  const [error, setError] = useState(false);

  useEffect(() => {
    // Already rendered — nothing to observe.
    if (cache.has(templateId)) {
      setSrc(cache.get(templateId)!);
      return;
    }
    const el = wrapRef.current;
    if (!el) return;

    let cancelled = false;
    const observer = new IntersectionObserver(
      (entries) => {
        if (!entries.some((e) => e.isIntersecting)) return;
        observer.disconnect();
        loadAndRender(templateId, size)
          .then((url) => {
            if (!cancelled) setSrc(url);
          })
          .catch((err) => {
            console.warn('thumbnail load failed', templateId, err);
            if (!cancelled) setError(true);
          });
      },
      // 100px rootMargin pre-renders the next row or two so scrolling
      // feels less choppy without firing requests for the entire list.
      { root: null, rootMargin: '100px', threshold: 0.01 }
    );
    observer.observe(el);
    return () => {
      cancelled = true;
      observer.disconnect();
    };
  }, [templateId, size]);

  const baseStyle: React.CSSProperties = {
    flexShrink: 0,
    width: `${size}px`,
    height: `${size}px`,
    border: '1px solid #ddd',
    borderRadius: '4px',
    overflow: 'hidden',
    backgroundColor: '#f8f9fa',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 11,
    color: '#888',
    textAlign: 'center',
  };

  if (src) {
    return (
      <div ref={wrapRef} style={baseStyle}>
        <img
          src={src}
          alt={alt}
          style={{ width: '100%', height: '100%', objectFit: 'contain', backgroundColor: '#fff' }}
        />
      </div>
    );
  }
  if (error) {
    return (
      <div ref={wrapRef} style={baseStyle}>
        No<br />Preview
      </div>
    );
  }
  return <div ref={wrapRef} style={baseStyle}>…</div>;
};
