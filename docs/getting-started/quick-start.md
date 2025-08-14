# Quick Start Guide

This guide will help you set up your first federated learning session using FL-GO.

## Prerequisites

- FL-GO installed (see [Installation Guide](./installation.md))
- Basic understanding of federated learning concepts

## Step 1: Create a Federation Plan

Create a basic federation plan file `my_first_federation.yaml`:

```yaml
federation:
  name: "my-first-federation"
  description: "A simple federated learning example"
  
  aggregator:
    host: "localhost"
    port: 8080
    
  collaborators:
    - name: "client-1"
      host: "localhost"
      port: 8081
    - name: "client-2"
      host: "localhost"
      port: 8082
      
  algorithm:
    type: "fedavg"
    rounds: 10
    epochs_per_round: 5
    
  model:
    type: "simple_nn"
    input_size: 784
    hidden_size: 128
    output_size: 10
```

## Step 2: Start the Aggregator

```bash
# Start the aggregator
fx aggregator start --config my_first_federation.yaml
```

## Step 3: Start Collaborators

In separate terminals, start the collaborators:

```bash
# Terminal 1 - Start collaborator 1
fx collaborator start --config my_first_federation.yaml --name client-1

# Terminal 2 - Start collaborator 2
fx collaborator start --config my_first_federation.yaml --name client-2
```

## Step 4: Monitor Progress

Open the web UI to monitor the federation progress:

```bash
# Start the monitoring server
fx monitor start --config configs/monitoring/development.yaml
```

Then visit `http://localhost:3000` in your browser.

## Step 5: Verify Results

Check the federation status:

```bash
# Check federation status
fx status --config my_first_federation.yaml

# View logs
fx logs --config my_first_federation.yaml
```

## Next Steps

- Explore [Advanced Examples](../examples/) for more complex scenarios
- Learn about [Federation Plans](../user-guide/federation-plans.md)
- Set up [Monitoring](../user-guide/monitoring.md) for production use
- Configure [Security](../user-guide/security.md) with mTLS
