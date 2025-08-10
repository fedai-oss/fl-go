package aggregator

import (
	"fmt"
	"math"
	"time"

	"github.com/ishaileshpant/fl-go/pkg/federation"
)

// AggregationAlgorithm defines the interface for all aggregation algorithms
type AggregationAlgorithm interface {
	// Initialize sets up the algorithm with configuration
	Initialize(config AlgorithmConfig) error

	// Aggregate performs the aggregation of client updates
	Aggregate(updates []ClientUpdate, globalModel []float32) ([]float32, error)

	// GetName returns the algorithm name
	GetName() string

	// GetHyperparameters returns algorithm-specific hyperparameters
	GetHyperparameters() map[string]interface{}

	// UpdateHyperparameters allows dynamic updates to hyperparameters
	UpdateHyperparameters(params map[string]interface{}) error
}

// ClientUpdate represents an update from a collaborator
type ClientUpdate struct {
	CollaboratorID string
	Weights        []float32
	Timestamp      time.Time
	Round          int
	Staleness      int
	NumSamples     int     // Number of training samples (for weighted aggregation)
	LearningRate   float32 // Client learning rate (for adaptive algorithms)
}

// AlgorithmConfig contains configuration for aggregation algorithms
type AlgorithmConfig struct {
	AlgorithmName   string                 `yaml:"algorithm"` // fedavg, fedopt, fedprox
	ModelSize       int                    `yaml:"model_size"`
	Hyperparameters map[string]interface{} `yaml:"hyperparameters"`
	Mode            federation.FLMode      `yaml:"mode"` // sync or async
}

// AlgorithmType represents supported aggregation algorithms
type AlgorithmType string

const (
	FedAvg  AlgorithmType = "fedavg"
	FedOpt  AlgorithmType = "fedopt"
	FedProx AlgorithmType = "fedprox"
)

// CreateAggregationAlgorithm creates an instance of the specified algorithm
func CreateAggregationAlgorithm(algType AlgorithmType) (AggregationAlgorithm, error) {
	switch algType {
	case FedAvg:
		return &FedAvgAlgorithm{}, nil
	case FedOpt:
		return &FedOptAlgorithm{}, nil
	case FedProx:
		return &FedProxAlgorithm{}, nil
	default:
		return nil, fmt.Errorf("unsupported aggregation algorithm: %s", algType)
	}
}

// =============================================================================
// FedAvg Algorithm (Vanilla Federated Averaging)
// =============================================================================

type FedAvgAlgorithm struct {
	name      string
	modelSize int
}

func (f *FedAvgAlgorithm) Initialize(config AlgorithmConfig) error {
	f.name = "FedAvg"
	f.modelSize = config.ModelSize
	return nil
}

func (f *FedAvgAlgorithm) GetName() string {
	return f.name
}

func (f *FedAvgAlgorithm) GetHyperparameters() map[string]interface{} {
	return map[string]interface{}{
		"algorithm":   "fedavg",
		"description": "Vanilla Federated Averaging",
	}
}

func (f *FedAvgAlgorithm) UpdateHyperparameters(params map[string]interface{}) error {
	// FedAvg has no hyperparameters to update
	return nil
}

func (f *FedAvgAlgorithm) Aggregate(updates []ClientUpdate, globalModel []float32) ([]float32, error) {
	if len(updates) == 0 {
		return globalModel, fmt.Errorf("no updates to aggregate")
	}

	// Simple averaging
	aggregated := make([]float32, f.modelSize)
	totalSamples := 0

	// Calculate total samples for weighted averaging
	for _, update := range updates {
		totalSamples += update.NumSamples
	}

	// Weighted aggregation based on number of samples
	for _, update := range updates {
		weight := float32(update.NumSamples) / float32(totalSamples)
		if totalSamples == 0 {
			weight = 1.0 / float32(len(updates)) // Equal weighting if no sample info
		}

		for i, v := range update.Weights {
			if i < len(aggregated) {
				aggregated[i] += weight * v
			}
		}
	}

	return aggregated, nil
}

// =============================================================================
// FedOpt Algorithm (Adaptive Server Optimization)
// Reference: "Adaptive Federated Optimization" (Reddi et al., 2020)
// =============================================================================

type FedOptAlgorithm struct {
	name      string
	modelSize int
	serverLR  float32
	beta1     float32
	beta2     float32
	epsilon   float32
	momentum  []float32 // First moment estimate
	velocity  []float32 // Second moment estimate
	round     int
}

func (f *FedOptAlgorithm) Initialize(config AlgorithmConfig) error {
	f.name = "FedOpt"
	f.modelSize = config.ModelSize

	// Default hyperparameters
	f.serverLR = 1.0
	f.beta1 = 0.9
	f.beta2 = 0.999
	f.epsilon = 1e-7
	f.round = 0

	// Initialize server optimizer state
	f.momentum = make([]float32, f.modelSize)
	f.velocity = make([]float32, f.modelSize)

	// Override with custom hyperparameters if provided
	if params := config.Hyperparameters; params != nil {
		if lr, ok := params["server_learning_rate"].(float64); ok {
			f.serverLR = float32(lr)
		}
		if beta1, ok := params["beta1"].(float64); ok {
			f.beta1 = float32(beta1)
		}
		if beta2, ok := params["beta2"].(float64); ok {
			f.beta2 = float32(beta2)
		}
		if eps, ok := params["epsilon"].(float64); ok {
			f.epsilon = float32(eps)
		}
	}

	return nil
}

func (f *FedOptAlgorithm) GetName() string {
	return f.name
}

