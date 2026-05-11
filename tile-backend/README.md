# Tile Template Backend (local-client branch)

> **Status:** this branch is the foundation for the **local-client** variant of the
> room template editor. The PostgreSQL backend has been removed; a filesystem-
> backed `Store` will land in **ORT-66**, and the standalone binary with browser
> auto-launch in **ORT-68**. Until then the binary builds and serves
> `/health`, but every template endpoint returns `ErrNotImplemented`.

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

Environment variables (see `.env.example`):

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8090` | HTTP listen port |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `CORS_ALLOWED_ORIGINS` | _(empty → `*`)_ | Comma-separated allowed origins |

## API

Base URL: `http://localhost:8090/api/v1`

| Method | Path | Description |
|---|---|---|
| `POST` | `/templates` | Create template (stub → 500 until ORT-66) |
| `GET` | `/templates` | List templates (stub → 500 until ORT-66) |
| `GET` | `/templates/{id}` | Get template (stub → 500 until ORT-66) |
| `DELETE` | `/templates/{id}` | Delete template (stub → 500 until ORT-66) |
| `POST` | `/templates/validate?strict=true` | Validate template payload |
| `POST` | `/generate/fullroom` | Generate full room |
| `POST` | `/generate/bridge` | Generate bridge room |
| `POST` | `/generate/platform` | Generate platform room |
| `GET` | `/stage-configs` | All stage type configurations |
| `GET` | `/health` | Health check (always 200 with the stub store) |

The `/validate` and `/generate/*` endpoints work today — they have no storage
dependency.

## Architecture

```
cmd/server/                Entry point + config + logger
internal/
  ├── http/               chi router, handlers, middleware
  ├── store/              Store interface + StubStore (filesystem impl in ORT-66)
  ├── model/              Template, request/response types
  ├── generate/           Room generators, stage rules, main-path planning
  └── validate/           Structural + logical validation
```

## Testing

```bash
make test-unit             # all unit tests
make test-coverage         # coverage report → coverage.html
```

## Where the cloud version lives

The PostgreSQL-backed cloud variant lives on `main`. This branch removes it
intentionally — see `ORT-65` for context and the rest of the
`local-client` epic (ORT-66..ORT-69) for what's next.
