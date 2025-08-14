# FL-GO Validation Guide

This document describes how to validate FL-GO implementations and test both synchronous and asynchronous federated learning flows.

## Overview

FL-GO includes comprehensive validation scripts that test both sync and async FL modes using daemon processes, automated collaborator execution, and log parsing to ensure complete functionality.

## Quick Validation

### Run All Validations

```bash
# Run comprehensive validation for both sync and async FL
./scripts/validate_fl_flows.sh
```

This single command will:
- Build the FL-GO binary
- Create test workspaces for both sync and async modes
- Start aggregators as daemon processes
- Run collaborators to completion
- Parse logs to validate functionality
- Report comprehensive results

### Expected Output

```
🧪 FL-GO Comprehensive Validation
==================================
🔨 Building FL-GO...
📁 Creating sync test workspace...
📁 Creating async test workspace...

🔄 Running Sync FL Validation...
🔄 Testing Synchronous Federated Learning
==============================================
✅ Found: Aggregator startup
✅ Found: Collaborator 1 completion
✅ Found: Collaborator 2 completion
✅ All 3 sync rounds completed
✅ Both collaborators completed all rounds
✅ Final sync model created
🎉 Synchronous FL validation PASSED

🔄 Running Async FL Validation...
🔄 Testing Asynchronous Federated Learning
==============================================
✅ Found: Async aggregator startup
✅ Async aggregation performed
✅ Collaborators received model updates
✅ Staleness-aware aggregation working
✅ 6 async model(s) created
🎉 Asynchronous FL validation PASSED

📊 VALIDATION SUMMARY
====================
✅ Synchronous FL: PASSED
✅ Asynchronous FL: PASSED

🎉 ALL VALIDATIONS PASSED! FL-GO is working correctly.
```

## Manual Validation Steps

If you prefer to run validation steps manually or debug specific issues:

### 1. Build FL-GO

```bash
make build
```

### 2. Create Test Workspace

```bash
# For sync FL testing
./build/fx plan init --name sync_test
cd sync_test

# Edit plan.yaml to set mode: "sync"
# Ensure initial model exists
```

### 3. Start Aggregator (Daemon)

```bash
# Start aggregator in background
./build/fx aggregator start --plan plan.yaml > aggregator.log 2>&1 &
AGGREGATOR_PID=$!

# Check if started successfully
tail -f aggregator.log
# Look for: "gRPC server listening on localhost:50051"
```

### 4. Start Collaborators

```bash
# Start collaborators (they will run to completion)
./build/fx collaborator start collaborator1 --plan plan.yaml > collab1.log 2>&1 &
./build/fx collaborator start collaborator2 --plan plan.yaml > collab2.log 2>&1 &

# Wait for completion and check logs
tail -f collab1.log
tail -f collab2.log
```

### 5. Validate Results

```bash
# Check aggregator completed all rounds
grep "All.*rounds completed successfully" aggregator.log

# Check collaborators completed
grep "Federated learning completed" collab1.log collab2.log

# Check generated models
ls -la save/
```

### 6. Cleanup

```bash
# Kill background processes
kill $AGGREGATOR_PID
pkill -f "fx aggregator"
pkill -f "fx collaborator"
```

## Validation Scenarios

### Synchronous FL Validation

The sync validation tests:

- **Round-based execution**: All collaborators must complete training before aggregation
- **Synchronized waiting**: Collaborators wait for each other
- **Sequential aggregation**: Models are aggregated only after all updates received
- **Model persistence**: Intermediate and final models are saved correctly

**Expected behavior:**
- 3 rounds of training (configurable)
- Both collaborators complete exactly 3 rounds
- Final model saved to `save/final_model.pt`
- Intermediate models: `save/round_N_model.pt`

### Asynchronous FL Validation

The async validation tests:

- **Continuous execution**: Collaborators train without waiting
- **Immediate aggregation**: Updates are aggregated as soon as received
- **Model distribution**: Latest models are sent to collaborators
- **Staleness handling**: Stale updates are handled according to configuration

**Expected behavior:**
- Multiple async aggregation rounds (typically 5-10 in 15 seconds)
- Collaborators receive updated models during training
- Async models saved as `save/async_round_N_model.pt`
- Staleness-aware aggregation logs visible

## Configuration Options

### Sync FL Configuration

```yaml
mode: "sync"
rounds: 3
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
```

### Async FL Configuration

```yaml
mode: "async"
rounds: 20  # Max rounds for async mode
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

async_config:
  max_staleness: 60        # 60 seconds max staleness
  min_updates: 1           # Aggregate immediately
  aggregation_delay: 3     # 3 seconds between aggregations
  staleness_weight: 0.95   # 95% weight decay per second
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

```bash
# Kill existing processes
pkill -f "fx aggregator"
pkill -f "fx collaborator"

# Or use different ports in plan.yaml
```

#### 2. Missing Initial Model

```bash
# Create initial model
python3 scripts/create_initial_model.py --output save/init_model.pt --size 100
```

#### 3. Python Script Errors

```bash
# Check if Python dependencies are available
python3 -c "import torch; print('PyTorch available')"

# Check training script
python3 src/taskrunner.py --help
```

#### 4. Log Analysis

```bash
# Check aggregator logs for errors
grep -i error aggregator.log

# Check collaborator logs
grep -i error collab*.log

# Check last 20 lines of all logs
tail -20 *.log
```

### Debug Mode

For detailed debugging, run components in foreground:

```bash
# Run aggregator in foreground (Ctrl+C to stop)
./build/fx aggregator start --plan plan.yaml

# In another terminal, run collaborator in foreground
./build/fx collaborator start collaborator1 --plan plan.yaml
```

## Continuous Integration

### Adding to CI/CD

```bash
# Add to your CI pipeline
name: FL Validation
run: |
  cd fl-go
  make build
  ./scripts/validate_fl_flows.sh
  if [ $? -eq 0 ]; then
    echo "FL validation passed"
  else
    echo "FL validation failed"
    exit 1
  fi
```

### Regression Testing

Use the validation script for regression testing:

```bash
# Before making changes
./scripts/validate_fl_flows.sh > baseline_results.txt

# After making changes
./scripts/validate_fl_flows.sh > new_results.txt

# Compare results
diff baseline_results.txt new_results.txt
```

## Performance Validation

### Measuring FL Performance

The validation script captures timing information:

```bash
# Sync FL typically takes 10-15 seconds for 3 rounds
# Async FL typically completes 5-10 aggregations in 15 seconds

# Extract timing from logs
grep "Round.*complete" validation_test/*/aggregator.log
grep "Async round.*complete" validation_test/*/aggregator.log
```

### Scalability Testing

To test with more collaborators:

1. Update `plan.yaml` with additional collaborators
2. Modify validation script to start more collaborator processes
3. Monitor resource usage and performance

## Custom Validation

### Creating Custom Tests

```bash
# Copy and modify the validation script
cp scripts/validate_fl_flows.sh scripts/my_custom_validation.sh

# Modify configuration parameters
# Add custom validation logic
# Run your custom validation
```

### Integration with Testing Frameworks

The validation script returns proper exit codes:
- `0`: All validations passed
- `1`: Some validations failed

This makes it easy to integrate with testing frameworks like pytest, jest, or custom CI systems.

## References

- [FL-GO Architecture](../README.md#architecture)
- [Async FL Documentation](ASYNC_FL.md)
- [Synchronous vs Asynchronous FL Comparison](../README.md#fl-modes-comparison)
