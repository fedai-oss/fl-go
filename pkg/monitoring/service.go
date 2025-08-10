package monitoring

import (
	"context"
	"time"
)

// MonitoringService defines the interface for the monitoring system
type MonitoringService interface {
	// Federation metrics
	RegisterFederation(ctx context.Context, metrics *FederationMetrics) error
	UpdateFederation(ctx context.Context, federationID string, metrics *FederationMetrics) error
	GetFederation(ctx context.Context, federationID string) (*FederationMetrics, error)
	GetActiveFederations(ctx context.Context) ([]*FederationMetrics, error)
	GetFederationHistory(ctx context.Context, filter *MetricsFilter) ([]*FederationMetrics, error)

	// Collaborator metrics
	RegisterCollaborator(ctx context.Context, metrics *CollaboratorMetrics) error
	UpdateCollaborator(ctx context.Context, collaboratorID string, metrics *CollaboratorMetrics) error
	GetCollaborator(ctx context.Context, collaboratorID string) (*CollaboratorMetrics, error)
	GetFederationCollaborators(ctx context.Context, federationID string) ([]*CollaboratorMetrics, error)
	GetCollaboratorHistory(ctx context.Context, filter *MetricsFilter) ([]*CollaboratorMetrics, error)

	// Round metrics
	RecordRoundStart(ctx context.Context, metrics *RoundMetrics) error
	RecordRoundEnd(ctx context.Context, roundID string, metrics *RoundMetrics) error
	GetRound(ctx context.Context, roundID string) (*RoundMetrics, error)
	GetFederationRounds(ctx context.Context, federationID string) ([]*RoundMetrics, error)
	GetRoundHistory(ctx context.Context, filter *MetricsFilter) ([]*RoundMetrics, error)

	// Model update metrics
	RecordModelUpdate(ctx context.Context, metrics *ModelUpdateMetrics) error
	GetModelUpdates(ctx context.Context, filter *MetricsFilter) ([]*ModelUpdateMetrics, error)
	GetUpdateStatistics(ctx context.Context, federationID string, roundNumber int) (*UpdateStatistics, error)

	// Aggregation metrics
	RecordAggregation(ctx context.Context, metrics *AggregationMetrics) error
	GetAggregations(ctx context.Context, filter *MetricsFilter) ([]*AggregationMetrics, error)
	GetAggregationStatistics(ctx context.Context, federationID string) (*AggregationStatistics, error)

	// Resource metrics
	RecordResourceMetrics(ctx context.Context, source string, metrics *ResourceMetrics) error
	GetResourceMetrics(ctx context.Context, source string, timeRange time.Duration) ([]*ResourceMetrics, error)
	GetSystemOverview(ctx context.Context, federationID string) (*SystemOverview, error)

	// Events and alerts
	RecordEvent(ctx context.Context, event *MonitoringEvent) error
	GetEvents(ctx context.Context, filter *MetricsFilter) ([]*MonitoringEvent, error)
	GetActiveAlerts(ctx context.Context, federationID string) ([]*Alert, error)

	// Analytics and insights
	GetPerformanceInsights(ctx context.Context, federationID string) (*PerformanceInsights, error)
	GetConvergenceAnalysis(ctx context.Context, federationID string) (*ConvergenceAnalysis, error)
	GetEfficiencyMetrics(ctx context.Context, federationID string) (*EfficiencyMetrics, error)

	// Dashboard management
	CreateDashboard(ctx context.Context, dashboard *Dashboard) error
	GetDashboard(ctx context.Context, dashboardID string) (*Dashboard, error)
	ListDashboards(ctx context.Context) ([]*Dashboard, error)
	UpdateDashboard(ctx context.Context, dashboardID string, dashboard *Dashboard) error
	DeleteDashboard(ctx context.Context, dashboardID string) error

	// Real-time subscriptions
	SubscribeToEvents(ctx context.Context, federationID string, eventTypes []MetricType) (<-chan *MonitoringEvent, error)
	UnsubscribeFromEvents(ctx context.Context, subscriptionID string) error

	// Health and status
	HealthCheck(ctx context.Context) error
	GetMetricsStats(ctx context.Context) (*MetricsStats, error)
}

// Additional types for analytics and insights
type UpdateStatistics struct {
	TotalUpdates     int     `json:"total_updates"`
	AverageLatency   float64 `json:"average_latency_ms"`
	AverageSize      float64 `json:"average_size_bytes"`
	SuccessRate      float64 `json:"success_rate"`
	QualityScore     float64 `json:"quality_score"`
	CompressionRatio float64 `json:"compression_ratio"`
}

