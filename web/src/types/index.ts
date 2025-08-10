export interface FederationMetrics {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'stopped'
  mode: string
  algorithm: string
  start_time: string
  end_time?: string
  current_round: number
  total_rounds: number
  active_collaborators: number
  total_collaborators: number
  model_size: number
  last_update: string
  aggregator_address: string
}

export interface CollaboratorMetrics {
  id: string
  federation_id: string
  address: string
  status: 'connected' | 'disconnected' | 'training' | 'idle' | 'error'
  join_time: string
  last_seen: string
  current_round: number
  updates_submitted: number
  training_time: number
  average_latency_ms: number
  error_count: number
  last_error?: string
  resource_metrics?: ResourceMetrics
}

export interface RoundMetrics {
  id: string
  federation_id: string
  round_number: number
  algorithm: string
  start_time: string
  end_time?: string
  duration_ms: number
  participant_count: number
  updates_received: number
  aggregation_time_ms: number
  model_accuracy?: number
  model_loss?: number
  convergence_rate?: number
  status: string
}

export interface ModelUpdateMetrics {
  id: string
  federation_id: string
  collaborator_id: string
  round_number: number
  timestamp: string
  update_size_bytes: number
  processing_time_ms: number
  staleness?: number
  weight?: number
  quality_score?: number
  compression_ratio?: number
}

export interface ResourceMetrics {
  timestamp: string
  cpu_usage_percent: number
  memory_usage_percent: number
  memory_used_bytes: number
  memory_total_bytes: number
  disk_usage_percent: number
  network_rx_rate_mbps: number
  network_tx_rate_mbps: number
  gpu_usage_percent?: number
  gpu_memory_percent?: number
}

export interface MonitoringEvent {
  id: string
  federation_id: string
  type: string
  timestamp: string
  source: string
  level: string
  message: string
  data?: Record<string, any>
}

export interface SystemOverview {
  federation_id: string
  status: string
  total_collaborators: number
  active_collaborators: number
  current_round: number
  total_rounds: number
  progress_percent: number
  average_resource_usage?: ResourceMetrics
  recent_events: MonitoringEvent[]
  active_alerts: Alert[]
}

export interface Alert {
  id: string
  federation_id: string
  type: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  title: string
  message: string
  source: string
  created_at: string
  resolved_at?: string
  data?: Record<string, any>
}

export interface PerformanceInsights {
  federation_id: string
  overall_performance_score: number
  training_efficiency: number
  communication_efficiency: number
  resource_utilization: number
  bottleneck_analysis: string[]
  recommendations: string[]
}

export interface ConvergenceAnalysis {
  federation_id: string
  convergence_rate: number
  estimated_completion?: string
  model_accuracy_trend: Array<{
    round: number
    timestamp: string
    accuracy: number
  }>
  model_loss_trend: Array<{
    round: number
    timestamp: string
    loss: number
  }>
  participation_rate: number
  quality_metrics: Record<string, number>
}

export interface APIResponse<T> {
  success: boolean
  data?: T
  error?: string
  meta?: {
    page?: number
    per_page?: number
    total?: number
    total_pages?: number
  }
}
