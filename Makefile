# Docker for Android Makefile
# Version: 28.0.1.10

# Variables
VERSION := 28.0.1.10
DOCKER_VERSION := 28.0.1
SUB_VERSION := 10

# CDN and Server URLs
# Note: CDN doesn't cache .txt files; version.txt is always fresh from CDN
CDN_URL := https://fw.kspeeder.com/binary/docker-for-android
ORIGIN_SERVER_URL := https://fw.koolcenter.com/binary/docker-for-android

# Directories
RELEASE_DIR := release
ARM64_BIN_DIR := docker/arm64_bin
X86_64_BIN_DIR := docker/x86_64_bin
DOCKER_DIR := docker

# Package names
ARM64_PACKAGE := docker-for-android-bin-$(VERSION)-arm64.tar.gz
X86_64_PACKAGE := docker-for-android-bin-$(VERSION)-x86_64.tar.gz
VERSION_FILE := version.txt

# Targets
.PHONY: all clean build-release arm64 x86_64 version help

# Default target
all: help

# Help target
help:
	@echo "Docker for Android Build System"
	@echo "================================"
	@echo "Version: $(VERSION)"
	@echo ""
	@echo "Available targets:"
	@echo "  make build-release - Build release packages (currently arm64 only)"
	@echo "  make arm64         - Build arm64 package only"
	@echo "  make x86_64        - Build x86_64 package (not implemented yet)"
	@echo "  make version       - Generate version file"
	@echo "  make clean         - Clean release directory"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "CDN URL: $(CDN_URL)"
	@echo "Origin Server URL: $(ORIGIN_SERVER_URL)"

# Create release directory
create-dir:
	@mkdir -p $(RELEASE_DIR)

# Build arm64 package
arm64: create-dir
	@echo "Building arm64 package..."
	@if [ ! -d "$(ARM64_BIN_DIR)" ]; then \
		echo "Error: $(ARM64_BIN_DIR) directory not found!"; \
		exit 1; \
	fi
	@echo "Packaging $(ARM64_BIN_DIR) to $(RELEASE_DIR)/$(ARM64_PACKAGE)..."
	@tar -czf $(RELEASE_DIR)/$(ARM64_PACKAGE) -C $(DOCKER_DIR) arm64_bin
	@echo "Generating sha256 checksum for arm64 package..."
	@cd $(RELEASE_DIR) && shasum -a 256 $(ARM64_PACKAGE) > $(ARM64_PACKAGE).sha256
	@echo "arm64 package created: $(RELEASE_DIR)/$(ARM64_PACKAGE)"
	@echo "SHA256: $$(cat $(RELEASE_DIR)/$(ARM64_PACKAGE).sha256)"

# Build x86_64 package (placeholder for future implementation)
x86_64: create-dir
	@echo "x86_64 package build not implemented yet"
	@echo "Please add x86_64_bin directory first"

# Generate version file
version: create-dir
	@echo "Generating version file..."
	@echo "# Docker for Android Version File" > $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "# Generated on $$(date '+%Y-%m-%d %H:%M:%S')" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "VERSION=$(VERSION)" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "DOCKER_VERSION=$(DOCKER_VERSION)" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "SUB_VERSION=$(SUB_VERSION)" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "# Download URLs" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "CDN_URL=$(CDN_URL)" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "ORIGIN_SERVER_URL=$(ORIGIN_SERVER_URL)" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@echo "# Package information" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@if [ -f "$(RELEASE_DIR)/$(ARM64_PACKAGE).sha256" ]; then \
		echo "ARM64_PACKAGE=$(ARM64_PACKAGE)" >> $(RELEASE_DIR)/$(VERSION_FILE); \
		echo "ARM64_SHA256=$$(cat $(RELEASE_DIR)/$(ARM64_PACKAGE).sha256 | cut -d' ' -f1)" >> $(RELEASE_DIR)/$(VERSION_FILE); \
	else \
		echo "ARM64_PACKAGE=$(ARM64_PACKAGE)" >> $(RELEASE_DIR)/$(VERSION_FILE); \
		echo "ARM64_SHA256=<not generated yet>" >> $(RELEASE_DIR)/$(VERSION_FILE); \
	fi
	@echo "" >> $(RELEASE_DIR)/$(VERSION_FILE)
	@if [ -f "$(RELEASE_DIR)/$(X86_64_PACKAGE).sha256" ]; then \
		echo "X86_64_PACKAGE=$(X86_64_PACKAGE)" >> $(RELEASE_DIR)/$(VERSION_FILE); \
		echo "X86_64_SHA256=$$(cat $(RELEASE_DIR)/$(X86_64_PACKAGE).sha256 | cut -d' ' -f1)" >> $(RELEASE_DIR)/$(VERSION_FILE); \
	else \
		echo "X86_64_PACKAGE=$(X86_64_PACKAGE)" >> $(RELEASE_DIR)/$(VERSION_FILE); \
		echo "X86_64_SHA256=" >> $(RELEASE_DIR)/$(VERSION_FILE); \
	fi
	@echo "Version file created: $(RELEASE_DIR)/$(VERSION_FILE)"
	@echo ""
	@cat $(RELEASE_DIR)/$(VERSION_FILE)

# Build full release
build-release: clean arm64 version
	@echo ""
	@echo "=========================================="
	@echo "Release build completed!"
	@echo "=========================================="
	@echo "Version: $(VERSION)"
	@echo "Output directory: $(RELEASE_DIR)/"
	@echo ""
	@echo "Files generated:"
	@ls -lh $(RELEASE_DIR)/
	@echo ""
	@echo "Next steps:"
	@echo "1. Upload files to server: $(ORIGIN_SERVER_URL)"
	@echo "2. Sync to CDN (except version file)"
	@echo "3. Upload version file to server (no CDN cache)"

# Clean release directory
clean:
	@echo "Cleaning release directory..."
	@rm -rf $(RELEASE_DIR)
	@echo "Release directory cleaned."
