# Tile Backend Makefile

.PHONY: build test test-unit test-integration clean run dev help deps lint

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the server binary
	@echo "Building server..."
	@go build -o bin/server cmd/server/main.go
	@echo "Server built to bin/server"

build-prod: ## Build production binary with optimizations
	@echo "Building production server..."
	@go build -ldflags="-s -w" -o bin/server cmd/server/main.go
	@echo "Production server built to bin/server"

# Dependency management
deps: ## Download and install dependencies
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Testing targets
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out ./internal/...

test-integration: ## Run integration tests (requires database)
	@echo "Running integration tests..."
	@TEST_INTEGRATION=1 go test -v -race ./tests/...

test-coverage: test-unit ## Run unit tests and show coverage
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@go test -bench=. -benchmem ./internal/...

# Development targets
run: ## Run the server in development mode
	@echo "Starting development server..."
	@go run cmd/server/main.go

dev: ## Run the server with auto-reload (requires air)
	@echo "Starting development server with hot reload..."
	@air

# Code quality targets
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Database targets
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@psql -d $(DATABASE_URL) -f migrations/001_create_room_templates.up.sql

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@psql -d $(DATABASE_URL) -f migrations/001_create_room_templates.down.sql

db-setup: ## Setup test database
	@echo "Setting up test database..."
	@createdb tile_templates_test || true
	@psql -d tile_templates_test -f migrations/001_create_room_templates.up.sql

# Cleanup targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

clean-all: clean ## Clean everything including dependencies
	@echo "Cleaning dependencies..."
	@go clean -modcache

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t tile-backend .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 tile-backend

# CI/CD targets
ci-test: deps lint vet test-unit ## Run CI tests
	@echo "CI tests completed"

# Load testing
load-test: ## Run load tests (requires hey)
	@echo "Running load tests..."
	@hey -n 1000 -c 10 http://localhost:8080/health

# Security scanning
security-scan: ## Run security scanner
	@echo "Running security scanner..."
	@gosec ./...

# Generate mocks (if using mockery)
generate-mocks: ## Generate mocks for testing
	@echo "Generating mocks..."
	@mockery --all

# Environment setup
env-example: ## Create .env.example file
	@echo "Creating .env.example file..."
	@echo "DATABASE_URL=postgres://user:password@localhost:5432/tile_templates?sslmode=disable" > .env.example
	@echo "PORT=8080" >> .env.example
	@echo "LOG_LEVEL=info" >> .env.example
	@echo "CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174" >> .env.example

# Performance profiling
profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./internal/...
	@go tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./internal/...
	@go tool pprof mem.prof