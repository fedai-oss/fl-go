package monitoring

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLStorage implements Storage interface using PostgreSQL
type PostgreSQLStorage struct {
	db     *sql.DB
	config DatabaseConfig
}

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int    `yaml:"max_connections"`
}

// NewPostgreSQLStorage creates a new PostgreSQL storage backend
func NewPostgreSQLStorage(config DatabaseConfig) (*PostgreSQLStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	if config.MaxConns > 0 {
		db.SetMaxOpenConns(config.MaxConns)
		db.SetMaxIdleConns(config.MaxConns / 2)
	}
	db.SetConnMaxLifetime(time.Hour)

	storage := &PostgreSQLStorage{
		db:     db,
		config: config,
	}

	// Initialize database schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary database tables
func (p *PostgreSQLStorage) initSchema() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS federations (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			mode VARCHAR(20) NOT NULL,
			algorithm VARCHAR(50) NOT NULL,
			current_round INTEGER NOT NULL DEFAULT 0,
			total_rounds INTEGER NOT NULL DEFAULT 0,
			active_collaborators INTEGER NOT NULL DEFAULT 0,
			total_collaborators INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS collaborators (
			id VARCHAR(255) PRIMARY KEY,
			federation_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			address VARCHAR(255),
			last_seen TIMESTAMP WITH TIME ZONE,
			updates_submitted INTEGER NOT NULL DEFAULT 0,
			errors INTEGER NOT NULL DEFAULT 0,
			avg_training_time REAL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (federation_id) REFERENCES federations(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS rounds (
			id VARCHAR(255) PRIMARY KEY,
			federation_id VARCHAR(255) NOT NULL,
			round_number INTEGER NOT NULL,
			algorithm VARCHAR(50) NOT NULL,
			participants INTEGER NOT NULL DEFAULT 0,
			start_time TIMESTAMP WITH TIME ZONE,
			end_time TIMESTAMP WITH TIME ZONE,
			duration_seconds REAL DEFAULT 0,
			updates_received INTEGER NOT NULL DEFAULT 0,
			accuracy REAL DEFAULT 0,
			loss REAL DEFAULT 0,
			convergence_rate REAL DEFAULT 0,
			communication_cost REAL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (federation_id) REFERENCES federations(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS model_updates (
			id VARCHAR(255) PRIMARY KEY,
			federation_id VARCHAR(255) NOT NULL,
			collaborator_id VARCHAR(255) NOT NULL,
			round_number INTEGER NOT NULL,
			update_size INTEGER NOT NULL DEFAULT 0,
			processing_time REAL DEFAULT 0,
			staleness INTEGER DEFAULT 0,
			weight REAL DEFAULT 1.0,
			accuracy REAL DEFAULT 0,
			loss REAL DEFAULT 0,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (federation_id) REFERENCES federations(id) ON DELETE CASCADE,
			FOREIGN KEY (collaborator_id) REFERENCES collaborators(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS resource_metrics (
			id SERIAL PRIMARY KEY,
			source_id VARCHAR(255) NOT NULL,
			source_type VARCHAR(50) NOT NULL,
			cpu_usage REAL DEFAULT 0,
			memory_usage REAL DEFAULT 0,
			disk_usage REAL DEFAULT 0,
			network_in REAL DEFAULT 0,
			network_out REAL DEFAULT 0,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS events (
			id SERIAL PRIMARY KEY,
			federation_id VARCHAR(255),
			event_type VARCHAR(100) NOT NULL,
			description TEXT,
			severity VARCHAR(20) DEFAULT 'info',
			metadata JSONB,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (federation_id) REFERENCES federations(id) ON DELETE CASCADE
		)`,

		// Indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_collaborators_federation ON collaborators(federation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rounds_federation ON rounds(federation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_model_updates_federation ON model_updates(federation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_model_updates_collaborator ON model_updates(collaborator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_metrics_source ON resource_metrics(source_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_events_federation ON events(federation_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp)`,
	}

	for _, schema := range schemas {
		if _, err := p.db.Exec(schema); err != nil {
			return fmt.Errorf("failed to execute schema: %s, error: %w", schema, err)
		}
	}

	return nil
}

// StoreFederationMetrics stores federation metrics in PostgreSQL
func (p *PostgreSQLStorage) StoreFederationMetrics(federation FederationMetrics) error {
	query := `
		INSERT INTO federations (id, name, status, mode, algorithm, current_round, total_rounds, active_collaborators, total_collaborators, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			mode = EXCLUDED.mode,
			algorithm = EXCLUDED.algorithm,
			current_round = EXCLUDED.current_round,
			total_rounds = EXCLUDED.total_rounds,
			active_collaborators = EXCLUDED.active_collaborators,
			total_collaborators = EXCLUDED.total_collaborators,
			updated_at = NOW()
	`

	_, err := p.db.Exec(query, federation.ID, federation.Name, federation.Status, federation.Mode,
		federation.Algorithm, federation.CurrentRound, federation.TotalRounds,
		federation.ActiveCollabs, federation.TotalCollabs)

	return err
}

// GetFederationMetrics retrieves federation metrics from PostgreSQL
func (p *PostgreSQLStorage) GetFederationMetrics(id string) (*FederationMetrics, error) {
	query := `
		SELECT id, name, status, mode, algorithm, current_round, total_rounds, 
		       active_collaborators, total_collaborators, created_at, updated_at
		FROM federations WHERE id = $1
	`

	var federation FederationMetrics
	var createdAt, updatedAt time.Time

	err := p.db.QueryRow(query, id).Scan(
		&federation.ID, &federation.Name, &federation.Status, &federation.Mode,
		&federation.Algorithm, &federation.CurrentRound, &federation.TotalRounds,
		&federation.ActiveCollabs, &federation.TotalCollabs,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	federation.StartTime = createdAt
	federation.LastUpdate = updatedAt

	return &federation, nil
}

// ListFederations lists all federations with optional filters
func (p *PostgreSQLStorage) ListFederations(activeOnly bool) ([]FederationMetrics, error) {
	query := `
		SELECT id, name, status, mode, algorithm, current_round, total_rounds,
		       active_collaborators, total_collaborators, created_at, updated_at
		FROM federations
	`

	if activeOnly {
		query += " WHERE status = 'running'"
	}

	query += " ORDER BY created_at DESC"

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var federations []FederationMetrics
	for rows.Next() {
		var federation FederationMetrics
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&federation.ID, &federation.Name, &federation.Status, &federation.Mode,
			&federation.Algorithm, &federation.CurrentRound, &federation.TotalRounds,
			&federation.ActiveCollabs, &federation.TotalCollabs,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		federation.StartTime = createdAt
		federation.LastUpdate = updatedAt
		federations = append(federations, federation)
	}

	return federations, rows.Err()
}

// StoreCollaboratorMetrics stores collaborator metrics in PostgreSQL
func (p *PostgreSQLStorage) StoreCollaboratorMetrics(collaborator CollaboratorMetrics) error {
	query := `
		INSERT INTO collaborators (id, federation_id, name, status, address, last_seen, updates_submitted, errors, avg_training_time, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (id) DO UPDATE SET
			federation_id = EXCLUDED.federation_id,
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			address = EXCLUDED.address,
			last_seen = EXCLUDED.last_seen,
			updates_submitted = EXCLUDED.updates_submitted,
			errors = EXCLUDED.errors,
			avg_training_time = EXCLUDED.avg_training_time,
			updated_at = NOW()
	`

	// Extract training time as float64 seconds
	trainingTimeSeconds := collaborator.TrainingTime.Seconds()

	_, err := p.db.Exec(query, collaborator.ID, collaborator.FederationID, collaborator.ID, // Use ID as name for now
		collaborator.Status, collaborator.Address, collaborator.LastSeen,
		collaborator.UpdatesSubmitted, collaborator.ErrorCount, trainingTimeSeconds)

	return err
}

// GetCollaboratorMetrics retrieves collaborator metrics from PostgreSQL
func (p *PostgreSQLStorage) GetCollaboratorMetrics(federationID string) ([]CollaboratorMetrics, error) {
	query := `
		SELECT id, federation_id, name, status, address, last_seen, updates_submitted, errors, avg_training_time, created_at, updated_at
		FROM collaborators WHERE federation_id = $1 ORDER BY created_at
	`

	rows, err := p.db.Query(query, federationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []CollaboratorMetrics
	for rows.Next() {
		var collaborator CollaboratorMetrics
		var lastSeen sql.NullTime
		var createdAt, updatedAt time.Time
		var name string
		var errors int
		var avgTrainingTimeSeconds float64

		err := rows.Scan(
			&collaborator.ID, &collaborator.FederationID, &name,
			&collaborator.Status, &collaborator.Address, &lastSeen,
			&collaborator.UpdatesSubmitted, &errors, &avgTrainingTimeSeconds,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastSeen.Valid {
			collaborator.LastSeen = lastSeen.Time
		}
		collaborator.JoinTime = createdAt
		collaborator.ErrorCount = errors
		collaborator.TrainingTime = time.Duration(avgTrainingTimeSeconds * float64(time.Second))

		collaborators = append(collaborators, collaborator)
	}

	return collaborators, rows.Err()
}

// StoreRoundMetrics stores round metrics in PostgreSQL
func (p *PostgreSQLStorage) StoreRoundMetrics(round RoundMetrics) error {
	query := `
		INSERT INTO rounds (id, federation_id, round_number, algorithm, participants, start_time, end_time, duration_seconds, updates_received, accuracy, loss, convergence_rate, communication_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			federation_id = EXCLUDED.federation_id,
			round_number = EXCLUDED.round_number,
			algorithm = EXCLUDED.algorithm,
			participants = EXCLUDED.participants,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			duration_seconds = EXCLUDED.duration_seconds,
			updates_received = EXCLUDED.updates_received,
			accuracy = EXCLUDED.accuracy,
			loss = EXCLUDED.loss,
			convergence_rate = EXCLUDED.convergence_rate,
			communication_cost = EXCLUDED.communication_cost
	`

	// Handle optional fields
	var accuracy, loss, convergenceRate interface{}
	if round.ModelAccuracy != nil {
		accuracy = *round.ModelAccuracy
	}
	if round.ModelLoss != nil {
		loss = *round.ModelLoss
	}
	if round.ConvergenceRate != nil {
		convergenceRate = *round.ConvergenceRate
	}

	_, err := p.db.Exec(query, round.ID, round.FederationID, round.RoundNumber, round.Algorithm,
		round.ParticipantCount, round.StartTime, round.EndTime, round.Duration.Seconds(),
		round.UpdatesReceived, accuracy, loss, convergenceRate, 0.0) // communication_cost placeholder

	return err
}

// GetRoundMetrics retrieves round metrics from PostgreSQL
func (p *PostgreSQLStorage) GetRoundMetrics(federationID string, limit int) ([]RoundMetrics, error) {
	query := `
		SELECT id, federation_id, round_number, algorithm, participants, start_time, end_time, duration_seconds, updates_received, accuracy, loss, convergence_rate, communication_cost, created_at
		FROM rounds WHERE federation_id = $1 ORDER BY round_number DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := p.db.Query(query, federationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rounds []RoundMetrics
	for rows.Next() {
		var round RoundMetrics
		var startTime, endTime sql.NullTime
		var durationSeconds float64
		var createdAt time.Time
		var accuracy, loss, convergenceRate sql.NullFloat64
		var communicationCost float64

		err := rows.Scan(
			&round.ID, &round.FederationID, &round.RoundNumber, &round.Algorithm,
			&round.ParticipantCount, &startTime, &endTime, &durationSeconds,
			&round.UpdatesReceived, &accuracy, &loss,
			&convergenceRate, &communicationCost, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		if startTime.Valid {
			round.StartTime = startTime.Time
		}
		if endTime.Valid {
			round.EndTime = &endTime.Time
		}
		round.Duration = time.Duration(durationSeconds * float64(time.Second))

		// Handle optional fields
		if accuracy.Valid {
			round.ModelAccuracy = &accuracy.Float64
		}
		if loss.Valid {
			round.ModelLoss = &loss.Float64
		}
		if convergenceRate.Valid {
			round.ConvergenceRate = &convergenceRate.Float64
		}

		rounds = append(rounds, round)
	}

	return rounds, rows.Err()
}

// StoreResourceMetrics stores resource metrics in PostgreSQL
func (p *PostgreSQLStorage) StoreResourceMetrics(metrics ResourceMetrics) error {
	query := `
		INSERT INTO resource_metrics (source_id, source_type, cpu_usage, memory_usage, disk_usage, network_in, network_out, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	// Use placeholder values for source_id and source_type since they're not in ResourceMetrics
	sourceID := "unknown"
	sourceType := "system"

	_, err := p.db.Exec(query, sourceID, sourceType, metrics.CPUUsage,
		metrics.MemoryUsage, metrics.DiskUsage, metrics.NetworkRxRate, metrics.NetworkTxRate, metrics.Timestamp)

	return err
}

// StoreEvent stores monitoring events in PostgreSQL
func (p *PostgreSQLStorage) StoreEvent(event MonitoringEvent) error {
	metadataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	query := `
		INSERT INTO events (federation_id, event_type, description, severity, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = p.db.Exec(query, event.FederationID, event.Type, event.Message,
		event.Level, metadataJSON, event.Timestamp)

	return err
}

// GetEvents retrieves monitoring events from PostgreSQL
func (p *PostgreSQLStorage) GetEvents(federationID string, limit int, offset int) ([]MonitoringEvent, error) {
	query := `
		SELECT federation_id, event_type, description, severity, metadata, timestamp
		FROM events
	`

	args := []interface{}{}
	argCount := 0

	if federationID != "" {
		argCount++
		query += fmt.Sprintf(" WHERE federation_id = $%d", argCount)
		args = append(args, federationID)
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []MonitoringEvent
	for rows.Next() {
		var event MonitoringEvent
		var metadataJSON []byte
		var federationID sql.NullString

		err := rows.Scan(
			&federationID, &event.Type, &event.Message,
			&event.Level, &metadataJSON, &event.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		if federationID.Valid {
			event.FederationID = federationID.String
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Data); err != nil {
				// Log error but don't fail the query
				event.Data = map[string]interface{}{"error": "failed to unmarshal metadata"}
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// Close closes the PostgreSQL database connection
func (p *PostgreSQLStorage) Close() error {
	return p.db.Close()
}

// Cleanup removes old data from the database
func (p *PostgreSQLStorage) Cleanup(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	queries := []string{
		"DELETE FROM resource_metrics WHERE timestamp < $1",
		"DELETE FROM events WHERE timestamp < $1",
		"DELETE FROM model_updates WHERE timestamp < $1",
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query, cutoff); err != nil {
			return fmt.Errorf("cleanup failed for query %s: %w", query, err)
		}
	}

	return nil
}
