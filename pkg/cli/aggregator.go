package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/ishaileshpant/fl-go/pkg/aggregator"
	"github.com/ishaileshpant/fl-go/pkg/federation"
)

// HandleAggregatorCommand handles all aggregator-related commands
func HandleAggregatorCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("aggregator command requires a subcommand (start, stop, etc.)")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "start":
		return handleAggregatorStart(subArgs)
	case "--help", "-h":
		printAggregatorUsage()
		return nil
	default:
		return fmt.Errorf("unknown aggregator subcommand: %s", subcommand)
	}
}

func handleAggregatorStart(args []string) error {
	// Parse flags
	planPath := "plan.yaml"

	for i, arg := range args {
		switch arg {
		case "--plan", "-p":
			if i+1 < len(args) {
				planPath = args[i+1]
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

	fmt.Printf("🚀 Starting aggregator...\n")
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   Mode: %s\n", plan.Mode)
	fmt.Printf("   Address: %s\n", plan.Aggregator.Address)
	
	// Display algorithm information
	algorithmName := "fedavg" // default
	if plan.Algorithm.Name != "" {
		algorithmName = plan.Algorithm.Name
	}
	fmt.Printf("   Algorithm: %s\n", algorithmName)
	
	if len(plan.Algorithm.Hyperparameters) > 0 {
		fmt.Printf("   Algorithm hyperparameters:\n")
		for key, value := range plan.Algorithm.Hyperparameters {
			fmt.Printf("     %s: %v\n", key, value)
		}
	}

	if plan.Mode == federation.ModeSync {
		fmt.Printf("   Rounds: %d\n", plan.Rounds)
	} else {
		fmt.Printf("   Async Config:\n")
		fmt.Printf("     Max Staleness: %d\n", plan.AsyncConfig.MaxStaleness)
		fmt.Printf("     Min Updates: %d\n", plan.AsyncConfig.MinUpdates)
		fmt.Printf("     Aggregation Delay: %ds\n", plan.AsyncConfig.AggregationDelay)
		fmt.Printf("     Staleness Weight: %.3f\n", plan.AsyncConfig.StalenessWeight)
	}

	fmt.Printf("   Collaborators: %d\n", len(plan.Collaborators))
	fmt.Printf("   Initial Model: %s\n", plan.InitialModel)
	fmt.Printf("   Output Model: %s\n", plan.OutputModel)

	agg := aggregator.NewAggregator(plan)

	fmt.Printf("\n🎯 Aggregator ready! Waiting for collaborators to connect...\n")
	fmt.Printf("💡 To start collaborators, run: fx collaborator start <name>\n\n")

	if err := agg.Start(context.Background()); err != nil {
		return fmt.Errorf("aggregator failed: %v", err)
	}

	fmt.Printf("✅ Federated learning completed successfully!\n")
	fmt.Printf("📄 Final model saved to: %s\n", plan.OutputModel)

	return nil
}

func printAggregatorUsage() {
	fmt.Println("Aggregator command - Start and manage aggregator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fx aggregator <subcommand> [options]")
	fmt.Println()
	fmt.Println("Available Subcommands:")
	fmt.Println("  start     Start the aggregator")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --plan, -p    Path to plan.yaml file (default: plan.yaml)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fx aggregator start                    # Start with plan.yaml")
	fmt.Println("  fx aggregator start --plan my_plan.yaml # Start with custom plan")
}
