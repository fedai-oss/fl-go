package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/ishaileshpant/openfl-go/pkg/federation"
)

// HandlePlanCommand handles all plan-related commands
func HandlePlanCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("plan command requires a subcommand (init, validate, etc.)")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "init":
		return handlePlanInit(subArgs)
	case "validate":
		return handlePlanValidate(subArgs)
	case "show":
		return handlePlanShow(subArgs)
	case "--help", "-h":
		printPlanUsage()
		return nil
	default:
		return fmt.Errorf("unknown plan subcommand: %s", subcommand)
	}
}

func handlePlanInit(args []string) error {
	// Parse flags
	planName := "fl_workspace"
	templateType := "basic"

	for i, arg := range args {
		switch arg {
		case "--name", "-n":
			if i+1 < len(args) {
				planName = args[i+1]
			}
		case "--template", "-t":
			if i+1 < len(args) {
				templateType = args[i+1]
			}
		}
	}

	fmt.Printf("🔄 Initializing OpenFL workspace: %s\n", planName)

	// Create workspace directory
	if err := os.MkdirAll(planName, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %v", err)
	}

	// Create subdirectories
	dirs := []string{
		"src",
		"data",
		"save",
		"logs",
		"cert",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(planName, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create plan.yaml
	planPath := filepath.Join(planName, "plan.yaml")
	if err := createDefaultPlan(planPath, templateType); err != nil {
		return fmt.Errorf("failed to create plan.yaml: %v", err)
	}

	// Create Python training script
	trainScriptPath := filepath.Join(planName, "src", "taskrunner.py")
	if err := createTrainingScript(trainScriptPath); err != nil {
		return fmt.Errorf("failed to create training script: %v", err)
	}

	// Create initial model using Python
	fmt.Println("🔄 Creating initial model...")
	if err := createInitialModel(planName); err != nil {
		return fmt.Errorf("failed to create initial model: %v", err)
	}

	fmt.Printf("✅ Successfully initialized OpenFL workspace: %s\n", planName)
	fmt.Printf("📁 Workspace structure:\n")
	fmt.Printf("   %s/\n", planName)
	fmt.Printf("   ├── plan.yaml          # Federated learning configuration\n")
	fmt.Printf("   ├── src/\n")
	fmt.Printf("   │   └── taskrunner.py  # Python training script\n")
	fmt.Printf("   ├── data/              # Local datasets\n")
	fmt.Printf("   ├── save/              # Model checkpoints\n")
	fmt.Printf("   └── logs/              # Training logs\n")
	fmt.Printf("\n")
	fmt.Printf("🚀 Next steps:\n")
	fmt.Printf("   1. cd %s\n", planName)
	fmt.Printf("   2. Edit plan.yaml to configure your FL experiment\n")
	fmt.Printf("   3. Add your training data to the data/ directory\n")
	fmt.Printf("   4. Start aggregator: fx aggregator start\n")
	fmt.Printf("   5. Start collaborators: fx collaborator start <name>\n")

	return nil
}

func createDefaultPlan(path string, templateType string) error {
	plan := &federation.FLPlan{
		Rounds: 3,
		Collaborators: []federation.Collaborator{
			{ID: "collaborator1", Address: "localhost:50052"},
			{ID: "collaborator2", Address: "localhost:50053"},
		},
		Aggregator: federation.AggregatorEntry{
			Address: "localhost:50051",
		},
		InitialModel: "save/init_model.pt",
		OutputModel:  "save/final_model.pt",
		Tasks: federation.TasksConfig{
			Train: federation.TaskConfig{
				Script: "src/taskrunner.py",
				Args: map[string]interface{}{
					"epochs":     5,
					"batch_size": 32,
					"lr":         0.001,
					"data_path":  "data",
				},
			},
		},
	}

	return federation.SavePlan(plan, path)
}

func createTrainingScript(path string) error {
	script := `#!/usr/bin/env python3
"""
OpenFL-Go TaskRunner - Python training script for federated learning
This script interfaces with the Go aggregator/collaborator components.
"""
import argparse
import os
import struct
import numpy as np
import sys

def load_model(model_path):
    """Load model weights from binary file."""
    if not os.path.exists(model_path):
        # Create simple initial model
        return np.array([0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0], dtype=np.float32)
    
    with open(model_path, 'rb') as f:
        data = f.read()
    
    # Convert bytes to float32 array
    weights = []
    for i in range(0, len(data), 4):
        weight = struct.unpack('<f', data[i:i+4])[0]
        weights.append(weight)
    
    return np.array(weights, dtype=np.float32)

def save_model(weights, model_path):
    """Save model weights to binary file."""
    os.makedirs(os.path.dirname(model_path), exist_ok=True)
    
    with open(model_path, 'wb') as f:
        for weight in weights:
            f.write(struct.pack('<f', float(weight)))

def train_model(weights, epochs, batch_size, lr, data_path):
    """
    Simulate training process. In a real implementation, this would:
    1. Load local training data from data_path
    2. Train the model for specified epochs
    3. Return updated weights
    """
    print(f"🔄 Training model for {epochs} epochs (batch_size={batch_size}, lr={lr})")
    print(f"📂 Data path: {data_path}")
    print(f"📊 Model size: {len(weights)} parameters")
    
    # Simulate training by adding small random updates
    np.random.seed(42)  # Reproducible for demo
    gradients = np.random.normal(0, 0.01, len(weights))
    updated_weights = weights + lr * gradients
    
    print(f"✅ Training completed")
    return updated_weights

def main():
    parser = argparse.ArgumentParser(description='OpenFL-Go TaskRunner')
    parser.add_argument('--model-in', required=True, help='Input model path')
    parser.add_argument('--model-out', required=True, help='Output model path')
    parser.add_argument('--epochs', type=int, default=5, help='Number of training epochs')
    parser.add_argument('--batch-size', type=int, default=32, help='Batch size')
    parser.add_argument('--lr', type=float, default=0.001, help='Learning rate')
    parser.add_argument('--data-path', default='data', help='Path to training data')
    
    args = parser.parse_args()
    
    try:
        # Load model
        print(f"📖 Loading model from: {args.model_in}")
        weights = load_model(args.model_in)
        
        # Train model
        updated_weights = train_model(
            weights, 
            args.epochs, 
            args.batch_size, 
            args.lr, 
            args.data_path
        )
        
        # Save updated model
        print(f"💾 Saving model to: {args.model_out}")
        save_model(updated_weights, args.model_out)
        
        print(f"🎯 Training completed successfully")
        
    except Exception as e:
        print(f"❌ Training failed: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
`
	return os.WriteFile(path, []byte(script), 0755)
}

func createInitialModel(workspacePath string) error {
	modelPath := filepath.Join(workspacePath, "save", "init_model.pt")

	// Create save directory
	if err := os.MkdirAll(filepath.Dir(modelPath), 0755); err != nil {
		return err
	}

	// Create a simple initial model with 10 float32 parameters
	// This mimics what the Python script would create
	weights := []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}

	// Convert to binary format (little-endian float32)
	buf := make([]byte, len(weights)*4)
	for i, weight := range weights {
		bits := math.Float32bits(weight)
		buf[i*4] = byte(bits)
		buf[i*4+1] = byte(bits >> 8)
		buf[i*4+2] = byte(bits >> 16)
		buf[i*4+3] = byte(bits >> 24)
	}

	return os.WriteFile(modelPath, buf, 0644)
}

func handlePlanValidate(args []string) error {
	planPath := "plan.yaml"
	if len(args) > 0 {
		planPath = args[0]
	}

	plan, err := federation.LoadPlan(planPath)
	if err != nil {
		return fmt.Errorf("failed to load plan: %v", err)
	}

	fmt.Printf("✅ Plan validation successful\n")
	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("   Rounds: %d\n", plan.Rounds)
	fmt.Printf("   Collaborators: %d\n", len(plan.Collaborators))
	fmt.Printf("   Aggregator: %s\n", plan.Aggregator.Address)
	fmt.Printf("   Initial Model: %s\n", plan.InitialModel)
	fmt.Printf("   Output Model: %s\n", plan.OutputModel)

	return nil
}

func handlePlanShow(args []string) error {
	planPath := "plan.yaml"
	if len(args) > 0 {
		planPath = args[0]
	}

	content, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("failed to read plan: %v", err)
	}

	fmt.Print(string(content))
	return nil
}

func printPlanUsage() {
	fmt.Println("Plan command - Manage federated learning plans")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fx plan <subcommand> [options]")
	fmt.Println()
	fmt.Println("Available Subcommands:")
	fmt.Println("  init      Initialize a new FL workspace")
	fmt.Println("  validate  Validate an existing plan")
	fmt.Println("  show      Display plan contents")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fx plan init --name my_experiment    # Create workspace 'my_experiment'")
	fmt.Println("  fx plan validate plan.yaml           # Validate plan.yaml")
	fmt.Println("  fx plan show                          # Show current plan")
}
