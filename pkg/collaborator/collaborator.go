package collaborator

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	pb "github.com/ishaileshpant/openfl-go/api"
	"github.com/ishaileshpant/openfl-go/pkg/federation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SimpleCollaborator struct {
	plan *federation.FLPlan
	id   string
	cli  pb.FederatedLearningClient
}

func NewCollaborator(plan *federation.FLPlan, id string) *SimpleCollaborator {
	return &SimpleCollaborator{plan: plan, id: id}
}

func (c *SimpleCollaborator) Connect() error {
	log.Printf("Connecting to aggregator at %s", c.plan.Aggregator.Address)
	conn, err := grpc.Dial(c.plan.Aggregator.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.cli = pb.NewFederatedLearningClient(conn)
	resp, err := c.cli.JoinFederation(context.Background(), &pb.JoinRequest{CollaboratorId: c.id})
	if err != nil {
		return err
	}

	// Create models directory if it doesn't exist
	if err := os.MkdirAll("models", 0755); err != nil {
		return err
	}

	return os.WriteFile("models/model_init.pt", resp.InitialModel, 0644)
}

func (c *SimpleCollaborator) RunTrainTask(task federation.TaskConfig) ([]byte, error) {
	args := []string{task.Script, "--model-in", "models/model_init.pt", "--model-out", "models/update.pt"}
	for k, v := range task.Args {
		// Convert snake_case to kebab-case for Python argparse
		kebabKey := strings.ReplaceAll(k, "_", "-")
		args = append(args, fmt.Sprintf("--%s", kebabKey), fmt.Sprint(v))
	}

	log.Printf("Running training task: python3 %v", args)
	cmd := exec.Command("python3", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return os.ReadFile("models/update.pt")
}

func (c *SimpleCollaborator) SubmitUpdate(weights []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := c.cli.SubmitUpdate(ctx, &pb.ModelUpdate{CollaboratorId: c.id, ModelWeights: weights})
	return err
}
