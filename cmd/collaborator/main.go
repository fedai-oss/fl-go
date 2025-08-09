package main

import (
	"flag"
	"log"

	"github.com/ishaileshpant/fl-go/pkg/collaborator"
	"github.com/ishaileshpant/fl-go/pkg/federation"
)

func main() {
	id := flag.String("id", "collab1", "ID")
	plan := flag.String("plan", "plans/example_plan.yaml", "Plan path")
	flag.Parse()

	pl, err := federation.LoadPlan(*plan)
	if err != nil {
		log.Fatal(err)
	}
	c := collaborator.NewCollaborator(pl, *id)
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect to aggregator: %v", err)
	}

	update, err := c.RunTrainTask(pl.Tasks.Train)
	if err != nil {
		log.Fatalf("Failed to run training task: %v", err)
	}

	if err := c.SubmitUpdate(update); err != nil {
		log.Fatalf("Failed to submit update: %v", err)
	}
}
