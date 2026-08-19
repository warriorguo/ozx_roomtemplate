# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **Branch:** `local-client`. Standalone, filesystem-backed desktop variant
> of the editor; the PostgreSQL/cloud variant lives on `main`. The bundled
> binary `ozx-roomeditor` ships the React SPA via `go:embed`, auto-opens
> the user's default browser, and lets the user switch projects in-place
> via the toolbar (PUT `/api/v1/config` hot-swaps the filesystem store).
> A native macOS wrapper lives in `swift-app/` (AppKit + WKWebView).
> Build either with `cd tile-backend && make build-local` (Go binary →
> opens a browser tab) or `cd swift-app && make build` (.app bundle →
> opens in a native window).

## Project Overview

This is a **room template editor** for game development, consisting of a React TypeScript frontend and a Go backend service. The editor creates tile-based room templates with a multi-layer system and rule-based validation for game room layouts.

**Layer System**: ground, softEdge, bridge, rail, mainPath, static, chaser, zoner, dps, mobAir
**Stage Types**: start, teaching, building, pressure, peak, release, boss

## Common Commands

### Frontend Development
```bash
# Install dependencies
npm install

# Start development server (port 5173)
npm run dev

# Build for production
npm run build

# Preview production build
npm preview
```

### Backend Development
```bash
cd tile-backend

# Install dependencies
go mod tidy

# Run API-only server (for use alongside `npm run dev`)
go run cmd/server/main.go

# Run with hot-reload (requires air: go install github.com/air-verse/air@latest)
make dev

# Build binaries
make build              # API-only server → bin/server
make build-prod         # API-only server, optimized
make build-local        # Standalone bundled binary → bin/ozx-roomeditor
make build-local-all    # Cross-compile for darwin/linux/windows

# Code quality
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint (requires golangci-lint)

# Run tests
make test              # All tests
make test-unit         # Unit tests only

# Run with coverage report
make test-coverage
```

### Testing
```bash
# Backend unit tests
go test -v -race ./internal/...

# No frontend tests currently configured
```

## Architecture

### Frontend Architecture (React + TypeScript + Zustand)

**State Management**: Zustand store (`src/store/newTemplateStore.ts`) manages:
- Template data (layers: ground, softEdge, bridge, rail, mainPath, static, chaser, zoner, dps, mobAir)
- UI state (active layer, drag operations, layer visibility)
- Validation results with error highlighting
- API state (loading, errors, last saved template)

**Component Structure**:
- `src/components/new/TileTemplateApp.tsx` - Main application container
- `src/components/new/LayerEditor.tsx` - Grid editor with click/drag support
- `src/components/new/GroundGenerator.tsx` - Auto-generation panel for ground layer
- `src/components/new/SaveLoadPanel.tsx` - Backend integration for save/load
- `src/components/new/ToolBar.tsx` - Main toolbar with validation and export

**Layer System**:
- Base layers: ground, softEdge, bridge, rail (generated in order)
- MainPath: center-biased path connecting doors (computed after bridge)
- Static: 2×2 obstacle blocks on ground
- Enemy layers: chaser (pressure), zoner (area control), dps (damage), mobAir (flying)
- Each layer is a 2D grid of 0s and 1s
- Ground layer can be auto-generated with room types (full, bridge, platform)

**Stage System**:
- Stage type determines enemy counts and room constraints
- start → teaching → building → pressure → peak → release → boss

**Validation**:
- Real-time validation in `src/utils/newTemplateUtils.ts`
- Backend validation via API endpoint with strict mode
- Visual error feedback (red borders on invalid cells)

### Backend Architecture (Go, storage pluggable)

