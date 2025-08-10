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
)

// ModularAggregator implements a flexible aggregator that can use different algorithms
type ModularAggregator struct {
	pb.UnimplementedFederatedLearningServer
	plan         *federation.FLPlan
	algorithm    AggregationAlgorithm
	mu           sync.Mutex
	updates      []ClientUpdate
	modelSize    int
	currentRound int
	srv          *grpc.Server
	globalModel  []float32
	lastUpdate   time.Time
	stopChan     chan struct{}
	isAsync      bool
}

// NewModularAggregator creates a new modular aggregator with the specified algorithm
func NewModularAggregator(plan *federation.FLPlan) (*ModularAggregator, error) {
	// Determine algorithm type
	algorithmName := "fedavg" // default
	if plan.Algorithm.Name != "" {
		algorithmName = plan.Algorithm.Name
	}

	// Create the aggregation algorithm
	algType := AlgorithmType(algorithmName)
	algorithm, err := CreateAggregationAlgorithm(algType)
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregation algorithm: %v", err)
	}

	// Determine if this is async mode
	isAsync := plan.Mode == federation.ModeAsync

	aggregator := &ModularAggregator{
		plan:         plan,
		algorithm:    algorithm,
		updates:      make([]ClientUpdate, 0),
		currentRound: 0,
		isAsync:      isAsync,
		stopChan:     make(chan struct{}),
	}

	return aggregator, nil
}

