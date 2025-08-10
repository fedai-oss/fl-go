package monitoring

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStorage implements MonitoringService using in-memory storage
type MemoryStorage struct {
	mu              sync.RWMutex
	federations     map[string]*FederationMetrics
	collaborators   map[string]*CollaboratorMetrics
	rounds          map[string]*RoundMetrics
	modelUpdates    []*ModelUpdateMetrics
	aggregations    []*AggregationMetrics
	resourceMetrics map[string][]*ResourceMetrics // key: source (aggregator/collaborator ID)
	events          []*MonitoringEvent
	alerts          []*Alert
	dashboards      map[string]*Dashboard
	subscriptions   map[string]*EventSubscription
	config          *MonitoringConfig
	startTime       time.Time
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage(config *MonitoringConfig) *MemoryStorage {
	return &MemoryStorage{
		federations:     make(map[string]*FederationMetrics),
		collaborators:   make(map[string]*CollaboratorMetrics),
		rounds:          make(map[string]*RoundMetrics),
		modelUpdates:    make([]*ModelUpdateMetrics, 0),
		aggregations:    make([]*AggregationMetrics, 0),
		resourceMetrics: make(map[string][]*ResourceMetrics),
		events:          make([]*MonitoringEvent, 0),
		alerts:          make([]*Alert, 0),
		dashboards:      make(map[string]*Dashboard),
		subscriptions:   make(map[string]*EventSubscription),
		config:          config,
		startTime:       time.Now(),
	}
}

// Federation metrics implementation
func (m *MemoryStorage) RegisterFederation(ctx context.Context, metrics *FederationMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.federations[metrics.ID] = metrics

	// Record event
	event := &MonitoringEvent{
		ID:           uuid.New().String(),
		FederationID: metrics.ID,
		Type:         MetricTypeRound,
		Timestamp:    time.Now(),
		Source:       "aggregator",
		Level:        "info",
		Message:      fmt.Sprintf("Federation %s registered", metrics.Name),
		Data: map[string]interface{}{
			"mode":      metrics.Mode,
			"algorithm": metrics.Algorithm,
			"rounds":    metrics.TotalRounds,
		},
	}
	m.events = append(m.events, event)
	m.notifySubscribers(event)

	return nil
}

func (m *MemoryStorage) UpdateFederation(ctx context.Context, federationID string, metrics *FederationMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.federations[federationID]; !exists {
		return fmt.Errorf("federation %s not found", federationID)
	}

	metrics.ID = federationID
	m.federations[federationID] = metrics

	return nil
}

func (m *MemoryStorage) GetFederation(ctx context.Context, federationID string) (*FederationMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	federation, exists := m.federations[federationID]
	if !exists {
		return nil, fmt.Errorf("federation %s not found", federationID)
	}

	// Return a copy to prevent external modification
	result := *federation
	return &result, nil
}

func (m *MemoryStorage) GetActiveFederations(ctx context.Context) ([]*FederationMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*FederationMetrics
	for _, federation := range m.federations {
		if federation.Status == StatusRunning {
			// Return a copy
			result := *federation
			active = append(active, &result)
		}
	}

	return active, nil
}

func (m *MemoryStorage) GetFederationHistory(ctx context.Context, filter *MetricsFilter) ([]*FederationMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*FederationMetrics
	for _, federation := range m.federations {
		if m.matchesFederationFilter(federation, filter) {
			result := *federation
			results = append(results, &result)
		}
	}

	// Sort by start time
	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.After(results[j].StartTime)
	})

	return m.paginateFederations(results, filter), nil
}

// Collaborator metrics implementation
func (m *MemoryStorage) RegisterCollaborator(ctx context.Context, metrics *CollaboratorMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.collaborators[metrics.ID] = metrics

	// Record event
	event := &MonitoringEvent{
		ID:           uuid.New().String(),
		FederationID: metrics.FederationID,
		Type:         MetricTypeCollaborator,
		Timestamp:    time.Now(),
		Source:       metrics.ID,
		Level:        "info",
		Message:      fmt.Sprintf("Collaborator %s joined federation", metrics.ID),
		Data: map[string]interface{}{
			"address": metrics.Address,
			"status":  metrics.Status,
		},
	}
	m.events = append(m.events, event)
	m.notifySubscribers(event)

	return nil
}

