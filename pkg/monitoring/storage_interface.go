package monitoring

import (
	"time"
)

// Storage defines the interface for different storage backends
type Storage interface {
	// Federation operations
	StoreFederationMetrics(federation FederationMetrics) error
	GetFederationMetrics(id string) (*FederationMetrics, error)
	ListFederations(activeOnly bool) ([]FederationMetrics, error)

	// Collaborator operations
	StoreCollaboratorMetrics(collaborator CollaboratorMetrics) error
	GetCollaboratorMetrics(federationID string) ([]CollaboratorMetrics, error)

	// Round operations
	StoreRoundMetrics(round RoundMetrics) error
	GetRoundMetrics(federationID string, limit int) ([]RoundMetrics, error)

	// Resource metrics operations
	StoreResourceMetrics(metrics ResourceMetrics) error

	// Event operations
	StoreEvent(event MonitoringEvent) error
	GetEvents(federationID string, limit int, offset int) ([]MonitoringEvent, error)

	// Cleanup operations
	Cleanup(maxAge time.Duration) error
	Close() error
}

// StorageConfig represents configuration for different storage backends
type StorageConfig struct {
	Backend    string         `yaml:"backend"` // memory, postgres, redis
	Memory     MemoryConfig   `yaml:"memory"`
	PostgreSQL DatabaseConfig `yaml:"postgresql"`
	Redis      RedisConfig    `yaml:"redis"`
}

// MemoryConfig represents configuration for in-memory storage
type MemoryConfig struct {
	MaxEntries int `yaml:"max_entries"`
}

// NewStorage creates a new storage backend based on configuration
func NewStorage(config StorageConfig) (Storage, error) {
	switch config.Backend {
	case "memory":
		return NewMemoryStorageBackend(config.Memory), nil
	case "postgres", "postgresql":
		return NewPostgreSQLStorage(config.PostgreSQL)
	case "redis":
		return NewRedisStorage(config.Redis)
	default:
		// Default to memory storage
		return NewMemoryStorageBackend(config.Memory), nil
	}
}
