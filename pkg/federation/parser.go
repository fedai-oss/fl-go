package federation

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadPlan loads a federated learning plan from a YAML file.
func LoadPlan(path string) (*FLPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var plan FLPlan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// SavePlan saves a federated learning plan to a YAML file.
func SavePlan(plan *FLPlan, path string) error {
	data, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
