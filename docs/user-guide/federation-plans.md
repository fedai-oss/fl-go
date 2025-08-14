# Federation Plans

Federation plans are YAML configuration files that define how federated learning sessions are conducted. They specify the participants, algorithms, models, and other parameters.

## Plan Structure

A federation plan consists of several key sections:

### Basic Structure

```yaml
federation:
  name: "string"           # Unique federation name
  description: "string"    # Human-readable description
  
  aggregator:              # Aggregator configuration
    host: "string"
    port: integer
    
  collaborators:           # List of collaborators
    - name: "string"
      host: "string"
      port: integer
      
  algorithm:               # Federated learning algorithm
    type: "string"
    parameters: {}
    
  model:                   # Model configuration
    type: "string"
    parameters: {}
    
  monitoring:              # Monitoring configuration (optional)
    enabled: boolean
    config: {}
```

## Algorithm Types

### FedAvg (Federated Averaging)
```yaml
algorithm:
  type: "fedavg"
  rounds: 10
  epochs_per_round: 5
  learning_rate: 0.01
```

### FedOpt (Federated Optimization)
```yaml
algorithm:
  type: "fedopt"
  rounds: 10
  epochs_per_round: 5
  learning_rate: 0.01
  beta1: 0.9
  beta2: 0.999
```

### FedProx (Federated Proximal)
```yaml
algorithm:
  type: "fedprox"
  rounds: 10
  epochs_per_round: 5
  learning_rate: 0.01
  mu: 0.001
```

## Model Types

### Simple Neural Network
```yaml
model:
  type: "simple_nn"
  input_size: 784
  hidden_size: 128
  output_size: 10
  activation: "relu"
```

### Convolutional Neural Network
```yaml
model:
  type: "cnn"
  input_channels: 1
  input_height: 28
  input_width: 28
  num_classes: 10
  layers: [32, 64, 128]
```

## Monitoring Configuration

```yaml
monitoring:
  enabled: true
  config:
    api_port: 8080
    webui_port: 3000
    metrics_retention: "24h"
    collection_interval: "5s"
```

## Security Configuration

### mTLS (Mutual TLS)
```yaml
security:
  mtls:
    enabled: true
    ca_cert: "path/to/ca.crt"
    cert: "path/to/cert.crt"
    key: "path/to/key.key"
```

## Example Plans

See the [examples directory](../../examples/plans/) for complete working examples:

- [Basic Plans](../../examples/plans/basic/) - Simple federations
- [Advanced Plans](../../examples/plans/advanced/) - Complex algorithms
- [Monitoring Plans](../../examples/plans/monitoring/) - With monitoring enabled

## Validation

Validate your plan before running:

```bash
fx plan validate my_federation.yaml
```

## Best Practices

1. **Use descriptive names** for federations and collaborators
2. **Set appropriate timeouts** for network operations
3. **Enable monitoring** for production deployments
4. **Use security features** for sensitive data
5. **Test with small datasets** before scaling up

## Troubleshooting

Common issues and solutions:

- **Connection refused**: Check host/port configuration
- **Invalid plan**: Use `fx plan validate` to check syntax
- **Performance issues**: Adjust algorithm parameters
- **Security errors**: Verify certificate paths and permissions
