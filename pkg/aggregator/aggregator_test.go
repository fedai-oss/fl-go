package aggregator

import (
	"fmt"
	"testing"

	"github.com/ishaileshpant/fl-go/pkg/federation"
)

func TestNewAggregator(t *testing.T) {
	tests := []struct {
		name     string
		mode     federation.FLMode
		expected string
	}{
		{
			name:     "Sync Mode",
			mode:     federation.ModeSync,
			expected: "*aggregator.FedAvgAggregator",
		},
		{
			name:     "Async Mode",
			mode:     federation.ModeAsync,
			expected: "*aggregator.AsyncFedAvgAggregator",
		},
		{
			name:     "Default Mode (Empty)",
			mode:     "",
			expected: "*aggregator.FedAvgAggregator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &federation.FLPlan{
				Mode: tt.mode,
				AsyncConfig: federation.AsyncConfig{
					MaxStaleness:     300,
					MinUpdates:       1,
					AggregationDelay: 10,
					StalenessWeight:  0.95,
				},
			}

			agg := NewAggregator(plan)
			aggType := fmt.Sprintf("%T", agg)

			if aggType != tt.expected {
				t.Errorf("NewAggregator() = %v, want %v", aggType, tt.expected)
			}
		})
	}
}

func TestAsyncConfigDefaults(t *testing.T) {
	plan := &federation.FLPlan{
		Mode: federation.ModeAsync,
		AsyncConfig: federation.AsyncConfig{
			MaxStaleness:     300,
			MinUpdates:       1,
			AggregationDelay: 10,
			StalenessWeight:  0.95,
		},
	}

	agg := NewAsyncFedAvgAggregator(plan)

	if agg.plan.AsyncConfig.MaxStaleness != 300 {
		t.Errorf("MaxStaleness = %d, want 300", agg.plan.AsyncConfig.MaxStaleness)
	}

	if agg.plan.AsyncConfig.MinUpdates != 1 {
		t.Errorf("MinUpdates = %d, want 1", agg.plan.AsyncConfig.MinUpdates)
	}

	if agg.plan.AsyncConfig.AggregationDelay != 10 {
		t.Errorf("AggregationDelay = %d, want 10", agg.plan.AsyncConfig.AggregationDelay)
	}

	if agg.plan.AsyncConfig.StalenessWeight != 0.95 {
		t.Errorf("StalenessWeight = %f, want 0.95", agg.plan.AsyncConfig.StalenessWeight)
	}
}
