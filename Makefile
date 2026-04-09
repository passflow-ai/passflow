.PHONY: all build test lint clean deps

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Binary names
BINDIR=build
CLI_BINARY=$(BINDIR)/passflow
EXECUTOR_BINARY=$(BINDIR)/passflow-executor
CHANNELS_BINARY=$(BINDIR)/passflow-channels
MCP_GW_BINARY=$(BINDIR)/passflow-mcp-gateway

# Build all binaries
all: build

build: build-cli build-executor build-channels build-mcp-gateway

build-cli:
	$(GOBUILD) -o $(CLI_BINARY) ./cmd/passflow-cli

build-executor:
	$(GOBUILD) -o $(EXECUTOR_BINARY) ./cmd/passflow-executor

build-channels:
	$(GOBUILD) -o $(CHANNELS_BINARY) ./cmd/passflow-channels

build-mcp-gateway:
	$(GOBUILD) -o $(MCP_GW_BINARY) ./cmd/passflow-mcp-gateway

# Run tests
test:
	$(GOTEST) -v -race -short ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	$(GOLINT) run ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	rm -rf $(BINDIR)
	rm -f coverage.out coverage.html

# Install CLI locally
install: build-cli
	cp $(CLI_BINARY) $(GOPATH)/bin/passflow

# Development helpers
dev-cli:
	$(GOCMD) run ./cmd/passflow-cli

dev-executor:
	$(GOCMD) run ./cmd/passflow-executor

dev-channels:
	$(GOCMD) run ./cmd/passflow-channels

dev-mcp-gateway:
	$(GOCMD) run ./cmd/passflow-mcp-gateway
