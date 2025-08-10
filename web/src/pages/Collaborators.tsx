import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Users, Activity, Clock, AlertTriangle } from 'lucide-react'
import { apiClient, queryKeys } from '../lib/api'

export default function Collaborators() {
  const { data: collaborators, isLoading } = useQuery({
    queryKey: queryKeys.collaborators(),
    queryFn: () => apiClient.getCollaborators(),
  })

  if (isLoading) {
    return <div className="animate-pulse">Loading collaborators...</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Collaborators</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <Users className="w-8 h-8 text-blue-600" />
            <div>
              <p className="text-sm text-gray-600">Total Collaborators</p>
              <p className="text-2xl font-semibold">{collaborators?.length || 0}</p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <Activity className="w-8 h-8 text-green-600" />
            <div>
              <p className="text-sm text-gray-600">Active</p>
              <p className="text-2xl font-semibold">
                {collaborators?.filter(c => c.status === 'connected' || c.status === 'training').length || 0}
              </p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <AlertTriangle className="w-8 h-8 text-red-600" />
            <div>
              <p className="text-sm text-gray-600">With Errors</p>
              <p className="text-2xl font-semibold">
                {collaborators?.filter(c => c.error_count > 0).length || 0}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="card">
        <h2 className="text-xl font-semibold mb-4">All Collaborators</h2>
        {collaborators && collaborators.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Address
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Updates
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Last Seen
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {collaborators.map((collaborator) => (
                  <tr key={collaborator.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {collaborator.id}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`status-${collaborator.status}`}>
                        {collaborator.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {collaborator.address}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {collaborator.updates_submitted}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(collaborator.last_seen).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <Users className="w-12 h-12 mx-auto mb-3 text-gray-300" />
            <p>No collaborators found</p>
          </div>
        )}
      </div>
    </div>
  )
}
