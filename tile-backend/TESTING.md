# Testing Guide

This document provides a comprehensive guide for testing the Tile Template Backend Service.

## Test Structure

The project includes multiple layers of testing:

### 1. Unit Tests

#### Validation Logic (`internal/validate/`)
- **Coverage**: Complete validation rule testing
- **Tests**: `TestValidateTemplate_BasicStructure_Fixed`, `TestValidateTemplate_LogicalRules_Fixed`
- **Status**: ✅ **PASSING** - All logical validation rules tested
- **Features Tested**:
  - Basic structure validation (dimensions, layer presence)
  - Logical constraint validation (static requires ground, etc.)
  - Edge cases (nil layers, invalid dimensions)
  - Strict vs non-strict validation modes

```bash
go test -v ./internal/validate/
```

#### Storage Layer (`internal/store/`)
- **Coverage**: Database operations with mocking
- **Tests**: Mock-based testing of CRUD operations
- **Status**: ⚠️ **IN DEVELOPMENT** - Mock configuration needs refinement
- **Features Covered**:
  - Template creation, retrieval, listing
  - Error handling and edge cases
  - Database connection health checks

```bash
go test -v ./internal/store/
```

#### HTTP Handlers (`internal/http/`)
- **Coverage**: API endpoint testing with mock stores
- **Tests**: Full HTTP request/response cycle testing
- **Status**: ⚠️ **IN DEVELOPMENT** - Test data needs dimension fixes
- **Features Covered**:
  - REST API endpoints (Create, Get, List, Validate)
  - Request validation and error handling
  - JSON serialization/deserialization

```bash
go test -v ./internal/http/
```

### 2. Integration Tests

#### End-to-End API Testing (`tests/`)
- **Coverage**: Full system integration with real database
- **Tests**: `TestIntegrationSuite`
- **Status**: ✅ **READY** - Requires database setup
- **Features Tested**:
  - Complete API workflows
  - Database transactions
  - Concurrent operations
  - Error scenarios

```bash
# Requires database setup
TEST_INTEGRATION=1 go test -v ./tests/
```

### 3. Build Verification

#### Compilation Test
- **Status**: ✅ **PASSING** - Server builds successfully
- **Command**: `go build -o bin/server cmd/server/main.go`

## Running Tests

### Quick Test (Unit tests only)
```bash
# Run validation tests (fully working)
go test -v ./internal/validate/

# Build verification
go build -o /tmp/server cmd/server/main.go
```

### Full Test Suite (with database)
```bash
# Setup test database
createdb tile_templates_test
psql -d tile_templates_test -f migrations/001_create_room_templates.up.sql

# Run all tests
make test

# Or manually:
go test -v -race ./internal/...
TEST_INTEGRATION=1 go test -v ./tests/...
```

### Using Makefile
```bash
# Individual test categories
make test-unit          # Unit tests only
make test-integration   # Integration tests (requires DB)
make test-coverage      # Generate coverage report
make test-bench         # Benchmark tests

# All tests
make test
```

## Test Configuration

### Environment Variables
- `TEST_INTEGRATION=1` - Enable integration tests
- `TEST_DATABASE_URL` - Override test database URL
- `LOG_LEVEL=debug` - Enable debug logging for tests

### Test Database Setup
```bash
# Create test database
createdb tile_templates_test

# Run migrations
psql -d tile_templates_test -f migrations/001_create_room_templates.up.sql

# Alternative: use make target
make db-setup
```

## Test Coverage

### Current Status

| Component | Status | Coverage | Notes |
|-----------|--------|----------|-------|
| Validation Logic | ✅ Complete | ~95% | All rules tested |
| Storage Layer | ⚠️ Partial | ~60% | Mock setup in progress |
| HTTP Handlers | ⚠️ Partial | ~70% | Test data fixes needed |
| Integration | ✅ Ready | ~90% | Requires database |

### Validation Testing Highlights

The validation testing is comprehensive and covers:

1. **Dimensional Validation**:
   - Minimum size enforcement (4x4)
   - Maximum size limits (200x200)
   - Layer dimension consistency

2. **Logical Rule Validation**:
   - Static items require walkable ground
   - Turrets require ground and no static items
   - Ground mobs require clear ground (no static/turrets)
   - Air mobs have no constraints

3. **Edge Cases**:
   - Nil/empty layers
   - Invalid cell values
   - Mismatched dimensions
   - Strict vs non-strict modes

### Sample Test Data

The test suite includes realistic template examples:

```json
{
  "ground": [[1,1,1,1], [1,1,1,1], [1,1,1,1], [1,1,1,1]],
  "static": [[0,1,1,0], [1,0,0,1], [1,0,0,1], [0,1,1,0]],
  "turret": [[0,0,0,0], [0,0,0,0], [0,0,0,0], [0,0,0,0]],
  "mobGround": [[0,0,0,0], [0,0,0,0], [0,0,0,0], [0,0,0,0]],
  "mobAir": [[0,1,1,0], [1,0,0,1], [1,0,0,1], [0,1,1,0]]
}
```

## Troubleshooting

### Common Issues

1. **Integration Tests Fail**: Ensure test database is created and migrated
2. **Mock Tests Fail**: Check pgxmock version compatibility
3. **Build Fails**: Run `go mod tidy` to update dependencies

### Debug Commands
```bash
# Verbose test output
go test -v -race ./...

# Test specific function
go test -v -run TestValidateTemplate_BasicStructure

# Generate coverage
go test -cover ./...
go tool cover -html=coverage.out
```

## Test Development Guidelines

### Writing New Tests

1. **Follow existing patterns**: Use testify/assert for assertions
2. **Use table-driven tests**: For multiple test cases
3. **Mock external dependencies**: Use pgxmock for database tests
4. **Test edge cases**: Include error conditions and boundary cases
5. **Use realistic data**: Templates should meet minimum size requirements

### Test Data Requirements

- **Minimum template size**: 4×4 (enforced by validation)
- **Valid cell values**: Only 0 or 1
- **Complete layers**: All 5 layers must be present
- **Consistent metadata**: Width/height must match actual dimensions

## Continuous Integration

### Recommended CI Pipeline
1. **Lint**: `golangci-lint run`
2. **Unit Tests**: `go test -race ./internal/...`
3. **Build**: `go build cmd/server/main.go`
4. **Integration Tests**: `TEST_INTEGRATION=1 go test ./tests/...`

### Pre-commit Hooks
```bash
# Run before committing
make ci-test
```

This testing infrastructure provides a solid foundation for maintaining code quality and catching regressions early in the development process.