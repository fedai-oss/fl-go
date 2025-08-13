package monitoring

import (
	"testing"
	"time"
)

func TestMemoryStorageBackend(t *testing.T) {
	storage := NewMemoryStorageBackend(MemoryConfig{MaxEntries: 1000})
	testStorageImplementation(t, storage)
}

func testStorageImplementation(t *testing.T, storage Storage) {
	t.Run("Federation operations", func(t *testing.T) {
		// Test storing and retrieving federation metrics
		federation := FederationMetrics{
			ID:            "test-federation",
			Name:          "Test Federation",
			Status:        "running",
			Mode:          "sync",
			Algorithm:     "fedavg",
			CurrentRound:  5,
			TotalRounds:   10,
			ActiveCollabs: 3,
			TotalCollabs:  5,
			StartTime:     time.Now().Add(-time.Hour),
		}

		err := storage.StoreFederationMetrics(federation)
		if err != nil {
			t.Fatalf("Failed to store federation metrics: %v", err)
		}

		retrieved, err := storage.GetFederationMetrics("test-federation")
		if err != nil {
			t.Fatalf("Failed to get federation metrics: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Retrieved federation is nil")
		}

		if retrieved.ID != federation.ID {
			t.Errorf("Federation ID mismatch: got %s, want %s", retrieved.ID, federation.ID)
		}

		// Test listing federations
		federations, err := storage.ListFederations(false)
		if err != nil {
			t.Fatalf("Failed to list federations: %v", err)
		}

		if len(federations) != 1 {
			t.Errorf("Expected 1 federation, got %d", len(federations))
		}

		// Test active filter
		activeFederations, err := storage.ListFederations(true)
		if err != nil {
			t.Fatalf("Failed to list active federations: %v", err)
		}

		if len(activeFederations) != 1 {
			t.Errorf("Expected 1 active federation, got %d", len(activeFederations))
		}
	})

	t.Run("Collaborator operations", func(t *testing.T) {
		collaborator := CollaboratorMetrics{
			ID:               "collab-1",
			FederationID:     "test-federation",
			Status:           "connected",
			Address:          "localhost:50052",
			LastSeen:         time.Now(),
			UpdatesSubmitted: 5,
			ErrorCount:       0,
			TrainingTime:     2500 * time.Millisecond,
			JoinTime:         time.Now().Add(-30 * time.Minute),
		}

		err := storage.StoreCollaboratorMetrics(collaborator)
		if err != nil {
			t.Fatalf("Failed to store collaborator metrics: %v", err)
		}

		collaborators, err := storage.GetCollaboratorMetrics("test-federation")
		if err != nil {
			t.Fatalf("Failed to get collaborator metrics: %v", err)
		}

		if len(collaborators) != 1 {
			t.Errorf("Expected 1 collaborator, got %d", len(collaborators))
		}

		if collaborators[0].ID != collaborator.ID {
			t.Errorf("Collaborator ID mismatch: got %s, want %s", collaborators[0].ID, collaborator.ID)
		}
	})

	t.Run("Round operations", func(t *testing.T) {
		endTime := time.Now().Add(-5 * time.Minute)
		accuracy := 0.85
		loss := 0.25
		convergenceRate := 0.02

		round := RoundMetrics{
			ID:               "round-1",
			FederationID:     "test-federation",
			RoundNumber:      1,
			Algorithm:        "fedavg",
			ParticipantCount: 3,
			StartTime:        time.Now().Add(-10 * time.Minute),
			EndTime:          &endTime,
			Duration:         5 * time.Minute,
			UpdatesReceived:  3,
			ModelAccuracy:    &accuracy,
			ModelLoss:        &loss,
			ConvergenceRate:  &convergenceRate,
		}

		err := storage.StoreRoundMetrics(round)
		if err != nil {
			t.Fatalf("Failed to store round metrics: %v", err)
		}

		rounds, err := storage.GetRoundMetrics("test-federation", 0)
		if err != nil {
			t.Fatalf("Failed to get round metrics: %v", err)
		}

		if len(rounds) != 1 {
			t.Errorf("Expected 1 round, got %d", len(rounds))
		}

		if rounds[0].ID != round.ID {
			t.Errorf("Round ID mismatch: got %s, want %s", rounds[0].ID, round.ID)
		}

		// Test limit
		limitedRounds, err := storage.GetRoundMetrics("test-federation", 1)
		if err != nil {
			t.Fatalf("Failed to get limited round metrics: %v", err)
		}

		if len(limitedRounds) != 1 {
			t.Errorf("Expected 1 round with limit, got %d", len(limitedRounds))
		}
	})

	t.Run("Resource metrics operations", func(t *testing.T) {
		metrics := ResourceMetrics{
			Timestamp:     time.Now(),
			CPUUsage:      75.5,
			MemoryUsage:   60.2,
			DiskUsage:     45.8,
			NetworkRxRate: 1024.0,
			NetworkTxRate: 2048.0,
		}

		err := storage.StoreResourceMetrics(metrics)
		if err != nil {
			t.Fatalf("Failed to store resource metrics: %v", err)
		}
	})

	t.Run("Event operations", func(t *testing.T) {
		event := MonitoringEvent{
			FederationID: "test-federation",
			Type:         MetricTypeRound,
			Message:      "Round 1 started",
			Level:        "info",
			Data: map[string]interface{}{
				"round_number": 1,
				"participants": 3,
			},
			Timestamp: time.Now(),
		}

		err := storage.StoreEvent(event)
		if err != nil {
			t.Fatalf("Failed to store event: %v", err)
		}

		events, err := storage.GetEvents("test-federation", 10, 0)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}

		if events[0].Type != event.Type {
			t.Errorf("Event type mismatch: got %s, want %s", events[0].Type, event.Type)
		}

		// Test pagination
		paginatedEvents, err := storage.GetEvents("", 5, 0)
		if err != nil {
			t.Fatalf("Failed to get paginated events: %v", err)
		}

		if len(paginatedEvents) == 0 {
			t.Error("Expected at least one event in paginated results")
		}
	})

	t.Run("Cleanup operations", func(t *testing.T) {
		// Test cleanup
		err := storage.Cleanup(24 * time.Hour)
		if err != nil {
			t.Fatalf("Failed to cleanup: %v", err)
		}

		// Test close
		err = storage.Close()
		if err != nil {
			t.Fatalf("Failed to close storage: %v", err)
		}
	})
}

