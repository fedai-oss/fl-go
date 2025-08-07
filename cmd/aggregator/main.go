package main

import (
	"context"
	"log"

	"github.com/ishaileshpant/openfl-go/pkg/aggregator"
	"github.com/ishaileshpant/openfl-go/pkg/federation"
)

func main() {
	log.Println("Loading federated learning plan...")
	plan, err := federation.LoadPlan("plans/example_plan.yaml")
	if err != nil {
		log.Fatalf("Failed to load plan: %v", err)
	}

	log.Println("Creating aggregator...")
	agg := aggregator.NewFedAvgAggregator(plan)

	log.Println("Starting aggregator...")
	if err := agg.Start(context.Background()); err != nil {
		log.Fatalf("Aggregator failed: %v", err)
	}

	log.Println("Aggregator completed successfully!")
}
