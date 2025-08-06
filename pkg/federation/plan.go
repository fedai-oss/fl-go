package federation

// FLPlan is the federated learning configuration.
type FLPlan struct {
	Rounds        int             `yaml:"rounds"`
	Collaborators []Collaborator  `yaml:"collaborators"`
	Aggregator    AggregatorEntry `yaml:"aggregator"`
	InitialModel  string          `yaml:"initial_model"`
	OutputModel   string          `yaml:"output_model"`
	Tasks         TasksConfig     `yaml:"tasks"`
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