**Layers**:
```
cmd/
  ├── server/             API-only entrypoint (frontend served separately)
  └── ozx-roomeditor/     Standalone bundled binary (go:embed SPA + browser auto-launch)
internal/
  ├── config/             User config file (project_root, port, auto_open_browser)
  ├── serve/              Shared startup helper (config → store → router → SIGINT)
  ├── http/               chi router, handlers, middleware, frontend mount
  ├── store/              Store interface + StubStore + fsstore (filesystem impl)
  ├── browser/            Cross-platform default-browser launcher
  ├── web/                go:embed dist/* — the SPA bundle
  ├── model/              Data models and domain types
  ├── generate/           Room generators (full, bridge, platform), stage rules
  └── validate/           Validation logic (structure + logical constraints)
```

**API Endpoints** (Base: `/api/v1`):
- `POST /templates` - Create template
- `GET /templates?limit&offset&name_like` - List with pagination/search
- `GET /templates/{id}` - Get specific template
- `DELETE /templates/{id}` - Delete template
- `POST /templates/validate?strict` - Validate payload
- `POST /generate/{fullroom|bridge|platform}` - Generate a room
- `GET /stage-configs` - Stage type configurations
- `GET /config` - Resolved user config (project_root, template_subdir, templates_dir, ...)
- `PUT /config` - Update config; hot-swaps the filesystem store to a new project folder
- `GET /health` - Health check

**Storage**:
- `store.Store` is the single storage interface (Create/Get/Update/Delete/List/HealthCheck).
- `store.fsstore.Store` persists each template as `<uuid>.json` under the
  resolved templates directory. Writes are atomic (`<file>.tmp` + rename) and
  a single `RWMutex` guards concurrent access.
- `internal/config` loads/saves `~/.config/ozx-roomeditor/config.json`
  (overridable with `--config`). The templates directory is computed as
  `project_root + template_subdir`, or a per-user fallback when
  `project_root` is empty.
- `store.StubStore` still exists as a fallback for when the configured
  directory cannot be created.

**Key Features**:
- CORS middleware for frontend integration
- Graceful shutdown with 30s timeout
- Structured logging with zap
- Request body size limit: 2MB
- Panic recovery middleware

### Validation Rules

**Layer Constraints** (enforced in strict mode):
1. **SoftEdge layer**: `softEdge==1` requires `ground==0` and an anchor to
   ground — see "SoftEdge anchoring" below
2. **Static layer**: `static==1` requires `ground==1`
3. **Chaser layer**: `chaser==1` requires `ground==1`, cannot overlap static/bridge/rail/zoner
4. **Zoner layer**: `zoner==1` requires `ground==1`, cannot overlap static/bridge/rail/chaser
5. **DPS layer**: `dps==1` requires `ground==1`, cannot overlap bridge/rail/zoner
6. **MobAir layer**: No ground requirement, cannot overlap other entity layers

**SoftEdge anchoring** (`computeSoftEdgeSupport`, ORT-116/117): a soft edge cell
is valid when it is *anchored* to ground, which is a least fixpoint over the whole
`softEdge` layer rather than a per-cell adjacency test: the **base** rule anchors
a cell orthogonally adjacent to `ground==1`, and a cell then borrows anchoring
when **both** the cell to its **visual right** and the one **below** it are
anchored, or **both** the one to its **visual left** and the one **above** it are.
Support is borrowed only from cells that are themselves anchored, so a patch
floating in void with no ground contact anywhere stays invalid however large it
is, and a diagonal neighbour alone never suffices. The effect is that a soft edge
patch may be thicker than the 1-cell rim the old adjacency rule allowed, as long
as its outer boundary reaches ground.

**The pairs are in visual space, the layer is in data space** (ORT-117). The
rule is stated against what the editor draws, and the canvas is rotated 90° CCW
(ORT-111), so the data-space offsets the code walks are the transpose:

| rule | visual | data offsets |
|------|--------|--------------|
| 1 | right + below | `(x, y+1)` and `(x-1, y)` |
| 2 | left + above | `(x, y-1)` and `(x+1, y)` |

