# Installation Guide

## Prerequisites

- Go 1.23 or later
- Git
- Docker (optional, for containerized deployment)

## Quick Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/fedai-oss/fl-go.git
cd fl-go

# Build the project
make build

# Install CLI tools
go install ./cmd/fx
```

### Using Docker

```bash
# Build the Docker image
docker build -t fl-go .

# Run the container
docker run -it fl-go
```

## Verification

After installation, verify that everything is working:

```bash
# Check if the CLI is available
fx --help

# Run tests
make test

# Check build status
make build
```

## Next Steps

- See [Quick Start Guide](./quick-start.md) for your first federation
- Check [CLI Reference](../user-guide/cli-reference.md) for available commands
- Explore [Examples](../examples/) for different use cases
