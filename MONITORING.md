# FL Monitoring System

A comprehensive monitoring solution for federated learning with an API-first approach and modern web UI.

## Features

### Core Monitoring
- **Real-time Federation Tracking**: Monitor active federations, rounds, and collaborators
- **Performance Metrics**: Track training efficiency, convergence rates, and resource utilization
- **Event Streaming**: Real-time events via WebSocket for live monitoring
- **Resource Monitoring**: CPU, memory, disk, and network usage tracking
- **Alert System**: Configurable alerts for performance and error conditions

### API-First Design
- **RESTful API**: Complete REST API for all monitoring data
- **Real-time Updates**: WebSocket support for live data streaming
- **Flexible Queries**: Advanced filtering and pagination
- **Multiple Backends**: Memory, SQLite, PostgreSQL storage options

### Modern Web UI
- **Responsive Dashboard**: Beautiful, responsive React-based interface
- **Real-time Charts**: Live updating charts and metrics
- **Federation Overview**: Detailed views of federations and collaborators
- **Event Timeline**: Comprehensive event logging and visualization
- **Mobile Friendly**: Works on desktop, tablet, and mobile devices

## Quick Start

### 1. Start the Monitoring Server

```bash
# Navigate to the FL-Go directory
cd go/fl-go

# Start the monitoring server (with sample data)
go run cmd/monitor/main.go
```

The monitoring server will start on:
- **API Server**: http://localhost:8080
- **Web UI**: http://localhost:3000 (when built)

### 2. Build and Start the Web UI

```bash
# Navigate to the web directory
cd web

# Install dependencies
npm install

# Start the development server
npm run dev
```

The web UI will be available at http://localhost:3000

### 3. Configure Federation Monitoring

Add monitoring configuration to your FL plan:

```yaml
monitoring:
  enabled: true
  monitoring_server_url: "http://localhost:8080"
  collect_resource_metrics: true
  report_interval: 30
  enable_realtime_events: true
```

## API Documentation

### Health Check
```bash
curl http://localhost:8080/api/v1/health
```

### Get Active Federations
```bash
curl http://localhost:8080/api/v1/federations?active=true
```

### Get Federation Details
```bash
curl http://localhost:8080/api/v1/federations/{federation_id}
```

### Get System Overview
```bash
curl http://localhost:8080/api/v1/federations/{federation_id}/overview
```

### Get Collaborators
```bash
curl http://localhost:8080/api/v1/collaborators?federation_id={federation_id}
```

### Get Training Rounds
```bash
curl http://localhost:8080/api/v1/rounds?federation_id={federation_id}
```

### Get Events
```bash
curl "http://localhost:8080/api/v1/events?federation_id={federation_id}&page=1&per_page=50"
```

### WebSocket Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?federation_id={federation_id}');
ws.onmessage = (event) => {
  const monitoringEvent = JSON.parse(event.data);
  console.log('Real-time event:', monitoringEvent);
};
```

## Configuration

### Monitoring Server Configuration

Create `monitoring_config.yaml`:

```yaml
enabled: true
api_port: 8080
webui_port: 3000
metrics_retention: "24h"
collection_interval: "30s"
enable_resource_metrics: true
enable_realtime_events: true
storage_backend: "memory"  # memory, sqlite, postgres

# Resource thresholds for alerts
resource_thresholds:
  cpu_warning: 80.0
  cpu_critical: 95.0
  memory_warning: 85.0
  memory_critical: 95.0
  disk_warning: 90.0
  disk_critical: 98.0

# Security settings
security:
  enable_cors: true
  allowed_origins: ["*"]  # Restrict in production
  api_key_required: false
```

### Federation Plan Configuration

Add to your FL plan YAML:

```yaml
monitoring:
  enabled: true
  monitoring_server_url: "http://localhost:8080"
  collect_resource_metrics: true
  report_interval: 30
  enable_realtime_events: true
```

## Architecture

### Components

1. **Monitoring Service Interface** (`pkg/monitoring/service.go`)
   - Defines the contract for monitoring operations
   - Supports multiple storage backends

2. **Storage Layer** (`pkg/monitoring/storage.go`)
   - In-memory storage for development/testing
   - Extensible for database backends

3. **REST API Server** (`pkg/monitoring/api.go`)
   - RESTful endpoints for all monitoring data
   - WebSocket support for real-time updates

4. **Integration Hooks** (`pkg/monitoring/hooks.go`)
   - Easy integration with existing FL components
   - Event-driven metric collection

5. **Web Frontend** (`web/`)
   - Modern React-based dashboard
   - Real-time updates via WebSocket
   - Responsive design

### Data Types

- **FederationMetrics**: Overall federation status and progress
- **CollaboratorMetrics**: Individual collaborator performance
- **RoundMetrics**: Training round details and results
- **ModelUpdateMetrics**: Model update timing and quality
- **ResourceMetrics**: System resource usage
- **MonitoringEvents**: Real-time event stream

## Integration Guide

### Adding Monitoring to Aggregator

```go
import "github.com/ishaileshpant/fl-go/pkg/monitoring"

// Create monitoring hooks
monitoringHooks := monitoring.NewMonitoringHooks(monitoringService, true)

// Record federation start
err := monitoringHooks.OnFederationStart(ctx, plan, aggregatorAddress)

// Record round start
roundID, err := monitoringHooks.OnRoundStart(ctx, federationID, roundNumber, algorithm, participantCount)

// Record round completion
err := monitoringHooks.OnRoundEnd(ctx, roundID, federationID, roundNumber, duration, updatesReceived, accuracy, loss)
```

### Adding Monitoring to Collaborator

```go
// Record collaborator join
err := monitoringHooks.OnCollaboratorJoin(ctx, collaboratorID, federationID, address)

// Record training start/end
err := monitoringHooks.OnTrainingStart(ctx, collaboratorID, roundNumber)
err := monitoringHooks.OnTrainingEnd(ctx, collaboratorID, roundNumber, duration, accuracy, loss)

// Record model update
err := monitoringHooks.OnModelUpdateReceived(ctx, federationID, collaboratorID, roundNumber, updateSize, processingTime, staleness, weight)
```

## Development

### Building the Web UI

```bash
cd web
npm install
npm run build
```

### Running Tests

```bash
go test ./pkg/monitoring/...
```

### Adding New Metrics

1. Define new metric types in `pkg/monitoring/types.go`
2. Add storage methods in `pkg/monitoring/storage.go`
3. Add API endpoints in `pkg/monitoring/api.go`
4. Update the web UI to display new metrics

## Production Deployment

### Database Backend

For production, use PostgreSQL:

```yaml
storage_backend: "postgres"
database_url: "postgres://user:password@localhost/fl_monitoring"
```

### Security

- Enable API key authentication
- Restrict CORS origins
- Use HTTPS in production
- Implement rate limiting

### Scaling

- Use PostgreSQL for persistent storage
- Consider Redis for caching
- Load balance the API server
- Use CDN for web assets

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure monitoring server is running on correct port
2. **CORS Errors**: Check allowed origins in configuration
3. **Memory Usage**: Adjust retention settings for large deployments
4. **WebSocket Disconnects**: Implement reconnection logic in clients

### Monitoring the Monitor

The monitoring system itself exposes metrics:
- API response times
- Memory usage
- Event processing rates
- Storage backend health

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Update documentation
5. Submit a pull request

## License

This monitoring system is part of the FL-Go project and follows the same license terms.
