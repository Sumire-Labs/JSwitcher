# JavaSwitcher Makefile

# Variables
APP_NAME=javaswitcher
VERSION=0.1.0
BUILD_DIR=build
DIST_DIR=dist

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: all build clean test deps windows linux macos

all: clean deps build

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for current platform
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd

# Build for Windows
windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd

# Build for Linux
linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd

# Build for macOS
macos:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd

# Build for all platforms
build-all: windows linux macos

# Create distribution packages
dist: build-all
	mkdir -p $(DIST_DIR)
	# Windows
	zip -j $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe README.md
	# Linux
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-amd64 -C .. README.md
	# macOS
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-amd64 -C .. README.md

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)

# Development run
run:
	$(GOCMD) run .

# Install locally
install: build
	cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build for current platform"
	@echo "  windows     - Build for Windows"
	@echo "  linux       - Build for Linux"
	@echo "  macos       - Build for macOS"
	@echo "  build-all   - Build for all platforms"
	@echo "  dist        - Create distribution packages"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  run         - Run in development mode"
	@echo "  install     - Install locally"
	@echo "  help        - Show this help"