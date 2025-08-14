package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// APIServer handles HTTP requests for the monitoring system
type APIServer struct {
	service  MonitoringService
	config   *MonitoringConfig
	router   *mux.Router
	upgrader websocket.Upgrader
}

// NewAPIServer creates a new API server instance
func NewAPIServer(service MonitoringService, config *MonitoringConfig) *APIServer {
	server := &APIServer{
		service: service,
		config:  config,
		router:  mux.NewRouter(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080", "http://127.0.0.1:3000", "http://127.0.0.1:8080"}
				if config.Production {
					allowedOrigins = config.AllowedOrigins
				}
				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}
				return false
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	server.setupRoutes()
	return server
}

// Start starts the API server
func (s *APIServer) Start() error {
	// Setup CORS with secure defaults
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080", "http://127.0.0.1:3000", "http://127.0.0.1:8080"}
	if s.config.Production {
		// In production, use specific origins
		allowedOrigins = s.config.AllowedOrigins
	}
	
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})

	handler := c.Handler(s.router)

	addr := fmt.Sprintf(":%d", s.config.APIPort)
	log.Printf("Starting monitoring API server on %s", addr)

	return http.ListenAndServe(addr, handler)
}

// setupRoutes configures all API routes
func (s *APIServer) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/stats", s.handleStats).Methods("GET")

	// Federation endpoints
	federations := api.PathPrefix("/federations").Subrouter()
	federations.HandleFunc("", s.handleListFederations).Methods("GET")
	federations.HandleFunc("", s.handleCreateFederation).Methods("POST")
	federations.HandleFunc("/{id}", s.handleGetFederation).Methods("GET")
	federations.HandleFunc("/{id}", s.handleUpdateFederation).Methods("PUT")
	federations.HandleFunc("/{id}/overview", s.handleGetSystemOverview).Methods("GET")
	federations.HandleFunc("/{id}/insights", s.handleGetPerformanceInsights).Methods("GET")
	federations.HandleFunc("/{id}/convergence", s.handleGetConvergenceAnalysis).Methods("GET")
	federations.HandleFunc("/{id}/efficiency", s.handleGetEfficiencyMetrics).Methods("GET")

	// Collaborator endpoints
	collaborators := api.PathPrefix("/collaborators").Subrouter()
	collaborators.HandleFunc("", s.handleListCollaborators).Methods("GET")
	collaborators.HandleFunc("", s.handleCreateCollaborator).Methods("POST")
	collaborators.HandleFunc("/{id}", s.handleGetCollaborator).Methods("GET")
	collaborators.HandleFunc("/{id}", s.handleUpdateCollaborator).Methods("PUT")

	// Round endpoints
	rounds := api.PathPrefix("/rounds").Subrouter()
	rounds.HandleFunc("", s.handleListRounds).Methods("GET")
	rounds.HandleFunc("", s.handleCreateRound).Methods("POST")
	rounds.HandleFunc("/{id}", s.handleGetRound).Methods("GET")
	rounds.HandleFunc("/{id}", s.handleUpdateRound).Methods("PUT")

	// Model update endpoints
	updates := api.PathPrefix("/updates").Subrouter()
	updates.HandleFunc("", s.handleListModelUpdates).Methods("GET")
	updates.HandleFunc("", s.handleCreateModelUpdate).Methods("POST")
	updates.HandleFunc("/statistics", s.handleGetUpdateStatistics).Methods("GET")

	// Aggregation endpoints
	aggregations := api.PathPrefix("/aggregations").Subrouter()
	aggregations.HandleFunc("", s.handleListAggregations).Methods("GET")
	aggregations.HandleFunc("", s.handleCreateAggregation).Methods("POST")
	aggregations.HandleFunc("/statistics", s.handleGetAggregationStatistics).Methods("GET")

	// Resource metrics endpoints
	resources := api.PathPrefix("/resources").Subrouter()
	resources.HandleFunc("/{source}", s.handleGetResourceMetrics).Methods("GET")
	resources.HandleFunc("/{source}", s.handleCreateResourceMetrics).Methods("POST")

	// Event endpoints
	events := api.PathPrefix("/events").Subrouter()
	events.HandleFunc("", s.handleListEvents).Methods("GET")
	events.HandleFunc("", s.handleCreateEvent).Methods("POST")
	events.HandleFunc("/alerts", s.handleGetActiveAlerts).Methods("GET")

	// Dashboard endpoints
	dashboards := api.PathPrefix("/dashboards").Subrouter()
	dashboards.HandleFunc("", s.handleListDashboards).Methods("GET")
	dashboards.HandleFunc("", s.handleCreateDashboard).Methods("POST")
	dashboards.HandleFunc("/{id}", s.handleGetDashboard).Methods("GET")
	dashboards.HandleFunc("/{id}", s.handleUpdateDashboard).Methods("PUT")
	dashboards.HandleFunc("/{id}", s.handleDeleteDashboard).Methods("DELETE")

	// WebSocket endpoint for real-time events
	api.HandleFunc("/ws", s.handleWebSocket).Methods("GET")

	// Serve static files for the web UI
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist/")))
}