func (m *MemoryStorage) UpdateCollaborator(ctx context.Context, collaboratorID string, metrics *CollaboratorMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.collaborators[collaboratorID]; !exists {
		return fmt.Errorf("collaborator %s not found", collaboratorID)
	}

	metrics.ID = collaboratorID
	m.collaborators[collaboratorID] = metrics

	return nil
}

func (m *MemoryStorage) GetCollaborator(ctx context.Context, collaboratorID string) (*CollaboratorMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	collaborator, exists := m.collaborators[collaboratorID]
	if !exists {
		return nil, fmt.Errorf("collaborator %s not found", collaboratorID)
	}

	result := *collaborator
	return &result, nil
}

func (m *MemoryStorage) GetFederationCollaborators(ctx context.Context, federationID string) ([]*CollaboratorMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var collaborators []*CollaboratorMetrics
	for _, collaborator := range m.collaborators {
		if collaborator.FederationID == federationID {
			result := *collaborator
			collaborators = append(collaborators, &result)
		}
	}

	return collaborators, nil
}

func (m *MemoryStorage) GetCollaboratorHistory(ctx context.Context, filter *MetricsFilter) ([]*CollaboratorMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*CollaboratorMetrics
	for _, collaborator := range m.collaborators {
		if m.matchesCollaboratorFilter(collaborator, filter) {
			result := *collaborator
			results = append(results, &result)
		}
	}

	// Sort by join time
	sort.Slice(results, func(i, j int) bool {
		return results[i].JoinTime.After(results[j].JoinTime)
	})

	return m.paginateCollaborators(results, filter), nil
}

// Round metrics implementation
func (m *MemoryStorage) RecordRoundStart(ctx context.Context, metrics *RoundMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics.ID == "" {
		metrics.ID = uuid.New().String()
	}

	m.rounds[metrics.ID] = metrics

	// Record event
	event := &MonitoringEvent{
		ID:           uuid.New().String(),
		FederationID: metrics.FederationID,
		Type:         MetricTypeRound,
		Timestamp:    time.Now(),
		Source:       "aggregator",
		Level:        "info",
		Message:      fmt.Sprintf("Round %d started", metrics.RoundNumber),
		Data: map[string]interface{}{
			"round_id":     metrics.ID,
			"algorithm":    metrics.Algorithm,
			"participants": metrics.ParticipantCount,
		},
	}
	m.events = append(m.events, event)
	m.notifySubscribers(event)

	return nil
}

func (m *MemoryStorage) RecordRoundEnd(ctx context.Context, roundID string, metrics *RoundMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rounds[roundID]; !exists {
		return fmt.Errorf("round %s not found", roundID)
	}

	metrics.ID = roundID
	m.rounds[roundID] = metrics

	// Record event
	event := &MonitoringEvent{
		ID:           uuid.New().String(),
		FederationID: metrics.FederationID,
		Type:         MetricTypeRound,
		Timestamp:    time.Now(),
		Source:       "aggregator",
		Level:        "info",
		Message:      fmt.Sprintf("Round %d completed", metrics.RoundNumber),
		Data: map[string]interface{}{
			"round_id":     metrics.ID,
			"duration_ms":  metrics.Duration.Milliseconds(),
			"participants": metrics.ParticipantCount,
			"updates":      metrics.UpdatesReceived,
		},
	}
	m.events = append(m.events, event)
	m.notifySubscribers(event)

	return nil
}

func (m *MemoryStorage) GetRound(ctx context.Context, roundID string) (*RoundMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	round, exists := m.rounds[roundID]
	if !exists {
		return nil, fmt.Errorf("round %s not found", roundID)
	}

	result := *round
	return &result, nil
}

func (m *MemoryStorage) GetFederationRounds(ctx context.Context, federationID string) ([]*RoundMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var rounds []*RoundMetrics
	for _, round := range m.rounds {
		if round.FederationID == federationID {
			result := *round
			rounds = append(rounds, &result)
		}
	}

	// Sort by round number
	sort.Slice(rounds, func(i, j int) bool {
		return rounds[i].RoundNumber < rounds[j].RoundNumber
	})

	return rounds, nil
}