func (f *FedOptAlgorithm) GetHyperparameters() map[string]interface{} {
	return map[string]interface{}{
		"algorithm":            "fedopt",
		"server_learning_rate": f.serverLR,
		"beta1":                f.beta1,
		"beta2":                f.beta2,
		"epsilon":              f.epsilon,
		"description":          "Adaptive Server Optimization (Adam-like)",
	}
}

func (f *FedOptAlgorithm) UpdateHyperparameters(params map[string]interface{}) error {
	if lr, ok := params["server_learning_rate"].(float64); ok {
		f.serverLR = float32(lr)
	}
	if beta1, ok := params["beta1"].(float64); ok {
		f.beta1 = float32(beta1)
	}
	if beta2, ok := params["beta2"].(float64); ok {
		f.beta2 = float32(beta2)
	}
	if eps, ok := params["epsilon"].(float64); ok {
		f.epsilon = float32(eps)
	}
	return nil
}

func (f *FedOptAlgorithm) Aggregate(updates []ClientUpdate, globalModel []float32) ([]float32, error) {
	if len(updates) == 0 {
		return globalModel, fmt.Errorf("no updates to aggregate")
	}

	f.round++

	// First, compute the pseudo-gradient (difference from global model)
	pseudoGradient := make([]float32, f.modelSize)
	totalSamples := 0

	// Calculate total samples for weighted averaging
	for _, update := range updates {
		totalSamples += update.NumSamples
	}

	// Compute weighted average of client updates
	clientAverage := make([]float32, f.modelSize)
	for _, update := range updates {
		weight := float32(update.NumSamples) / float32(totalSamples)
		if totalSamples == 0 {
			weight = 1.0 / float32(len(updates))
		}

		for i, v := range update.Weights {
			if i < len(clientAverage) {
				clientAverage[i] += weight * v
			}
		}
	}

	// Compute pseudo-gradient: difference between client average and global model
	for i := 0; i < f.modelSize && i < len(globalModel); i++ {
		pseudoGradient[i] = clientAverage[i] - globalModel[i]
	}

	// Apply Adam-like server optimization
	newModel := make([]float32, f.modelSize)
	copy(newModel, globalModel)

	for i := 0; i < f.modelSize; i++ {
		// Update momentum (first moment estimate)
		f.momentum[i] = f.beta1*f.momentum[i] + (1-f.beta1)*pseudoGradient[i]

		// Update velocity (second moment estimate)
		f.velocity[i] = f.beta2*f.velocity[i] + (1-f.beta2)*pseudoGradient[i]*pseudoGradient[i]

		// Bias correction
		momentumCorrected := f.momentum[i] / (1 - float32(math.Pow(float64(f.beta1), float64(f.round))))
		velocityCorrected := f.velocity[i] / (1 - float32(math.Pow(float64(f.beta2), float64(f.round))))

		// Apply Adam update
		if i < len(newModel) {
			newModel[i] += f.serverLR * momentumCorrected / (float32(math.Sqrt(float64(velocityCorrected))) + f.epsilon)
		}
	}

	return newModel, nil
}

// =============================================================================
// FedProx Algorithm (Federated Optimization with Proximal Term)
// Reference: "Federated Optimization in Heterogeneous Networks" (Li et al., 2020)
// =============================================================================

type FedProxAlgorithm struct {
	name      string
	modelSize int
	mu        float32 // Proximal term coefficient
}

func (f *FedProxAlgorithm) Initialize(config AlgorithmConfig) error {
	f.name = "FedProx"
	f.modelSize = config.ModelSize
	f.mu = 0.01 // Default proximal term

	// Override with custom hyperparameters if provided
	if params := config.Hyperparameters; params != nil {
		if mu, ok := params["mu"].(float64); ok {
			f.mu = float32(mu)
		}
	}

	return nil
}

func (f *FedProxAlgorithm) GetName() string {
	return f.name
}

func (f *FedProxAlgorithm) GetHyperparameters() map[string]interface{} {
	return map[string]interface{}{
		"algorithm":   "fedprox",
		"mu":          f.mu,
		"description": "Federated Optimization with Proximal Term",
	}
}

func (f *FedProxAlgorithm) UpdateHyperparameters(params map[string]interface{}) error {
	if mu, ok := params["mu"].(float64); ok {
		f.mu = float32(mu)
	}
	return nil
}

func (f *FedProxAlgorithm) Aggregate(updates []ClientUpdate, globalModel []float32) ([]float32, error) {
	if len(updates) == 0 {
		return globalModel, fmt.Errorf("no updates to aggregate")
	}

	// FedProx performs weighted aggregation with consideration for client heterogeneity
	aggregated := make([]float32, f.modelSize)
	totalWeight := float32(0)

	// Calculate weights based on number of samples and learning rates
	for _, update := range updates {
		// Weight based on samples and inverse of learning rate (more stable clients get higher weight)
		weight := float32(update.NumSamples)
		if update.LearningRate > 0 {
			// Clients with smaller learning rates (more conservative) get slightly higher weight
			weight *= (1.0 + f.mu/update.LearningRate)
		}
		totalWeight += weight

		for i, v := range update.Weights {
			if i < len(aggregated) {
				aggregated[i] += weight * v
			}
		}
	}

	// Normalize by total weight
	if totalWeight > 0 {
		for i := range aggregated {
			aggregated[i] /= totalWeight
		}
	}

	// Apply proximal term: blend with global model to ensure stability
	proximalBlend := make([]float32, f.modelSize)
	for i := 0; i < f.modelSize && i < len(globalModel); i++ {
		// Proximal update: new_model = (1-α) * aggregated + α * global_model
		// where α is determined by the proximal term mu
		alpha := f.mu / (1.0 + f.mu)
		proximalBlend[i] = (1-alpha)*aggregated[i] + alpha*globalModel[i]
	}

	return proximalBlend, nil
}