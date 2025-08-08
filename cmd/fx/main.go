package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ishaileshpant/fl-go/pkg/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "plan":
		if err := cli.HandlePlanCommand(args); err != nil {
			log.Fatalf("Plan command failed: %v", err)
		}
	case "aggregator":
		if err := cli.HandleAggregatorCommand(args); err != nil {
			log.Fatalf("Aggregator command failed: %v", err)
		}
	case "collaborator":
		if err := cli.HandleCollaboratorCommand(args); err != nil {
			log.Fatalf("Collaborator command failed: %v", err)
		}
	case "version":
		fmt.Println("FL-Go v1.0.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FL-Go - A Go implementation of OpenFL")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fx <command> [arguments]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  plan         Manage federated learning plans")
	fmt.Println("  aggregator   Start and manage aggregator")
	fmt.Println("  collaborator Start and manage collaborator")
	fmt.Println("  version      Show version information")
	fmt.Println("  help         Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fx plan init                    # Initialize a new FL workspace")
	fmt.Println("  fx aggregator start             # Start the aggregator")
	fmt.Println("  fx collaborator start collab1  # Start collaborator with ID 'collab1'")
	fmt.Println()
	fmt.Println("For more help on a specific command:")
	fmt.Println("  fx <command> --help")
}