func (m *MemoryStorage) GetRoundHistory(ctx context.Context, filter *MetricsFilter) ([]*RoundMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*RoundMetrics
	for _, round := range m.rounds {
		if m.matchesRoundFilter(round, filter) {
			result := *round
			results = append(results, &result)
		}
	}

	// Sort by start time
	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.After(results[j].StartTime)
	})

	return m.paginateRounds(results, filter), nil
}

// Model update metrics implementation
func (m *MemoryStorage) RecordModelUpdate(ctx context.Context, metrics *ModelUpdateMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics.ID == "" {
		metrics.ID = uuid.New().String()
	}

	m.modelUpdates = append(m.modelUpdates, metrics)

	// Record event
	event := &MonitoringEvent{
		ID:           uuid.New().String(),
		FederationID: metrics.FederationID,
		Type:         MetricTypeModelUpdate,
		Timestamp:    time.Now(),
		Source:       metrics.CollaboratorID,
		Level:        "info",
		Message:      fmt.Sprintf("Model update received from %s", metrics.CollaboratorID),
		Data: map[string]interface{}{
			"round":         metrics.RoundNumber,
			"size_bytes":    metrics.UpdateSize,
			"processing_ms": metrics.ProcessingTime,
		},
	}
	m.events = append(m.events, event)
	m.notifySubscribers(event)

	return nil
}

func (m *MemoryStorage) GetModelUpdates(ctx context.Context, filter *MetricsFilter) ([]*ModelUpdateMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*ModelUpdateMetrics
	for _, update := range m.modelUpdates {
		if m.matchesUpdateFilter(update, filter) {
			result := *update
			results = append(results, &result)
		}
	}

	// Sort by timestamp
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return m.paginateUpdates(results, filter), nil
}

func (m *MemoryStorage) GetUpdateStatistics(ctx context.Context, federationID string, roundNumber int) (*UpdateStatistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var updates []*ModelUpdateMetrics
	for _, update := range m.modelUpdates {
		if update.FederationID == federationID && update.RoundNumber == roundNumber {
			updates = append(updates, update)
		}
	}

	if len(updates) == 0 {
		return &UpdateStatistics{}, nil
	}

	// Calculate statistics
	stats := &UpdateStatistics{
		TotalUpdates: len(updates),
	}

	var totalLatency, totalSize, totalQuality, totalCompression float64
	successCount := 0

	for _, update := range updates {
		totalLatency += update.ProcessingTime
		totalSize += float64(update.UpdateSize)
		if update.QualityScore != nil {
			totalQuality += *update.QualityScore
		}
		if update.CompressionRatio != nil {
			totalCompression += *update.CompressionRatio
		}
		successCount++ // Assuming all recorded updates are successful
	}

	stats.AverageLatency = totalLatency / float64(len(updates))
	stats.AverageSize = totalSize / float64(len(updates))
	stats.SuccessRate = float64(successCount) / float64(len(updates)) * 100
	stats.QualityScore = totalQuality / float64(len(updates))
	stats.CompressionRatio = totalCompression / float64(len(updates))

	return stats, nil
}

// Aggregation metrics implementation
func (m *MemoryStorage) RecordAggregation(ctx context.Context, metrics *AggregationMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics.ID == "" {
		metrics.ID = uuid.New().String()
	}

	m.aggregations = append(m.aggregations, metrics)

	// Record event
	event := &MonitoringEvent{
		ID:           uuid.New().String(),
		FederationID: metrics.FederationID,
		Type:         MetricTypeAggregation,
		Timestamp:    time.Now(),
		Source:       "aggregator",
		Level:        "info",
		Message:      fmt.Sprintf("Aggregation completed for round %d", metrics.RoundNumber),
		Data: map[string]interface{}{
			"algorithm":   metrics.Algorithm,
			"duration_ms": metrics.Duration.Milliseconds(),
			"updates":     metrics.UpdatesAggregated,
		},
	}
	m.events = append(m.events, event)
	m.notifySubscribers(event)

	return nil
}

