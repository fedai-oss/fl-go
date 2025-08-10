.PHONY: build test clean run-monitor build-web install-web-deps start-web run-all

# Go build settings
BINARY_NAME=fl-go
MONITOR_BINARY=fl-monitor

# Build all binaries
build:
	go build -o bin/$(BINARY_NAME) cmd/aggregator/main.go
	go build -o bin/fl-collaborator cmd/collaborator/main.go
	go build -o bin/$(MONITOR_BINARY) cmd/monitor/main.go

# Build monitoring server only
build-monitor:
	go build -o bin/$(MONITOR_BINARY) cmd/monitor/main.go

# Run monitoring server with sample data
run-monitor: build-monitor
	./bin/$(MONITOR_BINARY) --port 8080 --web-port 3000

# Install web UI dependencies
install-web-deps:
	cd web && npm install

# Build web UI for production
build-web: install-web-deps
	cd web && npm run build

# Start web UI development server
start-web: install-web-deps
	cd web && npm run dev

# Run both monitoring server and web UI in development
run-all: build-monitor
	@echo "Starting monitoring server..."
	./bin/$(MONITOR_BINARY) --port 8080 --web-port 3000 &
	@echo "Starting web UI..."
	cd web && npm run dev

# Run tests
test:
	go test ./...

# Test monitoring package specifically
test-monitoring:
	go test ./pkg/monitoring/...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f *.log
	cd web && rm -rf dist/ node_modules/

# Format Go code
fmt:
	go fmt ./...

# Lint Go code
lint:
	golangci-lint run

# Generate protobuf files (if needed)
proto:
	protoc --go_out=. --go-grpc_out=. api/federation.proto

# Create sample federation plan with monitoring
sample-plan:
	cp plans/monitoring_example_plan.yaml /tmp/fl_plan.yaml
	@echo "Sample plan created at /tmp/fl_plan.yaml"

# Help
help:
	@echo "Available commands:"
	@echo "  build           - Build all binaries"
	@echo "  build-monitor   - Build monitoring server only"
	@echo "  run-monitor     - Run monitoring server with sample data"
	@echo "  install-web-deps- Install web UI dependencies"
	@echo "  build-web       - Build web UI for production"
	@echo "  start-web       - Start web UI development server"
	@echo "  run-all         - Run both monitoring server and web UI"
	@echo "  test            - Run all tests"
	@echo "  test-monitoring - Test monitoring package"
	@echo "  clean           - Clean build artifacts"
	@echo "  fmt             - Format Go code"
	@echo "  lint            - Lint Go code"
	@echo "  sample-plan     - Create sample federation plan"
	@echo "  help            - Show this help"

# Default target
default: help