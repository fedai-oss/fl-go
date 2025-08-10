package monitoring

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ishaileshpant/fl-go/pkg/federation"
)

// MonitoringHooks provides integration points for FL components to send metrics
type MonitoringHooks struct {
	service MonitoringService
	enabled bool
}

// NewMonitoringHooks creates a new monitoring hooks instance
func NewMonitoringHooks(service MonitoringService, enabled bool) *MonitoringHooks {
	return &MonitoringHooks{
		service: service,
		enabled: enabled,
	}
}

// IsEnabled returns whether monitoring is enabled
func (h *MonitoringHooks) IsEnabled() bool {
	return h.enabled
}

// Federation Lifecycle Hooks

// OnFederationStart records the start of a federation
func (h *MonitoringHooks) OnFederationStart(ctx context.Context, plan *federation.FLPlan, aggregatorAddress string) error {
	if !h.enabled {
		return nil
	}

	metrics := &FederationMetrics{
		ID:                fmt.Sprintf("fed_%d", time.Now().Unix()),
		Name:              fmt.Sprintf("Federation_%s", plan.Algorithm.Name),
		Status:            StatusRunning,
		Mode:              string(plan.Mode),
		Algorithm:         plan.Algorithm.Name,
		StartTime:         time.Now(),
		CurrentRound:      0,
		TotalRounds:       plan.Rounds,
		ActiveCollabs:     0,
		TotalCollabs:      len(plan.Collaborators),
		ModelSize:         0, // Will be updated when first model is loaded
		LastUpdate:        time.Now(),
		AggregatorAddress: aggregatorAddress,
	}

	if err := h.service.RegisterFederation(ctx, metrics); err != nil {
		log.Printf("Failed to record federation start: %v", err)
		return err
	}

	return nil
}

// OnFederationEnd records the completion or failure of a federation
func (h *MonitoringHooks) OnFederationEnd(ctx context.Context, federationID string, status FederationStatus, endTime time.Time) error {
	if !h.enabled {
		return nil
	}

	// Get current federation metrics
	currentMetrics, err := h.service.GetFederation(ctx, federationID)
	if err != nil {
		return err
	}

	// Update status and end time
	currentMetrics.Status = status
	currentMetrics.EndTime = &endTime
	currentMetrics.LastUpdate = time.Now()

	if err := h.service.UpdateFederation(ctx, federationID, currentMetrics); err != nil {
		log.Printf("Failed to record federation end: %v", err)
		return err
	}

	return nil
}

// Round Lifecycle Hooks

// OnRoundStart records the start of a training round
func (h *MonitoringHooks) OnRoundStart(ctx context.Context, federationID string, roundNumber int, algorithm string, participantCount int) (string, error) {
	if !h.enabled {
		return "", nil
	}

	roundID := fmt.Sprintf("round_%s_%d", federationID, roundNumber)
	metrics := &RoundMetrics{
		ID:               roundID,
		FederationID:     federationID,
		RoundNumber:      roundNumber,
		Algorithm:        algorithm,
		StartTime:        time.Now(),
		ParticipantCount: participantCount,
		UpdatesReceived:  0,
		Status:           "running",
	}

	if err := h.service.RecordRoundStart(ctx, metrics); err != nil {
		log.Printf("Failed to record round start: %v", err)
		return "", err
	}

	// Update federation current round
	if err := h.updateFederationRound(ctx, federationID, roundNumber); err != nil {
		log.Printf("Failed to update federation round: %v", err)
	}

	return roundID, nil
}

// OnRoundEnd records the completion of a training round
func (h *MonitoringHooks) OnRoundEnd(ctx context.Context, roundID string, federationID string, roundNumber int, duration time.Duration, updatesReceived int, accuracy *float64, loss *float64) error {
	if !h.enabled {
		return nil
	}

	endTime := time.Now()
	metrics := &RoundMetrics{
		ID:              roundID,
		FederationID:    federationID,
		RoundNumber:     roundNumber,
		EndTime:         &endTime,
		Duration:        duration,
		UpdatesReceived: updatesReceived,
		ModelAccuracy:   accuracy,
		ModelLoss:       loss,
		Status:          "completed",
	}

	if err := h.service.RecordRoundEnd(ctx, roundID, metrics); err != nil {
		log.Printf("Failed to record round end: %v", err)
		return err
	}

	return nil
}

// Collaborator Lifecycle Hooks

// OnCollaboratorJoin records when a collaborator joins a federation
func (h *MonitoringHooks) OnCollaboratorJoin(ctx context.Context, collaboratorID, federationID, address string) error {
	if !h.enabled {
		return nil
	}

	metrics := &CollaboratorMetrics{
		ID:               collaboratorID,
		FederationID:     federationID,
		Address:          address,
		Status:           CollabStatusConnected,
		JoinTime:         time.Now(),
		LastSeen:         time.Now(),
		CurrentRound:     0,
		UpdatesSubmitted: 0,
		ErrorCount:       0,
	}

	if err := h.service.RegisterCollaborator(ctx, metrics); err != nil {
		log.Printf("Failed to record collaborator join: %v", err)
		return err
	}

	// Update federation active collaborator count
	if err := h.updateFederationCollaboratorCount(ctx, federationID); err != nil {
		log.Printf("Failed to update federation collaborator count: %v", err)
	}

	return nil
}

