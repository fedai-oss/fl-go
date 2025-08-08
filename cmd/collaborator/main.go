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
	c.Connect()
	update, _ := c.RunTrainTask(pl.Tasks.Train)
	c.SubmitUpdate(update)
}
