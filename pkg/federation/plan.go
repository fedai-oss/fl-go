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
