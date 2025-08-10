import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { BarChart3, Clock, Users, TrendingUp } from 'lucide-react'
import { apiClient, queryKeys } from '../lib/api'

export default function Rounds() {
  const { data: rounds, isLoading } = useQuery({
    queryKey: queryKeys.rounds(),
    queryFn: () => apiClient.getRounds(),
  })

  if (isLoading) {
    return <div className="animate-pulse">Loading rounds...</div>
  }

  const completedRounds = rounds?.filter(r => r.status === 'completed') || []
  const totalDuration = completedRounds.reduce((sum, r) => sum + r.duration_ms, 0)
  const avgDuration = completedRounds.length > 0 ? totalDuration / completedRounds.length : 0

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Training Rounds</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <BarChart3 className="w-8 h-8 text-blue-600" />
            <div>
              <p className="text-sm text-gray-600">Total Rounds</p>
              <p className="text-2xl font-semibold">{rounds?.length || 0}</p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <Clock className="w-8 h-8 text-green-600" />
            <div>
              <p className="text-sm text-gray-600">Avg Duration</p>
              <p className="text-2xl font-semibold">{Math.round(avgDuration / 1000 / 60)}m</p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <Users className="w-8 h-8 text-purple-600" />
            <div>
              <p className="text-sm text-gray-600">Avg Participants</p>
              <p className="text-2xl font-semibold">
                {completedRounds.length > 0 
                  ? Math.round(completedRounds.reduce((sum, r) => sum + r.participant_count, 0) / completedRounds.length)
                  : 0
                }
              </p>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <div className="flex items-center space-x-3">
            <TrendingUp className="w-8 h-8 text-yellow-600" />
            <div>
              <p className="text-sm text-gray-600">Latest Accuracy</p>
              <p className="text-2xl font-semibold">
                {completedRounds.length > 0 && completedRounds[0].model_accuracy
                  ? `${(completedRounds[0].model_accuracy * 100).toFixed(1)}%`
                  : 'N/A'
                }
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="card">
        <h2 className="text-xl font-semibold mb-4">Round History</h2>
        {rounds && rounds.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Round
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Federation
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Duration
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Participants
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Accuracy
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {rounds.map((round) => (
                  <tr key={round.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {round.round_number}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {round.federation_id}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`status-${round.status}`}>
                        {round.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {Math.round(round.duration_ms / 1000 / 60)}m
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {round.participant_count}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {round.model_accuracy ? `${(round.model_accuracy * 100).toFixed(1)}%` : 'N/A'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <BarChart3 className="w-12 h-12 mx-auto mb-3 text-gray-300" />
            <p>No rounds found</p>
          </div>
        )}
      </div>
    </div>
  )
}
