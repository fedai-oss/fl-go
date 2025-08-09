# Federated Learning Algorithms

FL-GO supports multiple aggregation algorithms that can be easily configured and extended. This document provides an overview of the supported algorithms and how to use them.

## Supported Algorithms

### 1. FedAvg (Federated Averaging)

**Description**: The vanilla federated averaging algorithm as proposed in the original federated learning paper.

**Use Case**: Suitable for homogeneous environments where clients have similar data distributions and compute capabilities.

**Configuration**:
```yaml
algorithm:
  name: "fedavg"
  hyperparameters: {}
```

**How it works**:
- Performs weighted averaging of client updates based on the number of training samples
- Simple and robust baseline algorithm
- No server-side optimization or adaptive mechanisms

---

### 2. FedOpt (Adaptive Server Optimization)

**Description**: Implements adaptive server optimization using Adam-like momentum and variance estimation for better convergence in heterogeneous environments.

**Reference**: [Adaptive Federated Optimization (Reddi et al., 2020)](https://arxiv.org/abs/2003.00295)

**Use Case**: Recommended for environments with varying client capabilities and data heterogeneity where standard FedAvg converges slowly.

**Configuration**:
```yaml
algorithm:
  name: "fedopt"
  hyperparameters:
    server_learning_rate: 1.0    # Server learning rate (default: 1.0)
    beta1: 0.9                   # First moment decay rate (default: 0.9)
    beta2: 0.999                 # Second moment decay rate (default: 0.999)
    epsilon: 1e-7                # Small constant for numerical stability (default: 1e-7)
```

**How it works**:
1. Computes pseudo-gradients as the difference between client average and global model
2. Applies Adam-like optimization on the server side with momentum and adaptive learning rates
3. Performs bias correction for better initial convergence
4. Adapts to the optimization landscape for improved performance

**Hyperparameter Tuning**:
- `server_learning_rate`: Higher values (1.0-3.0) for faster convergence, lower values (0.1-0.8) for stability
- `beta1`: Controls momentum, typically kept at 0.9
- `beta2`: Controls variance estimation, typically kept at 0.999
- `epsilon`: Usually kept at default value

---

### 3. FedProx (Proximal Federated Optimization)

**Description**: Adds a proximal term to handle client heterogeneity and improve convergence stability in non-IID environments.

**Reference**: [Federated Optimization in Heterogeneous Networks (Li et al., 2020)](https://arxiv.org/abs/1812.06127)

**Use Case**: Ideal for highly heterogeneous environments with non-IID data distributions and varying client participation patterns.

**Configuration**:
```yaml
algorithm:
  name: "fedprox"
  hyperparameters:
    mu: 0.01    # Proximal term coefficient (default: 0.01)
```

**How it works**:
1. Performs weighted aggregation considering client learning rates and sample sizes
2. Applies a proximal term that blends the aggregated result with the global model
3. Provides stability by preventing large deviations from the global model
4. Handles client heterogeneity by adjusting weights based on client characteristics

**Hyperparameter Tuning**:
- `mu`: Proximal term coefficient
  - Small values (0.001-0.01): More aggressive updates, faster convergence
  - Medium values (0.01-0.1): Balanced stability and progress
  - Large values (0.1-1.0): Very conservative, high stability

---

## Algorithm Selection Guidelines

| Scenario | Recommended Algorithm | Reason |
|----------|---------------------|---------|
| Homogeneous clients, IID data | FedAvg | Simple and efficient baseline |
| Heterogeneous clients, moderate non-IID | FedOpt | Adaptive optimization handles varying capabilities |
| Highly heterogeneous, extreme non-IID | FedProx | Proximal term provides stability |
| Slow convergence with FedAvg | FedOpt | Server-side optimization accelerates convergence |
| Unstable training | FedProx | Proximal regularization improves stability |
| Async mode with varying latencies | FedOpt or FedProx | Better handling of staleness and heterogeneity |

## Configuration Examples

### Basic FedOpt Setup
```yaml
mode: "sync"
rounds: 10
algorithm:
  name: "fedopt"
  hyperparameters:
    server_learning_rate: 1.0
    beta1: 0.9
    beta2: 0.999
```

### FedProx for Non-IID Data
```yaml
mode: "sync"
rounds: 15
algorithm:
  name: "fedprox"
  hyperparameters:
    mu: 0.05  # Higher mu for more stability
```

### Async FedOpt for Heterogeneous Environment
```yaml
mode: "async"
algorithm:
  name: "fedopt"
  hyperparameters:
    server_learning_rate: 0.8  # Lower for async stability
    beta1: 0.9
    beta2: 0.999
async_config:
  max_staleness: 15
  min_updates: 2
  aggregation_delay: 3
```

## Performance Characteristics

| Algorithm | Convergence Speed | Memory Usage | Computational Overhead | Heterogeneity Handling |
|-----------|------------------|--------------|------------------------|----------------------|
| FedAvg    | Baseline         | Low          | Minimal                | Poor                 |
| FedOpt    | Fast             | Medium       | Low                    | Good                 |
| FedProx   | Moderate         | Low          | Minimal                | Excellent            |

## Adding New Algorithms

The FL-GO framework is designed for easy extension. To add a new algorithm:

1. **Implement the AggregationAlgorithm interface**:
```go
type MyNewAlgorithm struct {
    // Algorithm-specific state
}

func (a *MyNewAlgorithm) Initialize(config AlgorithmConfig) error {
    // Initialize algorithm state
}

func (a *MyNewAlgorithm) Aggregate(updates []ClientUpdate, globalModel []float32) ([]float32, error) {
    // Implement aggregation logic
}

// Implement other interface methods...
```

2. **Register the algorithm** in `CreateAggregationAlgorithm()`:
```go
case "mynew":
    return &MyNewAlgorithm{}, nil
```

3. **Add tests** in `algorithms_test.go`

4. **Update documentation** with algorithm details

## Troubleshooting

### Common Issues

**Slow Convergence**:
- Try FedOpt with higher server learning rate
- Increase the number of local epochs
- Check for data heterogeneity

**Training Instability**:
- Use FedProx with appropriate mu value
- Reduce learning rates
- Increase aggregation frequency in async mode

**Memory Issues**:
- FedOpt uses additional memory for momentum/velocity
- Consider model size when choosing algorithms

**Algorithm Not Found Error**:
- Check spelling in algorithm name
- Ensure algorithm is properly registered
- Verify YAML syntax in configuration

### Monitoring and Debugging

Enable detailed logging by setting the log level to see algorithm-specific information:
```bash
export LOG_LEVEL=debug
fx aggregator start --plan your_plan.yaml
```

The logs will show:
- Algorithm initialization parameters
- Aggregation progress and statistics
- Performance metrics per round
- Error details for troubleshooting

## Future Algorithms

Planned algorithm additions:
- **FedNova**: Normalized averaging for varying local steps
- **SCAFFOLD**: Control variates for client drift correction
- **FedCurv**: Curvature-aware aggregation
- **LAG**: Local adaptivity with global momentum

Contributions for new algorithms are welcome! See the development guide for implementation details.