func (m *MemoryStorage) GetAggregations(ctx context.Context, filter *MetricsFilter) ([]*AggregationMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*AggregationMetrics
	for _, aggregation := range m.aggregations {
		if m.matchesAggregationFilter(aggregation, filter) {
			result := *aggregation
			results = append(results, &result)
		}
	}

	// Sort by start time
	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.After(results[j].StartTime)
	})

	return m.paginateAggregations(results, filter), nil
}

func (m *MemoryStorage) GetAggregationStatistics(ctx context.Context, federationID string) (*AggregationStatistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var aggregations []*AggregationMetrics
	for _, agg := range m.aggregations {
		if agg.FederationID == federationID {
			aggregations = append(aggregations, agg)
		}
	}

	if len(aggregations) == 0 {
		return &AggregationStatistics{}, nil
	}

	// Calculate statistics
	stats := &AggregationStatistics{
		TotalAggregations: len(aggregations),
	}

	var totalTime, totalParticipants, totalConvergence, totalQuality float64

	for _, agg := range aggregations {
		totalTime += float64(agg.Duration.Milliseconds())
		totalParticipants += float64(agg.UpdatesAggregated)
		if agg.ModelConvergence != nil {
			totalConvergence += *agg.ModelConvergence
		}
		if agg.AggregationQuality != nil {
			totalQuality += *agg.AggregationQuality
		}
	}

	stats.AverageTime = totalTime / float64(len(aggregations))
	stats.AverageParticipants = totalParticipants / float64(len(aggregations))
	stats.ConvergenceRate = totalConvergence / float64(len(aggregations))
	stats.ModelQuality = totalQuality / float64(len(aggregations))

	return stats, nil
}

// Resource metrics implementation
func (m *MemoryStorage) RecordResourceMetrics(ctx context.Context, source string, metrics *ResourceMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.resourceMetrics[source] == nil {
		m.resourceMetrics[source] = make([]*ResourceMetrics, 0)
	}

	m.resourceMetrics[source] = append(m.resourceMetrics[source], metrics)

	// Keep only recent metrics to prevent memory overflow
	maxMetrics := 1000 // Keep last 1000 metrics per source
	if len(m.resourceMetrics[source]) > maxMetrics {
		m.resourceMetrics[source] = m.resourceMetrics[source][len(m.resourceMetrics[source])-maxMetrics:]
	}

	return nil
}

func (m *MemoryStorage) GetResourceMetrics(ctx context.Context, source string, timeRange time.Duration) ([]*ResourceMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, exists := m.resourceMetrics[source]
	if !exists {
		return []*ResourceMetrics{}, nil
	}

	cutoff := time.Now().Add(-timeRange)
	var results []*ResourceMetrics

	for _, metric := range metrics {
		if metric.Timestamp.After(cutoff) {
			result := *metric
			results = append(results, &result)
		}
	}

	return results, nil
}

func (m *MemoryStorage) GetSystemOverview(ctx context.Context, federationID string) (*SystemOverview, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	federation, exists := m.federations[federationID]
	if !exists {
		return nil, fmt.Errorf("federation %s not found", federationID)
	}

	// Get collaborators for this federation
	var totalCollabs, activeCollabs int
	for _, collab := range m.collaborators {
		if collab.FederationID == federationID {
			totalCollabs++
			if collab.Status == CollabStatusConnected || collab.Status == CollabStatusTraining {
				activeCollabs++
			}
		}
	}

	// Calculate progress
	progress := float64(federation.CurrentRound) / float64(federation.TotalRounds) * 100
	if federation.TotalRounds == 0 {
		progress = 0
	}

	// Get recent events
	var recentEvents []*MonitoringEvent
	eventCount := 0
	for i := len(m.events) - 1; i >= 0 && eventCount < 10; i-- {
		if m.events[i].FederationID == federationID {
			event := *m.events[i]
			recentEvents = append(recentEvents, &event)
			eventCount++
		}
	}

	// Get active alerts
	var activeAlerts []*Alert
	for _, alert := range m.alerts {
		if alert.FederationID == federationID && alert.ResolvedAt == nil {
			alertCopy := *alert
			activeAlerts = append(activeAlerts, &alertCopy)
		}
	}

	overview := &SystemOverview{
		FederationID:        federationID,
		Status:              federation.Status,
		TotalCollaborators:  totalCollabs,
		ActiveCollaborators: activeCollabs,
		CurrentRound:        federation.CurrentRound,
		TotalRounds:         federation.TotalRounds,
		Progress:            progress,
		RecentEvents:        recentEvents,
		Alerts:              activeAlerts,
	}

	return overview, nil
}

