import type {
  APIResponse,
  FederationMetrics,
  CollaboratorMetrics,
  RoundMetrics,
  MonitoringEvent,
  SystemOverview,
  PerformanceInsights,
  ConvergenceAnalysis,
  ResourceMetrics,
} from '../types'

const API_BASE = '/api/v1'

class APIClient {
  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${API_BASE}${endpoint}`
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
        ...options,
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result: APIResponse<T> = await response.json()
      
      if (!result.success) {
        throw new Error(result.error || 'API request failed')
      }

      return result.data as T
    } catch (error) {
      console.error(`API request failed for ${endpoint}:`, error)
      throw error
    }
  }

  // Health check
  async healthCheck() {
    return this.request<{ status: string; timestamp: string; version: string }>('/health')
  }

  // Federation endpoints
  async getFederations(active = false) {
    const endpoint = active ? '/federations?active=true' : '/federations'
    return this.request<FederationMetrics[]>(endpoint)
  }

  async getFederation(id: string) {
    return this.request<FederationMetrics>(`/federations/${id}`)
  }

  async getSystemOverview(federationId: string) {
    return this.request<SystemOverview>(`/federations/${federationId}/overview`)
  }

  async getPerformanceInsights(federationId: string) {
    return this.request<PerformanceInsights>(`/federations/${federationId}/insights`)
  }

  async getConvergenceAnalysis(federationId: string) {
    return this.request<ConvergenceAnalysis>(`/federations/${federationId}/convergence`)
  }

  // Collaborator endpoints
  async getCollaborators(federationId?: string) {
    const endpoint = federationId 
      ? `/collaborators?federation_id=${federationId}` 
      : '/collaborators'
    return this.request<CollaboratorMetrics[]>(endpoint)
  }

  async getCollaborator(id: string) {
    return this.request<CollaboratorMetrics>(`/collaborators/${id}`)
  }

  // Round endpoints
  async getRounds(federationId?: string) {
    const endpoint = federationId 
      ? `/rounds?federation_id=${federationId}` 
      : '/rounds'
    return this.request<RoundMetrics[]>(endpoint)
  }

  async getRound(id: string) {
    return this.request<RoundMetrics>(`/rounds/${id}`)
  }

  // Event endpoints
  async getEvents(filters?: {
    federation_id?: string
    metric_type?: string
    start_time?: string
    end_time?: string
    page?: number
    per_page?: number
  }) {
    const params = new URLSearchParams()
    
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, value.toString())
        }
      })
    }
    
    const endpoint = `/events${params.toString() ? `?${params.toString()}` : ''}`
    return this.request<MonitoringEvent[]>(endpoint)
  }

  // Resource metrics
  async getResourceMetrics(source: string, timeRange = '1h') {
    return this.request<ResourceMetrics[]>(`/resources/${source}?time_range=${timeRange}`)
  }

  // WebSocket connection for real-time updates
  createWebSocket(federationId?: string, eventTypes?: string[]) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsHost = window.location.host
    
    let wsUrl = `${wsProtocol}//${wsHost}${API_BASE}/ws`
    
    const params = new URLSearchParams()
    if (federationId) {
      params.append('federation_id', federationId)
    }
    if (eventTypes && eventTypes.length > 0) {
      params.append('event_types', eventTypes.join(','))
    }
    
    if (params.toString()) {
      wsUrl += `?${params.toString()}`
    }
    
    return new WebSocket(wsUrl)
  }
}

export const apiClient = new APIClient()

// React Query hooks
export const queryKeys = {
  health: () => ['health'] as const,
  federations: () => ['federations'] as const,
  activeFederations: () => ['federations', 'active'] as const,
  federation: (id: string) => ['federations', id] as const,
  systemOverview: (id: string) => ['federations', id, 'overview'] as const,
  performanceInsights: (id: string) => ['federations', id, 'insights'] as const,
  convergenceAnalysis: (id: string) => ['federations', id, 'convergence'] as const,
  collaborators: (federationId?: string) => 
    federationId ? ['collaborators', federationId] : ['collaborators'] as const,
  collaborator: (id: string) => ['collaborators', id] as const,
  rounds: (federationId?: string) => 
    federationId ? ['rounds', federationId] : ['rounds'] as const,
  round: (id: string) => ['rounds', id] as const,
  events: (filters?: any) => ['events', filters] as const,
  resourceMetrics: (source: string, timeRange?: string) => 
    ['resources', source, timeRange] as const,
}
