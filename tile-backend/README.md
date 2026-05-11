# Tile Template Backend (local-client branch)

> **Status:** local-client variant of the room template editor. The PostgreSQL
> backend has been removed. Template CRUD is backed by an on-disk store
> (`fsstore`, **ORT-66**) and driven by a user JSON config file pointing at an
> OZX project folder (**ORT-67**). The standalone binary with browser
> auto-launch lands in **ORT-68**.

A Go HTTP service that powers the room template editor: validation, room
generation (full/bridge/platform), and template CRUD. Storage is pluggable
behind the `store.Store` interface.

## Features

- **REST API** for template management (Create, List, Get, Delete, Validate)
- **Room generation** for full / bridge / platform layouts
- **Rule-based validation** for layers and stage constraints
- **CORS** for the React frontend
- **Health checks** and graceful shutdown

## Quick Start

### Prerequisites
- Go 1.22+

### Build & run
```bash
cd tile-backend
go mod tidy
go run cmd/server/main.go      # listens on :8090
# or
go build -o bin/server cmd/server/main.go && ./bin/server
```

### Configuration

The editor reads two layers of configuration:

**User config** (`~/.config/ozx-roomeditor/config.json` on macOS/Linux,
`%APPDATA%/ozx-roomeditor/config.json` on Windows). Created on first run:

```json
{
  "project_root": "",
  "template_subdir": "Assets/Resources/TilemapData",
  "port": 8090,
  "auto_open_browser": true
}
```

Point `project_root` at an OZX project (e.g. an `ozx_base` checkout) and the
editor will save templates under `project_root/template_subdir`. Leave it
empty to use `~/.local/share/ozx-roomeditor/templates` as a fallback. Pass
`--config <path>` to use a different file.

**Runtime env** (deployment knobs only; see `.env.example`):

| Variable | Default | Purpose |
|---|---|---|
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `CORS_ALLOWED_ORIGINS` | _(empty → `*`)_ | Comma-separated allowed origins |

## API

Base URL: `http://localhost:8090/api/v1`

| Method | Path | Description |
|---|---|---|
| `POST` | `/templates` | Create template |
| `GET` | `/templates` | List templates |
| `GET` | `/templates/{id}` | Get template |
| `DELETE` | `/templates/{id}` | Delete template |
| `POST` | `/templates/validate?strict=true` | Validate template payload |
| `POST` | `/generate/fullroom` | Generate full room |
| `POST` | `/generate/bridge` | Generate bridge room |
| `POST` | `/generate/platform` | Generate platform room |
| `GET` | `/stage-configs` | All stage type configurations |
| `GET` | `/config` | Resolved user config (project_root, templates_dir, ...) |
| `GET` | `/health` | Stats the templates dir and returns 200 if reachable |

## Architecture

```
cmd/server/                Entry point + flag parsing + logger
internal/
  ├── config/             User config file load/save (ORT-67)
  ├── http/               chi router, handlers, middleware
  ├── store/              Store interface, StubStore, fsstore (filesystem impl)
  ├── model/              Template, request/response types
  ├── generate/           Room generators, stage rules, main-path planning
  └── validate/           Structural + logical validation
```

### Filesystem store layout
- Each template is a single JSON file: `<TEMPLATES_DIR>/<uuid>.json`.
- Writes go to `<file>.json.tmp` then `rename`, so a crash mid-write never
  leaves a half-written `.json` behind.
- A single `RWMutex` serializes mutations; reads run in parallel.
- Listing reads every file in the directory, applies the same filters as the
  old SQL store, sorts by `UpdatedAt` desc, and paginates. Fine for the
  hundreds of templates a local user is realistically going to have.

## Testing

```bash
make test-unit             # all unit tests
make test-coverage         # coverage report → coverage.html
```

## Where the cloud version lives

The PostgreSQL-backed cloud variant lives on `main`. This branch removes it
intentionally — see `ORT-65` for context and the rest of the
`local-client` epic (ORT-66..ORT-69) for what's next.
