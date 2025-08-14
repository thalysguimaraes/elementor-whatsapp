# Elementor WhatsApp Manager Makefile
# Build configuration for ewctl

# Variables
BINARY_NAME := ewctl
MAIN_PATH := cmd/ewctl/main.go
BUILD_DIR := bin
DIST_DIR := dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GOFLAGS := -v
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: all help build run clean test lint fmt deps install dev release docker

## help: Display this help message
help:
	@echo "$(BLUE)Elementor WhatsApp Manager - Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make <target>"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## all: Build the application for the current platform
all: clean build

## build: Build the binary for the current platform
build:
	@echo "$(BLUE)Building $(BINARY_NAME) for current platform...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## run: Build and run the application
run: build
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME)

## dev: Run the application in development mode with live reload
dev:
	@echo "$(BLUE)Starting development mode...$(NC)"
	@if ! command -v air &> /dev/null; then \
		echo "$(YELLOW)Installing air for live reload...$(NC)"; \
		go install github.com/air-verse/air@latest; \
	fi
	@air

## clean: Remove build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "$(GREEN)✓ Clean complete$(NC)"

## test: Run tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	@go test $(GOFLAGS) -race -cover ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test $(GOFLAGS) -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## lint: Run linters
lint:
	@echo "$(BLUE)Running linters...$(NC)"
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run ./...
	@echo "$(GREEN)✓ Linting complete$(NC)"

## fmt: Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@go mod tidy
	@echo "$(GREEN)✓ Formatting complete$(NC)"

## deps: Download and tidy dependencies
deps:
	@echo "$(BLUE)Managing dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## install: Install the binary to GOPATH/bin
install: build
	@echo "$(BLUE)Installing $(BINARY_NAME) to GOPATH/bin...$(NC)"
	@go install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(MAIN_PATH)
	@echo "$(GREEN)✓ Installation complete$(NC)"

## uninstall: Remove the binary from GOPATH/bin
uninstall:
	@echo "$(BLUE)Uninstalling $(BINARY_NAME)...$(NC)"
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstallation complete$(NC)"

## release: Build releases for all platforms
release: clean
	@echo "$(BLUE)Building releases for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		OUTPUT_NAME=$(DIST_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}; \
		if [ "$${platform%/*}" = "windows" ]; then \
			OUTPUT_NAME=$$OUTPUT_NAME.exe; \
		fi; \
		echo "  Building for $$platform..."; \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} go build $(GOFLAGS) \
			-ldflags "$(LDFLAGS)" \
			-o $$OUTPUT_NAME \
			$(MAIN_PATH); \
	done
	@echo "$(GREEN)✓ All releases built in $(DIST_DIR)$(NC)"

## docker: Build Docker image
docker:
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "$(GREEN)✓ Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"

## worker-deploy: Deploy Cloudflare Worker
worker-deploy:
	@echo "$(BLUE)Deploying Cloudflare Worker...$(NC)"
	@cd worker && npm run deploy
	@echo "$(GREEN)✓ Worker deployed$(NC)"

## worker-dev: Run Cloudflare Worker in development mode
worker-dev:
	@echo "$(BLUE)Starting Cloudflare Worker development server...$(NC)"
	@cd worker && npm run dev

## setup: Set up development environment
setup: deps
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/goreleaser/goreleaser/v2@latest
	@echo "$(GREEN)✓ Development environment ready$(NC)"

## check: Run all checks (fmt, lint, test)
check: fmt lint test
	@echo "$(GREEN)✓ All checks passed$(NC)"

## version: Display version information
version:
	@echo "$(BINARY_NAME) version $(VERSION)"
	@echo "  commit: $(COMMIT)"
	@echo "  built:  $(DATE)"

# Default target
.DEFAULT_GOAL := help