#!/bin/bash

# Comprehensive FL Validation Script
# Tests both synchronous and asynchronous federated learning flows
# Validates through daemon processes and log parsing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
WORKSPACE_ROOT="validation_test"
SYNC_WORKSPACE="${WORKSPACE_ROOT}/sync_test"
ASYNC_WORKSPACE="${WORKSPACE_ROOT}/async_test"
AGGREGATOR_WAIT_TIME=5
COLLABORATOR_WAIT_TIME=12
SYNC_ROUNDS=2
ASYNC_ROUNDS=3

# Cleanup function
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up processes...${NC}"
    pkill -f "fx aggregator" 2>/dev/null || true
    pkill -f "fx collaborator" 2>/dev/null || true
    
    # Clean up workspace directories to prevent "File exists" errors
    echo -e "${YELLOW}🧹 Cleaning up workspaces...${NC}"
    rm -rf "${WORKSPACE_ROOT}" 2>/dev/null || true
    rm -f *.log 2>/dev/null || true
    
    sleep 2
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Initial cleanup to prevent conflicts
cleanup

# Helper function to wait for file and parse logs
wait_and_parse_logs() {
    local log_file="$1"
    local expected_pattern="$2"
    local timeout="$3"
    local description="$4"
    
    echo -e "${BLUE}⏳ Waiting for: ${description}${NC}"
    
    local count=0
    while [ $count -lt $timeout ]; do
        if [ -f "$log_file" ] && grep -q "$expected_pattern" "$log_file" 2>/dev/null; then
            echo -e "${GREEN}✅ Found: ${description}${NC}"
            return 0
        fi
        sleep 1
        ((count++))
    done
    
    echo -e "${RED}❌ Timeout waiting for: ${description}${NC}"
    if [ -f "$log_file" ]; then
        echo -e "${YELLOW}📄 Last 10 lines of ${log_file}:${NC}"
        tail -10 "$log_file"
    fi
    return 1
}

# Function to validate sync FL flow
validate_sync_fl() {
    echo -e "${BLUE}🔄 Testing Synchronous Federated Learning${NC}"
    echo "=============================================="
    
    # Setup sync workspace
    cd "$ROOT_DIR/$SYNC_WORKSPACE"
    
    # Start aggregator as daemon
    echo -e "${BLUE}🚀 Starting sync aggregator daemon...${NC}"
    "$FX_BINARY" aggregator start --plan plan.yaml > aggregator.log 2>&1 &
    AGGREGATOR_PID=$!
    echo "Aggregator PID: $AGGREGATOR_PID"
    
    # Wait for aggregator to be ready
    if ! wait_and_parse_logs "aggregator.log" "gRPC server listening" $AGGREGATOR_WAIT_TIME "Aggregator startup"; then
        echo -e "${RED}❌ Sync aggregator failed to start${NC}"
        return 1
    fi
    
    # Start collaborators
    echo -e "${BLUE}🤝 Starting sync collaborators...${NC}"
    "$FX_BINARY" collaborator start collaborator1 --plan plan.yaml > collab1.log 2>&1 &
    COLLAB1_PID=$!
    
    "$FX_BINARY" collaborator start collaborator2 --plan plan.yaml > collab2.log 2>&1 &
    COLLAB2_PID=$!
    
    echo "Collaborator PIDs: $COLLAB1_PID, $COLLAB2_PID"
    
    # Wait for training completion
    if ! wait_and_parse_logs "collab1.log" "Federated learning completed" $COLLABORATOR_WAIT_TIME "Collaborator 1 completion"; then
        echo -e "${RED}❌ Collaborator 1 did not complete${NC}"
        return 1
    fi
    
    if ! wait_and_parse_logs "collab2.log" "Federated learning completed" $COLLABORATOR_WAIT_TIME "Collaborator 2 completion"; then
        echo -e "${RED}❌ Collaborator 2 did not complete${NC}"
        return 1
    fi
    
    # Validate aggregator completed all rounds
    if ! wait_and_parse_logs "aggregator.log" "All.*rounds completed successfully" 10 "Aggregator completion"; then
        echo -e "${RED}❌ Aggregator did not complete all rounds${NC}"
        return 1
    fi
    
    # Parse and validate logs
    echo -e "${BLUE}📊 Parsing sync logs...${NC}"
    
    # Check round completion
    local completed_rounds=$(grep -c "Round.*complete" aggregator.log || echo "0")
    if [ "$completed_rounds" -eq "$SYNC_ROUNDS" ]; then
        echo -e "${GREEN}✅ All $SYNC_ROUNDS sync rounds completed${NC}"
    else
        echo -e "${RED}❌ Expected $SYNC_ROUNDS rounds, found $completed_rounds${NC}"
        return 1
    fi
    
    # Check collaborator participation
    local collab1_rounds=$(grep -c "Round.*completed" collab1.log || echo "0")
    local collab2_rounds=$(grep -c "Round.*completed" collab2.log || echo "0")
    
    if [ "$collab1_rounds" -eq "$SYNC_ROUNDS" ] && [ "$collab2_rounds" -eq "$SYNC_ROUNDS" ]; then
        echo -e "${GREEN}✅ Both collaborators completed all rounds${NC}"
    else
        echo -e "${RED}❌ Collaborator round mismatch: collab1=$collab1_rounds, collab2=$collab2_rounds${NC}"
        return 1
    fi
    
    # Check final model creation
    if [ -f "save/final_model.pt" ]; then
        echo -e "${GREEN}✅ Final sync model created${NC}"
    else
        echo -e "${RED}❌ Final sync model not found${NC}"
        return 1
    fi
    
    echo -e "${GREEN}🎉 Synchronous FL validation PASSED${NC}"
    return 0
}

