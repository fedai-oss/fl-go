import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { 
  Server, 
  Users, 
  Clock, 
  Activity,
  ExternalLink,
  Filter
} from 'lucide-react'
import { apiClient, queryKeys } from '../lib/api'
import type { FederationMetrics } from '../types'

export default function Federations() {
  const { data: federations, isLoading, error } = useQuery({
    queryKey: queryKeys.federations(),
    queryFn: () => apiClient.getFederations(),
  })

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-gray-900">Federations</h1>
        </div>
        <div className="animate-pulse space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="bg-gray-200 rounded-lg h-32"></div>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Federations</h1>
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-800">Error loading federations: {(error as Error).message}</p>
        </div>
      </div>
    )
  }

  const activeFederations = federations?.filter(f => f.status === 'running') || []
  const completedFederations = federations?.filter(f => f.status === 'completed') || []
  const otherFederations = federations?.filter(f => !['running', 'completed'].includes(f.status)) || []

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Federations</h1>
        <div className="flex items-center space-x-4">
          <button className="flex items-center space-x-2 px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50">
            <Filter className="w-4 h-4" />
            <span>Filter</span>
          </button>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-50 rounded-lg">
              <Server className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total</p>
              <p className="text-2xl font-semibold text-gray-900">{federations?.length || 0}</p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-green-50 rounded-lg">
              <Activity className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Active</p>
              <p className="text-2xl font-semibold text-gray-900">{activeFederations.length}</p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-purple-50 rounded-lg">
              <Clock className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Completed</p>
              <p className="text-2xl font-semibold text-gray-900">{completedFederations.length}</p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-yellow-50 rounded-lg">
              <Users className="w-5 h-5 text-yellow-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Collaborators</p>
              <p className="text-2xl font-semibold text-gray-900">
                {federations?.reduce((sum, f) => sum + f.total_collaborators, 0) || 0}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Active Federations */}
      {activeFederations.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold text-gray-900">Active Federations</h2>
          <div className="space-y-4">
            {activeFederations.map(federation => (
              <FederationCard key={federation.id} federation={federation} />
            ))}
          </div>
        </div>
      )}

      {/* Completed Federations */}
      {completedFederations.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold text-gray-900">Completed Federations</h2>
          <div className="space-y-4">
            {completedFederations.map(federation => (
              <FederationCard key={federation.id} federation={federation} />
            ))}
          </div>
        </div>
      )}

      {/* Other Federations */}
      {otherFederations.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold text-gray-900">Other Federations</h2>
          <div className="space-y-4">
            {otherFederations.map(federation => (
              <FederationCard key={federation.id} federation={federation} />
            ))}
          </div>
        </div>
      )}

      {!federations || federations.length === 0 ? (
        <div className="text-center py-12">
          <Server className="w-12 h-12 mx-auto mb-4 text-gray-300" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No federations found</h3>
          <p className="text-gray-500">Start a federated learning session to see it here.</p>
        </div>
      ) : null}
    </div>
  )
}

function FederationCard({ federation }: { federation: FederationMetrics }) {
  const progress = federation.total_rounds > 0 
    ? (federation.current_round / federation.total_rounds) * 100 
    : 0

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'bg-green-100 text-green-800'
      case 'completed':
        return 'bg-blue-100 text-blue-800'
      case 'failed':
        return 'bg-red-100 text-red-800'
      case 'pending':
        return 'bg-yellow-100 text-yellow-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDuration = (startTime: string, endTime?: string) => {
    const start = new Date(startTime)
    const end = endTime ? new Date(endTime) : new Date()
    const diffMs = end.getTime() - start.getTime()
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffMinutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
    
    if (diffHours > 0) {
      return `${diffHours}h ${diffMinutes}m`
    }
    return `${diffMinutes}m`
  }

  return (
    <div className="card hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1">
          <div className="flex items-center space-x-3 mb-2">
            <h3 className="text-lg font-semibold text-gray-900">{federation.name}</h3>
            <span className={`status-badge ${getStatusColor(federation.status)}`}>
              {federation.status}
            </span>
          </div>
          <p className="text-sm text-gray-600 mb-3">ID: {federation.id}</p>
          
          {/* Key metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wide">Algorithm</p>
              <p className="text-sm font-medium text-gray-900">{federation.algorithm}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wide">Mode</p>
              <p className="text-sm font-medium text-gray-900">{federation.mode}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wide">Collaborators</p>
              <p className="text-sm font-medium text-gray-900">
                {federation.active_collaborators}/{federation.total_collaborators}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wide">Duration</p>
              <p className="text-sm font-medium text-gray-900">
                {formatDuration(federation.start_time, federation.end_time)}
              </p>
            </div>
          </div>
        </div>
        
        <Link
          to={`/federations/${federation.id}`}
          className="ml-4 p-2 text-gray-400 hover:text-gray-600 transition-colors"
          title="View details"
        >
          <ExternalLink className="w-5 h-5" />
        </Link>
      </div>

      {/* Progress */}
      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span className="text-gray-600">Training Progress</span>
          <span className="font-medium text-gray-900">
            Round {federation.current_round} of {federation.total_rounds}
          </span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div 
            className="bg-primary-500 h-2 rounded-full transition-all duration-300"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
        <div className="flex justify-between text-xs text-gray-500">
          <span>Started {new Date(federation.start_time).toLocaleDateString()}</span>
          <span>{progress.toFixed(1)}% complete</span>
        </div>
      </div>

      {/* Additional info */}
      <div className="mt-4 pt-4 border-t border-gray-100 flex items-center justify-between text-sm text-gray-500">
        <span>Aggregator: {federation.aggregator_address}</span>
        <span>Updated {new Date(federation.last_update).toLocaleTimeString()}</span>
      </div>
    </div>
  )
}
