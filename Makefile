BINARY     := wollmilchsau
BUILD_DIR  := ./build
GO_FLAGS   := -trimpath
CGO_FLAGS  := CGO_ENABLED=1

# Use go list to get all packages
PACKAGES := $(shell go list ./...)

.PHONY: all build test clean install deps config help test-client fmt vet lint check

all: check build ## Run check and build the binary (default)

deps: ## Download dependencies and tidy mod file
	go mod download
	go mod tidy

fmt: ## Run go fmt on all source files
	go fmt $(PACKAGES)

vet: ## Run go vet on all source files
	$(CGO_FLAGS) go vet $(PACKAGES)

lint: ## Run golangci-lint with .golangci.yml config
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml ./...; \
	else \
		echo "golangci-lint not found, skipping linting. Install from: https://golangci-lint.run/"; \
	fi

check: fmt vet lint test ## Run fmt, vet, lint and unit tests

build: deps ## Build binary
	@mkdir -p $(BUILD_DIR)
	$(CGO_FLAGS) go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/main.go
	@echo "Built: $(BUILD_DIR)/$(BINARY)"

test: ## Run unit tests
	$(CGO_FLAGS) go test ./... -v -count=1

test-race: ## Run tests with race detector
	$(CGO_FLAGS) go test -race ./... -v -count=1

install: check ## Run check and install to ~/go/bin
	$(CGO_FLAGS) go install $(GO_FLAGS) ./cmd/main.go

clean: ## Clean build artifacts and temporary test files
	rm -rf $(BUILD_DIR)
	rm -f wollmilchsau

size: build ## Show binary size
	@ls -lh $(BUILD_DIR)/$(BINARY)

config: ## Print Claude Desktop config snippet
	@echo '{'
	@echo '  "mcpServers": {'
	@echo '    "wollmilchsau": {'
	@echo '      "command": "$(shell pwd)/build/$(BINARY)"'
	@echo '    }'
	@echo '  }'
	@echo '}'

test-client: build ## Run test client (stdio MCP integration)
	go run ./cmd/testclient/main.go

test-sse: build ## Run SSE integration test via curl/bash
	./scripts/test_sse.sh

# ##############################################################################
# # RELEASE & CROSS-COMPILATION (CGO CONSIDERATIONS)
# ##############################################################################
# Since this project uses 'v8go', it requires CGO_ENABLED=1. 
# Pure Go cross-compilation (e.g., GOOS=windows go build) will FAIL here because 
# CGO needs a platform-specific C++ toolchain (compiler & linker).
#
# RECOMMENDATION:
# 1. macOS Builds: Run on a Mac. It can build Intel & ARM and 'lipo' them.
# 2. Windows Builds: Run on Mac or Linux with 'mingw-w64' installed.
# 3. Linux Builds: Run on Linux (or use Docker on Mac/Windows).
#
# REQUIRED TOOLS for releases:
# - macOS: Xcode / Command Line Tools (for clang & lipo)
# - Windows: x86_64-w64-mingw32-gcc (brew install mingw-w64 / apt install mingw-w64)
# ##############################################################################

release-mac: deps check ## Build macOS Universal Binary (Intel + Apple Silicon)
	@mkdir -p $(BUILD_DIR)
	@echo "Building for macOS Intel (amd64)..."
	GOOS=darwin GOARCH=amd64 $(CGO_FLAGS) go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/main.go
	@echo "Building for macOS ARM (arm64)..."
	GOOS=darwin GOARCH=arm64 $(CGO_FLAGS) go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/main.go
	@echo "Creating Universal Binary..."
	lipo -create -output $(BUILD_DIR)/$(BINARY)-darwin-universal $(BUILD_DIR)/$(BINARY)-darwin-amd64 $(BUILD_DIR)/$(BINARY)-darwin-arm64
	@rm $(BUILD_DIR)/$(BINARY)-darwin-amd64 $(BUILD_DIR)/$(BINARY)-darwin-arm64
	@echo "Built: $(BUILD_DIR)/$(BINARY)-darwin-universal"

release-windows: deps check ## Build Windows Binary (requires mingw-w64)
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Windows (amd64) via MinGW..."
	GOOS=windows GOARCH=amd64 $(CGO_FLAGS) CC=x86_64-w64-mingw32-gcc go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY).exe ./cmd/main.go
	@echo "Built: $(BUILD_DIR)/$(BINARY).exe"

release-linux: deps check ## Build native Linux Binary
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(CGO_FLAGS) go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/main.go
	@echo "Built: $(BUILD_DIR)/$(BINARY)-linux-amd64"

release-all: release-mac release-windows release-linux ## Build for all platforms

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
