# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build variables
BINARY_NAME = linkding-mcp
MAIN_PATH = ./cmd/linkding-mcp
LDFLAGS = -ldflags="-w -s -X 'github.com/chickenzord/linkding-mcp/internal/version.Version=$(VERSION)' -X 'github.com/chickenzord/linkding-mcp/internal/version.GitCommit=$(GIT_COMMIT)' -X 'github.com/chickenzord/linkding-mcp/internal/version.BuildDate=$(BUILD_DATE)'"

# Docker variables
IMAGE_NAME = ghcr.io/chickenzord/linkding-mcp
IMAGE_TAG ?= $(VERSION)

.PHONY: help build build-linux build-windows build-darwin clean test docker-build docker-push version

help: ## Show help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build binary for current platform
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

build-linux: ## Build binary for Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-windows: ## Build binary for Windows  
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-darwin: ## Build binary for macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-all: build-linux build-windows build-darwin ## Build binaries for all platforms

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)*

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

docker-build: ## Build Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):latest .

docker-push: ## Push Docker image
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_NAME):latest

version: build ## Show version information
	./$(BINARY_NAME) version

install: build ## Install binary to $GOPATH/bin
	go install $(LDFLAGS) $(MAIN_PATH)

dev-run-stdio: ## Run in development stdio mode
	LINKDING_URL=http://localhost:9090 LINKDING_API_TOKEN=test-token go run $(MAIN_PATH) stdio

dev-run-http: ## Run in development HTTP mode  
	LINKDING_URL=http://localhost:9090 LINKDING_API_TOKEN=test-token go run $(MAIN_PATH) http

fmt: ## Format code
	go fmt ./...

lint: ## Lint code (requires golangci-lint)
	golangci-lint run

mod-tidy: ## Tidy go modules
	go mod tidy

release-build: clean build-all docker-build ## Build release artifacts