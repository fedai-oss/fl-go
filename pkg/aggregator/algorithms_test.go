package aggregator

import (
	"testing"
	"time"

	"github.com/ishaileshpant/fl-go/pkg/federation"
)

func TestCreateAggregationAlgorithm(t *testing.T) {
	tests := []struct {
		name    string
		algType AlgorithmType
		wantErr bool
	}{
		{"FedAvg", FedAvg, false},
		{"FedOpt", FedOpt, false},
		{"FedProx", FedProx, false},
		{"Invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alg, err := CreateAggregationAlgorithm(tt.algType)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAggregationAlgorithm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && alg == nil {
				t.Errorf("Expected algorithm instance, got nil")
			}
		})
	}
}

func TestFedAvgAlgorithm(t *testing.T) {
	alg := &FedAvgAlgorithm{}
	config := AlgorithmConfig{
		AlgorithmName: "fedavg",
		ModelSize:     10,
		Mode:         federation.ModeSync,
	}

	err := alg.Initialize(config)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	if alg.GetName() != "FedAvg" {
		t.Errorf("GetName() = %v, want FedAvg", alg.GetName())
	}

	// Test aggregation
	updates := []ClientUpdate{
		{
			CollaboratorID: "client1",
			Weights:        []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
			NumSamples:     100,
		},
		{
			CollaboratorID: "client2",
			Weights:        []float32{2.0, 4.0, 6.0, 8.0, 10.0, 12.0, 14.0, 16.0, 18.0, 20.0},
			NumSamples:     100,
		},
	}

	globalModel := make([]float32, 10)
	result, err := alg.Aggregate(updates, globalModel)
	if err != nil {
		t.Errorf("Aggregate() error = %v", err)
	}

	// Expected result: average of the two updates
	expected := []float32{1.5, 3.0, 4.5, 6.0, 7.5, 9.0, 10.5, 12.0, 13.5, 15.0}
	for i, v := range result {
		if i < len(expected) && v != expected[i] {
			t.Errorf("Aggregate() result[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestFedOptAlgorithm(t *testing.T) {
	alg := &FedOptAlgorithm{}
	config := AlgorithmConfig{
		AlgorithmName: "fedopt",
		ModelSize:     5,
		Hyperparameters: map[string]interface{}{
			"server_learning_rate": 1.0,
			"beta1":               0.9,
			"beta2":               0.999,
			"epsilon":             1e-7,
		},
		Mode: federation.ModeSync,
	}

	err := alg.Initialize(config)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	if alg.GetName() != "FedOpt" {
		t.Errorf("GetName() = %v, want FedOpt", alg.GetName())
	}

	// Test hyperparameters
	params := alg.GetHyperparameters()
	if params["server_learning_rate"] != float32(1.0) {
		t.Errorf("Expected server_learning_rate = 1.0, got %v", params["server_learning_rate"])
	}

	// Test aggregation with consistent updates
	updates := []ClientUpdate{
		{
			CollaboratorID: "client1",
			Weights:        []float32{1.0, 1.0, 1.0, 1.0, 1.0},
			NumSamples:     50,
		},
		{
			CollaboratorID: "client2",
			Weights:        []float32{1.0, 1.0, 1.0, 1.0, 1.0},
			NumSamples:     50,
		},
	}

	globalModel := []float32{0.0, 0.0, 0.0, 0.0, 0.0}
	result, err := alg.Aggregate(updates, globalModel)
	if err != nil {
		t.Errorf("Aggregate() error = %v", err)
	}

	// Result should be different from simple averaging due to Adam-like optimization
	if len(result) != 5 {
		t.Errorf("Expected result length 5, got %d", len(result))
	}

	// FedOpt should move the model in the direction of the pseudo-gradient
	for i, v := range result {
		if v <= globalModel[i] {
			t.Errorf("Expected FedOpt to improve model at index %d: %v <= %v", i, v, globalModel[i])
		}
	}
}

func TestFedProxAlgorithm(t *testing.T) {
	alg := &FedProxAlgorithm{}
	config := AlgorithmConfig{
		AlgorithmName: "fedprox",
		ModelSize:     5,
		Hyperparameters: map[string]interface{}{
			"mu": 0.1,
		},
		Mode: federation.ModeSync,
	}

	err := alg.Initialize(config)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	if alg.GetName() != "FedProx" {
		t.Errorf("GetName() = %v, want FedProx", alg.GetName())
	}

	// Test hyperparameter update
	newParams := map[string]interface{}{
		"mu": 0.05,
	}
	err = alg.UpdateHyperparameters(newParams)
	if err != nil {
		t.Errorf("UpdateHyperparameters() error = %v", err)
	}

	params := alg.GetHyperparameters()
	if params["mu"] != float32(0.05) {
		t.Errorf("Expected mu = 0.05 after update, got %v", params["mu"])
	}

	// Test aggregation with different learning rates
	updates := []ClientUpdate{
		{
			CollaboratorID: "client1",
			Weights:        []float32{2.0, 2.0, 2.0, 2.0, 2.0},
			NumSamples:     100,
			LearningRate:   0.01, // Conservative client
		},
		{
			CollaboratorID: "client2",
			Weights:        []float32{4.0, 4.0, 4.0, 4.0, 4.0},
			NumSamples:     50,
			LearningRate:   0.1, // Aggressive client
		},
	}

	globalModel := []float32{1.0, 1.0, 1.0, 1.0, 1.0}
	result, err := alg.Aggregate(updates, globalModel)
	if err != nil {
		t.Errorf("Aggregate() error = %v", err)
	}

	// Result should be between the updates and the global model due to proximal term
	for i, v := range result {
		if v <= globalModel[i] || v >= 4.0 {
			t.Errorf("Expected FedProx result to be between global model and max update at index %d: got %v", i, v)
		}
	}
}

func TestAlgorithmWithEmptyUpdates(t *testing.T) {
	algorithms := []AggregationAlgorithm{
		&FedAvgAlgorithm{},
		&FedOptAlgorithm{},
		&FedProxAlgorithm{},
	}

	config := AlgorithmConfig{
		ModelSize: 5,
		Mode:     federation.ModeSync,
	}

	for _, alg := range algorithms {
		t.Run(alg.GetName(), func(t *testing.T) {
			err := alg.Initialize(config)
			if err != nil {
				t.Errorf("Initialize() error = %v", err)
			}

			globalModel := []float32{1.0, 1.0, 1.0, 1.0, 1.0}
			updates := []ClientUpdate{} // Empty updates

			_, err = alg.Aggregate(updates, globalModel)
			if err == nil {
				t.Errorf("Expected error for empty updates, got nil")
			}
		})
	}
}

func TestClientUpdateStructure(t *testing.T) {
	update := ClientUpdate{
		CollaboratorID: "test_client",
		Weights:        []float32{1.0, 2.0, 3.0},
		Timestamp:      time.Now(),
		Round:          1,
		Staleness:      0,
		NumSamples:     100,
		LearningRate:   0.01,
	}

	if update.CollaboratorID != "test_client" {
		t.Errorf("Expected CollaboratorID = test_client, got %v", update.CollaboratorID)
	}

	if len(update.Weights) != 3 {
		t.Errorf("Expected 3 weights, got %d", len(update.Weights))
	}

	if update.NumSamples != 100 {
		t.Errorf("Expected NumSamples = 100, got %d", update.NumSamples)
	}

	if update.LearningRate != 0.01 {
		t.Errorf("Expected LearningRate = 0.01, got %v", update.LearningRate)
	}
}

func TestAlgorithmConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  AlgorithmConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: AlgorithmConfig{
				AlgorithmName: "fedavg",
				ModelSize:     10,
				Mode:         federation.ModeSync,
			},
			wantErr: false,
		},
		{
			name: "Zero model size",
			config: AlgorithmConfig{
				AlgorithmName: "fedavg",
				ModelSize:     0,
				Mode:         federation.ModeSync,
			},
			wantErr: false, // Should be handled gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alg := &FedAvgAlgorithm{}
			err := alg.Initialize(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