// Health check endpoint
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.service.HealthCheck(ctx); err != nil {
		s.sendError(w, http.StatusServiceUnavailable, "Service unhealthy", err)
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// Stats endpoint
func (s *APIServer) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := s.service.GetMetricsStats(ctx)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	s.sendSuccess(w, stats)
}

// Federation handlers
func (s *APIServer) handleListFederations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseMetricsFilter(r)

	if r.URL.Query().Get("active") == "true" {
		federations, err := s.service.GetActiveFederations(ctx)
		if err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to get active federations", err)
			return
		}
		s.sendSuccess(w, federations)
		return
	}

	federations, err := s.service.GetFederationHistory(ctx, filter)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get federation history", err)
		return
	}

	s.sendSuccess(w, federations)
}

func (s *APIServer) handleCreateFederation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var federation FederationMetrics
	if err := json.NewDecoder(r.Body).Decode(&federation); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RegisterFederation(ctx, &federation); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to register federation", err)
		return
	}

	s.sendSuccess(w, federation)
}

func (s *APIServer) handleGetFederation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	federation, err := s.service.GetFederation(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Federation not found", err)
		return
	}

	s.sendSuccess(w, federation)
}

func (s *APIServer) handleUpdateFederation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var federation FederationMetrics
	if err := json.NewDecoder(r.Body).Decode(&federation); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.UpdateFederation(ctx, id, &federation); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to update federation", err)
		return
	}

	s.sendSuccess(w, federation)
}

func (s *APIServer) handleGetSystemOverview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	overview, err := s.service.GetSystemOverview(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get system overview", err)
		return
	}

	s.sendSuccess(w, overview)
}

func (s *APIServer) handleGetPerformanceInsights(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	insights, err := s.service.GetPerformanceInsights(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get performance insights", err)
		return
	}

	s.sendSuccess(w, insights)
}

func (s *APIServer) handleGetConvergenceAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	analysis, err := s.service.GetConvergenceAnalysis(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get convergence analysis", err)
		return
	}

	s.sendSuccess(w, analysis)
}

func (s *APIServer) handleGetEfficiencyMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	metrics, err := s.service.GetEfficiencyMetrics(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get efficiency metrics", err)
		return
	}

	s.sendSuccess(w, metrics)
}

// Collaborator handlers
func (s *APIServer) handleListCollaborators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseMetricsFilter(r)

	// Check if filtering by federation
	if federationID := r.URL.Query().Get("federation_id"); federationID != "" {
		collaborators, err := s.service.GetFederationCollaborators(ctx, federationID)
		if err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to get federation collaborators", err)
			return
		}
		s.sendSuccess(w, collaborators)
		return
	}

	collaborators, err := s.service.GetCollaboratorHistory(ctx, filter)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get collaborator history", err)
		return
	}

	s.sendSuccess(w, collaborators)
}

func (s *APIServer) handleCreateCollaborator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var collaborator CollaboratorMetrics
	if err := json.NewDecoder(r.Body).Decode(&collaborator); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RegisterCollaborator(ctx, &collaborator); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to register collaborator", err)
		return
	}

	s.sendSuccess(w, collaborator)
}

