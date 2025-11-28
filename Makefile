.PHONY: help build test bench clean run server fmt lint tidy install-tools docker docker-up docker-down

# Default target
help:
	@echo "Mini KV Store Go - Make targets:"
	@echo ""
	@echo "  make build       - Build release binaries"
	@echo "  make test        - Run all tests"
	@echo "  make bench       - Run benchmarks"
	@echo "  make run         - Run CLI"
	@echo "  make server      - Run HTTP server"
	@echo "  make docker      - Build Docker image"
	@echo "  make docker-up   - Start Docker Compose cluster"
	@echo "  make docker-down - Stop Docker Compose cluster"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Run linter"
	@echo "  make tidy        - Tidy dependencies"
	@echo "  make clean       - Clean build artifacts"
	@echo ""

# Build binaries
build:
	@echo "Building binaries..."
	@go build -o bin/kvstore ./cmd/kvstore
	@go build -o bin/volume-server ./cmd/volume-server
	@echo "✓ Build complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Run CLI
run: build
	@./bin/kvstore

# Run server
server: build
	@./bin/volume-server

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Format complete"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✓ Tidy complete"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ Tools installed"

# Docker targets
docker:
	@echo "Building Docker image..."
	@docker build -t mini-kvstore-go:latest .
	@echo "✓ Docker image built"

docker-up:
	@echo "Starting Docker Compose cluster..."
	@docker-compose up -d
	@echo "✓ Cluster started"

docker-down:
	@echo "Stopping Docker Compose cluster..."
	@docker-compose down
	@echo "✓ Cluster stopped"

docker-logs:
	@docker-compose logs -f

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf db/
	@rm -rf data/
	@rm -rf testdata/
	@rm -rf examples/*/data/
	@rm -rf volume_data_*
	@echo "✓ Clean complete"

# CI target
ci: fmt lint test
	@echo "✓ CI checks passed"

# Pre-commit checks
pre-commit: fmt test
	@echo "✓ Pre-commit checks passed"
