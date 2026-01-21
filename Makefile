.PHONY: build test lint clean install dev release

# Build variables
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

# Default target
all: lint test build

# Build the binary
build:
	go build $(LDFLAGS) -o bin/fitwatch ./cmd/fitwatch

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/fitwatch-linux-amd64 ./cmd/fitwatch
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/fitwatch-linux-arm64 ./cmd/fitwatch
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/fitwatch-darwin-amd64 ./cmd/fitwatch
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/fitwatch-darwin-arm64 ./cmd/fitwatch
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/fitwatch-windows-amd64.exe ./cmd/fitwatch

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w -local github.com/user/fitwatch .

# Clean build artifacts
clean:
	rm -rf bin/ dist/ coverage.out coverage.html

# Install locally
install: build
	cp bin/fitwatch $(GOPATH)/bin/

# Run in development mode (watches for file changes)
dev:
	go run ./cmd/fitwatch -v

# Initialize config for development
dev-init:
	go run ./cmd/fitwatch --init

# Run as one-shot sync
dev-once:
	go run ./cmd/fitwatch --once -v

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  build-all   - Build for all platforms"
	@echo "  test        - Run tests"
	@echo "  test-integ  - Run integration tests (needs INTERVALS_* env vars)"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  clean       - Remove build artifacts"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  setup       - Install git hooks and dev dependencies"
	@echo "  dev         - Run in development mode"
	@echo "  dev-init    - Initialize dev config"
	@echo "  dev-once    - Run one-shot sync"
	@echo "  pre-commit  - Run pre-commit checks manually"

# Install git hooks and dev dependencies
setup:
	@echo "Installing git hooks..."
ifeq ($(OS),Windows_NT)
	@powershell -Command "Copy-Item -Force scripts/pre-commit.ps1 ../.git/hooks/pre-commit.ps1"
	@powershell -Command "Set-Content -Path ../.git/hooks/pre-commit -Value '#!/bin/sh\npowershell.exe -ExecutionPolicy Bypass -File \"$$(git rev-parse --show-toplevel)/fitwatch/scripts/pre-commit.ps1\"'"
else
	@cp scripts/pre-commit ../.git/hooks/pre-commit
	@chmod +x ../.git/hooks/pre-commit
endif
	@echo "Installing dev dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✓ Setup complete!"

# Run pre-commit checks manually
pre-commit:
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File scripts/pre-commit.ps1
else
	@./scripts/pre-commit
endif

# Run integration tests (requires INTERVALS_ATHLETE_ID and INTERVALS_API_KEY)
test-integ:
	go test -v ./tests/integration/... -run TestIntervalsAPI

# Run short tests only (skip integration tests)
test-short:
	go test -short ./...