ORT-116 shipped the literal data-space reading, `(x+1,y)+(x,y+1)`, which is the
*opposite* diagonal — the mirror image of the rule, rejecting exactly the
L-shaped corner the relaxation exists to accept. This is the same trap as
ORT-111/112/113: the rotation is invisible until a shape is asymmetric enough to
expose it. The rule is deliberately **not** symmetric across all four corners, so
"fixing" it to accept any two adjacent neighbours is wrong.

This is a pure relaxation — every payload that validated before still validates.
Generation is untouched: `layer_softedge.go` only ever emits 1xN / Nx1 strips in
concave notches, which are ground-adjacent by construction and satisfy the base
rule. The propagation rules exist for hand-edited rooms. `computeSoftEdgeSupport`
in `validate/validate.go` and the one in `src/utils/newTemplateUtils.ts` are
mirrors of each other and must stay in sync.

**Enemy Placement (Generation)**:
- Door forbidden zone: radius 2 (Manhattan distance) from all doors
- Chaser: 0-3 cells from main path, prefer low squishy score
- Zoner: 2×2 block (1×1 fallback), 0-5 cells from main path, prefer high squishy score, no static blocking LOS
- DPS: 0-4 cells from main path, prefers proximity to chaser/static
- MobAir: centre-seeded, evenly distributed outward, spacing >= 1

**Stage Rules**:
- Teaching: DPS (4-6) + Chaser (2) + Zoner (1) + MobAir (6)
- Building: DPS (4-6) + Chaser (4-6) + Zoner (1) + MobAir (6)
- Pressure: DPS (8-12) + Chaser (8-10) + Zoner (2) + MobAir (6-12), not bridge
- Peak: DPS (8-12) + Chaser (8-10) + Zoner (2-3) + MobAir (12-18), full only
- Release: light mix — DPS (2-4) + Chaser (2) + Zoner (1) + MobAir (6)
- Boss: requires 6×6 clear center area, restricted door configs

**Stage-driven static count** (`StageConfig.StaticRange`, ORT-100): when a stage
type is supplied it also supplies the static count, overriding the request's
`staticCount` in all three generators. Teaching/building/release place 6–9 2×2
blocks; pressure/peak place 2–3 (roughly one third, so the denser enemy waves
have room to move); start/boss place none. An empty stage type still honours the
request value verbatim.

**Stage-driven static placement** (`StagePlacementHints.StaticDisperse`, ORT-99):
start/teaching/building/release keep the default alternating centre-outward /
edge-inward scatter. Pressure and peak seed the first block from the room edge
and then pick each next one by farthest-point selection, so cover ends up on the
perimeter with the middle left open for the heavy enemy waves. The hints are
built in `buildPlacementHints` and passed into
`generateStaticLayerWithDebugAndRail` by all three generators.

**MobAir centre-out distribution** (ORT-104): mobAir is seeded at the valid cell
nearest the room centre, then fills a centre-anchored grid worked outward, each
slot snapping within a third of a cell. Placement is **deterministic** — the old
`rand.Intn(2)` strategy coin flip and its two strategies were dead code and are
gone. The zoner/chaser density preference survives only as a tiebreak inside one
slot; it no longer moves mobs across the room. Because mobAir is last in the
pipeline and Chaser/DPS do not check it, fullroom's grouped placement now runs
mobAir **once after the group loop** rather than inside it — the old order let
group 1's air mobs be overwritten by group 2's ground enemies.

**Zoner 2×2 footprint** (`zonerSize`, ORT-103): a zoner occupies a 2×2 block
wherever a valid site exists and falls back to a single cell only when none
does. Every cell of the block satisfies the same per-cell constraints the 1×1
path checks, and distinct blocks never touch in any of the 8 directions. The
game collapses a connected block into one spawn, so **`zonerCount` counts
spawns, not cells** — count 8-connected groups (`countZonerUnits`), not `1`s,
whenever comparing against a stage range. Measured 1×1 fallback rate is under 1%
of spawns at every stage minimum.