// OnCollaboratorLeave records when a collaborator leaves a federation
func (h *MonitoringHooks) OnCollaboratorLeave(ctx context.Context, collaboratorID string, reason string) error {
	if !h.enabled {
		return nil
	}

	// Get current collaborator metrics
	currentMetrics, err := h.service.GetCollaborator(ctx, collaboratorID)
	if err != nil {
		return err
	}

	// Update status
	currentMetrics.Status = CollabStatusDisconnected
	currentMetrics.LastSeen = time.Now()
	if reason != "" {
		currentMetrics.LastError = reason
	}

	if err := h.service.UpdateCollaborator(ctx, collaboratorID, currentMetrics); err != nil {
		log.Printf("Failed to record collaborator leave: %v", err)
		return err
	}

	// Update federation active collaborator count
	if err := h.updateFederationCollaboratorCount(ctx, currentMetrics.FederationID); err != nil {
		log.Printf("Failed to update federation collaborator count: %v", err)
	}

	return nil
}

// OnCollaboratorStatusChange records when a collaborator's status changes
func (h *MonitoringHooks) OnCollaboratorStatusChange(ctx context.Context, collaboratorID string, status CollaboratorStatus, errorMsg string) error {
	if !h.enabled {
		return nil
	}

	// Get current collaborator metrics
	currentMetrics, err := h.service.GetCollaborator(ctx, collaboratorID)
	if err != nil {
		return err
	}

	// Update status
	currentMetrics.Status = status
	currentMetrics.LastSeen = time.Now()

	if status == CollabStatusError {
		currentMetrics.ErrorCount++
		if errorMsg != "" {
			currentMetrics.LastError = errorMsg
		}
	}

	if err := h.service.UpdateCollaborator(ctx, collaboratorID, currentMetrics); err != nil {
		log.Printf("Failed to update collaborator status: %v", err)
		return err
	}

	return nil
}

// Model Update Hooks

// OnModelUpdateReceived records when a model update is received
func (h *MonitoringHooks) OnModelUpdateReceived(ctx context.Context, federationID, collaboratorID string, roundNumber int, updateSize int, processingTime time.Duration, staleness int, weight float64) error {
	if !h.enabled {
		return nil
	}

	metrics := &ModelUpdateMetrics{
		FederationID:   federationID,
		CollaboratorID: collaboratorID,
		RoundNumber:    roundNumber,
		Timestamp:      time.Now(),
		UpdateSize:     updateSize,
		ProcessingTime: float64(processingTime.Milliseconds()),
		Staleness:      staleness,
		Weight:         weight,
	}

	if err := h.service.RecordModelUpdate(ctx, metrics); err != nil {
		log.Printf("Failed to record model update: %v", err)
		return err
	}

	// Update collaborator metrics
	if err := h.updateCollaboratorUpdate(ctx, collaboratorID, roundNumber, processingTime); err != nil {
		log.Printf("Failed to update collaborator update metrics: %v", err)
	}

	return nil
}

// Aggregation Hooks

// OnAggregationStart records when aggregation starts
func (h *MonitoringHooks) OnAggregationStart(ctx context.Context, federationID string, roundNumber int, algorithm string, updatesCount int) (string, error) {
	if !h.enabled {
		return "", nil
	}

	aggregationID := fmt.Sprintf("agg_%s_%d_%d", federationID, roundNumber, time.Now().Unix())
	metrics := &AggregationMetrics{
		ID:                aggregationID,
		FederationID:      federationID,
		RoundNumber:       roundNumber,
		Algorithm:         algorithm,
		StartTime:         time.Now(),
		UpdatesAggregated: updatesCount,
	}

	if err := h.service.RecordAggregation(ctx, metrics); err != nil {
		log.Printf("Failed to record aggregation start: %v", err)
		return "", err
	}

	return aggregationID, nil
}

// OnAggregationEnd records when aggregation completes
func (h *MonitoringHooks) OnAggregationEnd(ctx context.Context, aggregationID string, duration time.Duration, convergence *float64, quality *float64) error {
	if !h.enabled {
		return nil
	}

	// This would need to be implemented to update the existing aggregation record
	// For now, we'll create a new record with the completion data
	endTime := time.Now()

	// In a real implementation, you'd update the existing record
	// Here we're showing the data structure for completion
	_ = &AggregationMetrics{
		ID:                 aggregationID,
		EndTime:            endTime,
		Duration:           duration,
		ModelConvergence:   convergence,
		AggregationQuality: quality,
	}

	// Log completion for now
	log.Printf("Aggregation %s completed in %v", aggregationID, duration)

	return nil
}

// Resource Monitoring Hooks

