# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **room template editor** for game development, consisting of a React TypeScript frontend and a Go backend service. The editor creates tile-based room templates with a 5-layer system (ground, static, turret, mobGround, mobAir) and rule-based validation for game room layouts.

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

# Run server (port 8090)
go run cmd/server/main.go

# Run with hot-reload (requires air: go install github.com/air-verse/air@latest)
make dev

# Build binary
make build              # Development build
make build-prod         # Production build with optimizations

# Code quality
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint (requires golangci-lint)

# Run tests
make test              # All tests
make test-unit         # Unit tests only
make test-integration  # Integration tests (requires database)

# Run with coverage report
make test-coverage

# Database operations
createdb tile_templates
psql -d tile_templates -f migrations/001_create_room_templates.up.sql
psql -d tile_templates -f migrations/002_add_thumbnail.up.sql
```

### Testing
```bash
# Backend unit tests
go test -v -race ./internal/...

# Backend integration tests (set TEST_INTEGRATION=1)
TEST_INTEGRATION=1 go test -v ./tests/...

# No frontend tests currently configured
```

## Architecture

### Frontend Architecture (React + TypeScript + Zustand)

**State Management**: Zustand store (`src/store/newTemplateStore.ts`) manages:
- Template data (5 layers: ground, static, turret, mobGround, mobAir)
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
- 5 layers with hierarchical constraints enforced in validation
- Each layer is a 2D grid of 0s and 1s
- Ground layer can be auto-generated with room types (rectangular, cross, custom)

**Validation**:
- Real-time validation in `src/utils/newTemplateUtils.ts`
- Backend validation via API endpoint with strict mode
- Visual error feedback (red borders on invalid cells)

### Backend Architecture (Go + PostgreSQL)

**Clean Architecture Layers**:
```
cmd/server/           - Entry point and configuration
internal/
  ├── http/          - HTTP handlers, routing, middleware (chi router)
  ├── store/         - Database layer (PostgreSQL with pgx)
  ├── model/         - Data models and domain types
  └── validate/      - Validation logic (structure + logical constraints)
```

**API Endpoints** (Base: `/api/v1`):
- `POST /templates` - Create template
- `GET /templates?limit&offset&name_like` - List with pagination/search
- `GET /templates/{id}` - Get specific template
- `POST /templates/validate?strict` - Validate without saving
- `GET /health` - Health check

**Database**:
- PostgreSQL with `room_templates` table
- JSONB payload column for flexible template storage
- Thumbnail column (TEXT) for base64-encoded preview images
- Indexes on created_at, name, GIN index on payload, and conditional index on thumbnail
- Connection pooling with pgx (MaxConns: 25, MinConns: 5)

**Key Features**:
- CORS middleware for frontend integration
- Graceful shutdown with 30s timeout
- Structured logging with zap
- Request body size limit: 2MB
- Panic recovery middleware

### Validation Rules

**Hierarchical Constraints** (enforced in strict mode):
1. **Static layer**: `static==1` requires `ground==1`
2. **Turret layer**: `turret==1` requires `ground==1 AND static==0`
3. **MobGround layer**: `mobGround==1` requires `ground==1 AND static==0 AND turret==0`
4. **MobAir layer**: No constraints

**Structure Validation**:
- Dimensions: 4-200 for width/height
- All 5 layers required
- Correct grid dimensions (height × width)
- Cell values: only 0 or 1

### Data Flow

**Frontend ↔ Backend Integration**:
1. **API Service**: `src/services/api.ts` - HTTP client for all backend operations
2. **Converter**: `src/services/templateConverter.ts` - Bidirectional conversion between frontend and backend formats
3. **Environment**: `.env` file configures `VITE_API_BASE_URL` (default: `http://localhost:8090/api/v1`)

**Save Operation**:
```
Frontend Template → templateConverter → API Request → Go Handler → PostgreSQL
```

**Load Operation**:
```
PostgreSQL → Go Handler → API Response → templateConverter → Frontend Template
```

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

**Backend** (`tile-backend/.env`):
```
DATABASE_URL=postgres://user:password@localhost:5432/tile_templates?sslmode=disable
PORT=8090
LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174
TEST_DATABASE_URL=postgres://user:password@localhost:5432/tile_templates_test?sslmode=disable
TEST_INTEGRATION=0  # Set to 1 to run integration tests
```

**Note**: The .env.example file shows PORT=8080, but the actual configuration uses 8090. Make sure frontend VITE_API_BASE_URL matches the backend PORT.

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

# 2. Setup database
createdb tile_templates
psql -d tile_templates -f tile-backend/migrations/001_create_room_templates.up.sql
psql -d tile_templates -f tile-backend/migrations/002_add_thumbnail.up.sql

# 3. Configure environment
cp .env.example .env
cp tile-backend/.env.example tile-backend/.env
# Edit both .env files with your database settings

# 4. Start services (see "Running Full Stack Locally" above)
```

### Running Tests Before Committing
```bash
# Frontend: No automated tests currently
npm run build  # Verify build works

# Backend: Run all tests
cd tile-backend
make fmt && make vet     # Format and check code
make test-unit           # Fast unit tests
make test-integration    # Requires database setup
```

## Development Notes

### When Working with Templates
- Template format is consistent across frontend/backend with converter layer
- Backend stores complete payload in JSONB column
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

### Database Migrations
- Migration files in `tile-backend/migrations/`
- Apply all migrations in order:
  ```bash
  psql -d tile_templates -f migrations/001_create_room_templates.up.sql
  psql -d tile_templates -f migrations/002_add_thumbnail.up.sql
  ```
- Rollback migrations in reverse order:
  ```bash
  psql -d tile_templates -f migrations/002_add_thumbnail.down.sql
  psql -d tile_templates -f migrations/001_create_room_templates.down.sql
  ```

## Testing Strategy

**Backend Tests**:
- Unit tests use mocks (pgxmock) for database operations
- Integration tests require real PostgreSQL database
- Run integration tests with `TEST_INTEGRATION=1` environment variable
- Coverage reports generated with `make test-coverage`

**Frontend Testing**:
- No automated tests currently configured
- Manual testing through UI during development

## Additional Documentation

- `README.md` - Original 3-layer system (deprecated in favor of 5-layer)
- `README-5Layer.md` - Current 5-layer system documentation
- `FRONTEND_API_INTEGRATION.md` - Detailed API integration guide
- `tile-backend/README.md` - Backend API documentation
- `tile-backend/TESTING.md` - Backend testing guide
