# Asynchronous Federated Learning in FL-GO

This document describes the implementation of asynchronous federated learning (Async FL) in FL-GO, based on the [Papaya paper](https://arxiv.org/abs/2111.04877).

## Overview

FL-GO now supports both synchronous and asynchronous federated learning modes:

- **Synchronous FL (Sync)**: Traditional round-based approach where all collaborators must complete training before aggregation
- **Asynchronous FL (Async)**: Non-round-based approach where collaborators can submit updates as they complete, without waiting for others

## Key Features

### 1. Mode Switching
The system can switch between sync and async modes using the `mode` field in the plan configuration:

```yaml
mode: "sync"   # or "async"
```

### 2. Asynchronous Aggregation
- **Immediate Processing**: Updates are processed as soon as they arrive
- **Staleness-Aware**: Updates are weighted based on their staleness
- **Configurable Bounds**: Maximum staleness and minimum updates can be configured

### 3. Staleness Handling
Based on the Papaya paper, the system implements staleness-aware aggregation:
- Updates are weighted using exponential decay: `weight = staleness_weight^staleness`
- Stale updates beyond `max_staleness` are dropped
- This prevents stale updates from negatively impacting convergence

## Configuration

### Async Configuration Parameters

```yaml
async_config:
  max_staleness: 300      # Maximum staleness in seconds (5 minutes)
  min_updates: 1          # Minimum updates before aggregation
  aggregation_delay: 10   # Delay in seconds before aggregating
  staleness_weight: 0.95  # Weight decay factor for stale updates
```

### Parameter Descriptions

- **`max_staleness`**: Maximum age (in seconds) of updates that will be accepted
- **`min_updates`**: Minimum number of updates required before aggregation occurs
- **`aggregation_delay`**: Time delay between aggregation attempts
- **`staleness_weight`**: Exponential decay factor for weighting stale updates

## Usage Examples

### Synchronous Mode (Default)

```yaml
# plan.yaml
mode: "sync"
rounds: 3
collaborators:
  - id: "collab1"
    address: "localhost:50052"
  - id: "collab2"
    address: "localhost:50053"
aggregator:
  address: "localhost:50051"
initial_model: "models/init_model.pt"
output_model: "models/aggregated_model.pt"
```

### Asynchronous Mode

```yaml
# async_plan.yaml
mode: "async"
rounds: 100  # Maximum rounds (async continues until stopped)
collaborators:
  - id: "collab1"
    address: "localhost:50052"
  - id: "collab2"
    address: "localhost:50053"
  - id: "collab3"
    address: "localhost:50054"
aggregator:
  address: "localhost:50051"
initial_model: "models/init_model.pt"
output_model: "models/async_aggregated_model.pt"

async_config:
  max_staleness: 600      # 10 minutes
  min_updates: 1          # Aggregate immediately
  aggregation_delay: 5    # 5 seconds
  staleness_weight: 0.98  # Gentle decay
```

## Running the System

### Start Aggregator

```bash
# Sync mode
fx aggregator start

# Async mode
fx aggregator start --plan async_plan.yaml
```

### Start Collaborators

```bash
# Sync mode
fx collaborator start collab1
fx collaborator start collab2

# Async mode
fx collaborator start collab1 --plan async_plan.yaml
fx collaborator start collab2 --plan async_plan.yaml
fx collaborator start collab3 --plan async_plan.yaml
```

## Implementation Details

### Aggregator Architecture

The system uses a factory pattern to create the appropriate aggregator:

```go
func NewAggregator(plan *federation.FLPlan) Aggregator {
    switch plan.Mode {
    case federation.ModeAsync:
        return NewAsyncFedAvgAggregator(plan)
    default:
        return NewFedAvgAggregator(plan)
    }
}
```

### Asynchronous Aggregation Algorithm

1. **Update Collection**: Updates are collected with timestamps
2. **Staleness Calculation**: Staleness is calculated as `current_time - update_time`
3. **Update Filtering**: Updates beyond `max_staleness` are dropped
4. **Weighted Aggregation**: Updates are weighted using exponential decay
5. **Model Update**: Global model is updated immediately

### Collaborator Behavior

- **Sync Mode**: Collaborators wait for all others before proceeding to next round
- **Async Mode**: Collaborators submit updates immediately and continue training

## Performance Benefits

Based on the Papaya paper, asynchronous FL provides:

1. **Faster Convergence**: 5x faster training in high concurrency settings
2. **Reduced Communication**: 8x less communication overhead
3. **Better Scalability**: Handles stragglers and system variability
4. **Improved Resource Utilization**: No idle waiting for slow collaborators

## Monitoring and Debugging

### Log Messages

The system provides detailed logging for both modes:

```
# Sync mode
Starting SYNC aggregator on localhost:50051
Starting round 1/3
Received updates from all 2 collaborators

# Async mode
Starting ASYNC aggregator on localhost:50051
Async config: max_staleness=600, min_updates=1, delay=5s
Performing async aggregation with 2 updates
Async round 1 complete, model saved to save/async_round_1_model.pt
```

### Model Files

- **Sync Mode**: `save/round_N_model.pt`
- **Async Mode**: `save/async_round_N_model.pt`

## Best Practices

### For Synchronous FL
- Use when all collaborators have similar compute resources
- Suitable for small to medium-scale deployments
- Predictable training time

### For Asynchronous FL
- Use when collaborators have varying compute resources
- Suitable for large-scale deployments (millions of devices)
- Better handling of stragglers and system heterogeneity
- Recommended for production environments

### Configuration Guidelines

1. **`max_staleness`**: Set based on expected training time variations
2. **`staleness_weight`**: Use 0.95-0.99 for gentle decay
3. **`aggregation_delay`**: Balance between responsiveness and overhead
4. **`min_updates`**: Use 1 for immediate aggregation, higher for batch processing

## References

- [Papaya: Practical, Private, and Scalable Federated Learning](https://arxiv.org/abs/2111.04877)
- [Federated Learning: Challenges, Methods, and Future Directions](https://arxiv.org/abs/1908.07873)
