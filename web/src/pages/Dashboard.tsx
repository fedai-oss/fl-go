import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { 
  Activity, 
  Users, 
  Server, 
  BarChart3, 
  TrendingUp, 
  Clock,
  AlertTriangle,
  CheckCircle
} from 'lucide-react'
import { apiClient, queryKeys } from '../lib/api'
import type { FederationMetrics } from '../types'

export default function Dashboard() {
  const { data: activeFederations, isLoading: federationsLoading } = useQuery({
    queryKey: queryKeys.activeFederations(),
    queryFn: () => apiClient.getFederations(true),
  })

  const { data: allFederations } = useQuery({
    queryKey: queryKeys.federations(),
    queryFn: () => apiClient.getFederations(),
  })

  // Calculate summary metrics
  const totalFederations = allFederations?.length || 0
  const activeFederationsCount = activeFederations?.length || 0
  const completedFederations = allFederations?.filter(f => f.status === 'completed').length || 0
  const totalCollaborators = activeFederations?.reduce((sum, f) => sum + f.total_collaborators, 0) || 0
  const activeCollaborators = activeFederations?.reduce((sum, f) => sum + f.active_collaborators, 0) || 0

  const stats = [
    {
      name: 'Active Federations',
      value: activeFederationsCount,
      total: totalFederations,
      icon: Server,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
      change: '+12%',
      changeType: 'increase',
    },
    {
      name: 'Active Collaborators',
      value: activeCollaborators,
      total: totalCollaborators,
      icon: Users,
      color: 'text-green-600',
      bgColor: 'bg-green-50',
      change: '+8%',
      changeType: 'increase',
    },
    {
      name: 'Completed Federations',
      value: completedFederations,
      total: totalFederations,
      icon: CheckCircle,
      color: 'text-purple-600',
      bgColor: 'bg-purple-50',
      change: '+5%',
      changeType: 'increase',
    },
    {
      name: 'Training Rounds',
      value: activeFederations?.reduce((sum, f) => sum + f.current_round, 0) || 0,
      total: activeFederations?.reduce((sum, f) => sum + f.total_rounds, 0) || 0,
      icon: BarChart3,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-50',
      change: '+15%',
      changeType: 'increase',
    },
  ]

  if (federationsLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        </div>
        <div className="animate-pulse">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="bg-gray-200 rounded-lg h-32"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <div className="flex items-center space-x-2 text-sm text-gray-500">
          <Clock className="w-4 h-4" />
          <span>Last updated: {new Date().toLocaleTimeString()}</span>
        </div>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => (
          <div key={stat.name} className="metric-card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">{stat.name}</p>
                <div className="flex items-baseline space-x-1">
                  <p className="text-2xl font-semibold text-gray-900">
                    {stat.value}
                  </p>
                  {stat.total > 0 && (
                    <p className="text-sm text-gray-500">/ {stat.total}</p>
                  )}
                </div>
                <div className="flex items-center space-x-1 mt-1">
                  <TrendingUp className="w-3 h-3 text-green-500" />
                  <span className="text-xs text-green-600">{stat.change}</span>
                </div>
              </div>
              <div className={`p-3 rounded-lg ${stat.bgColor}`}>
                <stat.icon className={`w-6 h-6 ${stat.color}`} />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Active Federations */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Active Federations</h2>
            <Activity className="w-5 h-5 text-gray-400" />
          </div>
          
          {activeFederations && activeFederations.length > 0 ? (
            <div className="space-y-3">
              {activeFederations.map((federation) => (
                <FederationCard key={federation.id} federation={federation} />
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <Server className="w-12 h-12 mx-auto mb-3 text-gray-300" />
              <p>No active federations</p>
            </div>
          )}
        </div>

        {/* System Status */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">System Status</h2>
            <CheckCircle className="w-5 h-5 text-green-500" />
          </div>
          
          <div className="space-y-4">
            <StatusItem 
              label="API Server" 
              status="healthy" 
              uptime="99.9%" 
            />
            <StatusItem 
              label="Monitoring Service" 
              status="healthy" 
              uptime="99.8%" 
            />
            <StatusItem 
              label="Storage Backend" 
              status="healthy" 
              uptime="100%" 
            />
            <StatusItem 
              label="Real-time Events" 
              status="healthy" 
              uptime="99.7%" 
            />
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Recent Activity</h2>
          <Activity className="w-5 h-5 text-gray-400" />
        </div>
        
        <div className="space-y-3">
          <ActivityItem 
            time="2 minutes ago"
            message="Federation 'Demo Federation' completed round 5"
            type="success"
          />
          <ActivityItem 
            time="5 minutes ago"
            message="Collaborator collab_002 connection timeout"
            type="warning"
          />
          <ActivityItem 
            time="8 minutes ago"
            message="Model aggregation completed for round 4"
            type="info"
          />
          <ActivityItem 
            time="12 minutes ago"
            message="New collaborator collab_003 joined federation"
            type="success"
          />
        </div>
      </div>
    </div>
  )
}

function FederationCard({ federation }: { federation: FederationMetrics }) {
  const progress = (federation.current_round / federation.total_rounds) * 100

  return (
    <div className="border border-gray-200 rounded-lg p-4 hover:shadow-sm transition-shadow">
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-medium text-gray-900">{federation.name}</h3>
        <span className={`status-${federation.status}`}>
          {federation.status}
        </span>
      </div>
      
      <div className="space-y-2">
        <div className="flex justify-between text-sm text-gray-600">
          <span>Progress</span>
          <span>{federation.current_round}/{federation.total_rounds} rounds</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div 
            className="bg-primary-500 h-2 rounded-full transition-all duration-300"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
        <div className="flex justify-between text-sm text-gray-600">
          <span>{federation.active_collaborators}/{federation.total_collaborators} collaborators</span>
          <span>{federation.algorithm}</span>
        </div>
      </div>
    </div>
  )
}

function StatusItem({ label, status, uptime }: { label: string; status: string; uptime: string }) {
  const isHealthy = status === 'healthy'
  
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center space-x-3">
        <div className={`w-2 h-2 rounded-full ${isHealthy ? 'bg-green-400' : 'bg-red-400'}`}></div>
        <span className="text-sm font-medium text-gray-900">{label}</span>
      </div>
      <div className="flex items-center space-x-2">
        <span className="text-sm text-gray-500">{uptime}</span>
        <span className={`text-xs px-2 py-1 rounded-full ${
          isHealthy ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
        }`}>
          {status}
        </span>
      </div>
    </div>
  )
}

function ActivityItem({ time, message, type }: { time: string; message: string; type: string }) {
  const getIcon = () => {
    switch (type) {
      case 'success':
        return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'warning':
        return <AlertTriangle className="w-4 h-4 text-yellow-500" />
      case 'error':
        return <AlertTriangle className="w-4 h-4 text-red-500" />
      default:
        return <Activity className="w-4 h-4 text-blue-500" />
    }
  }

  return (
    <div className="flex items-start space-x-3 py-2">
      {getIcon()}
      <div className="flex-1 min-w-0">
        <p className="text-sm text-gray-900">{message}</p>
        <p className="text-xs text-gray-500">{time}</p>
      </div>
    </div>
  )
}