// Events and alerts implementation
func (m *MemoryStorage) RecordEvent(ctx context.Context, event *MonitoringEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	m.events = append(m.events, event)
	m.notifySubscribers(event)

	// Keep only recent events to prevent memory overflow
	maxEvents := 10000
	if len(m.events) > maxEvents {
		m.events = m.events[len(m.events)-maxEvents:]
	}

	return nil
}

func (m *MemoryStorage) GetEvents(ctx context.Context, filter *MetricsFilter) ([]*MonitoringEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*MonitoringEvent
	for _, event := range m.events {
		if m.matchesEventFilter(event, filter) {
			result := *event
			results = append(results, &result)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return m.paginateEvents(results, filter), nil
}

func (m *MemoryStorage) GetActiveAlerts(ctx context.Context, federationID string) ([]*Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var alerts []*Alert
	for _, alert := range m.alerts {
		if alert.FederationID == federationID && alert.ResolvedAt == nil {
			result := *alert
			alerts = append(alerts, &result)
		}
	}

	return alerts, nil
}

// Placeholder implementations for advanced analytics
func (m *MemoryStorage) GetPerformanceInsights(ctx context.Context, federationID string) (*PerformanceInsights, error) {
	// This would contain complex analysis logic
	return &PerformanceInsights{
		FederationID:            federationID,
		OverallPerformance:      85.0,
		TrainingEfficiency:      78.5,
		CommunicationEfficiency: 92.3,
		ResourceUtilization:     67.8,
		BottleneckAnalysis:      []string{"Network latency between collaborators", "Heterogeneous compute capabilities"},
		Recommendations:         []string{"Consider increasing batch size", "Implement adaptive learning rates"},
	}, nil
}

func (m *MemoryStorage) GetConvergenceAnalysis(ctx context.Context, federationID string) (*ConvergenceAnalysis, error) {
	// This would analyze model convergence trends
	return &ConvergenceAnalysis{
		FederationID:      federationID,
		ConvergenceRate:   0.15,
		ParticipationRate: 95.0,
		QualityMetrics:    map[string]float64{"accuracy": 0.87, "f1_score": 0.82},
	}, nil
}

func (m *MemoryStorage) GetEfficiencyMetrics(ctx context.Context, federationID string) (*EfficiencyMetrics, error) {
	// This would calculate various efficiency metrics
	return &EfficiencyMetrics{
		FederationID:            federationID,
		ComputationalEfficiency: 78.5,
		CommunicationEfficiency: 85.2,
		ResourceOptimization:    72.1,
	}, nil
}

// Dashboard management
func (m *MemoryStorage) CreateDashboard(ctx context.Context, dashboard *Dashboard) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if dashboard.ID == "" {
		dashboard.ID = uuid.New().String()
	}
	dashboard.CreatedAt = time.Now()
	dashboard.UpdatedAt = time.Now()

	m.dashboards[dashboard.ID] = dashboard
	return nil
}

func (m *MemoryStorage) GetDashboard(ctx context.Context, dashboardID string) (*Dashboard, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dashboard, exists := m.dashboards[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard %s not found", dashboardID)
	}

	result := *dashboard
	return &result, nil
}

func (m *MemoryStorage) ListDashboards(ctx context.Context) ([]*Dashboard, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var dashboards []*Dashboard
	for _, dashboard := range m.dashboards {
		result := *dashboard
		dashboards = append(dashboards, &result)
	}

	// Sort by creation time
	sort.Slice(dashboards, func(i, j int) bool {
		return dashboards[i].CreatedAt.After(dashboards[j].CreatedAt)
	})

	return dashboards, nil
}

func (m *MemoryStorage) UpdateDashboard(ctx context.Context, dashboardID string, dashboard *Dashboard) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.dashboards[dashboardID]; !exists {
		return fmt.Errorf("dashboard %s not found", dashboardID)
	}

	dashboard.ID = dashboardID
	dashboard.UpdatedAt = time.Now()
	m.dashboards[dashboardID] = dashboard

	return nil
}

