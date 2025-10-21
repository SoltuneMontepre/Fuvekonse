# Makefile for Fuvekonse services

.PHONY: help build test clean install-swag generate-docs build-all test-all

# Default target
help:
	@echo "Available targets:"
	@echo "  help          - Show this help message"
	@echo "  install-swag  - Install swag CLI"
	@echo "  generate-docs - Generate Swagger docs for all services"
	@echo "  build         - Build all services"
	@echo "  build-general - Build general-service only"
	@echo "  build-ticket  - Build ticket-service only"
	@echo "  test          - Run tests for all services"
	@echo "  test-general  - Run tests for general-service only"
	@echo "  test-ticket   - Run tests for ticket-service only"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker images"

# Install swag CLI
install-swag:
	@echo "Installing swag CLI..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ swag CLI installed"

# Generate Swagger documentation for all services
generate-docs: install-swag
	@echo "Generating Swagger docs for general-service..."
	cd services/general-service && swag init -g cmd/main.go -o ./docs
	@echo "Generating Swagger docs for ticket-service..."
	cd services/ticket-service && swag init -g cmd/main.go -o ./docs
	@echo "✅ Swagger docs generated for all services"

# Build all services
build: generate-docs
	@echo "Building all services..."
	@$(MAKE) build-general
	@$(MAKE) build-ticket
	@echo "✅ All services built successfully"

# Build general-service
build-general: generate-docs
	@echo "Building general-service..."
	cd services/general-service && \
		go mod download && \
		go mod verify && \
		go build -o ./tmp/main ./cmd/main.go && \
		go build -o ./tmp/migrate ./cmd/migrate/main.go
	@echo "✅ general-service built successfully"

# Build ticket-service
build-ticket: generate-docs
	@echo "Building ticket-service..."
	cd services/ticket-service && \
		go mod download && \
		go mod verify && \
		go build -o ./tmp/main ./cmd/main.go && \
		go build -o ./tmp/migrate ./cmd/migrate/main.go
	@echo "✅ ticket-service built successfully"

# Run tests for all services
test:
	@echo "Running tests for all services..."
	@$(MAKE) test-general
	@$(MAKE) test-ticket
	@echo "✅ All tests completed"

# Run tests for general-service
test-general:
	@echo "Running tests for general-service..."
	cd services/general-service && \
		go test -v ./... && \
		go vet ./...
	@echo "✅ general-service tests passed"

# Run tests for ticket-service
test-ticket:
	@echo "Running tests for ticket-service..."
	cd services/ticket-service && \
		go test -v ./... && \
		go vet ./...
	@echo "✅ ticket-service tests passed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf services/*/tmp/
	rm -rf services/*/docs/
	@echo "✅ Build artifacts cleaned"

# Build Docker images
docker-build: build
	@echo "Building Docker images..."
	cd services/general-service && docker build -t general-service:latest .
	cd services/ticket-service && docker build -t ticket-service:latest .
	@echo "✅ Docker images built successfully"

# Development targets
dev-general:
	@echo "Starting general-service in development mode..."
	cd services/general-service && air

dev-ticket:
	@echo "Starting ticket-service in development mode..."
	cd services/ticket-service && air

# CI targets
ci-build: generate-docs build test
	@echo "✅ CI build completed successfully"

ci-test: generate-docs test
	@echo "✅ CI tests completed successfully"
