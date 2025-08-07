package aggregator

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/ishaileshpant/openfl-go/api"
	"github.com/ishaileshpant/openfl-go/pkg/federation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// FedAvgAggregator implements a multi-round FedAvg.
type FedAvgAggregator struct {
	pb.UnimplementedFederatedLearningServer
	plan         *federation.FLPlan
	mu           sync.Mutex
	updates      [][]float32
	modelSize    int
	currentRound int
	srv          *grpc.Server
}

func NewFedAvgAggregator(plan *federation.FLPlan) *FedAvgAggregator {
	return &FedAvgAggregator{plan: plan}
}

func (a *FedAvgAggregator) Start(ctx context.Context) error {
	log.Printf("Starting aggregator on %s", a.plan.Aggregator.Address)
	log.Printf("Expecting %d collaborators for %d rounds", len(a.plan.Collaborators), a.plan.Rounds)

	lis, err := net.Listen("tcp", a.plan.Aggregator.Address)
	if err != nil {
		return err
	}

	a.srv = grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	pb.RegisterFederatedLearningServer(a.srv, a)

	// Start gRPC server in background
	go func() {
		log.Printf("gRPC server listening on %s", a.plan.Aggregator.Address)
		if err := a.srv.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Read initial model to determine size
	data, err := os.ReadFile(a.plan.InitialModel)
	if err != nil {
		return err
	}
	a.modelSize = len(data) / 4
	log.Printf("Model size: %d parameters", a.modelSize)

	// Run federated learning for specified rounds
	for round := 1; round <= a.plan.Rounds; round++ {
		a.currentRound = round
		log.Printf("Starting round %d/%d", round, a.plan.Rounds)

		// Reset updates for new round
		a.mu.Lock()
		a.updates = make([][]float32, 0)
		a.mu.Unlock()

		// Wait for all collaborators to submit updates
		log.Printf("Waiting for %d collaborators to submit updates...", len(a.plan.Collaborators))
		for {
			a.mu.Lock()
			updateCount := len(a.updates)
			a.mu.Unlock()

			if updateCount >= len(a.plan.Collaborators) {
				log.Printf("Received updates from all %d collaborators", updateCount)
				break
			}

			log.Printf("Received %d/%d updates, waiting...", updateCount, len(a.plan.Collaborators))
			time.Sleep(2 * time.Second) // Check every 2 seconds
		}

		// Aggregate the updates
		log.Printf("Aggregating updates for round %d", round)
		avg := make([]float32, a.modelSize)
		a.mu.Lock()
		for _, upd := range a.updates {
			for i, v := range upd {
				avg[i] += v
			}
		}
		a.mu.Unlock()

		for i := range avg {
			avg[i] /= float32(len(a.updates))
		}

		// Save aggregated model
		buf := make([]byte, 4*a.modelSize)
		for i, v := range avg {
			binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
		}

		outputPath := a.plan.OutputModel
		if round < a.plan.Rounds {
			// For intermediate rounds, save to save directory
			outputPath = fmt.Sprintf("save/round_%d_model.pt", round)
		}

		if err := os.WriteFile(outputPath, buf, 0644); err != nil {
			return err
		}
		log.Printf("Round %d complete, model saved to %s", round, outputPath)
	}

	log.Printf("All %d rounds completed successfully", a.plan.Rounds)
	a.srv.Stop()
	return nil
}

func (a *FedAvgAggregator) JoinFederation(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	log.Printf("Collaborator %s joining federation", req.CollaboratorId)
	data, err := os.ReadFile(a.plan.InitialModel)
	if err != nil {
		log.Printf("Warning: Could not read initial model %s: %v", a.plan.InitialModel, err)
		// Return empty model if file doesn't exist
		return &pb.JoinResponse{InitialModel: []byte{}}, nil
	}
	return &pb.JoinResponse{InitialModel: data}, nil
}

func (a *FedAvgAggregator) SubmitUpdate(ctx context.Context, upd *pb.ModelUpdate) (*pb.Ack, error) {
	floats := make([]float32, len(upd.ModelWeights)/4)
	for i := range floats {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(upd.ModelWeights[i*4:]))
	}
	a.mu.Lock()
	a.updates = append(a.updates, floats)
	updateCount := len(a.updates)
	a.mu.Unlock()

	log.Printf("Received update %d/%d for round %d", updateCount, len(a.plan.Collaborators), a.currentRound)
	return &pb.Ack{Success: true}, nil
}