func (s *APIServer) handleGetCollaborator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	collaborator, err := s.service.GetCollaborator(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Collaborator not found", err)
		return
	}

	s.sendSuccess(w, collaborator)
}

func (s *APIServer) handleUpdateCollaborator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var collaborator CollaboratorMetrics
	if err := json.NewDecoder(r.Body).Decode(&collaborator); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.UpdateCollaborator(ctx, id, &collaborator); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to update collaborator", err)
		return
	}

	s.sendSuccess(w, collaborator)
}

// Round handlers
func (s *APIServer) handleListRounds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseMetricsFilter(r)

	// Check if filtering by federation
	if federationID := r.URL.Query().Get("federation_id"); federationID != "" {
		rounds, err := s.service.GetFederationRounds(ctx, federationID)
		if err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to get federation rounds", err)
			return
		}
		s.sendSuccess(w, rounds)
		return
	}

	rounds, err := s.service.GetRoundHistory(ctx, filter)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get round history", err)
		return
	}

	s.sendSuccess(w, rounds)
}

func (s *APIServer) handleCreateRound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var round RoundMetrics
	if err := json.NewDecoder(r.Body).Decode(&round); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RecordRoundStart(ctx, &round); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to record round start", err)
		return
	}

	s.sendSuccess(w, round)
}

func (s *APIServer) handleGetRound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	round, err := s.service.GetRound(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Round not found", err)
		return
	}

	s.sendSuccess(w, round)
}

func (s *APIServer) handleUpdateRound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var round RoundMetrics
	if err := json.NewDecoder(r.Body).Decode(&round); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RecordRoundEnd(ctx, id, &round); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to record round end", err)
		return
	}

	s.sendSuccess(w, round)
}

// Model update handlers
func (s *APIServer) handleListModelUpdates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseMetricsFilter(r)

	updates, err := s.service.GetModelUpdates(ctx, filter)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get model updates", err)
		return
	}

	s.sendSuccess(w, updates)
}

func (s *APIServer) handleCreateModelUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var update ModelUpdateMetrics
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RecordModelUpdate(ctx, &update); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to record model update", err)
		return
	}

	s.sendSuccess(w, update)
}

func (s *APIServer) handleGetUpdateStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	federationID := r.URL.Query().Get("federation_id")
	roundNumberStr := r.URL.Query().Get("round_number")

	if federationID == "" || roundNumberStr == "" {
		s.sendError(w, http.StatusBadRequest, "federation_id and round_number are required", nil)
		return
	}

	roundNumber, err := strconv.Atoi(roundNumberStr)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid round_number", err)
		return
	}

	stats, err := s.service.GetUpdateStatistics(ctx, federationID, roundNumber)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get update statistics", err)
		return
	}

	s.sendSuccess(w, stats)
}

// Aggregation handlers
func (s *APIServer) handleListAggregations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseMetricsFilter(r)

	aggregations, err := s.service.GetAggregations(ctx, filter)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get aggregations", err)
		return
	}

	s.sendSuccess(w, aggregations)
}

func (s *APIServer) handleCreateAggregation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var aggregation AggregationMetrics
	if err := json.NewDecoder(r.Body).Decode(&aggregation); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RecordAggregation(ctx, &aggregation); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to record aggregation", err)
		return
	}

	s.sendSuccess(w, aggregation)
}

func (s *APIServer) handleGetAggregationStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	federationID := r.URL.Query().Get("federation_id")
	if federationID == "" {
		s.sendError(w, http.StatusBadRequest, "federation_id is required", nil)
		return
	}

	stats, err := s.service.GetAggregationStatistics(ctx, federationID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get aggregation statistics", err)
		return
	}

	s.sendSuccess(w, stats)
}

// Resource metrics handlers
func (s *APIServer) handleGetResourceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source := mux.Vars(r)["source"]

	// Parse time range (default to last hour)
	timeRangeStr := r.URL.Query().Get("time_range")
	timeRange := time.Hour
	if timeRangeStr != "" {
		if parsed, err := time.ParseDuration(timeRangeStr); err == nil {
			timeRange = parsed
		}
	}

	metrics, err := s.service.GetResourceMetrics(ctx, source, timeRange)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get resource metrics", err)
		return
	}

	s.sendSuccess(w, metrics)
}