**No stage constrains room size** (ORT-109): ORT-102 gave pressure an 18×10
minimum and peak a 20×12 one, rejecting smaller rooms up front — an undersized
room cannot fit the stage's counts under the 8-directional spacing constraint,
and the relaxed placement fallback met them by dropping that constraint,
emitting adjacent same-category spawn tiles (which crash the game for Zoner —
ORT-93). ORT-108 cut those two stages' counts back instead, and the minimums,
`StageConfig.MinWidth`/`MinHeight`, and the `minWidth`/`minHeight` fields on
`/stage-configs` are all gone. Any room size may be paired with any stage.

Small rooms are permitted but not guaranteed clean: at ORT-108's counts, 16×8
still produces an adjacent same-category spawn pair in 6% of pressure rooms and
16.5% of peak rooms. That residue belongs to ORT-93 — placement should never
violate spacing whatever it is asked for — rather than to a size limit.

**The default room size is 16×10 in data space** (`DefaultRoomWidth` /
`DefaultRoomHeight` in `generate/rules.go`, ORT-114): a generate request that
omits `width` or `height` — or sends `0` — gets these; a *negative* value is
still rejected, so a malformed request cannot be silently rewritten into a valid
room. `DEFAULT_ROOM_WIDTH` / `DEFAULT_ROOM_HEIGHT` in `src/types/newTemplate.ts`
carry the same pair for a fresh template and the "New" dialog, and must be kept
in sync.

The numbers look transposed on purpose. OZX **ignores the payload's `meta`
block** (`RoomTilemapData` has no `meta` field) and derives the room size from
the shape of the `ground` array — **outer array = X, inner = Y**
(`RoomInstanceView.ComputeRoomSize`) — which is the transpose of this repo's
data space. So data-space 16 wide × 10 high is a **10-wide, 16-high** room in
the game, the size the fixed camera (x=5) is framed for. This is the same
data/OZX split as the door sides in ORT-111/112/113, except that no code applies
it: the array shape *is* the rotation.

Not every stage fits the default cleanly. Measured over 500 rooms per
combination, at 16×10 vs. the previous 16×8 default (shortfall = a room emitting
at least one ORT-110 warning): fullroom teaching 62 vs. 77, building 29 vs. 75,
pressure 1 vs. 205, peak 241 vs. 500, release 31 vs. 71 — 16×10 is a strict
improvement everywhere. Bridge and platform still fall short on `static` in
~90%+ of rooms at both sizes (a corridor room has nowhere to put 6–9 2×2
blocks), and boss cannot generate at either size at all: it needs a 6×6 area
more than 3 cells from every edge, which needs height ≥ 14. Those are ORT-108's
count ranges meeting small rooms, not a regression from the size change — pass
explicit dimensions for boss and for dense bridge/platform rooms.

**Placement shortfalls are reported** (`PlacementShortfall`, ORT-110): placement
is best-effort — the strict pass keeps the 8-directional spacing constraint and
stops when a room runs out of legal sites, and mobAir has no relaxed fallback at
all. All three generate responses therefore carry an optional
`warnings: [{layer, requested, placed, message}]`, built by
`collectPlacementShortfalls` from the per-layer debug counts and **absent when
every layer met its target**. Counts are spawns, not cells. A 200 response is
not by itself evidence the stage's counts were met; this is how an undersized
room announces itself now that ORT-109 no longer refuses one.

**Door walkability invariant** (`ensureDoorsWalkable`, ORT-105): the ground
generators carve after they fill, and their rollback guard
(`areAllDoorsConnected`) accepts a door whose own cell is void as long as one of
its 4 neighbours is reachable — so a corner erase or center pit could seal the
doorway, and two erases on the same wall could remove that wall's whole edge
line. `ensureGroundConnectivity` does not repair it (a void door cell is not an
island). All three generators therefore call `ensureDoorsWalkable` once their
ground layer is final: it re-opens each door anchor and reruns the connectivity
repair. **Post-condition: every requested door is walkable and reachable from
every other door**, so anything downstream may assume it.

