# OpenFL-Go E2E MVP

One-round federated MNIST training:
- Go for orchestration (gRPC + FedAvg)
- Python (PyTorch) for model training

## Prerequisites
- Go 1.21
- Python 3.8+ (`pip install torch torchvision`)

## Setup & Run
```bash
go mod download
# In one terminal:
go run cmd/aggregator/main.go
# In two others:
go run cmd/collaborator/main.go --id collab1
go run cmd/collaborator/main.go --id collab2

