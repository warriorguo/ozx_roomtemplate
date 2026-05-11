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
1. **Static layer**: `static==1` requires `ground==1`
2. **Chaser layer**: `chaser==1` requires `ground==1`, cannot overlap static/bridge/rail/zoner
3. **Zoner layer**: `zoner==1` requires `ground==1`, cannot overlap static/bridge/rail/chaser
4. **DPS layer**: `dps==1` requires `ground==1`, cannot overlap bridge/rail/zoner
5. **MobAir layer**: No ground requirement, cannot overlap other entity layers

**Enemy Placement (Generation)**:
- Door forbidden zone: radius 2 (Manhattan distance) from all doors
- Chaser: 0-3 cells from main path, prefer low squishy score
- Zoner: 0-5 cells from main path, prefer high squishy score, no static blocking LOS
- DPS: 0-4 cells from main path, prefers proximity to chaser/static
- MobAir: prefers zoner/chaser dense areas, spacing >= 1

**Stage Rules**:
- Teaching: DPS only (2-3)
- Building: DPS (2-3) + Chaser (2-3)
- Pressure: DPS (4-6) + Chaser (6-8) + Zoner (1) + MobAir (2-4), not bridge
- Peak: DPS (6-12) + Chaser (6-8) + Zoner (2-3) + MobAir (2-4), full only
- Release: minimal or no enemies
- Boss: requires 6×6 clear center area, restricted door configs

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

### Frontend State Management
- **Zustand store** is the single source of truth for template data
- All cell edits go through `setCellValue()` which triggers validation
- Drag operations track mode (set/clear) based on first clicked cell
- Validation runs after every edit and updates UI state

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
