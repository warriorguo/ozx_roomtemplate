# OZX Room Editor

> **Branch:** `local-client`. Standalone, filesystem-backed desktop variant of
> the room template editor that reads and writes the **OZX project's
> `Assets/StreamingAssets/TilemapData/` folder directly**. The
> PostgreSQL-backed cloud variant lives on `main`.

A visual editor for OZX game room templates. Open an `ozx_base` checkout,
browse all 260+ rooms in a left sidebar with live thumbnails, click to edit on
a multi-layer grid with rule-based validation, save back as JSON files Unity
can pick up on next import.

## What it is

- **Native macOS app** — `swift-app/`: AppKit window + WKWebView, embeds the
  Go server, opens to a portrait grid (transposed: data left → display top).
- **Embedded Go server** — `tile-backend/`: serves the SPA, the validation /
  generation / config API, and a filesystem-backed `Store` whose on-disk
  format is exactly what Unity's `Resources.Load` expects.
- **React + TypeScript SPA** — `src/`: layer editors, sidebar, save dialog,
  client-side thumbnail generator.

## Quick start (macOS)

```bash
git clone git@github.com:warriorguo/ozx_roomtemplate.git
cd ozx_roomtemplate
git checkout local-client

# Build the .app
cd swift-app
make build                            # → build/OZX Room Editor.app
open "build/OZX Room Editor.app"
```

First launch writes `~/.config/ozx-roomeditor/config.json`. If a conventional
OZX checkout exists at `~/Codes/github.com/warriorguo/ozx_base`, `project_root`
is pre-filled to it; otherwise edit the file (or call `PUT /api/v1/config`) and
restart. The window opens once the Go server's `/health` returns 200.

### Dev mode (no .app bundle)

```bash
# Terminal 1 — backend
cd tile-backend
go run ./cmd/server                   # listens on the config-file port (default 8090)

# Terminal 2 — frontend with HMR
npm install
npm run dev                           # Vite at http://localhost:5173
```

The browser-only flow (no native window) also works — `cd tile-backend &&
make build-local && ./bin/ozx-roomeditor` builds a single Go binary that
embeds the SPA and auto-opens your default browser.

## OZX file layout

The editor is a direct reader / writer of the OZX project's tilemap folder.

```
<project_root>/Assets/StreamingAssets/TilemapData/
├── normal/
│   ├── all_boss_3_01.json            # <shape>_<stage>_<openDoors>_<seq>.json
│   ├── all_boss_3_01.json.meta       # Unity-generated; we never touch it
│   └── …
├── basement/
├── cave/
└── test/
```

- **Filename pattern**: `<shape>_<stage>_<openDoors>_<seq>.json`
  - shape ∈ `all` / `bridge` / `platform` / `none`
  - stage ∈ `start` / `teaching` / `building` / `pressure` / `peak` /
    `release` / `boss` / `none`
  - openDoors: 1-15 bitmask (Top=1, Right=2, Bottom=4, Left=8)
  - seq: 01-99, allocated by the backend (gaps are reused)
- **File contents**: a bare `TemplatePayload` JSON — no envelope, no id, no
  timestamps. The editor synthesises a stable UUID v5 from the relative path
  so the frontend can still reference templates by ID.
- **Categories**: each first-level subfolder is a `roomCategory`. New rooms
  are routed to `payload.roomCategory` (default `normal`).
- **`.meta` sidecars**: skipped on read, untouched on update, removed
  alongside the `.json` on delete so Unity regenerates them cleanly.

## Layer system

10 layers, drawn in order; some have dependency constraints enforced by the
validator (red borders flag violations live):

| Layer | Required value | Constraint |
|---|---|---|
| `ground` | — | walkable floor |
| `softEdge` | — | informational |
| `bridge` | — | informational |
| `rail` | — | informational |
| `mainPath` | — | center-biased path connecting doors |
| `static` | `ground=1` | 2×2 obstacle blocks |
| `chaser` | `ground=1`, no overlap with static/bridge/rail/zoner | melee enemy |
| `zoner` | `ground=1`, no overlap with static/bridge/rail/chaser | area-control enemy |
| `dps` | `ground=1`, no overlap with bridge/rail/zoner | ranged enemy |
| `mobAir` | — (no ground requirement); no overlap with other enemy layers | flying enemy |

Stage types — `start`, `teaching`, `building`, `pressure`, `peak`, `release`,
`boss` — each constrain how many of each enemy a generated room may contain.
See `tile-backend/internal/generate/stage_rules.go` for the table.

## UI

- **Left sidebar** (~25 vw, persistent): every template in the configured
  folder, with lazy client-side thumbnails (`IntersectionObserver`-driven,
  cached per template id). Search box filters by name; the row of the
  currently-loaded template is highlighted. Each row has a × delete that
  drops both the `.json` and its `.meta`.
- **Right pane**: layer editors (one per layer), a composite read-only view,
  a threat heatmap, a toolbar with New / Save / Copy JSON.
