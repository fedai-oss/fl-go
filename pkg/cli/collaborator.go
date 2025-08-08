package cli

import (
	"fmt"
	"os"
	"time"

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
	fmt.Printf("   Aggregator: %s\n", plan.Aggregator.Address)
	fmt.Printf("   Training Script: %s\n", plan.Tasks.Train.Script)
	fmt.Printf("   Epochs: %v\n", plan.Tasks.Train.Args["epochs"])
	fmt.Printf("   Batch Size: %v\n", plan.Tasks.Train.Args["batch_size"])

	collab := collaborator.NewCollaborator(plan, collaboratorName)

	fmt.Printf("\n🔗 Connecting to aggregator...\n")
	if err := collab.Connect(); err != nil {
		return fmt.Errorf("failed to connect to aggregator: %v", err)
	}

	fmt.Printf("✅ Connected successfully!\n")
	fmt.Printf("🎯 Starting federated learning rounds...\n\n")

	// Start federated learning rounds
	for round := 1; round <= plan.Rounds; round++ {
		fmt.Printf("📍 Round %d/%d\n", round, plan.Rounds)

		// Run training task
		fmt.Printf("🔄 Running training task...\n")
		weights, err := collab.RunTrainTask(plan.Tasks.Train)
		if err != nil {
			return fmt.Errorf("training failed in round %d: %v", round, err)
		}

		// Submit update to aggregator
		fmt.Printf("📤 Submitting model update...\n")
		if err := collab.SubmitUpdate(weights); err != nil {
			return fmt.Errorf("failed to submit update in round %d: %v", round, err)
		}

		fmt.Printf("✅ Round %d completed\n", round)

		// Wait a bit before next round (if not last round)
		if round < plan.Rounds {
			fmt.Printf("⏳ Waiting for next round...\n\n")
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("\n🎉 Federated learning completed!\n")
	fmt.Printf("📊 Collaborator '%s' participated in %d rounds\n", collaboratorName, plan.Rounds)

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
