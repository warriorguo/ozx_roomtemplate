# Tile Backend Makefile (local-client branch)

.PHONY: build build-prod build-local build-local-all build-frontend test test-unit clean run dev help deps lint fmt vet test-coverage test-bench profile-cpu profile-mem

# Repo root, used to locate the Vite frontend build.
REPO_ROOT := $(abspath ..)
LOCAL_BIN := ozx-roomeditor
LOCAL_DIST := internal/web/dist

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

# Standalone local-client binary: embeds the built SPA via go:embed and
# launches the user's browser on start.
build-frontend: ## Build the React SPA into ../dist
	@echo "Building frontend (vite build)..."
	@cd $(REPO_ROOT) && npm run build

build-local: build-frontend ## Build the bundled local-client binary (frontend + Go)
	@echo "Staging frontend bundle into $(LOCAL_DIST)..."
	@rm -rf $(LOCAL_DIST)
	@mkdir -p $(LOCAL_DIST)
	@cp -R $(REPO_ROOT)/dist/. $(LOCAL_DIST)/
	@echo "Building $(LOCAL_BIN) for host..."
	@go build -ldflags="-s -w" -o bin/$(LOCAL_BIN) ./cmd/ozx-roomeditor
	@echo "Built bin/$(LOCAL_BIN) ($$(du -h bin/$(LOCAL_BIN) | cut -f1))"

build-local-all: build-frontend ## Cross-compile ozx-roomeditor for darwin/linux/windows
	@rm -rf $(LOCAL_DIST)
	@mkdir -p $(LOCAL_DIST)
	@cp -R $(REPO_ROOT)/dist/. $(LOCAL_DIST)/
	@for target in darwin/amd64 darwin/arm64 linux/amd64 windows/amd64; do \
	  GOOS=$${target%/*}; GOARCH=$${target#*/}; \
	  ext=""; [ "$$GOOS" = "windows" ] && ext=".exe"; \
	  echo "→ $$GOOS/$$GOARCH"; \
	  GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="-s -w" -o bin/$(LOCAL_BIN)-$$GOOS-$$GOARCH$$ext ./cmd/ozx-roomeditor || exit 1; \
	done
	@echo "Cross-compiled binaries in bin/"

# Dependency management
deps: ## Download and install dependencies
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Testing targets
test: test-unit ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out ./internal/...

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

# Cleanup targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

clean-all: clean ## Clean everything including dependencies
	@echo "Cleaning dependencies..."
	@go clean -modcache

# CI/CD targets
ci-test: deps lint vet test-unit ## Run CI tests
	@echo "CI tests completed"

# Performance profiling
profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./internal/...
	@go tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./internal/...
	@go tool pprof mem.prof