func (s *APIServer) handleCreateResourceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source := mux.Vars(r)["source"]

	var metrics ResourceMetrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RecordResourceMetrics(ctx, source, &metrics); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to record resource metrics", err)
		return
	}

	s.sendSuccess(w, metrics)
}

// Event handlers
func (s *APIServer) handleListEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseMetricsFilter(r)

	events, err := s.service.GetEvents(ctx, filter)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get events", err)
		return
	}

	s.sendSuccess(w, events)
}

func (s *APIServer) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var event MonitoringEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.RecordEvent(ctx, &event); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to record event", err)
		return
	}

	s.sendSuccess(w, event)
}

func (s *APIServer) handleGetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	federationID := r.URL.Query().Get("federation_id")
	if federationID == "" {
		s.sendError(w, http.StatusBadRequest, "federation_id is required", nil)
		return
	}

	alerts, err := s.service.GetActiveAlerts(ctx, federationID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get active alerts", err)
		return
	}

	s.sendSuccess(w, alerts)
}

// Dashboard handlers
func (s *APIServer) handleListDashboards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dashboards, err := s.service.ListDashboards(ctx)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get dashboards", err)
		return
	}

	s.sendSuccess(w, dashboards)
}

func (s *APIServer) handleCreateDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var dashboard Dashboard
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.CreateDashboard(ctx, &dashboard); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to create dashboard", err)
		return
	}

	s.sendSuccess(w, dashboard)
}

func (s *APIServer) handleGetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	dashboard, err := s.service.GetDashboard(ctx, id)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Dashboard not found", err)
		return
	}

	s.sendSuccess(w, dashboard)
}

func (s *APIServer) handleUpdateDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var dashboard Dashboard
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := s.service.UpdateDashboard(ctx, id, &dashboard); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to update dashboard", err)
		return
	}

	s.sendSuccess(w, dashboard)
}

func (s *APIServer) handleDeleteDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	if err := s.service.DeleteDashboard(ctx, id); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to delete dashboard", err)
		return
	}

	s.sendSuccess(w, map[string]string{"message": "Dashboard deleted successfully"})
}

// WebSocket handler for real-time events
func (s *APIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Parse query parameters
	federationID := r.URL.Query().Get("federation_id")
	eventTypesStr := r.URL.Query().Get("event_types")

	var eventTypes []MetricType
	if eventTypesStr != "" {
		for _, typeStr := range strings.Split(eventTypesStr, ",") {
			eventTypes = append(eventTypes, MetricType(strings.TrimSpace(typeStr)))
		}
	}

	// Subscribe to events
	ctx := context.Background()
	eventChan, err := s.service.SubscribeToEvents(ctx, federationID, eventTypes)
	if err != nil {
		log.Printf("Failed to subscribe to events: %v", err)
		return
	}

	// Handle WebSocket communication
	go func() {
		for {
			// Read message from client (for keepalive or unsubscribe)
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}
		}
	}()

	// Send events to client
	for event := range eventChan {
		if err := conn.WriteJSON(event); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// Helper methods
func (s *APIServer) parseMetricsFilter(r *http.Request) *MetricsFilter {
	filter := &MetricsFilter{}

	if federationID := r.URL.Query().Get("federation_id"); federationID != "" {
		filter.FederationID = federationID
	}

	if collaboratorID := r.URL.Query().Get("collaborator_id"); collaboratorID != "" {
		filter.CollaboratorID = collaboratorID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}

	if metricType := r.URL.Query().Get("metric_type"); metricType != "" {
		filter.MetricType = MetricType(metricType)
	}

	if roundNumberStr := r.URL.Query().Get("round_number"); roundNumberStr != "" {
		if roundNumber, err := strconv.Atoi(roundNumberStr); err == nil {
			filter.RoundNumber = &roundNumber
		}
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = page
		}
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil {
			filter.PerPage = perPage
		}
	}

	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}

	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	return filter
}

func (s *APIServer) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := APIResponse{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *APIServer) sendError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}

	response := APIResponse{
		Success: false,
		Error:   errorMsg,
	}

	json.NewEncoder(w).Encode(response)
}