type AggregationStatistics struct {
	TotalAggregations   int     `json:"total_aggregations"`
	AverageTime         float64 `json:"average_time_ms"`
	AverageParticipants float64 `json:"average_participants"`
	ConvergenceRate     float64 `json:"convergence_rate"`
	ModelQuality        float64 `json:"model_quality"`
}

type SystemOverview struct {
	FederationID         string             `json:"federation_id"`
	Status               FederationStatus   `json:"status"`
	TotalCollaborators   int                `json:"total_collaborators"`
	ActiveCollaborators  int                `json:"active_collaborators"`
	CurrentRound         int                `json:"current_round"`
	TotalRounds          int                `json:"total_rounds"`
	Progress             float64            `json:"progress_percent"`
	AverageResourceUsage *ResourceMetrics   `json:"average_resource_usage"`
	RecentEvents         []*MonitoringEvent `json:"recent_events"`
	Alerts               []*Alert           `json:"active_alerts"`
}

type Alert struct {
	ID           string                 `json:"id"`
	FederationID string                 `json:"federation_id"`
	Type         string                 `json:"type"`
	Severity     string                 `json:"severity"` // low/medium/high/critical
	Title        string                 `json:"title"`
	Message      string                 `json:"message"`
	Source       string                 `json:"source"`
	CreatedAt    time.Time              `json:"created_at"`
	ResolvedAt   *time.Time             `json:"resolved_at,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

type PerformanceInsights struct {
	FederationID            string     `json:"federation_id"`
	OverallPerformance      float64    `json:"overall_performance_score"`
	TrainingEfficiency      float64    `json:"training_efficiency"`
	CommunicationEfficiency float64    `json:"communication_efficiency"`
	ResourceUtilization     float64    `json:"resource_utilization"`
	BottleneckAnalysis      []string   `json:"bottleneck_analysis"`
	Recommendations         []string   `json:"recommendations"`
	TrendAnalysis           *TrendData `json:"trend_analysis"`
}

type ConvergenceAnalysis struct {
	FederationID        string              `json:"federation_id"`
	ConvergenceRate     float64             `json:"convergence_rate"`
	EstimatedCompletion *time.Time          `json:"estimated_completion,omitempty"`
	ModelAccuracy       []AccuracyDataPoint `json:"model_accuracy_trend"`
	ModelLoss           []LossDataPoint     `json:"model_loss_trend"`
	ParticipationRate   float64             `json:"participation_rate"`
	QualityMetrics      map[string]float64  `json:"quality_metrics"`
}

type EfficiencyMetrics struct {
	FederationID            string         `json:"federation_id"`
	ComputationalEfficiency float64        `json:"computational_efficiency"`
	CommunicationEfficiency float64        `json:"communication_efficiency"`
	EnergyEfficiency        float64        `json:"energy_efficiency,omitempty"`
	CostEfficiency          float64        `json:"cost_efficiency,omitempty"`
	TimeToConvergence       *time.Duration `json:"time_to_convergence,omitempty"`
	ResourceOptimization    float64        `json:"resource_optimization"`
}

type MetricsStats struct {
	TotalFederations    int       `json:"total_federations"`
	ActiveFederations   int       `json:"active_federations"`
	TotalCollaborators  int       `json:"total_collaborators"`
	ActiveCollaborators int       `json:"active_collaborators"`
	TotalRounds         int       `json:"total_rounds"`
	TotalUpdates        int       `json:"total_updates"`
	StorageUsed         int64     `json:"storage_used_bytes"`
	LastCleanup         time.Time `json:"last_cleanup"`
	UptimeSeconds       int64     `json:"uptime_seconds"`
}

type TrendData struct {
	TimeSeries []time.Time `json:"time_series"`
	Values     []float64   `json:"values"`
	Trend      string      `json:"trend"` // increasing/decreasing/stable
	ChangeRate float64     `json:"change_rate_percent"`
}

type AccuracyDataPoint struct {
	Round     int       `json:"round"`
	Timestamp time.Time `json:"timestamp"`
	Accuracy  float64   `json:"accuracy"`
}

type LossDataPoint struct {
	Round     int       `json:"round"`
	Timestamp time.Time `json:"timestamp"`
	Loss      float64   `json:"loss"`
}

// EventSubscription represents a real-time event subscription
type EventSubscription struct {
	ID           string                `json:"id"`
	FederationID string                `json:"federation_id"`
	EventTypes   []MetricType          `json:"event_types"`
	Channel      chan *MonitoringEvent `json:"-"`
	CreatedAt    time.Time             `json:"created_at"`
}
