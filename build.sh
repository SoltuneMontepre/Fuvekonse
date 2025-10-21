#!/bin/bash

# Build script for Fuvekonse services
# This script ensures swag CLI is installed and docs are generated before building

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to build a service
build_service() {
    local service=$1
    local service_dir="services/$service"
    
    print_status "Building $service..."
    
    # Check if service directory exists
    if [ ! -d "$service_dir" ]; then
        print_error "Service directory $service_dir does not exist"
        return 1
    fi
    
    # Change to service directory
    cd "$service_dir"
    
    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found in $service_dir"
        return 1
    fi
    
    # Install swag CLI if not already installed
    if ! command_exists swag; then
        print_status "Installing swag CLI..."
        go install github.com/swaggo/swag/cmd/swag@latest
        print_success "swag CLI installed"
    else
        print_status "swag CLI already installed"
    fi
    
    # Generate Swagger documentation
    print_status "Generating Swagger documentation for $service..."
    swag init -g cmd/main.go -o ./docs
    print_success "Swagger docs generated for $service"
    
    # Download dependencies
    print_status "Downloading dependencies for $service..."
    go mod download
    go mod verify
    
    # Build main service
    print_status "Building main service for $service..."
    go build -o ./tmp/main ./cmd/main.go
    print_success "$service main service built successfully"
    
    # Build migrate tool
    print_status "Building migrate tool for $service..."
    go build -o ./tmp/migrate ./cmd/migrate/main.go
    print_success "$service migrate tool built successfully"
    
    # Run tests
    print_status "Running tests for $service..."
    go test -v ./...
    print_success "$service tests passed"
    
    # Run go vet
    print_status "Running go vet for $service..."
    go vet ./...
    print_success "$service go vet passed"
    
    # Check formatting
    print_status "Checking code formatting for $service..."
    if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
        print_warning "The following files are not formatted:"
        gofmt -s -l .
        print_warning "Run 'gofmt -s -w .' to fix formatting"
    else
        print_success "$service code formatting is correct"
    fi
    
    # Return to root directory
    cd - > /dev/null
    
    print_success "$service build completed successfully!"
    echo ""
}

# Main execution
main() {
    print_status "Starting Fuvekonse services build process..."
    echo ""
    
    # Check if Go is installed
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.25 or later."
        exit 1
    fi
    
    # Check Go version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_status "Using Go version: $go_version"
    
    # Build all services
    services=("general-service" "ticket-service")
    
    for service in "${services[@]}"; do
        build_service "$service"
    done
    
    print_success "All services built successfully! 🎉"
    echo ""
    print_status "Build artifacts are available in:"
    for service in "${services[@]}"; do
        echo "  - services/$service/tmp/"
    done
}

# Run main function
main "$@"
