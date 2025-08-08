# OpenFL-Go

A Go implementation of [OpenFL](https://github.com/securefederatedai/openfl) - An Open Framework for Federated Learning.

OpenFL-Go provides the same CLI-driven workflow as the original OpenFL but with Go handling the orchestration and coordination while delegating ML operations to Python scripts.

## Features

- **OpenFL-Compatible CLI**: Same commands and workflow as the original OpenFL
- **Go Orchestration**: Fast, efficient coordination and communication in Go
- **Python ML Integration**: Seamless delegation of training/evaluation to Python
- **gRPC Communication**: Secure, efficient communication between components
- **Multi-round Federated Learning**: Support for multiple training rounds
- **Extensible Architecture**: Easy to add new aggregation algorithms

## Architecture

```
┌─────────────────┐    gRPC     ┌─────────────────┐
│   Aggregator    │◄────────────┤  Collaborator   │
│   (Go Server)   │             │   (Go Client)   │
└─────────────────┘             └─────────────────┘
         │                               │
         │                               │
         ▼                               ▼
┌─────────────────┐             ┌─────────────────┐
│ Model Averaging │             │ Training Script │
│   (Go Logic)    │             │   (Python ML)   │
└─────────────────┘             └─────────────────┘
```

## Quick Start

### 1. Initialize a Workspace

Create a new federated learning workspace:

```bash
fx plan init --name my_experiment
cd my_experiment
```

This creates:
```
my_experiment/
├── plan.yaml           # FL configuration
├── src/
│   └── taskrunner.py   # Python training script
├── data/               # Local datasets
├── save/               # Model checkpoints
└── logs/               # Training logs
```

### 2. Configure Your Experiment

Edit `plan.yaml` to configure your federated learning experiment:

```yaml
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
      epochs: 5
      batch_size: 32
      lr: 0.001
      data_path: "data"
```

### 3. Start the Aggregator

In one terminal:
```bash
fx aggregator start
```

### 4. Start Collaborators

In separate terminals:
```bash
# Terminal 2
fx collaborator start collaborator1

# Terminal 3  
fx collaborator start collaborator2
```

## CLI Commands

### Plan Management

```bash
# Initialize new workspace
fx plan init --name my_experiment

# Validate plan configuration
fx plan validate

# Show plan contents
fx plan show
```

### Aggregator

```bash
# Start aggregator with default plan.yaml
fx aggregator start

# Start with custom plan
fx aggregator start --plan custom_plan.yaml
```

### Collaborator

```bash
# Start collaborator
fx collaborator start <collaborator_name>

# Start with custom plan
fx collaborator start collab1 --plan custom_plan.yaml
```

## Python Integration

OpenFL-Go delegates all ML operations to Python scripts. The included `taskrunner.py` provides a template that:

1. **Loads models** from binary files
2. **Runs training** with your ML framework (PyTorch, TensorFlow, etc.)  
3. **Saves updated models** back to binary format
4. **Integrates seamlessly** with the Go orchestration layer

### Customizing Training

Modify `src/taskrunner.py` to implement your specific ML training:

```python
def train_model(weights, epochs, batch_size, lr, data_path):
    # Load your training data
    dataset = load_dataset(data_path)
    
    # Initialize your model with weights
    model = create_model(weights)
    
    # Train the model
    for epoch in range(epochs):
        train_epoch(model, dataset, lr, batch_size)
    
    # Return updated weights
    return model.get_weights()
```

## Development

### Using the Makefile

The project includes a comprehensive Makefile for common development tasks:

```bash
# Show all available commands
make help

# Build the binary
make build

# Initialize a new workspace
make workspace-init

# Run a complete demo
make run-demo

# Run demo with large model (1000 parameters, 10 rounds)
make run-demo-large

# Clean build artifacts
make clean

# Clean everything including workspaces
make clean-all
```

### Quick Commands

```bash
# Build and run basic demo
make demo

# Build and run large model demo
make demo-large

# Quick start individual components
make run-aggregator
make run-collaborator COLLAB_ID=collaborator1

# Development setup
make dev-setup    # Install deps, generate proto, lint, format
make dev-build    # Full development build with checks
```

### Testing

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Run tests with coverage
make test-coverage
```

### Docker Support

```bash
# Build Docker image
make docker-build

# Run in Docker
make docker-run
```

### Manual Building

```bash
go build -o fx cmd/fx/main.go
```

### Adding New Features

1. **New CLI commands**: Add to `pkg/cli/`
2. **New aggregation algorithms**: Extend `pkg/aggregator/`
3. **Custom protocols**: Modify `api/federation.proto`

## Comparison with Original OpenFL

| Feature | OpenFL (Python) | OpenFL-Go |
|---------|----------------|-----------|
| CLI Interface | ✅ | ✅ |
| Workspace Init | `fx plan init` | `fx plan init` |
| Aggregator | Python | Go + gRPC |
| Collaborator | Python | Go + Python |
| ML Training | Python | Python (delegated) |
| Communication | gRPC | gRPC |
| Performance | Good | Excellent |

## License

Apache License 2.0 - Same as original OpenFL

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Submit a pull request

## Roadmap

- [ ] Add mTLS security
- [ ] Implement additional aggregation algorithms (FedOpt, FedProx)
- [ ] Add TEE support
- [ ] Web UI for monitoring
- [ ] Integration with popular ML frameworks