func TestStorageFactory(t *testing.T) {
	tests := []struct {
		name     string
		config   StorageConfig
		wantType string
	}{
		{
			name: "memory storage",
			config: StorageConfig{
				Backend: "memory",
				Memory:  MemoryConfig{MaxEntries: 1000},
			},
			wantType: "*monitoring.MemoryStorageBackend",
		},
		{
			name: "default to memory",
			config: StorageConfig{
				Backend: "",
			},
			wantType: "*monitoring.MemoryStorageBackend",
		},
		{
			name: "invalid backend defaults to memory",
			config: StorageConfig{
				Backend: "invalid",
			},
			wantType: "*monitoring.MemoryStorageBackend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewStorage(tt.config)
			if err != nil {
				t.Fatalf("NewStorage() error = %v", err)
			}

			if storage == nil {
				t.Fatal("Storage is nil")
			}

			// Basic functionality test
			federation := FederationMetrics{
				ID:     "test-federation",
				Name:   "Test Federation",
				Status: "running",
			}

			err = storage.StoreFederationMetrics(federation)
			if err != nil {
				t.Errorf("Failed to store federation metrics: %v", err)
			}

			retrieved, err := storage.GetFederationMetrics("test-federation")
			if err != nil {
				t.Errorf("Failed to get federation metrics: %v", err)
			}

			if retrieved == nil {
				t.Error("Retrieved federation is nil")
			}

			storage.Close()
		})
	}
}
