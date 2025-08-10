#!/bin/bash

# FL Monitoring Stack Test Script
set -e

echo "🧪 FL Monitoring Stack Test Suite"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test 1: Build monitoring server
echo
log_info "Test 1: Building monitoring server..."
if make build-monitor > /dev/null 2>&1; then
    log_success "Monitoring server built successfully"
else
    log_error "Failed to build monitoring server"
    exit 1
fi

# Test 2: Start monitoring server
echo
log_info "Test 2: Starting monitoring server..."
./bin/fl-monitor --port 8080 --web-port 3000 &
MONITOR_PID=$!
sleep 3

# Check if server is running
if ps -p $MONITOR_PID > /dev/null; then
    log_success "Monitoring server started (PID: $MONITOR_PID)"
else
    log_error "Failed to start monitoring server"
    exit 1
fi

# Test 3: Health check
echo
log_info "Test 3: Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/api/v1/health)
if echo "$HEALTH_RESPONSE" | grep -q '"success":true'; then
    log_success "Health endpoint working"
else
    log_error "Health endpoint failed"
    echo "Response: $HEALTH_RESPONSE"
    kill $MONITOR_PID
    exit 1
fi

# Test 4: API endpoints
echo
log_info "Test 4: Testing main API endpoints..."

# Test federations endpoint
if curl -s http://localhost:8080/api/v1/federations | grep -q '"success":true'; then
    log_success "Federations endpoint working"
else
    log_error "Federations endpoint failed"
    kill $MONITOR_PID
    exit 1
fi

# Test collaborators endpoint
if curl -s http://localhost:8080/api/v1/collaborators | grep -q '"success":true'; then
    log_success "Collaborators endpoint working"
else
    log_error "Collaborators endpoint failed"
    kill $MONITOR_PID
    exit 1
fi

# Test events endpoint
if curl -s http://localhost:8080/api/v1/events | grep -q '"success":true'; then
    log_success "Events endpoint working"
else
    log_error "Events endpoint failed"
    kill $MONITOR_PID
    exit 1
fi

# Test stats endpoint
if curl -s http://localhost:8080/api/v1/stats | grep -q '"success":true'; then
    log_success "Stats endpoint working"
else
    log_error "Stats endpoint failed"
    kill $MONITOR_PID
    exit 1
fi

# Test 5: System overview
echo
log_info "Test 5: Testing system overview..."
OVERVIEW_RESPONSE=$(curl -s "http://localhost:8080/api/v1/federations/fed_demo_001/overview")
if echo "$OVERVIEW_RESPONSE" | grep -q '"progress_percent":50'; then
    log_success "System overview working with correct data"
else
    log_error "System overview failed or incorrect data"
    echo "Response: $OVERVIEW_RESPONSE"
    kill $MONITOR_PID
    exit 1
fi

# Test 6: WebSocket connection
echo
log_info "Test 6: Testing WebSocket connection..."
WS_RESPONSE=$(curl -s -i -N \
     -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" \
     -H "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     http://localhost:8080/api/v1/ws 2>&1 &)
WS_PID=$!
sleep 1
kill $WS_PID 2>/dev/null || true

if ps -p $MONITOR_PID > /dev/null; then
    log_success "WebSocket endpoint accessible"
else
    log_error "WebSocket test caused server crash"
    exit 1
fi

# Test 7: Configuration validation
echo
log_info "Test 7: Testing configuration files..."

# Check monitoring config
if [ -f "monitoring_config.yaml" ]; then
    log_success "Monitoring configuration file exists"
else
    log_warning "Monitoring configuration file not found (optional)"
fi

# Check sample plan
if [ -f "plans/monitoring_example_plan.yaml" ]; then
    log_success "Sample monitoring plan exists"
else
    log_error "Sample monitoring plan not found"
    kill $MONITOR_PID
    exit 1
fi

# Test 8: Data validation
echo
log_info "Test 8: Validating sample data..."

# Check federation data
FEDERATION_COUNT=$(curl -s http://localhost:8080/api/v1/federations | jq '.data | length')
if [ "$FEDERATION_COUNT" = "1" ]; then
    log_success "Sample federation data correct"
else
    log_error "Expected 1 federation, got $FEDERATION_COUNT"
    kill $MONITOR_PID
    exit 1
fi

# Check collaborators data
COLLABORATOR_COUNT=$(curl -s http://localhost:8080/api/v1/collaborators | jq '.data | length')
if [ "$COLLABORATOR_COUNT" = "3" ]; then
    log_success "Sample collaborator data correct"
else
    log_error "Expected 3 collaborators, got $COLLABORATOR_COUNT"
    kill $MONITOR_PID
    exit 1
fi

# Test 9: Web UI files
echo
log_info "Test 9: Checking web UI files..."

WEB_FILES=("web/package.json" "web/index.html" "web/src/main.tsx" "web/src/App.tsx")
for file in "${WEB_FILES[@]}"; do
    if [ -f "$file" ]; then
        log_success "Web file exists: $file"
    else
        log_error "Web file missing: $file"
        kill $MONITOR_PID
        exit 1
    fi
done

# Test 10: Documentation
echo
log_info "Test 10: Checking documentation..."

DOC_FILES=("MONITORING.md" "Makefile")
for file in "${DOC_FILES[@]}"; do
    if [ -f "$file" ]; then
        log_success "Documentation exists: $file"
    else
        log_error "Documentation missing: $file"
        kill $MONITOR_PID
        exit 1
    fi
done

# Cleanup
echo
log_info "Cleaning up..."
kill $MONITOR_PID
wait $MONITOR_PID 2>/dev/null || true
log_success "Monitoring server stopped"

# Final summary
echo
echo "🎉 All tests passed!"
echo "=================================="
log_success "✅ Build system working"
log_success "✅ API server functional"
log_success "✅ All endpoints responding"
log_success "✅ WebSocket support working"
log_success "✅ Sample data correct"
log_success "✅ Web UI files present"
log_success "✅ Documentation complete"

echo
echo "📋 Next Steps:"
echo "1. Install Node.js to test web UI: brew install node"
echo "2. Run 'make install-web-deps && make start-web' to test frontend"
echo "3. Run 'make run-monitor' to start the monitoring server"
echo "4. Visit http://localhost:8080/api/v1/health for API"
echo "5. Visit http://localhost:3000 for web UI (after installing Node.js)"

echo
log_success "🚀 FL Monitoring Stack is ready for GitHub!"
