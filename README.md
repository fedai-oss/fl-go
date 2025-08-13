# FL-Go

<!-- Badges -->
<div align="center">

[![CI Status](https://github.com/fedai-oss/fl-go/actions/workflows/ci.yml/badge.svg)](https://github.com/fedai-oss/fl-go/actions/workflows/ci.yml)
[![Security](https://github.com/fedai-oss/fl-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/fedai-oss/fl-go/actions/workflows/codeql.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/fedai-oss/fl-go?color=blue&logo=go)](https://golang.org/dl/)
[![License](https://img.shields.io/github/license/fedai-oss/fl-go?color=green)](LICENSE)
[![Release](https://img.shields.io/github/v/release/fedai-oss/fl-go?include_prereleases&sort=semver)](https://github.com/fedai-oss/fl-go/releases)

[![Go Report Card](https://goreportcard.com/badge/github.com/fedai-oss/fl-go)](https://goreportcard.com/report/github.com/fedai-oss/fl-go)
[![Contributors](https://img.shields.io/github/contributors/fedai-oss/fl-go?color=orange)](https://github.com/fedai-oss/fl-go/graphs/contributors)
[![Stars](https://img.shields.io/github/stars/fedai-oss/fl-go?style=social)](https://github.com/fedai-oss/fl-go/stargazers)
[![Forks](https://img.shields.io/github/forks/fedai-oss/fl-go?style=social)](https://github.com/fedai-oss/fl-go/network/members)
[![Issues](https://img.shields.io/github/issues/fedai-oss/fl-go)](https://github.com/fedai-oss/fl-go/issues)

</div>

---

A Go implementation of [OpenFL](https://github.com/securefederatedai/openfl) - An Open Framework for Federated Learning.

FL-Go provides the same CLI-driven workflow as the original OpenFL but with Go handling the orchestration and coordination while delegating ML operations to Python scripts.

## Features

- **OpenFL-Compatible CLI**: Same commands and workflow as the original OpenFL
- **Go Orchestration**: Fast, efficient coordination and communication in Go
- **Python ML Integration**: Seamless delegation of training/evaluation to Python
- **gRPC Communication**: Secure, efficient communication between components
- **Multi-round Federated Learning**: Support for multiple training rounds
- **Asynchronous Federated Learning**: Based on [Papaya paper](https://arxiv.org/abs/2111.04877) for scalable FL
- **Mode Switching**: Easy switching between synchronous and asynchronous FL modes
- **Staleness-Aware Aggregation**: Intelligent handling of stale updates in async mode
- **Multiple Aggregation Algorithms**: Support for FedAvg, FedOpt, and FedProx algorithms
- **Modular Algorithm Framework**: Easy to add new aggregation algorithms
- **Hyperparameter Configuration**: Fine-tune algorithm behavior via YAML configuration
- **Comprehensive Monitoring**: Real-time web UI with REST API for tracking FL metrics
- **Event Streaming**: WebSocket-based real-time monitoring of federation progress
- **Resource Monitoring**: Track system performance and collaborator health

## Architecture

```
┌─────────────────┐    gRPC     ┌─────────────────┐
│   Aggregator    │◄────────────┤  Collaborator   │
│   (Go Server)   │             │   (Go Client)   │
└─────────────────┘             └─────────────────┘
         │                               │
         │ metrics                       │ metrics
         ▼                               ▼
┌─────────────────┐             ┌─────────────────┐
│ Model Averaging │             │ Training Script │
│   (Go Logic)    │             │   (Python ML)   │
└─────────────────┘             └─────────────────┘
         │
         │ monitoring
         ▼
┌─────────────────┐      REST/WS    ┌─────────────────┐
│ Monitoring API  │◄────────────────┤   Web Dashboard │
│  (Go Server)    │                 │  (React TypeScript)│
└─────────────────┘                 └─────────────────┘
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

#### Synchronous Mode (Default)
```yaml
mode: "sync"  # Traditional round-based FL
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

#### Asynchronous Mode (Papaya-style)
```yaml
mode: "async"  # Non-round-based FL
rounds: 100  # Maximum rounds (async continues until stopped)
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
      epochs: 5
      batch_size: 32
      lr: 0.001
      data_path: "data"

# Async-specific configuration
async_config:
  max_staleness: 300      # Maximum staleness in seconds
  min_updates: 1          # Minimum updates before aggregation
  aggregation_delay: 10   # Delay in seconds before aggregating
  staleness_weight: 0.95  # Weight decay factor for stale updates

# Monitoring configuration
monitoring:
  enabled: true
  monitoring_server_url: "http://localhost:8080"
  collect_resource_metrics: true
  report_interval: 30
  enable_realtime_events: true
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

## Validation

To quickly validate that FL-GO is working correctly:

```bash
# Run comprehensive validation (tests both sync and async modes)
./scripts/validate_fl_flows.sh
```

This automated script will:
- ✅ Test synchronous FL with round-based training
- ✅ Test asynchronous FL with continuous training  
- ✅ Validate model aggregation and distribution
- ✅ Check gRPC communication
- ✅ Parse logs and report results

For detailed validation procedures, troubleshooting, and custom testing, see the **[Validation Guide](docs/VALIDATION.md)**.

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

FL-Go delegates all ML operations to Python scripts. The included `taskrunner.py` provides a template that:

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

# Start monitoring server
make run-monitor

# Build and start web UI
make install-web-deps
make start-web

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

### FL Validation

FL-GO includes comprehensive validation scripts to test both sync and async FL flows:

```bash
# Run comprehensive FL validation (recommended)
./scripts/validate_fl_flows.sh

# This validates:
# ✅ Synchronous FL with round-based training
# ✅ Asynchronous FL with continuous training
# ✅ Model aggregation and distribution
# ✅ gRPC communication
# ✅ Log parsing and result validation
```

For detailed validation procedures, troubleshooting, and custom testing scenarios, see the **[Validation Guide](docs/VALIDATION.md)**.

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

## FL Modes Comparison

| Feature | Synchronous FL | Asynchronous FL |
|---------|----------------|-----------------|
| **Training Style** | Round-based | Continuous |
| **Synchronization** | All collaborators wait | No waiting |
| **Scalability** | Limited by slowest | Handles stragglers |
| **Convergence** | Predictable | Faster (5x) |
| **Communication** | Synchronized rounds | On-demand |
| **Resource Usage** | Idle waiting | Full utilization |
| **Best For** | Small-medium scale | Large scale, production |

## Comparison with Original OpenFL

| Feature | OpenFL (Python) | FL-Go |
|---------|----------------|-----------|
| CLI Interface | ✅ | ✅ |
| Workspace Init | `fx plan init` | `fx plan init` |
| Aggregator | Python | Go + gRPC |
| Collaborator | Python | Go + Python |
| ML Training | Python | Python (delegated) |
| Communication | gRPC | gRPC |
| Performance | Good | Excellent |
| **Async FL Support** | ❌ | ✅ (Papaya-based) |
| **Mode Switching** | ❌ | ✅ |
| **Staleness Handling** | ❌ | ✅ |
| **Web UI Monitoring** | ❌ | ✅ (Real-time) |
| **REST API** | ❌ | ✅ (Comprehensive) |
| **Event Streaming** | ❌ | ✅ (WebSocket) |
| **Resource Monitoring** | ❌ | ✅ |

## License

Apache License 2.0 - Same as original OpenFL

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Submit a pull request

## Aggregation Algorithms

FL-GO supports multiple state-of-the-art aggregation algorithms that can be easily configured:

### Supported Algorithms

| Algorithm | Description | Use Case |
|-----------|-------------|----------|
| **FedAvg** | Vanilla federated averaging | Homogeneous environments, IID data |
| **FedOpt** | Adaptive server optimization with Adam-like momentum | Heterogeneous clients, faster convergence |
| **FedProx** | Proximal term for handling client heterogeneity | Non-IID data, unstable training environments |

### Configuration Examples

**FedOpt Algorithm:**
```yaml
algorithm:
  name: "fedopt"
  hyperparameters:
    server_learning_rate: 1.0
    beta1: 0.9
    beta2: 0.999
    epsilon: 1e-7
```

**FedProx Algorithm:**
```yaml
algorithm:
  name: "fedprox"
  hyperparameters:
    mu: 0.01  # Proximal term coefficient
```

### Algorithm Selection Guidelines

- **FedAvg**: Start here for baseline performance
- **FedOpt**: Use when FedAvg converges slowly or in heterogeneous environments
- **FedProx**: Choose for highly non-IID data or when training is unstable

For comprehensive documentation on algorithms, hyperparameter tuning, and implementation details, see [Algorithm Guide](docs/ALGORITHMS.md).

## Monitoring & Observability

FL-Go includes a comprehensive monitoring system for tracking federated learning progress in real-time:

### Features

- **🌐 Modern Web Dashboard**: Beautiful, responsive React-based interface
- **📊 Real-time Metrics**: Live federation progress, collaborator status, and training metrics
- **🔗 REST API**: Complete API for programmatic access to all monitoring data
- **⚡ WebSocket Streaming**: Real-time event updates for live monitoring
- **📈 Performance Analytics**: Track convergence, efficiency, and resource utilization
- **🎯 Event Timeline**: Comprehensive logging of all federation activities

### Quick Start

```bash
# Start the monitoring server
make run-monitor

# Or manually
go run cmd/monitor/main.go --port 8080
```

**Access the monitoring system:**
- **API**: http://localhost:8080/api/v1/health
- **Web UI**: http://localhost:3000 (after `make install-web-deps && make start-web`)

### Configuration

Add monitoring to your FL plan:

```yaml
monitoring:
  enabled: true
  monitoring_server_url: "http://localhost:8080"
  collect_resource_metrics: true
  report_interval: 30
  enable_realtime_events: true
```

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/federations` | List all federations |
| `GET /api/v1/collaborators` | Get collaborator metrics |
| `GET /api/v1/rounds` | Training round statistics |
| `GET /api/v1/events` | Event timeline |
| `WS /api/v1/ws` | Real-time event stream |

For complete documentation, API reference, and integration guide, see **[Monitoring Guide](MONITORING.md)**.

## Security

FL-Go includes comprehensive security features for production federated learning deployments:

### 🔒 mTLS (Mutual TLS) Security

Secure gRPC communication between aggregators and collaborators with mutual authentication:

**Features:**
- **Auto-generated certificates** for development and testing
- **Custom certificate support** for production deployments
- **Mutual authentication** ensures both client and server identity verification
- **TLS 1.2+ encryption** for all federated learning communications
- **Certificate validation** with proper hostname verification

**Configuration:**
```yaml
security:
  tls:
    enabled: true
    auto_generate_cert: true  # Auto-generate for development
    server_name: "fl-go-server"
    insecure_skip_tls: false  # Never true in production
    # Optional: Custom certificate paths
    # cert_path: "certs/server.crt"
    # key_path: "certs/server.key" 
    # ca_path: "certs/ca.crt"
```

**Quick Start with mTLS:**
```bash
# Use the secure example plan
cp plans/secure_example_plan.yaml my_secure_plan.yaml

# Start aggregator with mTLS
fx aggregator start --plan my_secure_plan.yaml

# Start collaborators with mTLS
fx collaborator start collab1 --plan my_secure_plan.yaml
```

**Production Deployment:**
- Generate proper certificates signed by your organization's CA
- Set `auto_generate_cert: false` and provide certificate paths
- Enable certificate revocation checking
- Use proper DNS names in certificates

### 🔐 Enhanced Security Features

**API Key & JWT Authentication:**
```yaml
# monitoring_config.yaml
auth:
  enabled: true
  required_role: "readonly"  # admin, monitor, readonly
  api_key:
    enabled: true
    header_name: "X-API-Key"
    keys:
      "admin-key-12345": "admin"
      "monitor-key-67890": "monitor"
      "readonly-key-abcde": "readonly"
  jwt:
    enabled: true
    secret: "your-jwt-secret-key"
    token_expiry: "24h"
    issuer: "fl-go-monitoring"
```

**Using API Keys:**
```bash
# Monitor with API key
curl -H "X-API-Key: monitor-key-67890" http://localhost:8080/api/v1/federations

# Admin access with API key
curl -H "X-API-Key: admin-key-12345" http://localhost:8080/api/v1/federations
```

**Role-Based Access Control:**
- **Admin**: Full access to all endpoints and operations
- **Monitor**: Read/write access to monitoring data
- **ReadOnly**: Read-only access to monitoring data

### 💾 Database Backends

**Memory Storage (Default):**
```yaml
storage:
  backend: "memory"
  memory:
    max_entries: 10000
```

**PostgreSQL Storage:**
```yaml
storage:
  backend: "postgresql"
  postgresql:
    host: "localhost"
    port: 5432
    user: "fl_user"
    password: "fl_password"
    database: "fl_monitoring"
    ssl_mode: "require"
    max_connections: 20
```

**Redis Storage:**
```yaml
storage:
  backend: "redis"
  redis:
    address: "localhost:6379"
    password: ""
    database: 0
    pool_size: 10
    ttl: "24h"
```

**Production Recommendations:**
- Use PostgreSQL for persistent, queryable storage
- Use Redis for high-performance, distributed scenarios
- Configure proper backup strategies for database backends
- Set appropriate TTL values for Redis to manage memory usage

## Roadmap

- [x] **Add mTLS security** ✅
- [x] **Implement additional aggregation algorithms (FedOpt, FedProx)** ✅
- [x] **Web UI for monitoring** ✅
- [x] **Enhanced security features (API keys, OAuth)** ✅
- [x] **Database backends for monitoring (PostgreSQL, Redis)** ✅
- [ ] Add more algorithms (FedNova, SCAFFOLD, LAG)
- [ ] Add TEE support
- [ ] Integration with popular ML frameworks
- [ ] Mobile-responsive monitoring dashboard
- [ ] Advanced analytics and ML insights