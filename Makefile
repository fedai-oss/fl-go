# FL-Go Makefile
# A Go implementation of OpenFL - An Open Framework for Federated Learning

.PHONY: help build clean test run-aggregator run-collaborator run-demo workspace-init workspace-clean deps proto lint format docker-build docker-run validate validate-sync validate-async

# Variables
BINARY_NAME=fx
BUILD_DIR=build
WORKSPACE_NAME=fl_workspace
DOCKER_IMAGE=fl-go
DOCKER_TAG=latest

# Default target
help: ## Show this help message
	@echo "FL-Go - A Go implementation of OpenFL"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build the fx binary"
	@echo "  make validate                 # Run comprehensive FL validation"
	@echo "  make workspace-init           # Initialize a new FL workspace"
	@echo "  make run-demo                 # Run a complete FL demo"
	@echo "  make clean                    # Clean build artifacts"

# Build targets
build: ## Build the fx binary
	@echo "🔨 Building FL-Go binary..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/fx/main.go
	@echo "✅ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-race: ## Build with race detection
	@echo "🔨 Building with race detection..."
	@mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME) cmd/fx/main.go
	@echo "✅ Binary built with race detection"

build-debug: ## Build with debug symbols
	@echo "🔨 Building with debug symbols..."
	@mkdir -p $(BUILD_DIR)
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/fx/main.go
	@echo "✅ Binary built with debug symbols"

# Clean targets
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✅ Cleaned build artifacts"

workspace-clean: ## Clean workspace directories
	@echo "🧹 Cleaning workspace directories..."
	@rm -rf $(WORKSPACE_NAME) large_model_test test_workspace
	@echo "✅ Cleaned workspace directories"

clean-all: clean workspace-clean ## Clean everything
	@echo "✅ Cleaned everything"

# Development targets
deps: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

proto: ## Generate protobuf code
	@echo "🔧 Generating protobuf code..."
	@if ! command -v protoc > /dev/null; then \
		echo "❌ protoc not found. Please install Protocol Buffers compiler."; \
		exit 1; \
	fi
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/federation.proto
	@echo "✅ Protobuf code generated"

lint: ## Run linter
	@echo "🔍 Running linter..."
	@if ! command -v staticcheck > /dev/null; then \
		echo "❌ staticcheck not found. Installing..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	@echo "🔍 Running go vet..."
	go vet ./...
	@echo "🔍 Running staticcheck..."
	staticcheck ./...
	@echo "✅ Linting completed"

format: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Workspace management
workspace-init: build ## Initialize a new FL workspace
	@echo "🚀 Initializing new FL workspace..."
	./$(BUILD_DIR)/$(BINARY_NAME) plan init --name $(WORKSPACE_NAME)
	@echo "✅ Workspace initialized: $(WORKSPACE_NAME)"

workspace-init-large: build ## Initialize workspace with large model
	@echo "🚀 Initializing workspace with large model..."
	./$(BUILD_DIR)/$(BINARY_NAME) plan init --name large_model_test
	@echo "✅ Large model workspace initialized: large_model_test"

# Run targets
run-aggregator: build ## Run aggregator (requires workspace)
	@echo "🎯 Starting aggregator..."
	@if [ ! -f "plan.yaml" ]; then \
		echo "❌ No plan.yaml found. Run 'make workspace-init' first."; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY_NAME) aggregator start

run-collaborator: build ## Run collaborator (requires workspace and collaborator name)
	@echo "🤝 Starting collaborator..."
	@if [ ! -f "plan.yaml" ]; then \
		echo "❌ No plan.yaml found. Run 'make workspace-init' first."; \
		exit 1; \
	fi
	@if [ -z "$(COLLAB_ID)" ]; then \
		echo "❌ Please specify COLLAB_ID. Example: make run-collaborator COLLAB_ID=collaborator1"; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY_NAME) collaborator start $(COLLAB_ID)