- **Rendering**: rooms render transposed so 20×12 OZX data displays as a
  12-wide × 20-tall portrait grid. Data top/bottom edges map to the left /
  right of the display; data left/right map to the top / bottom. The
  on-disk arrays are never transposed.

## API

Base URL: `http://localhost:<port>/api/v1` (port from `config.json`, or a
random free port chosen by the Swift launcher).

| Method | Path | Notes |
|---|---|---|
| `GET` | `/health` | 200 when the templates dir is reachable |
| `GET` | `/config` | resolved config: `project_root`, `template_subdir`, computed `templates_dir`, `config_path`, `uses_fallback` |
| `PUT` | `/config` | partial update; hot-swaps the fsstore to the new directory |
| `GET` | `/templates?limit&offset&name_like&room_type&stage_type&…` | summary list with computed counts and door connectivity |
| `POST` | `/templates` | create; filename derived from payload metadata |
| `GET` | `/templates/{id}` | full envelope; id is either the synthesised UUID or `<category>__<basename>` |
| `DELETE` | `/templates/{id}` | removes `.json` plus the sibling `.meta` |
| `POST` | `/templates/validate?strict=true` | validation only |
| `POST` | `/generate/{fullroom\|bridge\|platform}` | room generators |
| `GET` | `/stage-configs` | enemy count ranges per stage |

## Project layout

```
ozx_roomtemplate/
├── src/                              # React frontend
│   ├── components/new/
│   │   ├── TileTemplateApp.tsx       # main editor pane
│   │   ├── TemplateSidebar.tsx       # persistent left sidebar
│   │   ├── LayerEditor.tsx           # per-layer grid (transposed render)
│   │   ├── CompositeLayerEditor.tsx
│   │   ├── HeatmapLayerEditor.tsx
│   │   ├── LazyThumbnail.tsx         # IntersectionObserver thumbnails
│   │   ├── SaveLoadPanel.tsx
│   │   └── ToolBar.tsx
│   ├── services/{api,templateConverter}.ts
│   ├── store/newTemplateStore.ts     # Zustand
│   └── utils/{thumbnailGenerator,newTemplateUtils}.ts
│
├── tile-backend/                     # Go server (no DB on this branch)
│   ├── cmd/
│   │   ├── server/                   # API-only, for `npm run dev` workflow
│   │   └── ozx-roomeditor/           # standalone binary; go:embed of SPA
│   └── internal/
│       ├── config/                   # ~/.config/ozx-roomeditor/config.json
│       ├── serve/                    # shared startup helper
│       ├── http/                     # chi router + handlers + frontend mount
│       ├── store/                    # Store interface
│       │   └── fsstore/              # OZX-native filesystem store
│       ├── browser/                  # cross-platform open(url)
│       ├── web/dist/                 # //go:embed dist/* (Vite output)
│       ├── model/                    # template + payload types
│       ├── generate/                 # full/bridge/platform generators
│       └── validate/                 # structural + stage-rule checks
│
└── swift-app/                        # macOS .app wrapper
    ├── Package.swift                 # SPM, no Xcode required
    ├── Makefile                      # `make build` / `make icon`
    ├── Resources/
    │   ├── Info.plist                # ATS localhost exception
    │   └── AppIcon.icns              # generated from tools/MakeIcon.swift
    ├── Sources/OZXRoomEditor/
    │   ├── main.swift                # NSApplication entry
    │   ├── AppDelegate.swift         # lifecycle + menu bar
    │   ├── MainWindow.swift          # NSWindow + WKWebView + WKUIDelegate
    │   └── GoServer.swift            # subprocess manager (port + health)
    └── tools/MakeIcon.swift          # icon generator (Core Graphics)
```

## Building

```bash
# Frontend
npm install
npm run build

# Backend (API only)
cd tile-backend
make build              # → bin/server

# Standalone Go binary that embeds the SPA
cd tile-backend
make build-local        # → bin/ozx-roomeditor
make build-local-all    # cross-compile darwin amd64/arm64, linux amd64, windows amd64

# macOS .app bundle
cd swift-app
make build              # → build/OZX Room Editor.app
make icon               # regenerate AppIcon.icns
```

## Tech stack

- **Frontend**: React 18 + TypeScript + Zustand + Vite
- **Backend**: Go 1.22 + chi + zap + go:embed
- **Desktop wrapper**: Swift (Package Manager, no Xcode) + AppKit + WKWebView

## Configuration

`~/.config/ozx-roomeditor/config.json` (or `%APPDATA%/ozx-roomeditor/config.json`
on Windows). Override with `--config <path>`.

```json
{
  "project_root": "/Users/andrew/Codes/github.com/warriorguo/ozx_base",
  "template_subdir": "Assets/StreamingAssets/TilemapData",
  "port": 8090,
  "auto_open_browser": true
}
```

When `project_root` is empty the editor falls back to
`~/.local/share/ozx-roomeditor/templates` so it still launches before you've
pointed it at an OZX checkout.

## See also

- `tile-backend/README.md` — backend details and API
- `swift-app/README.md` — macOS wrapper internals
- `CLAUDE.md` — Claude Code notes for working in this repo
- `tile-backend/documents/` — room generation rule references
