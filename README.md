# FL-Go

<!-- Badges -->
<div align="center">

[![CI Status](https://github.com/fedai-oss/fl-go/actions/workflows/ci.yml/badge.svg)](https://github.com/fedai-oss/fl-go/actions/workflows/ci.yml)
[![Security](https://github.com/fedai-oss/fl-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/fedai-oss/fl-go/actions/workflows/codeql.yml)
[![FOSSA](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ffedai-oss%2Ffl-go.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Ffedai-oss%2Ffl-go)
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

## 🚀 Features

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
- **Security Features**: mTLS support for secure communication
- **Production Ready**: Comprehensive testing, CI/CD, and monitoring

## 🏗️ Architecture

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

## 📁 Project Structure

```
fl-go/
├── 📁 api/                    # Protocol definitions
├── 📁 cmd/                    # Application entry points
├── 📁 pkg/                    # Core packages
├── 📁 web/                    # Web UI
├── 📁 examples/               # Examples and samples
│   ├── 📁 plans/              # Federation plan examples
│   ├── 📁 workspaces/         # Complete workspace examples
│   ├── 📁 monitoring/         # Monitoring examples
│   └── 📁 scripts/            # Example scripts
├── 📁 docs/                   # Documentation
├── 📁 scripts/                # Build and utility scripts
├── 📁 configs/                # Configuration files
├── 📁 deploy/                 # Deployment configurations
├── 📁 tests/                  # Test files and test data
└── 📁 tools/                  # Development tools
```

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or later
- Git
- Docker (optional)
- Node.js and npm (for web UI development)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/fedai-oss/fl-go.git
   cd fl-go
   ```

2. **Setup development environment**
   ```bash
   make setup-dev
   ```

3. **Build the project**
   ```bash
   make build
   ```

4. **Run tests to verify installation**
   ```bash
   make test
   ```

### Your First Federation

1. **Create a federation plan**
   ```bash
   # Use a sample plan
   cp examples/plans/basic/sync_plan.yaml my_federation.yaml
   ```

2. **Start the aggregator**
   ```bash
   ./bin/fx aggregator start --config my_federation.yaml
   ```

3. **Start collaborators** (in separate terminals)
   ```bash
   ./bin/fx collaborator start --config my_federation.yaml --name client-1
   ./bin/fx collaborator start --config my_federation.yaml --name client-2
   ```

4. **Monitor progress**
   ```bash
   ./bin/fx monitor start --config configs/monitoring/development.yaml
   ```
   Then visit `http://localhost:3000` in your browser.

## 📚 Documentation

### Getting Started
- [Installation Guide](docs/getting-started/installation.md)
- [Quick Start Guide](docs/getting-started/quick-start.md)

### User Guides
- [Federation Plans](docs/user-guide/federation-plans.md)
- [CLI Reference](docs/user-guide/cli-reference.md)
- [Monitoring](docs/user-guide/monitoring.md)
- [Security](docs/user-guide/security.md)

### Examples
- [Basic Federation](docs/examples/basic-federation.md)
- [Async Federation](docs/examples/async-federation.md)
- [Secure Federation](docs/examples/secure-federation.md)
- [Monitoring Setup](docs/examples/monitoring-setup.md)

### Development
- [Contributing Guide](docs/development/contributing.md)
- [Architecture](docs/development/architecture.md)
- [Testing](docs/development/testing.md)
- [Deployment](docs/development/deployment.md)

### Security
- [FOSSA Setup](docs/security/fossa-setup.md)
- [mTLS Configuration](docs/security/mTLS-configuration.md)
- [Security Best Practices](docs/security/security-best-practices.md)

## 🔧 Development

### Building

```bash
# Build all components
make build

# Build specific components
make build-monitor
make build-web

# Build Docker image
make docker-build
```

### Testing

```bash
# Run all tests
make test

# Run specific test types
make test-unit
make test-integration
make test-security
make test-performance

# Run with coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make format

# Run linting
make lint

# Validate federation flows
make validate
```

## 🛠️ Examples

### Basic Federation
```bash
# Use the basic example
cd examples/workspaces/basic_fl_workspace
./bin/fx aggregator start --config plan.yaml
```

### Secure Federation with mTLS
```bash
# Use the secure example
cd examples/workspaces/secure_mtls_workspace
./bin/fx security generate-certs --output-dir certs
./bin/fx aggregator start --config plan.yaml
```

### Async Federation
```bash
# Use the async example
cd examples/workspaces/async_fl_workspace
./bin/fx aggregator start --config plan.yaml
```

## 🔒 Security & Compliance

FL-GO includes comprehensive security features and compliance tools:

### Security Features
- **mTLS Support**: Mutual TLS authentication for secure communication
- **Certificate Management**: Automated certificate generation and validation
- **Secure Communication**: Encrypted gRPC communication
- **Access Control**: Role-based access control for monitoring

### Compliance Tools
- **FOSSA Integration**: Automated dependency analysis and license compliance
- **Security Scanning**: Integrated gosec and govulncheck for vulnerability detection
- **Code Quality**: Automated linting and code quality checks
- **Audit Trail**: Comprehensive logging and monitoring

### Setup Security Features
```bash
# Generate certificates for mTLS
./bin/fx security generate-certs --output-dir certs

# Verify certificates
./bin/fx security verify-certs --cert certs/client.crt --key certs/client.key
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/development/contributing.md) for details.

### Development Setup
```bash
# Setup development environment
make setup-dev

# Run tests
make test

# Format code
make format
```

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and documentation
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [OpenFL](https://github.com/securefederatedai/openfl) - The original OpenFL framework
- [Papaya](https://arxiv.org/abs/2111.04877) - Asynchronous federated learning paper
- [gRPC](https://grpc.io/) - High-performance RPC framework
- [React](https://reactjs.org/) - Web UI framework

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Examples**: [examples/](examples/)
- **Issues**: [GitHub Issues](https://github.com/fedai-oss/fl-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fedai-oss/fl-go/discussions)

---

<div align="center">

**FL-Go** - Empowering Federated Learning with Go

[Get Started](#quick-start) • [Documentation](docs/) • [Examples](examples/) • [Contributing](docs/development/contributing.md)

</div>