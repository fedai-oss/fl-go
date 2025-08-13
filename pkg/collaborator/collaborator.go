package collaborator

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	pb "github.com/ishaileshpant/fl-go/api"
	"github.com/ishaileshpant/fl-go/pkg/federation"
	"github.com/ishaileshpant/fl-go/pkg/security"
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

	// Initialize TLS manager for secure communication
	tlsManager, err := security.NewTLSManager(security.TLSConfig(c.plan.Security.TLS), "certs")
	if err != nil {
		return fmt.Errorf("failed to initialize TLS manager: %w", err)
	}

	// Get client dial options with TLS support
	dialOpts, err := tlsManager.NewClientDialOptions()
	if err != nil {
		return fmt.Errorf("failed to get client dial options: %w", err)
	}

	// Fallback to insecure credentials if TLS is not enabled
	if len(dialOpts) == 0 {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	conn, err := grpc.NewClient(c.plan.Aggregator.Address, dialOpts...)
	if err != nil {
		return err
	}
	c.cli = pb.NewFederatedLearningClient(conn)
	resp, err := c.cli.JoinFederation(context.Background(), &pb.JoinRequest{CollaboratorId: c.id})
	if err != nil {
		return err
	}

	// Create models directory if it doesn't exist
	if err := os.MkdirAll("models", 0750); err != nil {
		return err
	}

	return os.WriteFile("models/model_init.pt", resp.InitialModel, 0600)
}

func (c *SimpleCollaborator) RunTrainTask(task federation.TaskConfig) ([]byte, error) {
	args := []string{task.Script, "--model-in", "models/model_init.pt", "--model-out", "models/update.pt"}
	for k, v := range task.Args {
		// Validate key and value to prevent injection
		if !isValidArgument(k) || !isValidArgument(fmt.Sprint(v)) {
			return nil, fmt.Errorf("invalid argument detected: key=%s, value=%v", k, v)
		}

		// Convert snake_case to kebab-case for Python argparse
		kebabKey := strings.ReplaceAll(k, "_", "-")
		args = append(args, fmt.Sprintf("--%s", kebabKey), fmt.Sprint(v))
	}

	log.Printf("Running training task: python3 %v", args)
	cmd := exec.Command("python3", args...) // #nosec G204 - Arguments validated with whitelist above
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

func (c *SimpleCollaborator) GetLatestModel() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := c.cli.GetLatestModel(ctx, &pb.GetModelRequest{CollaboratorId: c.id})
	if err != nil {
		return nil, err
	}
	return resp.ModelWeights, nil
}

// RunSyncMode runs the traditional synchronous FL mode
func (c *SimpleCollaborator) RunSyncMode(task federation.TaskConfig) error {
	log.Printf("Starting SYNC mode training for %d rounds", c.plan.Rounds)

	for round := 1; round <= c.plan.Rounds; round++ {
		log.Printf("Starting round %d/%d", round, c.plan.Rounds)

		// Train on current model
		weights, err := c.RunTrainTask(task)
		if err != nil {
			return fmt.Errorf("training failed in round %d: %v", round, err)
		}

		// Submit update
		if err := c.SubmitUpdate(weights); err != nil {
			return fmt.Errorf("failed to submit update in round %d: %v", round, err)
		}

		log.Printf("Round %d/%d completed", round, c.plan.Rounds)

		// Wait for next round (in sync mode, we wait for all collaborators)
		if round < c.plan.Rounds {
			log.Printf("Waiting for next round...")
			time.Sleep(5 * time.Second)
		}
	}

	log.Printf("SYNC mode training completed")
	return nil
}

// RunAsyncMode runs the asynchronous FL mode based on Papaya paper
func (c *SimpleCollaborator) RunAsyncMode(task federation.TaskConfig) error {
	log.Printf("Starting ASYNC mode training (continuous)")

	round := 1
	for {
		log.Printf("Starting async round %d", round)

		// Train on current model
		weights, err := c.RunTrainTask(task)
		if err != nil {
			return fmt.Errorf("training failed in async round %d: %v", round, err)
		}

		// Submit update immediately
		if err := c.SubmitUpdate(weights); err != nil {
			return fmt.Errorf("failed to submit update in async round %d: %v", round, err)
		}

		log.Printf("Async round %d completed", round)

		// In async mode, get the latest model from aggregator after each round
		log.Printf("Getting latest model from aggregator...")
		latestModel, err := c.GetLatestModel()
		if err != nil {
			log.Printf("Warning: failed to get latest model: %v", err)
		} else {
			// Update the local model with the latest from aggregator
			if err := os.WriteFile("models/model_init.pt", latestModel, 0600); err != nil {
				log.Printf("Warning: failed to save latest model: %v", err)
			} else {
				log.Printf("Updated local model with latest from aggregator")
			}
		}

		// In async mode, we can continue immediately without waiting
		// But we add a small delay to prevent overwhelming the system
		time.Sleep(2 * time.Second)

		round++

		// Optional: add a maximum round limit for async mode
		if c.plan.Rounds > 0 && round > c.plan.Rounds {
			log.Printf("Reached maximum rounds (%d), stopping async training", c.plan.Rounds)
			break
		}
	}

	log.Printf("ASYNC mode training completed")
	return nil
}

// Run starts the collaborator in the appropriate mode
func (c *SimpleCollaborator) Run(task federation.TaskConfig) error {
	// Set default mode if not specified
	if c.plan.Mode == "" {
		c.plan.Mode = federation.ModeSync
	}

	switch c.plan.Mode {
	case federation.ModeAsync:
		return c.RunAsyncMode(task)
	default:
		return c.RunSyncMode(task)
	}
}

// isValidArgument validates command line arguments to prevent injection attacks
func isValidArgument(arg string) bool {
	// Allow alphanumeric characters, dots, slashes, dashes, underscores, and equals
	// This is a whitelist approach for security
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._/\-=]+$`)
	return validPattern.MatchString(arg) && len(arg) < 256 // Also limit length
}