# Demo targets
run-demo: build workspace-init ## Run a complete FL demo
	@echo "🎬 Starting FL-Go demo..."
	@cd $(WORKSPACE_NAME) && \
	echo "🚀 Starting aggregator..." && \
	../$(BUILD_DIR)/$(BINARY_NAME) aggregator start > aggregator.log 2>&1 & \
	AGG_PID=$$! && \
	sleep 3 && \
	echo "🤝 Starting collaborator 1..." && \
	../$(BUILD_DIR)/$(BINARY_NAME) collaborator start collaborator1 > collab1.log 2>&1 & \
	COLLAB1_PID=$$! && \
	sleep 2 && \
	echo "🤝 Starting collaborator 2..." && \
	../$(BUILD_DIR)/$(BINARY_NAME) collaborator start collaborator2 > collab2.log 2>&1 & \
	COLLAB2_PID=$$! && \
	echo "⏳ Waiting for training to complete..." && \
	wait $$COLLAB1_PID $$COLLAB2_PID && \
	kill $$AGG_PID 2>/dev/null || true && \
	echo "✅ Demo completed! Check logs in $(WORKSPACE_NAME)/"

run-demo-large: build workspace-init-large ## Run demo with large model
	@echo "🎬 Starting FL-Go large model demo..."
	@cd large_model_test && \
	echo "🚀 Starting aggregator..." && \
	../$(BUILD_DIR)/$(BINARY_NAME) aggregator start > aggregator.log 2>&1 & \
	AGG_PID=$$! && \
	sleep 3 && \
	echo "🤝 Starting collaborator 1..." && \
	../$(BUILD_DIR)/$(BINARY_NAME) collaborator start collaborator1 > collab1.log 2>&1 & \
	COLLAB1_PID=$$! && \
	sleep 2 && \
	echo "🤝 Starting collaborator 2..." && \
	../$(BUILD_DIR)/$(BINARY_NAME) collaborator start collaborator2 > collab2.log 2>&1 & \
	COLLAB2_PID=$$! && \
	echo "⏳ Waiting for training to complete..." && \
	wait $$COLLAB1_PID $$COLLAB2_PID && \
	kill $$AGG_PID 2>/dev/null || true && \
	echo "✅ Large model demo completed! Check logs in large_model_test/"

# Test targets
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests completed"

test-race: ## Run tests with race detection
	@echo "🧪 Running tests with race detection..."
	go test -race -v ./...
	@echo "✅ Tests with race detection completed"

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Docker targets
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: docker-build ## Run in Docker
	@echo "🐳 Running in Docker..."
	docker run -it --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start FL-Go with Docker Compose
	@echo "🐳 Starting FL-Go with Docker Compose..."
	docker-compose up --build

docker-compose-down: ## Stop Docker Compose services
	@echo "🐳 Stopping Docker Compose services..."
	docker-compose down

# =============================================================================
# Validation Commands
# =============================================================================

validate: build ## Run comprehensive FL validation (sync + async)
	@echo "🧪 Running comprehensive FL validation..."
	@./scripts/validate_fl_flows.sh

validate-sync: build ## Run synchronous FL validation only
	@echo "🔄 Running sync FL validation..."
	@./scripts/validate_fl_flows.sh sync

validate-async: build ## Run asynchronous FL validation only  
	@echo "🔄 Running async FL validation..."
	@./scripts/validate_fl_flows.sh async

docker-compose-logs: ## Show Docker Compose logs
	@echo "🐳 Showing Docker Compose logs..."
	docker-compose logs -f

docker-shell: ## Start development shell in Docker
	@echo "🐳 Starting development shell in Docker..."
	docker-compose run --rm dev-shell

# Utility targets
install: build ## Install binary to system
	@echo "📦 Installing binary..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Binary installed to /usr/local/bin/$(BINARY_NAME)"

uninstall: ## Uninstall binary
	@echo "🗑️ Uninstalling binary..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Binary uninstalled"

version: ## Show version information
	@echo "FL-Go version information:"
	@echo "  Go version: $(shell go version)"
	@echo "  Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Build time: $(shell date)"

# Development workflow
dev-setup: deps proto lint format ## Setup development environment
	@echo "✅ Development environment setup complete"

dev-build: dev-setup build ## Build with all development checks
	@echo "✅ Development build complete"

# Quick commands
demo: run-demo ## Alias for run-demo
demo-large: run-demo-large ## Alias for run-demo-large
collab1: build ## Quick start collaborator1
	@cd $(WORKSPACE_NAME) && ../$(BUILD_DIR)/$(BINARY_NAME) collaborator start collaborator1
collab2: build ## Quick start collaborator2
	@cd $(WORKSPACE_NAME) && ../$(BUILD_DIR)/$(BINARY_NAME) collaborator start collaborator2