**MainPath is a hard requirement** (ORT-106): `ComputeMainPath` returns an
`error` — a door with no walkable cell on its wall, or a door pair with no route
between them, fails the whole generation (surfaced as `400 Generation failed`
with a `main path: ...` reason). Both used to be recorded as debug misses only,
so an untraversable room was returned as a success; worse, `findCenterBiasedPath`
snaps an unwalkable endpoint to the nearest walkable cell, so the room came back
with a plausible-looking path between two interior cells that never crossed a
doorway. Doors are visited in a fixed order (`doorOrder`) rather than by map
iteration, so pairing and error text are deterministic. A single-door room is
still legal and yields an empty main path.

**Structure Validation**:
- Dimensions: 4-200 for width/height
- Required layers: ground, static, chaser, zoner, dps, mobAir
- Correct grid dimensions (height × width)
- Cell values: only 0 or 1

### Data Flow

**Frontend ↔ Backend Integration**:
1. **API Service**: `src/services/api.ts` - HTTP client for all backend operations
2. **Converter**: `src/services/templateConverter.ts` - Bidirectional conversion between frontend and backend formats
3. **Environment**: `.env` file configures `VITE_API_BASE_URL` (default: `http://localhost:8090/api/v1`)

**Save Operation**:
```
Frontend Template → templateConverter → API Request → Go Handler → store.Store
```

**Load Operation**:
```
store.Store → Go Handler → API Response → templateConverter → Frontend Template
```

The concrete `store.Store` is `StubStore` on this branch; the filesystem-backed
implementation (per-template JSON files under a configured OZX project folder)
lands in ORT-66.

## Key Implementation Details

### Door Sides Live in Two Spaces (ORT-111)

**Data space** is the in-memory representation — the grid layers, `doors{}`,
`doorOverrides{}` and the `openDoors` bitmask as the store, the API, the
converters and the generators all see them. `ground[0]` is the data top edge.
**Visual space** is what the canvas draws: ORT-76 renders the grid rotated 90°
CCW, the same transpose + Y-flip OZX applies, so the editor shows the room the
way the game will — which means visual space is also **OZX's** space.

| data | visual |
|------|--------|
| top | left |
| right | top |
| bottom | right |
| left | bottom |

ORT-76 rotated only the canvas; every label stayed in data space, so an opening
drawn at the visual top was reported and saved as "Right" and the door controls
were 90° off from what the user was looking at. ORT-111 rotated **the labels**,
not the storage: nothing on disk changed, so existing templates and the OZX
importer are unaffected.

`dataToVisual` / `visualToData` / `VISUAL_DOOR_SIDES` in
`src/utils/newTemplateUtils.ts` are the **only** place the rotation is written
down — `LayerEditor.getDoorBorderSide` calls `dataToVisual` for exactly that
reason. Every UI surface that names a door side (status panel, generator
checkboxes, BFS tooltip, sidebar filters and row labels) goes through them.
Never open-code the mapping; the store, the converters, `detectDoorConnectivity`
and the whole backend stay data-space. The rotation lives at the view boundary
only.

**The on-disk file is in OZX space, not data space (ORT-112, ORT-113).** The
`.json` under an OZX project is a game artifact, so `fsstore` rotates the door
metadata at the disk boundary and nowhere else:

