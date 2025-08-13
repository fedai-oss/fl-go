package federation

// FLPlan is the federated learning configuration.
type FLPlan struct {
	Rounds        int             `yaml:"rounds"`
	Collaborators []Collaborator  `yaml:"collaborators"`
	Aggregator    AggregatorEntry `yaml:"aggregator"`
	InitialModel  string          `yaml:"initial_model"`
	OutputModel   string          `yaml:"output_model"`
	Tasks         TasksConfig     `yaml:"tasks"`
	// New fields for async FL support
	Mode        FLMode      `yaml:"mode"`         // sync or async
	AsyncConfig AsyncConfig `yaml:"async_config"` // async-specific settings
	// New field for aggregation algorithm support
	Algorithm AlgorithmConfig `yaml:"algorithm"` // aggregation algorithm configuration
	// Monitoring configuration
	Monitoring MonitoringConfig `yaml:"monitoring"` // monitoring configuration
	// Security configuration
	Security SecurityConfig `yaml:"security"` // security configuration
}

type FLMode string

const (
	ModeSync  FLMode = "sync"
	ModeAsync FLMode = "async"
)

type AsyncConfig struct {
	MaxStaleness     int     `yaml:"max_staleness"`     // Maximum staleness allowed for updates
	MinUpdates       int     `yaml:"min_updates"`       // Minimum updates before aggregation
	AggregationDelay int     `yaml:"aggregation_delay"` // Delay in seconds before aggregating
	StalenessWeight  float64 `yaml:"staleness_weight"`  // Weight decay factor for stale updates
}

type Collaborator struct {
	ID      string `yaml:"id"`
	Address string `yaml:"address"`
}

type AggregatorEntry struct {
	Address string `yaml:"address"`
}

type TasksConfig struct {
	Train TaskConfig `yaml:"train"`
}

type TaskConfig struct {
	Script string                 `yaml:"script"`
	Args   map[string]interface{} `yaml:"args"`
}

type AlgorithmConfig struct {
	Name            string                 `yaml:"name"`            // fedavg, fedopt, fedprox
	Hyperparameters map[string]interface{} `yaml:"hyperparameters"` // algorithm-specific parameters
}

// MonitoringConfig contains monitoring configuration for a federation
type MonitoringConfig struct {
	Enabled                bool   `yaml:"enabled"`                  // Enable monitoring for this federation
	MonitoringServerURL    string `yaml:"monitoring_server_url"`    // URL of the monitoring server
	CollectResourceMetrics bool   `yaml:"collect_resource_metrics"` // Collect system resource metrics
	ReportInterval         int    `yaml:"report_interval"`          // Interval in seconds for metric reporting
	EnableRealTimeEvents   bool   `yaml:"enable_realtime_events"`   // Enable real-time event streaming
}

// SecurityConfig contains security configuration for a federation
type SecurityConfig struct {
	TLS TLSConfig `yaml:"tls"` // TLS configuration
}

// TLSConfig represents the TLS configuration for mTLS
type TLSConfig struct {
	Enabled          bool   `yaml:"enabled"`            // Enable TLS/mTLS
	CertPath         string `yaml:"cert_path"`          // Path to server certificate
	KeyPath          string `yaml:"key_path"`           // Path to server private key
	CAPath           string `yaml:"ca_path"`            // Path to CA certificate
	ServerName       string `yaml:"server_name"`        // Server name for certificate validation
	InsecureSkipTLS  bool   `yaml:"insecure_skip_tls"`  // Skip TLS verification (development only)
	AutoGenerateCert bool   `yaml:"auto_generate_cert"` // Auto-generate self-signed certificates
}
