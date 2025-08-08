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

	pb "github.com/ishaileshpant/fl-go/api"
	"github.com/ishaileshpant/fl-go/pkg/federation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Aggregator interface defines the contract for both sync and async aggregators
type Aggregator interface {
	Start(ctx context.Context) error
	JoinFederation(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error)
	SubmitUpdate(ctx context.Context, upd *pb.ModelUpdate) (*pb.Ack, error)
	GetLatestModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error)
}

// UpdateInfo tracks update metadata for async FL
type UpdateInfo struct {
	CollaboratorID string
	Weights        []float32
	Timestamp      time.Time
	Round          int
	Staleness      int
}

// FedAvgAggregator implements synchronous multi-round FedAvg (existing implementation)
type FedAvgAggregator struct {
	pb.UnimplementedFederatedLearningServer
	plan         *federation.FLPlan
	mu           sync.Mutex
	updates      [][]float32
	modelSize    int
	currentRound int
	srv          *grpc.Server
}

// AsyncFedAvgAggregator implements asynchronous FedAvg based on Papaya paper
type AsyncFedAvgAggregator struct {
	pb.UnimplementedFederatedLearningServer
	plan         *federation.FLPlan
	mu           sync.Mutex
	updates      []UpdateInfo
	modelSize    int
	currentRound int
	srv          *grpc.Server
	globalModel  []float32
	lastUpdate   time.Time
	stopChan     chan struct{}
}

// NewAggregator creates the appropriate aggregator based on mode
func NewAggregator(plan *federation.FLPlan) Aggregator {
	switch plan.Mode {
	case federation.ModeAsync:
		return NewAsyncFedAvgAggregator(plan)
	default:
		return NewFedAvgAggregator(plan)
	}
}

func NewFedAvgAggregator(plan *federation.FLPlan) *FedAvgAggregator {
	return &FedAvgAggregator{plan: plan}
}

func NewAsyncFedAvgAggregator(plan *federation.FLPlan) *AsyncFedAvgAggregator {
	return &AsyncFedAvgAggregator{
		plan:     plan,
		stopChan: make(chan struct{}),
	}
}

// Synchronous Aggregator Implementation (existing)
func (a *FedAvgAggregator) Start(ctx context.Context) error {
	log.Printf("Starting SYNC aggregator on %s", a.plan.Aggregator.Address)
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

func (a *FedAvgAggregator) GetLatestModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error) {
	// In sync mode, return the initial model since rounds are synchronized
	data, err := os.ReadFile(a.plan.InitialModel)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial model: %v", err)
	}

	return &pb.GetModelResponse{
		ModelWeights: data,
		CurrentRound: int32(a.currentRound),
	}, nil
}

// Asynchronous Aggregator Implementation (new)
func (a *AsyncFedAvgAggregator) Start(ctx context.Context) error {
	log.Printf("Starting ASYNC aggregator on %s", a.plan.Aggregator.Address)
	log.Printf("Async config: max_staleness=%d, min_updates=%d, delay=%ds",
		a.plan.AsyncConfig.MaxStaleness, a.plan.AsyncConfig.MinUpdates, a.plan.AsyncConfig.AggregationDelay)

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

	// Read initial model to determine size and set as global model
	data, err := os.ReadFile(a.plan.InitialModel)
	if err != nil {
		return err
	}
	a.modelSize = len(data) / 4
	a.globalModel = make([]float32, a.modelSize)
	for i := range a.globalModel {
		a.globalModel[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	log.Printf("Model size: %d parameters", a.modelSize)

	// Start async aggregation loop
	go a.asyncAggregationLoop()

	// Wait for completion signal (could be based on time, rounds, or other criteria)
	<-ctx.Done()

	log.Printf("Async FL completed")
	a.srv.Stop()
	return nil
}

func (a *AsyncFedAvgAggregator) asyncAggregationLoop() {
	ticker := time.NewTicker(time.Duration(a.plan.AsyncConfig.AggregationDelay) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			updateCount := len(a.updates)
			a.mu.Unlock()

			if updateCount >= a.plan.AsyncConfig.MinUpdates {
				a.performAsyncAggregation()
			}
		case <-a.stopChan:
			return
		}
	}
}

func (a *AsyncFedAvgAggregator) performAsyncAggregation() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.updates) == 0 {
		return
	}

	log.Printf("Performing async aggregation with %d updates", len(a.updates))

	// Calculate staleness for each update
	currentTime := time.Now()
	for i := range a.updates {
		a.updates[i].Staleness = int(currentTime.Sub(a.updates[i].Timestamp).Seconds())
	}

	// Filter out updates that are too stale
	validUpdates := make([]UpdateInfo, 0)
	for _, update := range a.updates {
		if update.Staleness <= a.plan.AsyncConfig.MaxStaleness {
			validUpdates = append(validUpdates, update)
		} else {
			log.Printf("Dropping stale update from %s (staleness: %d)",
				update.CollaboratorID, update.Staleness)
		}
	}

	if len(validUpdates) == 0 {
		log.Printf("No valid updates to aggregate")
		return
	}

	// Perform staleness-aware aggregation
	newModel := make([]float32, a.modelSize)
	totalWeight := 0.0

	for _, update := range validUpdates {
		// Apply staleness weight decay
		weight := math.Pow(a.plan.AsyncConfig.StalenessWeight, float64(update.Staleness))
		totalWeight += weight

		for i, v := range update.Weights {
			newModel[i] += float32(weight) * v
		}
	}

	// Normalize by total weight
	for i := range newModel {
		newModel[i] /= float32(totalWeight)
	}

	// Update global model
	a.globalModel = newModel
	a.currentRound++
	a.lastUpdate = currentTime

	// Save updated model
	buf := make([]byte, 4*a.modelSize)
	for i, v := range a.globalModel {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}

	outputPath := fmt.Sprintf("save/async_round_%d_model.pt", a.currentRound)
	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		log.Printf("Error saving async model: %v", err)
	} else {
		log.Printf("Async round %d complete, model saved to %s", a.currentRound, outputPath)
	}

	// Clear processed updates
	a.updates = make([]UpdateInfo, 0)
}

func (a *AsyncFedAvgAggregator) JoinFederation(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	log.Printf("Collaborator %s joining async federation", req.CollaboratorId)

	// Return current global model
	buf := make([]byte, 4*a.modelSize)
	for i, v := range a.globalModel {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}

	return &pb.JoinResponse{InitialModel: buf}, nil
}

func (a *AsyncFedAvgAggregator) SubmitUpdate(ctx context.Context, upd *pb.ModelUpdate) (*pb.Ack, error) {
	floats := make([]float32, len(upd.ModelWeights)/4)
	for i := range floats {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(upd.ModelWeights[i*4:]))
	}

	updateInfo := UpdateInfo{
		CollaboratorID: upd.CollaboratorId,
		Weights:        floats,
		Timestamp:      time.Now(),
		Round:          a.currentRound,
	}

	a.mu.Lock()
	a.updates = append(a.updates, updateInfo)
	updateCount := len(a.updates)
	a.mu.Unlock()

	log.Printf("Received async update %d from %s (round %d)", updateCount, upd.CollaboratorId, a.currentRound)
	return &pb.Ack{Success: true}, nil
}

func (a *AsyncFedAvgAggregator) GetLatestModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Return current global model
	buf := make([]byte, 4*a.modelSize)
	for i, v := range a.globalModel {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}

	log.Printf("Providing latest model to %s (round %d)", req.CollaboratorId, a.currentRound)

	return &pb.GetModelResponse{
		ModelWeights: buf,
		CurrentRound: int32(a.currentRound),
	}, nil
}