# Function to validate async FL flow
validate_async_fl() {
    echo -e "${BLUE}🔄 Testing Asynchronous Federated Learning${NC}"
    echo "=============================================="
    
    # Setup async workspace
    cd "$ROOT_DIR/$ASYNC_WORKSPACE"
    
    # Start aggregator as daemon
    echo -e "${BLUE}🚀 Starting async aggregator daemon...${NC}"
    "$FX_BINARY" aggregator start --plan plan.yaml > aggregator.log 2>&1 &
    AGGREGATOR_PID=$!
    echo "Aggregator PID: $AGGREGATOR_PID"
    
    # Wait for aggregator to be ready
    if ! wait_and_parse_logs "aggregator.log" "gRPC server listening" $AGGREGATOR_WAIT_TIME "Async aggregator startup"; then
        echo -e "${RED}❌ Async aggregator failed to start${NC}"
        return 1
    fi
    
    # Start collaborators
    echo -e "${BLUE}🤝 Starting async collaborators...${NC}"
    "$FX_BINARY" collaborator start collaborator1 --plan plan.yaml > collab1.log 2>&1 &
    COLLAB1_PID=$!
    
    "$FX_BINARY" collaborator start collaborator2 --plan plan.yaml > collab2.log 2>&1 &
    COLLAB2_PID=$!
    
    echo "Collaborator PIDs: $COLLAB1_PID, $COLLAB2_PID"
    
    # In async mode, let them run for a while then stop
    echo -e "${BLUE}⏳ Letting async FL run for ${ASYNC_ROUNDS} rounds...${NC}"
    sleep $COLLABORATOR_WAIT_TIME
    
    # Stop collaborators (async mode runs continuously)
    kill $COLLAB1_PID $COLLAB2_PID 2>/dev/null || true
    sleep 2
    
    # Parse and validate logs
    echo -e "${BLUE}📊 Parsing async logs...${NC}"
    
    # Check that async aggregation occurred
    if grep -q "Performing async aggregation" aggregator.log; then
        echo -e "${GREEN}✅ Async aggregation performed${NC}"
    else
        echo -e "${RED}❌ No async aggregation found${NC}"
        return 1
    fi
    
    # Check that collaborators received updated models
    if grep -q "Updated local model with latest from aggregator" collab1.log && 
       grep -q "Updated local model with latest from aggregator" collab2.log; then
        echo -e "${GREEN}✅ Collaborators received model updates${NC}"
    else
        echo -e "${RED}❌ Model updates not properly distributed${NC}"
        return 1
    fi
    
    # Check staleness handling
    if grep -q "staleness" aggregator.log; then
        echo -e "${GREEN}✅ Staleness-aware aggregation working${NC}"
    else
        echo -e "${YELLOW}⚠️  Staleness handling not clearly visible (may be ok)${NC}"
    fi
    
    # Check async model files
    local async_models=$(ls save/async_round_*.pt 2>/dev/null | wc -l || echo "0")
    if [ "$async_models" -gt "0" ]; then
        echo -e "${GREEN}✅ $async_models async model(s) created${NC}"
    else
        echo -e "${RED}❌ No async models found${NC}"
        return 1
    fi
    
    echo -e "${GREEN}🎉 Asynchronous FL validation PASSED${NC}"
    return 0
}

