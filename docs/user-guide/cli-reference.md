# CLI Reference

The FL-GO CLI (`fx`) provides commands for managing federated learning sessions, monitoring, and system administration.

## Command Overview

```bash
fx [command] [subcommand] [options]
```

## Global Options

- `--config <file>`: Specify configuration file
- `--verbose`: Enable verbose output
- `--log-level <level>`: Set log level (debug, info, warn, error)
- `--help`: Show help for command

## Commands

### Aggregator Commands

#### `fx aggregator start`
Start the federated learning aggregator.

```bash
fx aggregator start [options]
```

**Options:**
- `--config <file>`: Federation plan file (required)
- `--host <host>`: Host to bind to (default: localhost)
- `--port <port>`: Port to bind to (default: 8080)
- `--timeout <duration>`: Request timeout (default: 30s)

**Example:**
```bash
fx aggregator start --config examples/plans/basic/sync_plan.yaml
```

#### `fx aggregator stop`
Stop the aggregator gracefully.

```bash
fx aggregator stop [options]
```

**Options:**
- `--force`: Force stop without graceful shutdown

### Collaborator Commands

#### `fx collaborator start`
Start a federated learning collaborator.

```bash
fx collaborator start [options]
```

**Options:**
- `--config <file>`: Federation plan file (required)
- `--name <name>`: Collaborator name (required)
- `--host <host>`: Host to bind to (default: localhost)
- `--port <port>`: Port to bind to (auto-assigned if not specified)
- `--data-dir <path>`: Directory containing training data
- `--model-dir <path>`: Directory for model storage

**Example:**
```bash
fx collaborator start --config examples/plans/basic/sync_plan.yaml --name client-1
```

#### `fx collaborator stop`
Stop a collaborator gracefully.

```bash
fx collaborator stop [options]
```

**Options:**
- `--name <name>`: Collaborator name
- `--force`: Force stop without graceful shutdown

### Monitoring Commands

#### `fx monitor start`
Start the monitoring server.

```bash
fx monitor start [options]
```

**Options:**
- `--config <file>`: Monitoring configuration file
- `--api-port <port>`: API server port (default: 8080)
- `--webui-port <port>`: Web UI port (default: 3000)
- `--storage <backend>`: Storage backend (memory, postgres, redis)

**Example:**
```bash
fx monitor start --config configs/monitoring/development.yaml
```

#### `fx monitor status`
Show monitoring server status.

```bash
fx monitor status [options]
```

### Plan Commands

#### `fx plan validate`
Validate a federation plan file.

```bash
fx plan validate <file> [options]
```

**Options:**
- `--strict`: Enable strict validation
- `--output <format>`: Output format (text, json, yaml)

**Example:**
```bash
fx plan validate examples/plans/basic/sync_plan.yaml
```

#### `fx plan generate`
Generate a federation plan template.

```bash
fx plan generate [options]
```

**Options:**
- `--type <type>`: Plan type (basic, async, secure, monitoring)
- `--output <file>`: Output file path
- `--name <name>`: Federation name

**Example:**
```bash
fx plan generate --type basic --output my_plan.yaml --name my-federation
```

### Federation Commands

#### `fx status`
Show federation status.

```bash
fx status [options]
```

**Options:**
- `--config <file>`: Federation plan file
- `--format <format>`: Output format (text, json, yaml)

#### `fx logs`
Show federation logs.

```bash
fx logs [options]
```

**Options:**
- `--config <file>`: Federation plan file
- `--follow`: Follow log output
- `--tail <lines>`: Number of lines to show (default: 100)
- `--level <level>`: Minimum log level

#### `fx stop`
Stop all federation participants.

```bash
fx stop [options]
```

**Options:**
- `--config <file>`: Federation plan file
- `--force`: Force stop without graceful shutdown

### Security Commands

#### `fx security generate-certs`
Generate mTLS certificates.

```bash
fx security generate-certs [options]
```

**Options:**
- `--output-dir <path>`: Output directory for certificates
- `--ca-name <name>`: CA certificate name
- `--validity <days>`: Certificate validity in days (default: 365)

**Example:**
```bash
fx security generate-certs --output-dir certs --ca-name fl-ca
```

#### `fx security verify-certs`
Verify certificate validity.

```bash
fx security verify-certs [options]
```

**Options:**
- `--cert <file>`: Certificate file to verify
- `--key <file>`: Private key file
- `--ca <file>`: CA certificate file

### Utility Commands

#### `fx version`
Show version information.

```bash
fx version
```

#### `fx help`
Show help information.

```bash
fx help [command]
```

## Configuration Files

### Federation Plan Format
See [Federation Plans](./federation-plans.md) for detailed format.

### Monitoring Configuration Format
```yaml
enabled: true
api_port: 8080
webui_port: 3000
metrics_retention: "24h"
collection_interval: "5s"
storage_backend: "memory"
database_url: ""
production: false
allowed_origins: []
```

## Environment Variables

- `FL_GO_LOG_LEVEL`: Set log level
- `FL_GO_CONFIG_PATH`: Default config file path
- `FL_GO_DATA_DIR`: Default data directory
- `FL_GO_MODEL_DIR`: Default model directory

## Examples

### Basic Federation Workflow

```bash
# 1. Validate the plan
fx plan validate examples/plans/basic/sync_plan.yaml

# 2. Start aggregator
fx aggregator start --config examples/plans/basic/sync_plan.yaml

# 3. Start collaborators (in separate terminals)
fx collaborator start --config examples/plans/basic/sync_plan.yaml --name client-1
fx collaborator start --config examples/plans/basic/sync_plan.yaml --name client-2

# 4. Monitor progress
fx monitor start --config configs/monitoring/development.yaml

# 5. Check status
fx status --config examples/plans/basic/sync_plan.yaml

# 6. Stop federation
fx stop --config examples/plans/basic/sync_plan.yaml
```

### Secure Federation with mTLS

```bash
# 1. Generate certificates
fx security generate-certs --output-dir certs

# 2. Start secure federation
fx aggregator start --config examples/plans/advanced/secure_plan.yaml
fx collaborator start --config examples/plans/advanced/secure_plan.yaml --name client-1
```

## Troubleshooting

### Common Issues

1. **Port already in use**: Use different ports or stop existing processes
2. **Invalid plan**: Use `fx plan validate` to check syntax
3. **Certificate errors**: Verify certificate paths and permissions
4. **Connection refused**: Check host/port configuration

### Debug Mode

Enable debug logging for troubleshooting:

```bash
export FL_GO_LOG_LEVEL=debug
fx aggregator start --config plan.yaml --verbose
```
