.PHONY: build clean install test help

# Binary name
BINARY_NAME=pg_user_admin

# Build directory
BUILD_DIR=./bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main package path
MAIN_PATH=./cmd/pg_user_admin

help:
	@echo "PostgreSQL User Management Admin Tool - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build        Build the binary for current platform"
	@echo "  build-all    Build binaries for all platforms"
	@echo "  clean        Remove build artifacts"
	@echo "  install      Install the binary to GOPATH/bin"
	@echo "  test         Run tests"
	@echo "  tidy         Tidy go.mod"
	@echo "  help         Show this help message"

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: build-linux build-darwin build-windows
	@echo "All builds complete"

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Linux build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "macOS build complete: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Windows build complete: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(MAIN_PATH)
	@echo "Install complete"

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

tidy:
	@echo "Tidying go.mod..."
	$(GOMOD) tidy
	@echo "Tidy complete"
