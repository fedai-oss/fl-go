package main

import (
	"context"
	"github.com/ishaileshpant/openfl-go/pkg/aggregator"
	"github.com/ishaileshpant/openfl-go/pkg/federation"
	"log"
)

func main() {
	plan, err := federation.LoadPlan("plans/example_plan.yaml")
	if err != nil {
		log.Fatal(err)
	}
	agg := aggregator.NewFedAvgAggregator(plan)
	agg.Start(context.Background())
}