- `writePayload` rotates `doors{}` and `openDoors` data→visual before marshalling
  (on a copy — the caller's payload stays data-space).
- `readPayload` rotates them back visual→data, so the API, the converters and the
  editor never see OZX space.
- `doorsOf` derives the `<shape>_<stage>_<doors>_<NN>` filename mask through the
  same rotation, so **the name and the body always agree**.

`rotateDoorFields` / `dataToVisualSide` / `visualToDataSide` in `fsstore.go` are
the Go-side counterpart to `dataToVisual` / `visualToData` on the frontend — the
only place the rotation is written down in the backend.

This is what the game requires: OZX reads `RoomTilemapData.OpenDoors` from the
JSON body and matches it against `RoomQueryHelper.DeriveOpenDoors`, which is in
OZX space. A data-space body makes `RoomTilemapQuery.Query` throw
`No tilemap matches openDoors=N`. The `room-sync` skill has always applied this
same remap on export (`remap_doors`); ORT-113 brought the editor's own save path
in line, so a room saved from the editor and a room exported by sync are now
byte-compatible.

**Only the door metadata rotates — the grid layers do not.** `ground` and the
rest stay in data-space row order on disk; the game applies that rotation itself
at render time. `doorOverrides` is an editor-only whitelist the game never reads
and sync never remaps, so it stays data-space in the file too.

Renaming files by hand does not stick: `Create`/`Update` re-derive the name.

### Frontend State Management
- **Zustand store** is the single source of truth for template data
- All cell edits go through `setCellValue()` which triggers validation
- Drag operations track mode (set/clear) based on first clicked cell
- Validation runs after every edit and updates UI state
- `apiState.lastSaved` is the **save identity of the template currently in the
  editor**: `saveTemplate` updates that stored template in place when it is set
  and creates a new one when it is not (ORT-83). Only `loadTemplateFromBackend`
  and a successful save may set it; every action that swaps the editor's content
  for something with no stored counterpart — `createNewTemplate`,
  `loadTemplate`, `loadTemplateFromJSON` (Generate Room, paste, import) — must
  clear it via `withoutSaveIdentity`, or the next save overwrites the previous
  room (ORT-107). Because `fsstore.Update` relocates on a changed derived name
  (ORT-87), that overwrite deletes the old file rather than just rewriting it.

### Backend Request Processing
1. Request received by chi router
2. Middleware: CORS → Request logging → Panic recovery
3. Handler: Parse JSON → Validate → Store operation → Response
4. Error handling with appropriate HTTP status codes

### Template Validation
- Frontend performs real-time validation for immediate feedback
- Backend performs authoritative validation before saving
- Strict mode (`?strict=true`) enables logical constraint checking
- Validation errors include layer, position (x, y), and reason

### Ground Auto-Generation
- Rectangular rooms: Configurable wall thickness with door positions
- Cross-shaped rooms: For intersection layouts
- Door system: Specify position and direction (north/south/east/west)
- Auto-placement creates walkable paths through walls

## Environment Configuration

**Frontend** (`.env`):
```
VITE_API_BASE_URL=http://localhost:8090/api/v1
VITE_NODE_ENV=development
```

**Backend runtime env** (`tile-backend/.env`):
```
LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174
```

**Backend user config** (`~/.config/ozx-roomeditor/config.json`, written on first run):
```json
{
  "project_root": "",
  "template_subdir": "Assets/Resources/TilemapData",
  "port": 8090,
  "auto_open_browser": true
}
```

`project_root` is the absolute path to an OZX project (e.g. an `ozx_base`
checkout); leave it empty to use a per-user fallback templates directory. Pass
`--config <path>` to the binary to use a different config file. `GET
/api/v1/config` returns the resolved config (with computed `templates_dir`,
`config_path`, and `uses_fallback`) so the frontend can display the current
project folder.

**Note**: Make sure frontend `VITE_API_BASE_URL` matches the backend `PORT`.

## Common Development Workflows

### Running Full Stack Locally
```bash
# Terminal 1 - Backend
cd tile-backend
go run cmd/server/main.go

# Terminal 2 - Frontend
npm run dev
```

### Setting Up From Scratch
```bash
# 1. Clone repository and install dependencies
npm install
cd tile-backend && go mod tidy && cd ..

# 2. Configure environment
cp .env.example .env
cp tile-backend/.env.example tile-backend/.env

# 3. Start services (see "Running Full Stack Locally" above)
```

No database setup is needed on this branch — the backend uses the in-process
`StubStore` until ORT-66 lands the filesystem-backed implementation.

### Running Tests Before Committing
```bash
# Frontend: No automated tests currently
npm run build  # Verify build works

# Backend: Run all tests
cd tile-backend
make fmt && make vet     # Format and check code
make test-unit           # Fast unit tests
```

## Development Notes

### When Working with Templates
- Template format is consistent across frontend/backend with converter layer
- Backend persists the full payload as a `<uuid>.json` file via
  `store.fsstore.Store`; computed fields are regenerated by
  `model.ComputeTemplateStats` on every write
- Frontend maintains separate layers for UI editing
- Version field is always `1` in current implementation

### When Adding API Endpoints
- Add handler in `tile-backend/internal/http/`
- Register route in `router.go`
- Update API client in `src/services/api.ts`
- Add converter functions if data format differs

### When Modifying Validation
- Update logic in `tile-backend/internal/validate/validate.go`
- Mirror changes in `src/utils/newTemplateUtils.ts` for real-time feedback
- Add tests in both locations

### When Modifying Room Generation
- Generation pipeline order: ground → softEdge → bridge → rail → **stageRules** → **mainPath** → static → zoner → chaser → dps → mobAir
  - **All three generators follow this order** (ORT-98). Each layer may only be
    constrained by layers earlier in the sequence, so anything reading stage
    data — including stage-driven static behaviour — must sit after `stageRules`.
  - The bridge layer is only produced by the **bridge** room type. `fullroom` and
    `platform` deliberately skip it: both run `ensureGroundConnectivity`, so no
    floating islands remain to span, and calling the bridge generator would
    trigger its "force at least one bridge" fallback and emit spurious tiles.
- `tile-backend/internal/generate/` contains all generation logic:
  - `fullroom.go`, `bridge.go`, `platform.go` — room type generators
  - `mainpath.go` — center-biased pathfinding + squishy score computation
  - `layer_chaser.go`, `layer_zoner.go`, `layer_dps.go` — enemy placement
  - `stage_rules.go` — stage type validation and enemy count ranges
  - `rules.go` — shared validation functions and constants
- Documents in `tile-backend/documents/` describe generation rules
- When updating generation logic, keep the documentation in sync with the implementation

## Testing Strategy

**Backend Tests**:
- Unit tests use a `testify/mock`-based `MockStore` implementing `store.Store`
  (see `internal/http/handlers_test.go`)
- No integration tests on this branch — there is no database
- Coverage reports generated with `make test-coverage`

**Frontend Testing**:
- No automated tests currently configured
- Manual testing through UI during development

## Skill Sync Rule

**IMPORTANT**: When you modify any of the following backend files, you MUST also update the corresponding Claude Code skills in `.claude/skills/` to stay in sync:

| Backend files (trigger) | Skills to update |
|------------------------|-----------------|
| `tile-backend/internal/model/types.go` (TemplatePayload, request/response structs) | `room-generator`, `room-test` |
| `tile-backend/internal/generate/types.go`, `platform.go`, `fullroom.go`, `bridge.go` (request types, generation output) | `room-generator`, `room-test` |
| `tile-backend/internal/generate/stage_rules.go` (enemy count ranges, stage constraints) | `room-test` (references/validation.md) |
| `tile-backend/internal/http/router.go`, `handlers.go` (API endpoints) | `room-generator`, `room-test` |

**What to update in skills:**
- Field names in request/response examples (e.g. JSON keys, parameter tables)
- Allowed values for enum-like fields (roomCategory, roomShape, stageType)
- Layer names in ASCII visualization and validation scripts
- Enemy count ranges in stage rules tables
- API endpoint URLs

The skills are symlinked from `~/.claude/skills/` → `.claude/skills/` in this project, so edits here propagate to the machine-level skills automatically.

## Additional Documentation

- `README.md` - Original 3-layer system (deprecated in favor of 5-layer)
- `README-5Layer.md` - Current 5-layer system documentation
- `FRONTEND_API_INTEGRATION.md` - Detailed API integration guide
- `tile-backend/README.md` - Backend API documentation
- `tile-backend/TESTING.md` - Backend testing guide
