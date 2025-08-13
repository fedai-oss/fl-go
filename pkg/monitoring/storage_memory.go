package monitoring

import (
	"sort"
	"sync"
	"time"
)

// MemoryStorageBackend implements Storage interface using in-memory storage
type MemoryStorageBackend struct {
	mu              sync.RWMutex
	federations     map[string]FederationMetrics
	collaborators   map[string][]CollaboratorMetrics // federationID -> collaborators
	rounds          map[string][]RoundMetrics        // federationID -> rounds
	resourceMetrics []ResourceMetrics
	events          []MonitoringEvent
	config          MemoryConfig
}

// NewMemoryStorageBackend creates a new in-memory storage backend
func NewMemoryStorageBackend(config MemoryConfig) *MemoryStorageBackend {
	if config.MaxEntries <= 0 {
		config.MaxEntries = 10000 // Default max entries
	}

	return &MemoryStorageBackend{
		federations:   make(map[string]FederationMetrics),
		collaborators: make(map[string][]CollaboratorMetrics),
		rounds:        make(map[string][]RoundMetrics),
		config:        config,
	}
}

// StoreFederationMetrics stores federation metrics in memory
func (m *MemoryStorageBackend) StoreFederationMetrics(federation FederationMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	federation.LastUpdate = time.Now()
	m.federations[federation.ID] = federation
	return nil
}

// GetFederationMetrics retrieves federation metrics from memory
func (m *MemoryStorageBackend) GetFederationMetrics(id string) (*FederationMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if federation, exists := m.federations[id]; exists {
		return &federation, nil
	}
	return nil, nil
}

// ListFederations lists all federations with optional filters
func (m *MemoryStorageBackend) ListFederations(activeOnly bool) ([]FederationMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var federations []FederationMetrics
	for _, federation := range m.federations {
		if activeOnly && federation.Status != "running" {
			continue
		}
		federations = append(federations, federation)
	}

	// Sort by start time (newest first)
	sort.Slice(federations, func(i, j int) bool {
		return federations[i].StartTime.After(federations[j].StartTime)
	})

	return federations, nil
}

// StoreCollaboratorMetrics stores collaborator metrics in memory
func (m *MemoryStorageBackend) StoreCollaboratorMetrics(collaborator CollaboratorMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	collaborators := m.collaborators[collaborator.FederationID]

	// Find existing collaborator or add new one
	found := false
	for i, existing := range collaborators {
		if existing.ID == collaborator.ID {
			collaborators[i] = collaborator
			found = true
			break
		}
	}

	if !found {
		collaborators = append(collaborators, collaborator)
	}

	m.collaborators[collaborator.FederationID] = collaborators
	return nil
}

// GetCollaboratorMetrics retrieves collaborator metrics from memory
func (m *MemoryStorageBackend) GetCollaboratorMetrics(federationID string) ([]CollaboratorMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if collaborators, exists := m.collaborators[federationID]; exists {
		// Create a copy to avoid race conditions
		result := make([]CollaboratorMetrics, len(collaborators))
		copy(result, collaborators)
		return result, nil
	}

	return []CollaboratorMetrics{}, nil
}

// StoreRoundMetrics stores round metrics in memory
func (m *MemoryStorageBackend) StoreRoundMetrics(round RoundMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rounds := m.rounds[round.FederationID]

	// Find existing round or add new one
	found := false
	for i, existing := range rounds {
		if existing.ID == round.ID {
			rounds[i] = round
			found = true
			break
		}
	}

	if !found {
		rounds = append(rounds, round)
	}

	// Sort by round number (descending)
	sort.Slice(rounds, func(i, j int) bool {
		return rounds[i].RoundNumber > rounds[j].RoundNumber
	})

	m.rounds[round.FederationID] = rounds
	return nil
}

// GetRoundMetrics retrieves round metrics from memory
func (m *MemoryStorageBackend) GetRoundMetrics(federationID string, limit int) ([]RoundMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if rounds, exists := m.rounds[federationID]; exists {
		if limit > 0 && len(rounds) > limit {
			result := make([]RoundMetrics, limit)
			copy(result, rounds[:limit])
			return result, nil
		}

		// Create a copy to avoid race conditions
		result := make([]RoundMetrics, len(rounds))
		copy(result, rounds)
		return result, nil
	}

	return []RoundMetrics{}, nil
}

// StoreResourceMetrics stores resource metrics in memory
func (m *MemoryStorageBackend) StoreResourceMetrics(metrics ResourceMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.resourceMetrics = append(m.resourceMetrics, metrics)

	// Cleanup old entries if we exceed max entries
	if len(m.resourceMetrics) > m.config.MaxEntries {
		// Keep only the most recent entries
		keep := m.config.MaxEntries / 2
		m.resourceMetrics = m.resourceMetrics[len(m.resourceMetrics)-keep:]
	}

	return nil
}

// StoreEvent stores monitoring events in memory
func (m *MemoryStorageBackend) StoreEvent(event MonitoringEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, event)

	// Cleanup old entries if we exceed max entries
	if len(m.events) > m.config.MaxEntries {
		// Keep only the most recent entries
		keep := m.config.MaxEntries / 2
		m.events = m.events[len(m.events)-keep:]
	}

	return nil
}

// GetEvents retrieves monitoring events from memory
func (m *MemoryStorageBackend) GetEvents(federationID string, limit int, offset int) ([]MonitoringEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filteredEvents []MonitoringEvent

	// Filter by federation ID if specified
	for i := len(m.events) - 1; i >= 0; i-- { // Reverse order (newest first)
		event := m.events[i]
		if federationID == "" || event.FederationID == federationID {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Apply offset and limit
	start := offset
	if start > len(filteredEvents) {
		return []MonitoringEvent{}, nil
	}

	end := start + limit
	if limit <= 0 || end > len(filteredEvents) {
		end = len(filteredEvents)
	}

	result := make([]MonitoringEvent, end-start)
	copy(result, filteredEvents[start:end])
	return result, nil
}

// Cleanup removes old data from memory
func (m *MemoryStorageBackend) Cleanup(maxAge time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	// Clean up resource metrics
	var newResourceMetrics []ResourceMetrics
	for _, metrics := range m.resourceMetrics {
		if metrics.Timestamp.After(cutoff) {
			newResourceMetrics = append(newResourceMetrics, metrics)
		}
	}
	m.resourceMetrics = newResourceMetrics

	// Clean up events
	var newEvents []MonitoringEvent
	for _, event := range m.events {
		if event.Timestamp.After(cutoff) {
			newEvents = append(newEvents, event)
		}
	}
	m.events = newEvents

	return nil
}

// Close closes the memory storage (no-op for memory backend)
func (m *MemoryStorageBackend) Close() error {
	return nil
}

// GetStats returns memory storage statistics
func (m *MemoryStorageBackend) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"storage_type":        "memory",
		"federations_count":   len(m.federations),
		"total_collaborators": m.getTotalCollaborators(),
		"total_rounds":        m.getTotalRounds(),
		"resource_metrics":    len(m.resourceMetrics),
		"events_count":        len(m.events),
		"max_entries":         m.config.MaxEntries,
	}
}

func (m *MemoryStorageBackend) getTotalCollaborators() int {
	total := 0
	for _, collaborators := range m.collaborators {
		total += len(collaborators)
	}
	return total
}

func (m *MemoryStorageBackend) getTotalRounds() int {
	total := 0
	for _, rounds := range m.rounds {
		total += len(rounds)
	}
	return total
}
