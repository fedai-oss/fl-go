package federation

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// LoadPlan loads a YAML file into FLPlan.
func LoadPlan(path string) (*FLPlan, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var plan FLPlan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}
