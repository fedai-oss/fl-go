package collaborator

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	pb "github.com/ishaileshpant/openfl-go/api/federation"
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
	conn, err := grpc.Dial(c.plan.Aggregator.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.cli = pb.NewFederatedLearningClient(conn)
	resp, err := c.cli.JoinFederation(context.Background(), &pb.JoinRequest{CollaboratorId: c.id})
	if err != nil {
		return err
	}
	return ioutil.WriteFile("models/model_init.pt", resp.InitialModel, 0644)
}

func (c *SimpleCollaborator) RunTrainTask(task federation.TaskConfig) ([]byte, error) {
	args := []string{task.Script, "--model-in", "models/model_init.pt", "--model-out", "models/update.pt"}
	for k, v := range task.Args {
		args = append(args, fmt.Sprintf("--%s", k), fmt.Sprint(v))
	}
	cmd := exec.Command("python3", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return ioutil.ReadFile("models/update.pt")
}

func (c *SimpleCollaborator) SubmitUpdate(weights []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := c.cli.SubmitUpdate(ctx, &pb.ModelUpdate{CollaboratorId: c.id, ModelWeights: weights})
	return err
}
