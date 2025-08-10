import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Activity, AlertTriangle, CheckCircle, Info, Clock } from 'lucide-react'
import { apiClient, queryKeys } from '../lib/api'

export default function Events() {
  const { data: events, isLoading } = useQuery({
    queryKey: queryKeys.events(),
    queryFn: () => apiClient.getEvents({ per_page: 100 }),
  })

  if (isLoading) {
    return <div className="animate-pulse">Loading events...</div>
  }

  const getEventIcon = (level: string) => {
    switch (level) {
      case 'error':
        return <AlertTriangle className="w-5 h-5 text-red-500" />
      case 'warning':
        return <AlertTriangle className="w-5 h-5 text-yellow-500" />
      case 'info':
        return <Info className="w-5 h-5 text-blue-500" />
      default:
        return <Activity className="w-5 h-5 text-gray-500" />
    }
  }

  const getEventBgColor = (level: string) => {
    switch (level) {
      case 'error':
        return 'bg-red-50 border-red-200'
      case 'warning':
        return 'bg-yellow-50 border-yellow-200'
      case 'info':
        return 'bg-blue-50 border-blue-200'
      default:
        return 'bg-gray-50 border-gray-200'
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Events</h1>
        <div className="flex items-center space-x-2 text-sm text-gray-500">
          <Clock className="w-4 h-4" />
          <span>Real-time updates</span>
          <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <Activity className="w-8 h-8 text-blue-600" />
            <div>
              <p className="text-sm text-gray-600">Total Events</p>
              <p className="text-2xl font-semibold">{events?.length || 0}</p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <Info className="w-8 h-8 text-blue-600" />
            <div>
              <p className="text-sm text-gray-600">Info</p>
              <p className="text-2xl font-semibold">
                {events?.filter(e => e.level === 'info').length || 0}
              </p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <AlertTriangle className="w-8 h-8 text-yellow-600" />
            <div>
              <p className="text-sm text-gray-600">Warnings</p>
              <p className="text-2xl font-semibold">
                {events?.filter(e => e.level === 'warning').length || 0}
              </p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <AlertTriangle className="w-8 h-8 text-red-600" />
            <div>
              <p className="text-sm text-gray-600">Errors</p>
              <p className="text-2xl font-semibold">
                {events?.filter(e => e.level === 'error').length || 0}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="card">
        <h2 className="text-xl font-semibold mb-4">Recent Events</h2>
        {events && events.length > 0 ? (
          <div className="space-y-3">
            {events.map((event) => (
              <div 
                key={event.id} 
                className={`border rounded-lg p-4 ${getEventBgColor(event.level)}`}
              >
                <div className="flex items-start space-x-3">
                  {getEventIcon(event.level)}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-1">
                      <p className="text-sm font-medium text-gray-900">
                        {event.type.replace('_', ' ').toUpperCase()}
                      </p>
                      <span className="text-xs text-gray-500">
                        {new Date(event.timestamp).toLocaleString()}
                      </span>
                    </div>
                    <p className="text-sm text-gray-700 mb-2">{event.message}</p>
                    <div className="flex items-center space-x-4 text-xs text-gray-500">
                      <span>Source: {event.source}</span>
                      <span>Federation: {event.federation_id}</span>
                    </div>
                    {event.data && Object.keys(event.data).length > 0 && (
                      <details className="mt-2">
                        <summary className="text-xs text-gray-500 cursor-pointer hover:text-gray-700">
                          Event data
                        </summary>
                        <pre className="mt-1 text-xs bg-gray-100 p-2 rounded overflow-x-auto">
                          {JSON.stringify(event.data, null, 2)}
                        </pre>
                      </details>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <Activity className="w-12 h-12 mx-auto mb-3 text-gray-300" />
            <p>No events found</p>
          </div>
        )}
      </div>
    </div>
  )
}
