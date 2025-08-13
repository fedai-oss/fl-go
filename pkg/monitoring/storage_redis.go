package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implements Storage interface using Redis
type RedisStorage struct {
	client *redis.Client
	config RedisConfig
	ctx    context.Context
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
	PoolSize int    `yaml:"pool_size"`
	TTL      string `yaml:"ttl"` // Default TTL for keys
}

// NewRedisStorage creates a new Redis storage backend
func NewRedisStorage(config RedisConfig) (*RedisStorage, error) {
	opts := &redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.Database,
	}

	if config.PoolSize > 0 {
		opts.PoolSize = config.PoolSize
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client: client,
		config: config,
		ctx:    ctx,
	}, nil
}

// getDefaultTTL returns the default TTL for Redis keys
func (r *RedisStorage) getDefaultTTL() time.Duration {
	if r.config.TTL == "" {
		return 24 * time.Hour // Default 24 hours
	}

	duration, err := time.ParseDuration(r.config.TTL)
	if err != nil {
		return 24 * time.Hour // Fallback to default
	}

	return duration
}

// StoreFederationMetrics stores federation metrics in Redis
func (r *RedisStorage) StoreFederationMetrics(federation FederationMetrics) error {
	key := fmt.Sprintf("federation:%s", federation.ID)

	// Update last updated time
	federation.LastUpdate = time.Now()

	data, err := json.Marshal(federation)
	if err != nil {
		return fmt.Errorf("failed to marshal federation metrics: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, r.getDefaultTTL()).Err(); err != nil {
		return fmt.Errorf("failed to store federation metrics: %w", err)
	}

	// Add to federations list
	listKey := "federations:list"
	if err := r.client.SAdd(r.ctx, listKey, federation.ID).Err(); err != nil {
		return fmt.Errorf("failed to add federation to list: %w", err)
	}

	// Set TTL for the list as well
	r.client.Expire(r.ctx, listKey, r.getDefaultTTL())

	return nil
}

// GetFederationMetrics retrieves federation metrics from Redis
func (r *RedisStorage) GetFederationMetrics(id string) (*FederationMetrics, error) {
	key := fmt.Sprintf("federation:%s", id)

	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get federation metrics: %w", err)
	}

	var federation FederationMetrics
	if err := json.Unmarshal([]byte(data), &federation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal federation metrics: %w", err)
	}

	return &federation, nil
}

// ListFederations lists all federations with optional filters
func (r *RedisStorage) ListFederations(activeOnly bool) ([]FederationMetrics, error) {
	listKey := "federations:list"

	federationIDs, err := r.client.SMembers(r.ctx, listKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get federation list: %w", err)
	}

	var federations []FederationMetrics
	for _, id := range federationIDs {
		federation, err := r.GetFederationMetrics(id)
		if err != nil {
			continue // Skip failed federations
		}
		if federation == nil {
			continue // Skip not found federations
		}

		// Apply active filter
		if activeOnly && federation.Status != "running" {
			continue
		}

		federations = append(federations, *federation)
	}

	return federations, nil
}

