.PHONY: help build test test-verbose test-coverage clean run-backend run-frontend version deps update-deps export

# Read version from git tags (fallback to "dev" if no tags exist)
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")

# Default target
help:
	@echo "Darts Web Application - Available targets:"
	@echo ""
	@echo "  make build           - Build Docker image (version $(VERSION))"
	@echo "  make export          - Build and export Docker image as tar"
	@echo "  make test            - Run all tests"
	@echo "  make test-verbose    - Run tests with verbose output"
	@echo "  make run-backend     - Run backend server locally"
	@echo "  make run-frontend    - Run frontend dev server"
	@echo "  make deps            - Install dependencies"
	@echo "  make update-deps     - Update all Go and JS dependencies"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make version         - Show current version"
	@echo ""

# Show version
version:
	@echo "Current version: $(VERSION)"

# Run backend tests
test:
	@echo "Running Go tests..."
	go test ./...

test-verbose:
	@echo "Running Go tests (verbose)..."
	go test -v ./...

test-coverage:
	@echo "Running Go tests with coverage..."
	go test -cover ./...

# Run backend locally
run-backend:
	@echo "Starting backend server..."
	go run cmd/server/main.go

# Run frontend dev server
run-frontend:
	@echo "Starting frontend dev server..."
	cd frontend && npm run dev

# Build Docker image
build:
	@echo "Building Docker image version $(VERSION)..."
	docker build --platform linux/amd64 \
		-t darts-app:$(VERSION) \
		-t darts-app:latest .
	@echo "✓ Image tagged as: darts-app:$(VERSION) and darts-app:latest"

# Build and export Docker image
export: build
	@echo "Exporting image to darts-app-$(VERSION).tar..."
	docker save darts-app:$(VERSION) -o darts-app-$(VERSION).tar
	@echo ""
	@echo "✓ Build complete!"
	@echo "✓ Image tagged as: darts-app:$(VERSION) and darts-app:latest"
	@echo "✓ Exported to: darts-app-$(VERSION).tar"
	@echo ""
	@echo "Next steps:"
	@echo "1. scp darts-app-$(VERSION).tar <user>@mini-pc:/tmp/"
	@echo "2. ssh <user>@mini-pc"
	@echo "3. minikube image load /tmp/darts-app-$(VERSION).tar"
	@echo "4. Update charts/darts-web/values.yaml with tag: \"$(VERSION)\""
	@echo "5. helm upgrade darts ./charts/darts-web"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f darts-server
	rm -f darts-app-*.tar
	rm -rf frontend/dist
	rm -rf frontend/node_modules/.vite
	@echo "✓ Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "✓ Dependencies installed"

# Update all dependencies
update-deps:
	@echo "Updating all dependencies..."
	@echo "Updating Go dependencies..."
	go get -u ./...
	go mod tidy
	@echo "Updating frontend dependencies..."
	cd frontend && npm update
	@echo "✓ All dependencies updated"