func (m *MemoryStorage) DeleteDashboard(ctx context.Context, dashboardID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.dashboards[dashboardID]; !exists {
		return fmt.Errorf("dashboard %s not found", dashboardID)
	}

	delete(m.dashboards, dashboardID)
	return nil
}

// Real-time subscriptions
func (m *MemoryStorage) SubscribeToEvents(ctx context.Context, federationID string, eventTypes []MetricType) (<-chan *MonitoringEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscription := &EventSubscription{
		ID:           uuid.New().String(),
		FederationID: federationID,
		EventTypes:   eventTypes,
		Channel:      make(chan *MonitoringEvent, 100), // Buffered channel
		CreatedAt:    time.Now(),
	}

	m.subscriptions[subscription.ID] = subscription
	return subscription.Channel, nil
}

func (m *MemoryStorage) UnsubscribeFromEvents(ctx context.Context, subscriptionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscription, exists := m.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription %s not found", subscriptionID)
	}

	close(subscription.Channel)
	delete(m.subscriptions, subscriptionID)
	return nil
}

// Health and status
func (m *MemoryStorage) HealthCheck(ctx context.Context) error {
	// Simple health check - could be extended with more sophisticated checks
	return nil
}

func (m *MemoryStorage) GetMetricsStats(ctx context.Context) (*MetricsStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeFederations := 0
	for _, federation := range m.federations {
		if federation.Status == StatusRunning {
			activeFederations++
		}
	}

	activeCollaborators := 0
	for _, collaborator := range m.collaborators {
		if collaborator.Status == CollabStatusConnected || collaborator.Status == CollabStatusTraining {
			activeCollaborators++
		}
	}

	stats := &MetricsStats{
		TotalFederations:    len(m.federations),
		ActiveFederations:   activeFederations,
		TotalCollaborators:  len(m.collaborators),
		ActiveCollaborators: activeCollaborators,
		TotalRounds:         len(m.rounds),
		TotalUpdates:        len(m.modelUpdates),
		StorageUsed:         0,          // Would calculate actual memory usage
		LastCleanup:         time.Now(), // Would track last cleanup
		UptimeSeconds:       int64(time.Since(m.startTime).Seconds()),
	}

	return stats, nil
}

// Helper methods for filtering and pagination
func (m *MemoryStorage) matchesFederationFilter(federation *FederationMetrics, filter *MetricsFilter) bool {
	if filter == nil {
		return true
	}

	if filter.FederationID != "" && federation.ID != filter.FederationID {
		return false
	}

	if filter.Status != "" && string(federation.Status) != filter.Status {
		return false
	}

	if filter.StartTime != nil && federation.StartTime.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && federation.StartTime.After(*filter.EndTime) {
		return false
	}

	return true
}

func (m *MemoryStorage) matchesCollaboratorFilter(collaborator *CollaboratorMetrics, filter *MetricsFilter) bool {
	if filter == nil {
		return true
	}

	if filter.FederationID != "" && collaborator.FederationID != filter.FederationID {
		return false
	}

	if filter.CollaboratorID != "" && collaborator.ID != filter.CollaboratorID {
		return false
	}

	if filter.Status != "" && string(collaborator.Status) != filter.Status {
		return false
	}

	if filter.StartTime != nil && collaborator.JoinTime.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && collaborator.JoinTime.After(*filter.EndTime) {
		return false
	}

	return true
}

func (m *MemoryStorage) matchesRoundFilter(round *RoundMetrics, filter *MetricsFilter) bool {
	if filter == nil {
		return true
	}

	if filter.FederationID != "" && round.FederationID != filter.FederationID {
		return false
	}

	if filter.RoundNumber != nil && round.RoundNumber != *filter.RoundNumber {
		return false
	}

	if filter.Status != "" && round.Status != filter.Status {
		return false
	}

	if filter.StartTime != nil && round.StartTime.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && round.StartTime.After(*filter.EndTime) {
		return false
	}

	return true
}

