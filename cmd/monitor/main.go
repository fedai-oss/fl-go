package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ishaileshpant/fl-go/pkg/monitoring"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		configPath = flag.String("config", "monitoring_config.yaml", "Path to monitoring configuration file")
		port       = flag.Int("port", 8080, "API server port")
		webPort    = flag.Int("web-port", 3000, "Web UI port")
	)
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		config = &monitoring.MonitoringConfig{
			Enabled:               true,
			APIPort:               *port,
			WebUIPort:             *webPort,
			MetricsRetention:      24 * time.Hour,
			CollectionInterval:    30 * time.Second,
			EnableResourceMetrics: true,
			EnableRealTimeEvents:  true,
			StorageBackend:        "memory",
		}
	}

	// Override with command line arguments
	if *port != 8080 {
		config.APIPort = *port
	}
	if *webPort != 3000 {
		config.WebUIPort = *webPort
	}

	log.Printf("Starting FL Monitoring Server")
	log.Printf("API Port: %d", config.APIPort)
	log.Printf("Web UI Port: %d", config.WebUIPort)
	log.Printf("Storage Backend: %s", config.StorageBackend)

	// Create storage backend
	var storage monitoring.MonitoringService
	switch config.StorageBackend {
	case "memory":
		storage = monitoring.NewMemoryStorage(config)
	default:
		log.Fatalf("Unsupported storage backend: %s", config.StorageBackend)
	}

	// Create API server
	apiServer := monitoring.NewAPIServer(storage, config)

	// Start resource monitoring if enabled
	if config.EnableResourceMetrics {
		go startResourceMonitoring(storage, config)
	}

	// Create sample data for demonstration
	if err := createSampleData(storage); err != nil {
		log.Printf("Failed to create sample data: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Start API server
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	log.Println("FL Monitoring Server started successfully")
	log.Printf("API available at: http://localhost:%d/api/v1", config.APIPort)
	log.Printf("Health check: http://localhost:%d/api/v1/health", config.APIPort)
	log.Printf("Web UI will be available at: http://localhost:%d", config.WebUIPort)

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("FL Monitoring Server stopped")
}

// loadConfig loads monitoring configuration from file
func loadConfig(configPath string) (*monitoring.MonitoringConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config monitoring.MonitoringConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// startResourceMonitoring starts a goroutine to collect system resource metrics
func startResourceMonitoring(storage monitoring.MonitoringService, config *monitoring.MonitoringConfig) {
	ticker := time.NewTicker(config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Collect and record resource metrics for the monitoring server itself
			ctx := context.Background()

			// Simulate resource metrics collection
			// In a real implementation, you would use system calls or libraries to get actual metrics
			cpuUsage := 15.5                             // Percentage
			memoryUsage := 42.8                          // Percentage
			diskUsage := 67.2                            // Percentage
			memoryUsed := int64(1024 * 1024 * 512)       // 512MB
			memoryTotal := int64(1024 * 1024 * 1024 * 4) // 4GB
			networkRx := 1.2                             // Mbps
			networkTx := 0.8                             // Mbps

			metrics := &monitoring.ResourceMetrics{
				Timestamp:     time.Now(),
				CPUUsage:      cpuUsage,
				MemoryUsage:   memoryUsage,
				MemoryUsed:    memoryUsed,
				MemoryTotal:   memoryTotal,
				DiskUsage:     diskUsage,
				NetworkRxRate: networkRx,
				NetworkTxRate: networkTx,
			}

			if err := storage.RecordResourceMetrics(ctx, "monitoring-server", metrics); err != nil {
				log.Printf("Failed to record resource metrics: %v", err)
			}
		}
	}
}

// createSampleData creates some sample monitoring data for demonstration
func createSampleData(storage monitoring.MonitoringService) error {
	ctx := context.Background()

	// Create sample federation
	federation := &monitoring.FederationMetrics{
		ID:                "fed_demo_001",
		Name:              "Demo Federation",
		Status:            monitoring.StatusRunning,
		Mode:              "async",
		Algorithm:         "fedavg",
		StartTime:         time.Now().Add(-2 * time.Hour),
		CurrentRound:      5,
		TotalRounds:       10,
		ActiveCollabs:     3,
		TotalCollabs:      5,
		ModelSize:         1024000,
		LastUpdate:        time.Now().Add(-5 * time.Minute),
		AggregatorAddress: "localhost:8000",
	}

	if err := storage.RegisterFederation(ctx, federation); err != nil {
		return err
	}

	// Create sample collaborators
	collaborators := []*monitoring.CollaboratorMetrics{
		{
			ID:               "collab_001",
			FederationID:     "fed_demo_001",
			Address:          "192.168.1.100:8001",
			Status:           monitoring.CollabStatusConnected,
			JoinTime:         time.Now().Add(-2 * time.Hour),
			LastSeen:         time.Now().Add(-2 * time.Minute),
			CurrentRound:     5,
			UpdatesSubmitted: 5,
			TrainingTime:     45 * time.Minute,
			AverageLatency:   250.5,
			ErrorCount:       0,
		},
		{
			ID:               "collab_002",
			FederationID:     "fed_demo_001",
			Address:          "192.168.1.101:8001",
			Status:           monitoring.CollabStatusTraining,
			JoinTime:         time.Now().Add(-2 * time.Hour),
			LastSeen:         time.Now().Add(-1 * time.Minute),
			CurrentRound:     5,
			UpdatesSubmitted: 4,
			TrainingTime:     42 * time.Minute,
			AverageLatency:   380.2,
			ErrorCount:       1,
			LastError:        "Connection timeout",
		},
		{
			ID:               "collab_003",
			FederationID:     "fed_demo_001",
			Address:          "192.168.1.102:8001",
			Status:           monitoring.CollabStatusConnected,
			JoinTime:         time.Now().Add(-90 * time.Minute),
			LastSeen:         time.Now().Add(-30 * time.Second),
			CurrentRound:     5,
			UpdatesSubmitted: 3,
			TrainingTime:     28 * time.Minute,
			AverageLatency:   195.8,
			ErrorCount:       0,
		},
	}

	for _, collaborator := range collaborators {
		if err := storage.RegisterCollaborator(ctx, collaborator); err != nil {
			return err
		}
	}

	// Create sample rounds
	for i := 1; i <= 5; i++ {
		startTime := time.Now().Add(-time.Duration(6-i) * 20 * time.Minute)
		endTime := startTime.Add(15 * time.Minute)

		round := &monitoring.RoundMetrics{
			ID:               fmt.Sprintf("round_fed_demo_001_%d", i),
			FederationID:     "fed_demo_001",
			RoundNumber:      i,
			Algorithm:        "fedavg",
			StartTime:        startTime,
			EndTime:          &endTime,
			Duration:         15 * time.Minute,
			ParticipantCount: 3,
			UpdatesReceived:  3,
			AggregationTime:  2 * time.Minute,
			Status:           "completed",
		}

		// Add some accuracy progression
		accuracy := 0.65 + float64(i)*0.05
		loss := 0.8 - float64(i)*0.08
		round.ModelAccuracy = &accuracy
		round.ModelLoss = &loss

		if err := storage.RecordRoundStart(ctx, round); err != nil {
			return err
		}
		if err := storage.RecordRoundEnd(ctx, round.ID, round); err != nil {
			return err
		}
	}

	// Create sample model updates
	for i := 1; i <= 5; i++ {
		for j, collabID := range []string{"collab_001", "collab_002", "collab_003"} {
			if i == 2 && j == 1 {
				continue // Simulate collab_002 missing round 2
			}

			update := &monitoring.ModelUpdateMetrics{
				FederationID:   "fed_demo_001",
				CollaboratorID: collabID,
				RoundNumber:    i,
				Timestamp:      time.Now().Add(-time.Duration(6-i) * 20 * time.Minute),
				UpdateSize:     1024000 + j*50000,
				ProcessingTime: 150.0 + float64(j*50),
				Staleness:      0,
				Weight:         1.0,
			}

			if err := storage.RecordModelUpdate(ctx, update); err != nil {
				return err
			}
		}
	}

	// Create sample events
	events := []*monitoring.MonitoringEvent{
		{
			FederationID: "fed_demo_001",
			Type:         monitoring.MetricTypeRound,
			Timestamp:    time.Now().Add(-10 * time.Minute),
			Source:       "aggregator",
			Level:        "info",
			Message:      "Round 5 completed successfully",
			Data: map[string]interface{}{
				"round":        5,
				"participants": 3,
				"duration_ms":  900000,
			},
		},
		{
			FederationID: "fed_demo_001",
			Type:         monitoring.MetricTypeCollaborator,
			Timestamp:    time.Now().Add(-15 * time.Minute),
			Source:       "collab_002",
			Level:        "warning",
			Message:      "Connection timeout during model update",
			Data: map[string]interface{}{
				"timeout_ms":  5000,
				"retry_count": 1,
			},
		},
		{
			FederationID: "fed_demo_001",
			Type:         monitoring.MetricTypeAggregation,
			Timestamp:    time.Now().Add(-20 * time.Minute),
			Source:       "aggregator",
			Level:        "info",
			Message:      "Model aggregation completed",
			Data: map[string]interface{}{
				"algorithm":   "fedavg",
				"updates":     3,
				"convergence": 0.15,
			},
		},
	}

	for _, event := range events {
		if err := storage.RecordEvent(ctx, event); err != nil {
			return err
		}
	}

	log.Println("Sample monitoring data created successfully")
	return nil
}
