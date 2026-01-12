.PHONY: help build build-linux build-darwin build-darwin-amd64 build-windows install test lint format clean run deps tidy

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Build variables
BINARY_NAME=guitargame
MAIN_PATH=.
BUILD_DIR=./dist
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

# Note: This project uses CGO (portaudio, aubio), so cross-compilation requires
# the appropriate C libraries and cross-compilers installed on the build machine.

build: ## Build for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: ## Build for Linux (amd64)
	@echo "Building $(BINARY_NAME) for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-linux-arm64: ## Build for Linux (arm64)
	@echo "Building $(BINARY_NAME) for Linux (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"

build-darwin: ## Build for macOS (arm64/Apple Silicon)
	@echo "Building $(BINARY_NAME) for macOS (arm64)..."
	@echo "Note: Requires portaudio and aubio installed via: brew install portaudio aubio"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

build-darwin-amd64: ## Build for macOS (amd64/Intel)
	@echo "Building $(BINARY_NAME) for macOS (amd64)..."
	@echo "Note: Requires portaudio and aubio installed via: brew install portaudio aubio"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

build-windows: ## Build for Windows (amd64) - requires MinGW cross-compiler
	@echo "Building $(BINARY_NAME) for Windows (amd64)..."
	@echo "Note: Requires MinGW and Windows versions of portaudio/aubio"
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

build-all: build-linux build-linux-arm64 build-darwin build-darwin-amd64 ## Build for all platforms (requires cross-compilers)

install: build ## Install binary to ~/.local/bin
	@echo "Installing $(BINARY_NAME)..."
	@mkdir -p $$HOME/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) $$HOME/.local/bin/
	@echo "✓ $(BINARY_NAME) installed to $$HOME/.local/bin/$(BINARY_NAME)"
	@case ":$$PATH:" in \
		*":$$HOME/.local/bin:"*) ;; \
		*) echo ""; \
		   echo "⚠ Warning: $$HOME/.local/bin is not in your PATH"; \
		   echo "Add this to your ~/.bashrc or ~/.zshrc:"; \
		   echo "  export PATH=\"\$$PATH:$$HOME/.local/bin\"" ;; \
	esac

test: ## Run tests
	@echo "Running tests..."
	@go test -race -cover ./...

lint: ## Run linters (requires golangci-lint)
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	@golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	@gofmt -s -w .
	@go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

run: build ## Build and run
	@$(BUILD_DIR)/$(BINARY_NAME)

# Module management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@go mod tidy

verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify

# System dependencies
deps-linux: ## Install system dependencies (Linux/Debian/Ubuntu)
	@echo "Installing system dependencies..."
	sudo apt-get update
	sudo apt-get install -y portaudio19-dev libaubio-dev libxkbcommon-dev libwayland-dev libvulkan-dev libxkbcommon-x11-dev libx11-xcb-dev

deps-macos: ## Install system dependencies (macOS)
	@echo "Installing system dependencies..."
	brew install portaudio aubio

prepush: format lint test build ## Run before pushing (format, lint, test, build)
