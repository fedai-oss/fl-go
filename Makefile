.PHONY: build test clean run-monitor build-web install-web-deps start-web run-all help

# Go build settings
BINARY_NAME=fl-go
MONITOR_BINARY=fl-monitor
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
LDFLAGS=-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

# Directories
BIN_DIR=bin
BUILD_DIR=build
WEB_DIR=web
SCRIPTS_DIR=scripts
EXAMPLES_DIR=examples
CONFIGS_DIR=configs

# Build all binaries
build:
	@echo "Building FL-GO project..."
	@mkdir -p $(BIN_DIR) $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) cmd/aggregator/main.go
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/fl-collaborator cmd/collaborator/main.go
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(MONITOR_BINARY) cmd/monitor/main.go
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/fx cmd/fx/main.go
	@echo "Build completed successfully!"

# Build monitoring server only
build-monitor:
	@echo "Building monitoring server..."
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(MONITOR_BINARY) cmd/monitor/main.go

# Build web UI for production
build-web: install-web-deps
	@echo "Building web UI..."
	cd $(WEB_DIR) && npm run build

# Install web UI dependencies
install-web-deps:
	@echo "Installing web UI dependencies..."
	cd $(WEB_DIR) && npm install

# Start web UI development server
start-web: install-web-deps
	@echo "Starting web UI development server..."
	cd $(WEB_DIR) && npm run dev

# Run monitoring server with sample data
run-monitor: build-monitor
	@echo "Starting monitoring server..."
	./$(BIN_DIR)/$(MONITOR_BINARY) --port 8080 --web-port 3000

# Run both monitoring server and web UI in development
run-all: build-monitor
	@echo "Starting monitoring server..."
	./$(BIN_DIR)/$(MONITOR_BINARY) --port 8080 --web-port 3000 &
	@echo "Starting web UI..."
	cd $(WEB_DIR) && npm run dev

# Run tests using the test script
test:
	@echo "Running test suite..."
	./$(SCRIPTS_DIR)/ci/run_tests.sh all

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	./$(SCRIPTS_DIR)/ci/run_tests.sh unit

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	./$(SCRIPTS_DIR)/ci/run_tests.sh integration

# Run security tests only
test-security:
	@echo "Running security tests..."
	./$(SCRIPTS_DIR)/ci/run_tests.sh security

# Run performance tests only
test-performance:
	@echo "Running performance tests..."
	./$(SCRIPTS_DIR)/ci/run_tests.sh performance

# Test monitoring package specifically
test-monitoring:
	go test ./pkg/monitoring/...

# Format Go code
fmt:
	go fmt ./...

# Format code with goimports
format: fmt
	@command -v goimports >/dev/null 2>&1 || { echo "Installing goimports..."; go install golang.org/x/tools/cmd/goimports@latest; }
	@command -v goimports >/dev/null 2>&1 && goimports -w . || $(HOME)/go/bin/goimports -w .

# Lint Go code
lint:
	@echo "Running linting..."
	./$(SCRIPTS_DIR)/ci/run_tests.sh lint

# Test with race detection
test-race:
	go test -race ./...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	./$(SCRIPTS_DIR)/ci/run_tests.sh coverage

# Validate the FL flows
validate:
	@echo "Validating FL flows..."
	./$(SCRIPTS_DIR)/ci/validate_fl_flows.sh

# Test monitoring system
test-monitoring-system:
	@echo "Testing monitoring system..."
	./$(SCRIPTS_DIR)/ci/test_monitoring.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/ $(BUILD_DIR)/
	rm -f *.log *.out coverage.html fx
	cd $(WEB_DIR) && rm -rf dist/ node_modules/
	rm -rf coverage/ test-results/

# Generate protobuf files (if needed)
proto:
	@echo "Generating protobuf files..."
	protoc --go_out=. --go-grpc_out=. api/federation.proto

# Create sample federation plan
sample-plan:
	@echo "Creating sample federation plan..."
	@if [ -f "$(EXAMPLES_DIR)/plans/basic/sync_plan.yaml" ]; then \
		cp $(EXAMPLES_DIR)/plans/basic/sync_plan.yaml /tmp/fl_plan.yaml; \
		echo "Sample plan created at /tmp/fl_plan.yaml"; \
	else \
		echo "Sample plan not found"; \
	fi

# Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	@command -v goimports >/dev/null 2>&1 || { echo "Installing goimports..."; go install golang.org/x/tools/cmd/goimports@latest; }
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest; }
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	@echo "Development environment setup completed!"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -it $(BINARY_NAME):latest

# Generate documentation
docs:
	@echo "Generating documentation..."
	@echo "Documentation is located in the docs/ directory"

# Show project structure
structure:
	@echo "FL-GO Project Structure:"
	@echo "├── api/                    # Protocol definitions"
	@echo "├── cmd/                    # Application entry points"
	@echo "├── pkg/                    # Core packages"
	@echo "├── web/                    # Web UI"
	@echo "├── examples/               # Examples and samples"
	@echo "│   ├── plans/              # Federation plan examples"
	@echo "│   ├── workspaces/         # Complete workspace examples"
	@echo "│   ├── monitoring/         # Monitoring examples"
	@echo "│   └── scripts/            # Example scripts"
	@echo "├── docs/                   # Documentation"
	@echo "├── scripts/                # Build and utility scripts"
	@echo "├── configs/                # Configuration files"
	@echo "├── deploy/                 # Deployment configurations"
	@echo "├── tests/                  # Test files and test data"
	@echo "└── tools/                  # Development tools"

# Help
help:
	@echo "FL-GO Makefile - Available commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build           - Build all binaries"
	@echo "  build-monitor   - Build monitoring server only"
	@echo "  build-web       - Build web UI for production"
	@echo "  docker-build    - Build Docker image"
	@echo ""
	@echo "Development Commands:"
	@echo "  setup-dev       - Setup development environment"
	@echo "  install-web-deps- Install web UI dependencies"
	@echo "  start-web       - Start web UI development server"
	@echo "  run-monitor     - Run monitoring server with sample data"
	@echo "  run-all         - Run both monitoring server and web UI"
	@echo "  docker-run      - Run Docker container"
	@echo ""
	@echo "Testing Commands:"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run only unit tests"
	@echo "  test-integration- Run only integration tests"
	@echo "  test-security   - Run only security tests"
	@echo "  test-performance- Run only performance tests"
	@echo "  test-monitoring - Test monitoring package"
	@echo "  test-monitoring-system - Test monitoring system"
	@echo "  test-race       - Test with race detection"
	@echo "  test-coverage   - Test with coverage"
	@echo ""
	@echo "Code Quality Commands:"
	@echo "  fmt             - Format Go code"
	@echo "  format          - Format code with goimports"
	@echo "  lint            - Lint Go code"
	@echo "  validate        - Validate the FL flows"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean           - Clean build artifacts"
	@echo "  proto           - Generate protobuf files"
	@echo "  sample-plan     - Create sample federation plan"
	@echo "  docs            - Generate documentation"
	@echo "  structure       - Show project structure"
	@echo "  help            - Show this help"

# Default target
default: help