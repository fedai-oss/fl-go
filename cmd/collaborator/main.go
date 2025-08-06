package main

import (
	"flag"
	"github.com/ishaileshpant/openfl-go/pkg/collaborator"
	"github.com/ishaileshpant/openfl-go/pkg/federation"
	"log"
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