// StoreCollaboratorMetrics stores collaborator metrics in Redis
func (r *RedisStorage) StoreCollaboratorMetrics(collaborator CollaboratorMetrics) error {
	key := fmt.Sprintf("collaborator:%s", collaborator.ID)

	data, err := json.Marshal(collaborator)
	if err != nil {
		return fmt.Errorf("failed to marshal collaborator metrics: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, r.getDefaultTTL()).Err(); err != nil {
		return fmt.Errorf("failed to store collaborator metrics: %w", err)
	}

	// Add to federation's collaborators list
	federationKey := fmt.Sprintf("federation:%s:collaborators", collaborator.FederationID)
	if err := r.client.SAdd(r.ctx, federationKey, collaborator.ID).Err(); err != nil {
		return fmt.Errorf("failed to add collaborator to federation list: %w", err)
	}

	// Set TTL for the federation list
	r.client.Expire(r.ctx, federationKey, r.getDefaultTTL())

	return nil
}

// GetCollaboratorMetrics retrieves collaborator metrics from Redis
func (r *RedisStorage) GetCollaboratorMetrics(federationID string) ([]CollaboratorMetrics, error) {
	federationKey := fmt.Sprintf("federation:%s:collaborators", federationID)

	collaboratorIDs, err := r.client.SMembers(r.ctx, federationKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator list: %w", err)
	}

	var collaborators []CollaboratorMetrics
	for _, id := range collaboratorIDs {
		key := fmt.Sprintf("collaborator:%s", id)

		data, err := r.client.Get(r.ctx, key).Result()
		if err != nil {
			continue // Skip failed collaborators
		}

		var collaborator CollaboratorMetrics
		if err := json.Unmarshal([]byte(data), &collaborator); err != nil {
			continue // Skip invalid data
		}

		collaborators = append(collaborators, collaborator)
	}

	return collaborators, nil
}

// StoreRoundMetrics stores round metrics in Redis
func (r *RedisStorage) StoreRoundMetrics(round RoundMetrics) error {
	key := fmt.Sprintf("round:%s", round.ID)

	data, err := json.Marshal(round)
	if err != nil {
		return fmt.Errorf("failed to marshal round metrics: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, r.getDefaultTTL()).Err(); err != nil {
		return fmt.Errorf("failed to store round metrics: %w", err)
	}

	// Add to federation's rounds sorted set (by round number)
	federationKey := fmt.Sprintf("federation:%s:rounds", round.FederationID)
	score := float64(round.RoundNumber)

	if err := r.client.ZAdd(r.ctx, federationKey, redis.Z{
		Score:  score,
		Member: round.ID,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add round to federation rounds: %w", err)
	}

	// Set TTL for the federation rounds list
	r.client.Expire(r.ctx, federationKey, r.getDefaultTTL())

	return nil
}

// GetRoundMetrics retrieves round metrics from Redis
func (r *RedisStorage) GetRoundMetrics(federationID string, limit int) ([]RoundMetrics, error) {
	federationKey := fmt.Sprintf("federation:%s:rounds", federationID)

	// Get round IDs from sorted set (highest round numbers first)
	var roundIDs []string
	var err error

	if limit > 0 {
		roundIDs, err = r.client.ZRevRange(r.ctx, federationKey, 0, int64(limit-1)).Result()
	} else {
		roundIDs, err = r.client.ZRevRange(r.ctx, federationKey, 0, -1).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get round list: %w", err)
	}

	var rounds []RoundMetrics
	for _, id := range roundIDs {
		key := fmt.Sprintf("round:%s", id)

		data, err := r.client.Get(r.ctx, key).Result()
		if err != nil {
			continue // Skip failed rounds
		}

		var round RoundMetrics
		if err := json.Unmarshal([]byte(data), &round); err != nil {
			continue // Skip invalid data
		}

		rounds = append(rounds, round)
	}

	return rounds, nil
}

// StoreResourceMetrics stores resource metrics in Redis using time series
func (r *RedisStorage) StoreResourceMetrics(metrics ResourceMetrics) error {
	// Use Redis Streams for time series data
	streamKey := "resource_metrics:system" // Use a fixed key since SourceID is not available

	values := map[string]interface{}{
		"source_type":  "system",
		"cpu_usage":    metrics.CPUUsage,
		"memory_usage": metrics.MemoryUsage,
		"disk_usage":   metrics.DiskUsage,
		"network_rx":   metrics.NetworkRxRate,
		"network_tx":   metrics.NetworkTxRate,
		"timestamp":    metrics.Timestamp.Unix(),
	}

	if err := r.client.XAdd(r.ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: values,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store resource metrics: %w", err)
	}

	// Set TTL for the stream
	r.client.Expire(r.ctx, streamKey, r.getDefaultTTL())

	// Trim stream to keep only recent entries (last 1000 entries)
	r.client.XTrimMaxLen(r.ctx, streamKey, 1000)

	return nil
}

// StoreEvent stores monitoring events in Redis
func (r *RedisStorage) StoreEvent(event MonitoringEvent) error {
	// Use Redis Streams for events
	streamKey := "events"
	if event.FederationID != "" {
		streamKey = fmt.Sprintf("events:%s", event.FederationID)
	}

	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	values := map[string]interface{}{
		"federation_id": event.FederationID,
		"type":          event.Type,
		"message":       event.Message,
		"level":         event.Level,
		"data":          string(dataJSON),
		"timestamp":     event.Timestamp.Unix(),
	}

	if err := r.client.XAdd(r.ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: values,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Set TTL for the stream
	r.client.Expire(r.ctx, streamKey, r.getDefaultTTL())

	// Trim stream to keep only recent entries (last 10000 events)
	r.client.XTrimMaxLen(r.ctx, streamKey, 10000)

	return nil
}

// GetEvents retrieves monitoring events from Redis
func (r *RedisStorage) GetEvents(federationID string, limit int, offset int) ([]MonitoringEvent, error) {
	streamKey := "events"
	if federationID != "" {
		streamKey = fmt.Sprintf("events:%s", federationID)
	}

	// Redis Streams don't support offset directly, so we'll get more and slice
	count := int64(limit + offset)
	if count <= 0 {
		count = 100 // Default limit
	}

	// Get events from stream (newest first)
	streams, err := r.client.XRevRangeN(r.ctx, streamKey, "+", "-", count).Result()
	if err != nil {
		if err == redis.Nil {
			return []MonitoringEvent{}, nil
		}
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	var events []MonitoringEvent

	// Apply offset
	start := offset
	if start > len(streams) {
		return []MonitoringEvent{}, nil
	}

	end := start + limit
	if end > len(streams) {
		end = len(streams)
	}

	for i := start; i < end; i++ {
		stream := streams[i]

		var event MonitoringEvent

		// Parse values from stream
		for field, value := range stream.Values {
			switch field {
			case "federation_id":
				if str, ok := value.(string); ok {
					event.FederationID = str
				}
			case "type":
				if str, ok := value.(string); ok {
					event.Type = MetricType(str)
				}
			case "message":
				if str, ok := value.(string); ok {
					event.Message = str
				}
			case "level":
				if str, ok := value.(string); ok {
					event.Level = str
				}
			case "data":
				if str, ok := value.(string); ok && str != "" {
					var data map[string]interface{}
					if err := json.Unmarshal([]byte(str), &data); err == nil {
						event.Data = data
					}
				}
			case "timestamp":
				if str, ok := value.(string); ok {
					if timestamp, err := strconv.ParseInt(str, 10, 64); err == nil {
						event.Timestamp = time.Unix(timestamp, 0)
					}
				}
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// Cleanup removes old data from Redis
func (r *RedisStorage) Cleanup(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge).Unix()

	// Find all resource metric streams
	keys, err := r.client.Keys(r.ctx, "resource_metrics:*").Result()
	if err != nil {
		return fmt.Errorf("failed to find resource metric keys: %w", err)
	}

	for _, key := range keys {
		// Remove old entries from streams
		if err := r.client.XTrimMinID(r.ctx, key, fmt.Sprintf("%d-0", cutoff)).Err(); err != nil {
			continue // Skip errors for individual streams
		}
	}

	// Find all event streams
	eventKeys, err := r.client.Keys(r.ctx, "events*").Result()
	if err != nil {
		return fmt.Errorf("failed to find event keys: %w", err)
	}

	for _, key := range eventKeys {
		// Remove old entries from event streams
		if err := r.client.XTrimMinID(r.ctx, key, fmt.Sprintf("%d-0", cutoff)).Err(); err != nil {
			continue // Skip errors for individual streams
		}
	}

	return nil
}

// GetStats returns Redis storage statistics
func (r *RedisStorage) GetStats() (map[string]interface{}, error) {
	info, err := r.client.Info(r.ctx, "memory", "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Count keys
	federationCount, _ := r.client.SCard(r.ctx, "federations:list").Result()

	stats := map[string]interface{}{
		"storage_type":      "redis",
		"server_info":       info,
		"federation_count":  federationCount,
		"connection_status": "connected",
	}

	return stats, nil
}
