# Room Template Editor

A full-stack visual editor for creating game room tile templates with a multi-layer system, rule-based validation, and backend storage.

## Features

- **Multi-Layer System**: 7 layers (Ground, SoftEdge, Bridge, Static, Turret, MobGround, MobAir) with hierarchical constraints
- **Visual Grid Editor**: Click and drag painting with real-time validation
- **Ground Auto-Generation**: Rule-driven room layout generation with door placement
- **Room Attributes**: Boss, Elite, Mob, Treasure, Teleport, Story tags
- **Room Types**: Full Room, Bridge, Platform classifications
- **Backend Storage**: PostgreSQL-based template persistence with thumbnails
- **Advanced Filtering**: Filter templates by doors, room type, and attributes
- **Docker Support**: Ready for containerized deployment

## Quick Start

### Prerequisites

- Node.js 18+
- Go 1.22+
- PostgreSQL 12+

### Setup

```bash
# 1. Install frontend dependencies
npm install

# 2. Setup backend
cd tile-backend
go mod tidy

# 3. Create database and run migrations
createdb tile_templates
psql -d tile_templates -f migrations/001_create_room_templates.up.sql
psql -d tile_templates -f migrations/002_add_thumbnail.up.sql
psql -d tile_templates -f migrations/003_add_computed_fields.up.sql

# 4. Configure environment
cd ..
cp .env.example .env
# Edit .env: VITE_API_BASE_URL=http://localhost:8090/api/v1

cp tile-backend/.env.example tile-backend/.env
# Edit tile-backend/.env with your database settings
```

### Run

```bash
# Terminal 1 - Backend (port 8090)
cd tile-backend
go run cmd/server/main.go

# Terminal 2 - Frontend (port 5173)
npm run dev
```

Open http://localhost:5173 in your browser.

## Layer System

### Layer Hierarchy

| Layer | Description | Constraints |
|-------|-------------|-------------|
| **Ground** | Walkable floor tiles | None |
| **SoftEdge** | Soft edge markers | `ground=1` |
| **Bridge** | Bridge tiles | `ground=1` |
| **Static** | Item placement zones | `ground=1` |
| **Turret** | Defensive positions | `ground=1 AND static=0` |
| **MobGround** | Ground enemy spawns | `ground=1 AND static=0 AND turret=0` |
| **MobAir** | Flying enemy spawns | None |

### Validation Rules

```
Static:    valid if static=0    OR ground=1
Turret:    valid if turret=0    OR (ground=1 AND static=0)
MobGround: valid if mobGround=0 OR (ground=1 AND static=0 AND turret=0)
MobAir:    always valid
```

Invalid cells are highlighted with red borders in real-time.

## Room Configuration

### Room Types

- **Full Room**: Standard enclosed room
- **Bridge**: Connecting corridor
- **Platform**: Floating platform area

### Room Attributes

- Boss, Elite, Mob, Treasure, Teleport, Story

### Door Connectivity

Doors can be placed on four sides (Top, Right, Bottom, Left) and automatically tracked for filtering.

## Data Format

```json
{
  "name": "example-room",
  "payload": {
    "ground": [[0,1,1,...], ...],
    "softEdge": [[0,0,0,...], ...],
    "bridge": [[0,0,0,...], ...],
    "static": [[0,0,1,...], ...],
    "turret": [[0,0,0,...], ...],
    "mobGround": [[0,0,0,...], ...],
    "mobAir": [[0,1,0,...], ...],
    "doors": { "top": 1, "right": 0, "bottom": 1, "left": 0 },
    "attributes": {
      "boss": false, "elite": true, "mob": true,
      "treasure": false, "teleport": false, "story": false
    },
    "roomType": "full",
    "meta": {
      "name": "example-room",
      "version": 1,
      "width": 20,
      "height": 12
    }
  }
}
```

## API Endpoints

Base URL: `http://localhost:8090/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/templates` | Create template |
| GET | `/templates` | List templates with filters |
| GET | `/templates/{id}` | Get template by ID |
| DELETE | `/templates/{id}` | Delete template |
| POST | `/templates/validate` | Validate template |
| GET | `/health` | Health check |

### Filter Parameters

```
GET /templates?room_type=full&has_boss=true&top_door_connected=true
```

- `name_like` - Search by name
- `room_type` - full, bridge, platform
- `has_boss`, `has_elite`, `has_mob`, `has_treasure`, `has_teleport`, `has_story`
- `top_door_connected`, `right_door_connected`, `bottom_door_connected`, `left_door_connected`

## Project Structure

```
ozx_roomtemplate/
├── src/                          # Frontend (React + TypeScript)
│   ├── components/new/           # UI components
│   │   ├── TileTemplateApp.tsx   # Main application
│   │   ├── LayerEditor.tsx       # Grid editor
│   │   ├── SaveLoadPanel.tsx     # Save/Load with filters
│   │   └── ToolBar.tsx           # Main toolbar
│   ├── services/
│   │   ├── api.ts                # Backend API client
│   │   └── templateConverter.ts  # Data format conversion
│   ├── store/
│   │   └── newTemplateStore.ts   # Zustand state management
│   └── utils/
│       └── newTemplateUtils.ts   # Validation logic
│
├── tile-backend/                 # Backend (Go + PostgreSQL)
│   ├── cmd/server/               # Entry point
│   ├── internal/
│   │   ├── http/                 # HTTP handlers
│   │   ├── store/                # Database layer
│   │   ├── model/                # Data models
│   │   └── validate/             # Validation logic
│   └── migrations/               # Database migrations
│
├── Dockerfile                    # Docker build
└── .env.example                  # Environment template
```

## Docker Deployment

```bash
# Build and run
docker build -t room-template-editor .
docker run -p 80:80 -e DATABASE_URL=... room-template-editor
```

The Dockerfile includes nginx for serving the frontend and proxying API requests to the backend.

## Development

### Frontend Commands

```bash
npm install          # Install dependencies
npm run dev          # Development server
npm run build        # Production build
npm run preview      # Preview build
```

### Backend Commands

```bash
cd tile-backend
go mod tidy          # Install dependencies
go run cmd/server/main.go  # Run server
make test            # Run tests
make build           # Build binary
```

### Testing

```bash
# Backend tests
cd tile-backend
make test-unit       # Unit tests
make test-integration # Integration tests (requires DB)
```

## Tech Stack

**Frontend:**
- React 18 + TypeScript
- Zustand (state management)
- Vite (build tool)

**Backend:**
- Go 1.22
- Chi (HTTP router)
- pgx (PostgreSQL driver)
- Zap (logging)

**Database:**
- PostgreSQL with JSONB storage
