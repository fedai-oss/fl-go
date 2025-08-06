package aggregator

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"os"
	"sync"

	pb "github.com/ishaileshpant/openfl-go/api"
	"github.com/ishaileshpant/openfl-go/pkg/federation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// FedAvgAggregator implements a simple one-round FedAvg.
type FedAvgAggregator struct {
	pb.UnimplementedFederatedLearningServer
	plan      *federation.FLPlan
	mu        sync.Mutex
	updates   [][]float32
	modelSize int
}

func NewFedAvgAggregator(plan *federation.FLPlan) *FedAvgAggregator {
	return &FedAvgAggregator{plan: plan}
}

func (a *FedAvgAggregator) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", a.plan.Aggregator.Address)
	if err != nil {
		return err
	}
	srv := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	pb.RegisterFederatedLearningServer(srv, a)
	go srv.Serve(lis)
	data, err := os.ReadFile(a.plan.InitialModel)
	if err != nil {
		return err
	}
	a.modelSize = len(data) / 4
	for {
		a.mu.Lock()
		if len(a.updates) >= len(a.plan.Collaborators) {
			a.mu.Unlock()
			break
		}
		a.mu.Unlock()
	}
	avg := make([]float32, a.modelSize)
	for _, upd := range a.updates {
		for i, v := range upd {
			avg[i] += v
		}
	}
	for i := range avg {
		avg[i] /= float32(len(a.updates))
	}
	buf := make([]byte, 4*a.modelSize)
	for i, v := range avg {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	if err := os.WriteFile(a.plan.OutputModel, buf, 0644); err != nil {
		return err
	}
	srv.Stop()
	return nil
}

func (a *FedAvgAggregator) JoinFederation(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	data, _ := os.ReadFile(a.plan.InitialModel)
	return &pb.JoinResponse{InitialModel: data}, nil
}

func (a *FedAvgAggregator) SubmitUpdate(ctx context.Context, upd *pb.ModelUpdate) (*pb.Ack, error) {
	floats := make([]float32, len(upd.ModelWeights)/4)
	for i := range floats {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(upd.ModelWeights[i*4:]))
	}
	a.mu.Lock()
	a.updates = append(a.updates, floats)
	a.mu.Unlock()
	return &pb.Ack{Success: true}, nil
}
