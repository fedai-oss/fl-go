package monitoring

import (
	"time"
)

// MetricType represents the type of metric being tracked
type MetricType string

const (
	MetricTypeRound          MetricType = "round"
	MetricTypeCollaborator   MetricType = "collaborator"
	MetricTypeModelUpdate    MetricType = "model_update"
	MetricTypeAggregation    MetricType = "aggregation"
	MetricTypeTraining       MetricType = "training"
	MetricTypeSystemResource MetricType = "system_resource"
	MetricTypePerformance    MetricType = "performance"
)

// FederationStatus represents the current status of a federation
type FederationStatus string

const (
	StatusPending   FederationStatus = "pending"
	StatusRunning   FederationStatus = "running"
	StatusCompleted FederationStatus = "completed"
	StatusFailed    FederationStatus = "failed"
	StatusStopped   FederationStatus = "stopped"
)

// CollaboratorStatus represents the status of a collaborator
type CollaboratorStatus string

const (
	CollabStatusConnected    CollaboratorStatus = "connected"
	CollabStatusDisconnected CollaboratorStatus = "disconnected"
	CollabStatusTraining     CollaboratorStatus = "training"
	CollabStatusIdle         CollaboratorStatus = "idle"
	CollabStatusError        CollaboratorStatus = "error"
)

// FederationMetrics contains overall federation statistics
type FederationMetrics struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Status            FederationStatus `json:"status"`
	Mode              string           `json:"mode"` // sync/async
	Algorithm         string           `json:"algorithm"`
	StartTime         time.Time        `json:"start_time"`
	EndTime           *time.Time       `json:"end_time,omitempty"`
	CurrentRound      int              `json:"current_round"`
	TotalRounds       int              `json:"total_rounds"`
	ActiveCollabs     int              `json:"active_collaborators"`
	TotalCollabs      int              `json:"total_collaborators"`
	ModelSize         int              `json:"model_size"`
	LastUpdate        time.Time        `json:"last_update"`
	AggregatorAddress string           `json:"aggregator_address"`
}

// CollaboratorMetrics contains metrics for a specific collaborator
type CollaboratorMetrics struct {
	ID               string             `json:"id"`
	FederationID     string             `json:"federation_id"`
	Address          string             `json:"address"`
	Status           CollaboratorStatus `json:"status"`
	JoinTime         time.Time          `json:"join_time"`
	LastSeen         time.Time          `json:"last_seen"`
	CurrentRound     int                `json:"current_round"`
	UpdatesSubmitted int                `json:"updates_submitted"`
	TrainingTime     time.Duration      `json:"training_time"`
	AverageLatency   float64            `json:"average_latency_ms"`
	ErrorCount       int                `json:"error_count"`
	LastError        string             `json:"last_error,omitempty"`
	ResourceMetrics  *ResourceMetrics   `json:"resource_metrics,omitempty"`
}