# Main validation function
main() {
    echo -e "${BLUE}🧪 FL-GO Comprehensive Validation${NC}"
    echo "=================================="
    
    # Store the root directory
    ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
    FX_BINARY="$ROOT_DIR/build/fx"
    
    echo "Root directory: $ROOT_DIR"
    echo "FX binary: $FX_BINARY"
    
    # Clean up any existing processes
    cleanup
    
    # Build the project
    echo -e "${BLUE}🔨 Building FL-GO...${NC}"
    cd "$ROOT_DIR"
    make build
    
    # Remove existing test workspace
    rm -rf "$WORKSPACE_ROOT" 2>/dev/null || true
    
    # Create sync test workspace
    echo -e "${BLUE}📁 Creating sync test workspace...${NC}"
    mkdir -p "$WORKSPACE_ROOT"
    cd "$WORKSPACE_ROOT"
    
    "$FX_BINARY" plan init --name sync_test
    cd sync_test
    
    # Update sync plan
    cat > plan.yaml << 'EOF'
mode: "sync"
rounds: 2
collaborators:
  - id: "collaborator1"
    address: "localhost:50052"
  - id: "collaborator2"
    address: "localhost:50053"
aggregator:
  address: "localhost:50051"
initial_model: "save/init_model.pt"
output_model: "save/final_model.pt"
tasks:
  train:
    script: "src/taskrunner.py"
    args:
      epochs: 1
      batch_size: 32
      lr: 0.001
      data_path: "data"
EOF
    
    # Create initial model
    python3 "$ROOT_DIR/scripts/create_initial_model.py" --output models/init_model.pt --size 10
    cp models/init_model.pt save/init_model.pt
    
    cd ..
    
    # Create async test workspace
    echo -e "${BLUE}📁 Creating async test workspace...${NC}"
    "$FX_BINARY" plan init --name async_test
    cd async_test
    
    # Update async plan
    cat > plan.yaml << 'EOF'
mode: "async"
rounds: 20
collaborators:
  - id: "collaborator1"
    address: "localhost:50052"
  - id: "collaborator2"
    address: "localhost:50053"
aggregator:
  address: "localhost:50051"
initial_model: "save/init_model.pt"
output_model: "save/async_final_model.pt"
tasks:
  train:
    script: "src/taskrunner.py"
    args:
      epochs: 1
      batch_size: 32
      lr: 0.001
      data_path: "data"

async_config:
  max_staleness: 60
  min_updates: 1
  aggregation_delay: 3
  staleness_weight: 0.95
EOF
    
    # Create initial model
    python3 "$ROOT_DIR/scripts/create_initial_model.py" --output models/init_model.pt --size 10
    cp models/init_model.pt save/init_model.pt
    
    cd ..
    
    # Run validations
    local sync_result=0
    local async_result=0
    
    # Test sync FL
    echo -e "\n${BLUE}🔄 Running Sync FL Validation...${NC}"
    if validate_sync_fl; then
        echo -e "${GREEN}✅ SYNC FL VALIDATION PASSED${NC}"
    else
        echo -e "${RED}❌ SYNC FL VALIDATION FAILED${NC}"
        sync_result=1
    fi
    
    # Clean up between tests
    cleanup
    sleep 3
    
    # Test async FL
    echo -e "\n${BLUE}🔄 Running Async FL Validation...${NC}"
    if validate_async_fl; then
        echo -e "${GREEN}✅ ASYNC FL VALIDATION PASSED${NC}"
    else
        echo -e "${RED}❌ ASYNC FL VALIDATION FAILED${NC}"
        async_result=1
    fi
    
    # Final results
    echo -e "\n${BLUE}📊 VALIDATION SUMMARY${NC}"
    echo "===================="
    
    if [ $sync_result -eq 0 ]; then
        echo -e "${GREEN}✅ Synchronous FL: PASSED${NC}"
    else
        echo -e "${RED}❌ Synchronous FL: FAILED${NC}"
    fi
    
    if [ $async_result -eq 0 ]; then
        echo -e "${GREEN}✅ Asynchronous FL: PASSED${NC}"
    else
        echo -e "${RED}❌ Asynchronous FL: FAILED${NC}"
    fi
    
    if [ $sync_result -eq 0 ] && [ $async_result -eq 0 ]; then
        echo -e "\n${GREEN}🎉 ALL VALIDATIONS PASSED! FL-GO is working correctly.${NC}"
        return 0
    else
        echo -e "\n${RED}❌ SOME VALIDATIONS FAILED. Check logs above.${NC}"
        return 1
    fi
}

# Run main function
main "$@"
