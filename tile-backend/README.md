# Tile Template Backend Service

A lightweight backend service for storing and managing room tile templates using Golang and PostgreSQL.

## Features

- **REST API** for template management (Create, List, Get)
- **PostgreSQL storage** with JSONB for flexible template data
- **Comprehensive validation** with rule-based constraints
- **CORS support** for frontend integration
- **Health checks** and monitoring
- **Graceful shutdown** and error recovery

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 12+
- Docker (optional)

### Setup

1. **Clone and build**:
   ```bash
   cd tile-backend
   go mod tidy
   go build -o bin/server cmd/server/main.go
   ```

2. **Setup PostgreSQL**:
   ```bash
   # Create database
   createdb tile_templates
   
   # Run migrations
   psql -d tile_templates -f migrations/001_create_room_templates.up.sql
   ```

3. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your database settings
   ```

4. **Run server**:
   ```bash
   ./bin/server
   # or
   go run cmd/server/main.go
   ```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### 1. Create Template
**POST** `/templates`

Create a new room template.

**Request Body:**
```json
{
  "name": "room-20x12-v1",
  "payload": {
    "ground": [[0,1,0,...], ...],
    "static": [[0,0,1,...], ...],
    "turret": [[0,0,0,...], ...],
    "mobGround": [[0,0,0,...], ...],
    "mobAir": [[0,1,0,...], ...],
    "meta": {
      "name": "room-20x12-v1",
      "version": 1,
      "width": 20,
      "height": 12
    }
  }
}
```

**Response (201):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "room-20x12-v1",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### 2. List Templates
**GET** `/templates?limit=20&offset=0&name_like=room`

List templates with pagination and optional filtering.

**Query Parameters:**
- `limit` (optional): Number of results (default: 20, max: 100)
- `offset` (optional): Number of results to skip (default: 0)
- `name_like` (optional): Filter by name (case-insensitive partial match)

**Response (200):**
```json
{
  "total": 123,
  "items": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "room-20x12-v1",
      "version": 1,
      "width": 20,
      "height": 12,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

#### 3. Get Template
**GET** `/templates/{id}`

Retrieve a template by ID.

**Response (200):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "room-20x12-v1",
  "version": 1,
  "width": 20,
  "height": 12,
  "payload": {
    "ground": [[0,1,0,...], ...],
    "static": [[0,0,1,...], ...],
    "turret": [[0,0,0,...], ...],
    "mobGround": [[0,0,0,...], ...],
    "mobAir": [[0,1,0,...], ...]
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### 4. Validate Template
**POST** `/templates/validate?strict=true`

Validate a template without saving it.

**Query Parameters:**
- `strict` (optional): Enable strict logical validation (default: false)

**Request Body:** Same as Create Template payload

**Response (200):**
```json
{
  "valid": true,
  "errors": []
}
```

#### 5. Health Check
**GET** `/health`

Check service health and database connectivity.

**Response (200):**
```json
{
  "status": "healthy"
}
```

## Validation Rules

### Basic Structure Validation
- Width and height must be between 4 and 200
- All 5 layers must be present
- Each layer must have correct dimensions (height×width)
- All cell values must be 0 or 1

### Logical Validation (Strict Mode)
- **Static**: `static==1` requires `ground==1`
- **Turret**: `turret==1` requires `ground==1 AND static==0`
- **MobGround**: `mobGround==1` requires `ground==1 AND static==0 AND turret==0`
- **MobAir**: No constraints (can be placed anywhere)

## Database Schema

### Table: `room_templates`

| Column | Type | Description |
|--------|------|-------------|
| id | uuid (PK) | Unique identifier |
| name | text | Template name |
| version | int | Template version |
| width | int | Grid width |
| height | int | Grid height |
| payload | jsonb | Complete template data |
| created_at | timestamptz | Creation timestamp |
| updated_at | timestamptz | Last update timestamp |

### Indexes
- `idx_room_templates_created_at` (created_at DESC)
- `idx_room_templates_name` (name)
- `idx_room_templates_payload_gin` (payload GIN)

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | See .env.example | PostgreSQL connection string |
| `PORT` | 8080 | HTTP server port |
| `LOG_LEVEL` | info | Logging level (debug, info, warn, error) |
| `CORS_ALLOWED_ORIGINS` | localhost origins | Comma-separated CORS origins |

## Error Handling

### HTTP Status Codes
- **200**: Success
- **201**: Created
- **400**: Bad Request (validation errors, malformed JSON)
- **404**: Not Found
- **413**: Request Entity Too Large (>2MB)
- **500**: Internal Server Error
- **503**: Service Unavailable (database connection failed)

### Error Response Format
```json
{
  "error": "Bad Request",
  "message": "Template validation failed",
  "details": {
    "static_5_10": "static items require walkable ground",
    "turret_8_3": "turrets cannot be placed on static items"
  }
}
```

## Development

### Project Structure
```
tile-backend/
├── cmd/server/           # Application entry point
├── internal/
│   ├── http/            # HTTP handlers and middleware
│   ├── store/           # Database storage layer
│   ├── model/           # Data models and types
│   └── validate/        # Validation logic
├── migrations/          # Database migration files
├── go.mod              # Go module definition
└── .env.example        # Environment configuration example
```

### Building
```bash
# Development build
go build -o bin/server cmd/server/main.go

# Production build with optimizations
go build -ldflags="-s -w" -o bin/server cmd/server/main.go
```

### Running Migrations
```bash
# Up migration
psql -d tile_templates -f migrations/001_create_room_templates.up.sql

# Down migration (if needed)
psql -d tile_templates -f migrations/001_create_room_templates.down.sql
```

### Testing

The project includes comprehensive unit tests and integration tests:

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests (requires database)
make test-integration

# Run tests with coverage report
make test-coverage

# Run benchmark tests
make test-bench

# Run tests manually
go test -v -race ./internal/...                    # Unit tests
TEST_INTEGRATION=1 go test -v ./tests/...          # Integration tests
```

#### Test Configuration

1. **Unit Tests**: Mock-based tests that don't require external dependencies
2. **Integration Tests**: End-to-end tests that require a PostgreSQL database

For integration tests, set up a test database:
```bash
# Setup test database
make db-setup

# Or manually:
createdb tile_templates_test
psql -d tile_templates_test -f migrations/001_create_room_templates.up.sql
```

#### Test Coverage

The test suite provides:
- **Validation Logic**: Complete coverage of template validation rules
- **Storage Layer**: Database operations with mocked connections
- **HTTP Handlers**: API endpoint testing with mock stores
- **Integration Tests**: Full API workflows with real database
- **Concurrent Operations**: Multi-threaded safety testing
- **Error Scenarios**: Comprehensive error handling validation

#### Example API Test

```bash
# Create a template
curl -X POST http://localhost:8080/api/v1/templates \
  -H "Content-Type: application/json" \
  -d @tests/example_templates.json

# Validate a template
curl -X POST http://localhost:8080/api/v1/templates/validate?strict=true \
  -H "Content-Type: application/json" \
  -d @tests/example_templates.json
```

## Docker Support

### Dockerfile
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

### Docker Compose
```yaml
version: '3.8'
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_DB: tile_templates
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
  
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://postgres:password@db:5432/tile_templates?sslmode=disable
    depends_on:
      - db
```

## Production Considerations

### Security
- Use environment variables for sensitive configuration
- Enable HTTPS in production
- Consider adding authentication/authorization
- Implement rate limiting

### Performance
- Connection pooling configured for optimal performance
- GIN indexes on JSONB for fast metadata queries
- Request body size limited to 2MB
- Graceful shutdown with connection draining

### Monitoring
- Structured JSON logging with zap
- Health check endpoint for load balancer
- Request logging middleware
- Panic recovery middleware

This backend service provides a solid foundation for storing and managing tile templates with room for future enhancements like authentication, template versioning, and advanced querying capabilities.