func (m *MemoryStorage) matchesUpdateFilter(update *ModelUpdateMetrics, filter *MetricsFilter) bool {
	if filter == nil {
		return true
	}

	if filter.FederationID != "" && update.FederationID != filter.FederationID {
		return false
	}

	if filter.CollaboratorID != "" && update.CollaboratorID != filter.CollaboratorID {
		return false
	}

	if filter.RoundNumber != nil && update.RoundNumber != *filter.RoundNumber {
		return false
	}

	if filter.StartTime != nil && update.Timestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && update.Timestamp.After(*filter.EndTime) {
		return false
	}

	return true
}

func (m *MemoryStorage) matchesAggregationFilter(aggregation *AggregationMetrics, filter *MetricsFilter) bool {
	if filter == nil {
		return true
	}

	if filter.FederationID != "" && aggregation.FederationID != filter.FederationID {
		return false
	}

	if filter.RoundNumber != nil && aggregation.RoundNumber != *filter.RoundNumber {
		return false
	}

	if filter.StartTime != nil && aggregation.StartTime.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && aggregation.StartTime.After(*filter.EndTime) {
		return false
	}

	return true
}

func (m *MemoryStorage) matchesEventFilter(event *MonitoringEvent, filter *MetricsFilter) bool {
	if filter == nil {
		return true
	}

	if filter.FederationID != "" && event.FederationID != filter.FederationID {
		return false
	}

	if filter.MetricType != "" && event.Type != filter.MetricType {
		return false
	}

	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}

	return true
}

// Pagination helpers
func (m *MemoryStorage) paginateFederations(results []*FederationMetrics, filter *MetricsFilter) []*FederationMetrics {
	if filter == nil || filter.Page <= 0 {
		return results
	}

	page := filter.Page
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 20 // default
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(results) {
		return []*FederationMetrics{}
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

func (m *MemoryStorage) paginateCollaborators(results []*CollaboratorMetrics, filter *MetricsFilter) []*CollaboratorMetrics {
	if filter == nil || filter.Page <= 0 {
		return results
	}

	page := filter.Page
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(results) {
		return []*CollaboratorMetrics{}
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

func (m *MemoryStorage) paginateRounds(results []*RoundMetrics, filter *MetricsFilter) []*RoundMetrics {
	if filter == nil || filter.Page <= 0 {
		return results
	}

	page := filter.Page
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(results) {
		return []*RoundMetrics{}
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

func (m *MemoryStorage) paginateUpdates(results []*ModelUpdateMetrics, filter *MetricsFilter) []*ModelUpdateMetrics {
	if filter == nil || filter.Page <= 0 {
		return results
	}

	page := filter.Page
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 50
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(results) {
		return []*ModelUpdateMetrics{}
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

func (m *MemoryStorage) paginateAggregations(results []*AggregationMetrics, filter *MetricsFilter) []*AggregationMetrics {
	if filter == nil || filter.Page <= 0 {
		return results
	}

	page := filter.Page
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(results) {
		return []*AggregationMetrics{}
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

func (m *MemoryStorage) paginateEvents(results []*MonitoringEvent, filter *MetricsFilter) []*MonitoringEvent {
	if filter == nil || filter.Page <= 0 {
		return results
	}

	page := filter.Page
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 100
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(results) {
		return []*MonitoringEvent{}
	}

	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

// notifySubscribers sends events to all relevant subscribers
func (m *MemoryStorage) notifySubscribers(event *MonitoringEvent) {
	for _, subscription := range m.subscriptions {
		// Check if subscription matches the event
		if subscription.FederationID != "" && subscription.FederationID != event.FederationID {
			continue
		}

		// Check if event type matches subscription
		if len(subscription.EventTypes) > 0 {
			matches := false
			for _, eventType := range subscription.EventTypes {
				if eventType == event.Type {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}

		// Send event to subscriber (non-blocking)
		select {
		case subscription.Channel <- event:
		default:
			// Channel is full, skip this event to prevent blocking
		}
	}
}