// OnResourceMetrics records system resource usage
func (h *MonitoringHooks) OnResourceMetrics(ctx context.Context, source string, cpuUsage, memoryUsage, diskUsage float64, memoryUsed, memoryTotal int64, networkRx, networkTx float64) error {
	if !h.enabled {
		return nil
	}

	metrics := &ResourceMetrics{
		Timestamp:     time.Now(),
		CPUUsage:      cpuUsage,
		MemoryUsage:   memoryUsage,
		MemoryUsed:    memoryUsed,
		MemoryTotal:   memoryTotal,
		DiskUsage:     diskUsage,
		NetworkRxRate: networkRx,
		NetworkTxRate: networkTx,
	}

	if err := h.service.RecordResourceMetrics(ctx, source, metrics); err != nil {
		log.Printf("Failed to record resource metrics: %v", err)
		return err
	}

	return nil
}

// Event Hooks

// OnEvent records a monitoring event
func (h *MonitoringHooks) OnEvent(ctx context.Context, federationID, source, level, message string, eventType MetricType, data map[string]interface{}) error {
	if !h.enabled {
		return nil
	}

	event := &MonitoringEvent{
		FederationID: federationID,
		Type:         eventType,
		Timestamp:    time.Now(),
		Source:       source,
		Level:        level,
		Message:      message,
		Data:         data,
	}

	if err := h.service.RecordEvent(ctx, event); err != nil {
		log.Printf("Failed to record event: %v", err)
		return err
	}

	return nil
}

// Alert Hooks

// OnAlert creates an alert for significant events
func (h *MonitoringHooks) OnAlert(ctx context.Context, federationID, alertType, severity, title, message, source string, data map[string]interface{}) error {
	if !h.enabled {
		return nil
	}

	// For now, log the alert - in a real implementation, this would create an alert record
	log.Printf("ALERT [%s] %s: %s - %s", severity, title, message, source)

	// Also record as an event
	return h.OnEvent(ctx, federationID, source, "alert", fmt.Sprintf("[%s] %s: %s", severity, title, message), MetricTypeRound, data)
}

// Training Performance Hooks

// OnTrainingStart records when training starts on a collaborator
func (h *MonitoringHooks) OnTrainingStart(ctx context.Context, collaboratorID string, roundNumber int) error {
	if !h.enabled {
		return nil
	}

	return h.OnCollaboratorStatusChange(ctx, collaboratorID, CollabStatusTraining, "")
}

// OnTrainingEnd records when training completes on a collaborator
func (h *MonitoringHooks) OnTrainingEnd(ctx context.Context, collaboratorID string, roundNumber int, duration time.Duration, accuracy *float64, loss *float64) error {
	if !h.enabled {
		return nil
	}

	// Update collaborator training time
	currentMetrics, err := h.service.GetCollaborator(ctx, collaboratorID)
	if err != nil {
		return err
	}

	currentMetrics.TrainingTime += duration
	currentMetrics.Status = CollabStatusIdle
	currentMetrics.CurrentRound = roundNumber
	currentMetrics.LastSeen = time.Now()

	if err := h.service.UpdateCollaborator(ctx, collaboratorID, currentMetrics); err != nil {
		log.Printf("Failed to update collaborator training metrics: %v", err)
		return err
	}

	return nil
}

// Helper methods for updating related metrics

func (h *MonitoringHooks) updateFederationRound(ctx context.Context, federationID string, roundNumber int) error {
	currentMetrics, err := h.service.GetFederation(ctx, federationID)
	if err != nil {
		return err
	}

	currentMetrics.CurrentRound = roundNumber
	currentMetrics.LastUpdate = time.Now()

	return h.service.UpdateFederation(ctx, federationID, currentMetrics)
}

func (h *MonitoringHooks) updateFederationCollaboratorCount(ctx context.Context, federationID string) error {
	collaborators, err := h.service.GetFederationCollaborators(ctx, federationID)
	if err != nil {
		return err
	}

	activeCount := 0
	for _, collab := range collaborators {
		if collab.Status == CollabStatusConnected || collab.Status == CollabStatusTraining {
			activeCount++
		}
	}

	currentMetrics, err := h.service.GetFederation(ctx, federationID)
	if err != nil {
		return err
	}

	currentMetrics.ActiveCollabs = activeCount
	currentMetrics.LastUpdate = time.Now()

	return h.service.UpdateFederation(ctx, federationID, currentMetrics)
}

func (h *MonitoringHooks) updateCollaboratorUpdate(ctx context.Context, collaboratorID string, roundNumber int, latency time.Duration) error {
	currentMetrics, err := h.service.GetCollaborator(ctx, collaboratorID)
	if err != nil {
		return err
	}

	currentMetrics.UpdatesSubmitted++
	currentMetrics.CurrentRound = roundNumber
	currentMetrics.LastSeen = time.Now()

	// Update average latency (simple moving average)
	if currentMetrics.AverageLatency == 0 {
		currentMetrics.AverageLatency = float64(latency.Milliseconds())
	} else {
		currentMetrics.AverageLatency = (currentMetrics.AverageLatency + float64(latency.Milliseconds())) / 2
	}

	return h.service.UpdateCollaborator(ctx, collaboratorID, currentMetrics)
}