// RoundMetrics contains metrics for a specific training round
type RoundMetrics struct {
	ID               string        `json:"id"`
	FederationID     string        `json:"federation_id"`
	RoundNumber      int           `json:"round_number"`
	Algorithm        string        `json:"algorithm"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          *time.Time    `json:"end_time,omitempty"`
	Duration         time.Duration `json:"duration_ms"`
	ParticipantCount int           `json:"participant_count"`
	UpdatesReceived  int           `json:"updates_received"`
	AggregationTime  time.Duration `json:"aggregation_time_ms"`
	ModelAccuracy    *float64      `json:"model_accuracy,omitempty"`
	ModelLoss        *float64      `json:"model_loss,omitempty"`
	ConvergenceRate  *float64      `json:"convergence_rate,omitempty"`
	Status           string        `json:"status"`
}

// ModelUpdateMetrics contains metrics for model updates
type ModelUpdateMetrics struct {
	ID               string    `json:"id"`
	FederationID     string    `json:"federation_id"`
	CollaboratorID   string    `json:"collaborator_id"`
	RoundNumber      int       `json:"round_number"`
	Timestamp        time.Time `json:"timestamp"`
	UpdateSize       int       `json:"update_size_bytes"`
	ProcessingTime   float64   `json:"processing_time_ms"`
	Staleness        int       `json:"staleness,omitempty"` // for async FL
	Weight           float64   `json:"weight,omitempty"`    // aggregation weight
	QualityScore     *float64  `json:"quality_score,omitempty"`
	CompressionRatio *float64  `json:"compression_ratio,omitempty"`
}

// ResourceMetrics contains system resource usage metrics
type ResourceMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	CPUUsage      float64   `json:"cpu_usage_percent"`
	MemoryUsage   float64   `json:"memory_usage_percent"`
	MemoryUsed    int64     `json:"memory_used_bytes"`
	MemoryTotal   int64     `json:"memory_total_bytes"`
	DiskUsage     float64   `json:"disk_usage_percent"`
	NetworkRxRate float64   `json:"network_rx_rate_mbps"`
	NetworkTxRate float64   `json:"network_tx_rate_mbps"`
	GPUUsage      *float64  `json:"gpu_usage_percent,omitempty"`
	GPUMemory     *float64  `json:"gpu_memory_percent,omitempty"`
}

// AggregationMetrics contains metrics specific to aggregation operations
type AggregationMetrics struct {
	ID                 string        `json:"id"`
	FederationID       string        `json:"federation_id"`
	RoundNumber        int           `json:"round_number"`
	Algorithm          string        `json:"algorithm"`
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	Duration           time.Duration `json:"duration_ms"`
	UpdatesAggregated  int           `json:"updates_aggregated"`
	ModelConvergence   *float64      `json:"model_convergence,omitempty"`
	AggregationQuality *float64      `json:"aggregation_quality,omitempty"`
	ComputationCost    *float64      `json:"computation_cost,omitempty"`
}

// MonitoringEvent represents a real-time event in the FL system
type MonitoringEvent struct {
	ID           string                 `json:"id"`
	FederationID string                 `json:"federation_id"`
	Type         MetricType             `json:"type"`
	Timestamp    time.Time              `json:"timestamp"`
	Source       string                 `json:"source"` // aggregator/collaborator ID
	Level        string                 `json:"level"`  // info/warning/error
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// MonitoringConfig contains configuration for the monitoring system
type MonitoringConfig struct {
	Enabled               bool          `yaml:"enabled" json:"enabled"`
	APIPort               int           `yaml:"api_port" json:"api_port"`
	WebUIPort             int           `yaml:"webui_port" json:"webui_port"`
	MetricsRetention      time.Duration `yaml:"metrics_retention" json:"metrics_retention"`
	CollectionInterval    time.Duration `yaml:"collection_interval" json:"collection_interval"`
	EnableResourceMetrics bool          `yaml:"enable_resource_metrics" json:"enable_resource_metrics"`
	EnableRealTimeEvents  bool          `yaml:"enable_realtime_events" json:"enable_realtime_events"`
	StorageBackend        string        `yaml:"storage_backend" json:"storage_backend"` // memory/sqlite/postgres
	DatabaseURL           string        `yaml:"database_url,omitempty" json:"database_url,omitempty"`
}

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// MetaInfo contains pagination and additional response metadata
type MetaInfo struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// MetricsFilter contains filtering options for metrics queries
type MetricsFilter struct {
	FederationID   string     `json:"federation_id,omitempty"`
	CollaboratorID string     `json:"collaborator_id,omitempty"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	MetricType     MetricType `json:"metric_type,omitempty"`
	RoundNumber    *int       `json:"round_number,omitempty"`
	Status         string     `json:"status,omitempty"`
	Page           int        `json:"page,omitempty"`
	PerPage        int        `json:"per_page,omitempty"`
}

// Dashboard represents a monitoring dashboard configuration
type Dashboard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Widgets     []Widget  `json:"widgets"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"` // chart/table/metric/alert
	Title  string                 `json:"title"`
	Config map[string]interface{} `json:"config"`
	X      int                    `json:"x"`
	Y      int                    `json:"y"`
	Width  int                    `json:"width"`
	Height int                    `json:"height"`
}
