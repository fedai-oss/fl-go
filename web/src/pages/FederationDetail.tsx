import React from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { apiClient, queryKeys } from '../lib/api'

export default function FederationDetail() {
  const { id } = useParams<{ id: string }>()
  
  const { data: federation, isLoading } = useQuery({
    queryKey: queryKeys.federation(id!),
    queryFn: () => apiClient.getFederation(id!),
    enabled: !!id,
  })

  if (isLoading) {
    return <div className="animate-pulse">Loading federation details...</div>
  }

  if (!federation) {
    return <div className="text-red-600">Federation not found</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Federation Details</h1>
      <div className="card">
        <h2 className="text-xl font-semibold mb-4">{federation.name}</h2>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-gray-600">ID</p>
            <p className="font-medium">{federation.id}</p>
          </div>
          <div>
            <p className="text-sm text-gray-600">Status</p>
            <p className="font-medium">{federation.status}</p>
          </div>
          <div>
            <p className="text-sm text-gray-600">Algorithm</p>
            <p className="font-medium">{federation.algorithm}</p>
          </div>
          <div>
            <p className="text-sm text-gray-600">Mode</p>
            <p className="font-medium">{federation.mode}</p>
          </div>
        </div>
      </div>
    </div>
  )
}