func (a *ModularAggregator) Start(ctx context.Context) error {
	log.Printf("Starting Modular Aggregator with %s algorithm in %s mode",
		a.algorithm.GetName(), a.plan.Mode)

	// Initialize the algorithm
	algConfig := AlgorithmConfig{
		AlgorithmName:   a.plan.Algorithm.Name,
		ModelSize:       a.modelSize,
		Hyperparameters: a.plan.Algorithm.Hyperparameters,
		Mode:            a.plan.Mode,
	}

	if err := a.algorithm.Initialize(algConfig); err != nil {
		return fmt.Errorf("failed to initialize algorithm: %v", err)
	}

	// Load initial model to determine model size
	if err := a.loadInitialModel(); err != nil {
		return fmt.Errorf("failed to load initial model: %v", err)
	}

	// Update algorithm config with actual model size
	algConfig.ModelSize = a.modelSize
	if err := a.algorithm.Initialize(algConfig); err != nil {
		return fmt.Errorf("failed to reinitialize algorithm with model size: %v", err)
	}

	// Log algorithm hyperparameters
	hyperparams := a.algorithm.GetHyperparameters()
	log.Printf("Algorithm hyperparameters: %+v", hyperparams)

	// Start gRPC server
	lis, err := net.Listen("tcp", a.plan.Aggregator.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	a.srv = grpc.NewServer()
	pb.RegisterFederatedLearningServer(a.srv, a)

	// Start server in background
	go func() {
		log.Printf("Modular aggregator server listening on %s", a.plan.Aggregator.Address)
		if err := a.srv.Serve(lis); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Run federation based on mode
	if a.isAsync {
		return a.runAsyncFederation(ctx)
	} else {
		return a.runSyncFederation(ctx)
	}
}

func (a *ModularAggregator) loadInitialModel() error {
	data, err := os.ReadFile(a.plan.InitialModel)
	if err != nil {
		log.Printf("Warning: Could not read initial model %s: %v", a.plan.InitialModel, err)
		// Create a dummy model for testing
		a.modelSize = 1000 // Default model size
		a.globalModel = make([]float32, a.modelSize)
		return nil
	}

	// Determine model size from file
	a.modelSize = len(data) / 4 // 4 bytes per float32
	a.globalModel = make([]float32, a.modelSize)

	// Load initial weights
	for i := 0; i < a.modelSize; i++ {
		a.globalModel[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}

	log.Printf("Loaded initial model with %d parameters", a.modelSize)
	return nil
}

func (a *ModularAggregator) runSyncFederation(ctx context.Context) error {
	log.Printf("Running synchronous federation with %s for %d rounds",
		a.algorithm.GetName(), a.plan.Rounds)

	// Run federated learning for specified rounds
	for round := 1; round <= a.plan.Rounds; round++ {
		a.currentRound = round
		log.Printf("Starting round %d/%d with %s algorithm", round, a.plan.Rounds, a.algorithm.GetName())

		// Reset updates for new round
		a.mu.Lock()
		a.updates = make([]ClientUpdate, 0)
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
			time.Sleep(2 * time.Second)
		}

		// Perform aggregation using the selected algorithm
		log.Printf("Aggregating updates for round %d using %s", round, a.algorithm.GetName())
		a.mu.Lock()
		newModel, err := a.algorithm.Aggregate(a.updates, a.globalModel)
		a.mu.Unlock()

		if err != nil {
			return fmt.Errorf("aggregation failed in round %d: %v", round, err)
		}

		// Update global model
		a.globalModel = newModel

		// Save aggregated model
		if err := a.saveModel(round); err != nil {
			return fmt.Errorf("failed to save model in round %d: %v", round, err)
		}

		log.Printf("Round %d complete using %s algorithm", round, a.algorithm.GetName())
	}

	log.Printf("All %d rounds completed successfully with %s", a.plan.Rounds, a.algorithm.GetName())
	a.srv.Stop()
	return nil
}

func (a *ModularAggregator) runAsyncFederation(ctx context.Context) error {
	log.Printf("Running asynchronous federation with %s algorithm", a.algorithm.GetName())

	// Start async aggregation goroutine
	go a.asyncAggregationLoop()

	// Keep server running
	select {
	case <-ctx.Done():
		close(a.stopChan)
		a.srv.Stop()
		return ctx.Err()
	}
}

func (a *ModularAggregator) asyncAggregationLoop() {
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

func (a *ModularAggregator) performAsyncAggregation() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.updates) == 0 {
		return
	}

	log.Printf("Performing async aggregation with %d updates using %s",
		len(a.updates), a.algorithm.GetName())

	// Calculate staleness for each update
	currentTime := time.Now()
	validUpdates := make([]ClientUpdate, 0)

	for _, update := range a.updates {
		staleness := int(currentTime.Sub(update.Timestamp).Seconds())
		update.Staleness = staleness

		if staleness <= a.plan.AsyncConfig.MaxStaleness {
			validUpdates = append(validUpdates, update)
		} else {
			log.Printf("Dropping stale update from %s (staleness: %d)",
				update.CollaboratorID, staleness)
		}
	}

	if len(validUpdates) == 0 {
		log.Printf("No valid updates to aggregate")
		return
	}

	// Perform aggregation using the selected algorithm
	newModel, err := a.algorithm.Aggregate(validUpdates, a.globalModel)
	if err != nil {
		log.Printf("Async aggregation failed: %v", err)
		return
	}

	// Update global model
	a.globalModel = newModel
	a.currentRound++
	a.lastUpdate = currentTime

	// Save updated model
	if err := a.saveAsyncModel(); err != nil {
		log.Printf("Failed to save async model: %v", err)
	} else {
		log.Printf("Async round %d complete using %s, model saved",
			a.currentRound, a.algorithm.GetName())
	}

	// Clear processed updates
	a.updates = make([]ClientUpdate, 0)
}

func (a *ModularAggregator) saveModel(round int) error {
	buf := make([]byte, 4*a.modelSize)
	for i, v := range a.globalModel {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}

	outputPath := a.plan.OutputModel
	if round < a.plan.Rounds {
		outputPath = fmt.Sprintf("save/round_%d_model.pt", round)
	}

	if err := os.WriteFile(outputPath, buf, 0600); err != nil {
		return err
	}

	log.Printf("Model saved to %s", outputPath)
	return nil
}

func (a *ModularAggregator) saveAsyncModel() error {
	buf := make([]byte, 4*a.modelSize)
	for i, v := range a.globalModel {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}

	outputPath := fmt.Sprintf("save/async_%s_round_%d_model.pt",
		a.algorithm.GetName(), a.currentRound)
	return os.WriteFile(outputPath, buf, 0600)
}

// gRPC service implementations

func (a *ModularAggregator) JoinFederation(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	log.Printf("Collaborator %s joining %s federation with %s algorithm",
		req.CollaboratorId, a.plan.Mode, a.algorithm.GetName())

	// Return current global model
	buf := make([]byte, 4*a.modelSize)
	for i, v := range a.globalModel {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}

	return &pb.JoinResponse{InitialModel: buf}, nil
}

func (a *ModularAggregator) SubmitUpdate(ctx context.Context, upd *pb.ModelUpdate) (*pb.Ack, error) {
	floats := make([]float32, len(upd.ModelWeights)/4)
	for i := range floats {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(upd.ModelWeights[i*4:]))
	}

	update := ClientUpdate{
		CollaboratorID: upd.CollaboratorId,
		Weights:        floats,
		Timestamp:      time.Now(),
		Round:          a.currentRound,
		NumSamples:     100,  // Default value - could be passed from client
		LearningRate:   0.01, // Default value - could be passed from client
	}

	a.mu.Lock()
	a.updates = append(a.updates, update)
	updateCount := len(a.updates)
	a.mu.Unlock()

	mode := "sync"
	if a.isAsync {
		mode = "async"
	}

	log.Printf("Received %s update %d from %s (round %d) for %s algorithm",
		mode, updateCount, upd.CollaboratorId, a.currentRound, a.algorithm.GetName())

	return &pb.Ack{Success: true}, nil
}

func (a *ModularAggregator) GetLatestModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Return current global model
	buf := make([]byte, 4*a.modelSize)
	for i, v := range a.globalModel {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}

	log.Printf("Providing latest %s model to %s (round %d)",
		a.algorithm.GetName(), req.CollaboratorId, a.currentRound)

	// Safely convert int to int32 to prevent overflow
	var currentRound int32
	if a.currentRound > math.MaxInt32 {
		log.Printf("Warning: current round %d exceeds int32 max, capping at %d", a.currentRound, math.MaxInt32)
		currentRound = math.MaxInt32
	} else {
		currentRound = int32(a.currentRound) // #nosec G115 - Safe conversion with bounds check above
	}

	return &pb.GetModelResponse{
		ModelWeights: buf,
		CurrentRound: currentRound,
	}, nil
}
