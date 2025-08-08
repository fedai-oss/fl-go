package cli

import (
	"fmt"
	"os"

	"github.com/ishaileshpant/fl-go/pkg/collaborator"
	"github.com/ishaileshpant/fl-go/pkg/federation"
)

// HandleCollaboratorCommand handles all collaborator-related commands
func HandleCollaboratorCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("collaborator command requires a subcommand (start, etc.)")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "start":
		return handleCollaboratorStart(subArgs)
	case "--help", "-h":
		printCollaboratorUsage()
		return nil
	default:
		return fmt.Errorf("unknown collaborator subcommand: %s", subcommand)
	}
}

func handleCollaboratorStart(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("collaborator start requires a collaborator name")
	}

	collaboratorName := args[0]

	// Parse flags
	planPath := "plan.yaml"

	for i, arg := range args[1:] {
		switch arg {
		case "--plan", "-p":
			if i+2 < len(args) {
				planPath = args[i+2]
			}
		}
	}

	// Check if plan exists
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		return fmt.Errorf("plan file not found: %s\nRun 'fx plan init' to create a workspace first", planPath)
	}

	fmt.Printf("📋 Loading federated learning plan: %s\n", planPath)
	plan, err := federation.LoadPlan(planPath)
	if err != nil {
		return fmt.Errorf("failed to load plan: %v", err)
	}

	// Set default mode if not specified
	if plan.Mode == "" {
		plan.Mode = federation.ModeSync
	}

	// Find this collaborator in the plan
	var found bool
	for _, collab := range plan.Collaborators {
		if collab.ID == collaboratorName {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("⚠️  Warning: Collaborator '%s' not found in plan. Available collaborators:\n", collaboratorName)
		for _, collab := range plan.Collaborators {
			fmt.Printf("   - %s\n", collab.ID)
		}
		fmt.Printf("Continuing anyway...\n\n")
	}

	fmt.Printf("🤝 Starting collaborator: %s\n", collaboratorName)
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   Mode: %s\n", plan.Mode)
	fmt.Printf("   Aggregator: %s\n", plan.Aggregator.Address)

	if plan.Mode == federation.ModeSync {
		fmt.Printf("   Rounds: %d\n", plan.Rounds)
	} else {
		fmt.Printf("   Async Config:\n")
		fmt.Printf("     Max Staleness: %d\n", plan.AsyncConfig.MaxStaleness)
		fmt.Printf("     Min Updates: %d\n", plan.AsyncConfig.MinUpdates)
		fmt.Printf("     Aggregation Delay: %ds\n", plan.AsyncConfig.AggregationDelay)
		fmt.Printf("     Staleness Weight: %.3f\n", plan.AsyncConfig.StalenessWeight)
	}

	fmt.Printf("   Training Script: %s\n", plan.Tasks.Train.Script)
	fmt.Printf("   Epochs: %v\n", plan.Tasks.Train.Args["epochs"])
	fmt.Printf("   Batch Size: %v\n", plan.Tasks.Train.Args["batch_size"])

	collab := collaborator.NewCollaborator(plan, collaboratorName)

	fmt.Printf("\n🔗 Connecting to aggregator...\n")
	if err := collab.Connect(); err != nil {
		return fmt.Errorf("failed to connect to aggregator: %v", err)
	}

	fmt.Printf("✅ Connected successfully!\n")
	fmt.Printf("🎯 Starting federated learning...\n\n")

	// Use the new Run method that handles both sync and async modes
	if err := collab.Run(plan.Tasks.Train); err != nil {
		return fmt.Errorf("federated learning failed: %v", err)
	}

	fmt.Printf("\n🎉 Federated learning completed!\n")
	fmt.Printf("📊 Collaborator '%s' completed training in %s mode\n", collaboratorName, plan.Mode)

	return nil
}

func printCollaboratorUsage() {
	fmt.Println("Collaborator command - Start and manage collaborator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fx collaborator <subcommand> [options]")
	fmt.Println()
	fmt.Println("Available Subcommands:")
	fmt.Println("  start     Start a collaborator")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --plan, -p    Path to plan.yaml file (default: plan.yaml)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fx collaborator start collaborator1           # Start collaborator1")
	fmt.Println("  fx collaborator start collab1 --plan my.yaml  # Start with custom plan")
}
