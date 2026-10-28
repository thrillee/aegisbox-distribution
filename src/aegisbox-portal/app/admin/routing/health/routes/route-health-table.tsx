'use client'

import * as React from 'react'
import { useState, useEffect } from 'react'
import { HealthScoreBadge } from '@/components/routing/health-score-badge'
import { RouteStatusBadge } from '@/components/routing/route-status-badge'
import { DataCard } from '@/components/dashboard/data-card'
import { CheckCircle, XCircle, RefreshCw, Filter } from 'lucide-react'

interface RouteHealthData {
  assignmentId: number
  healthScore: string | null
  successCount: number
  failCount: number
  totalAttempts: number
  consecutiveFailures: number
  isAvailable: boolean
  lastAttemptAt: Date | null
  updatedAt: Date
  routingGroupId: number | null
  routingGroupName: string | null
  prefixGroupName: string | null
  mnoName: string | null
  priority: number
  weight: number
}

interface RouteHealthTableProps {
  initialData: RouteHealthData[]
  routingGroups: { id: number; name: string }[]
  mnos: { id: number; name: string }[]
}

export function RouteHealthTable({ initialData, routingGroups, mnos }: RouteHealthTableProps) {
  const [data, setData] = useState<RouteHealthData[]>(initialData)
  const [filterRoutingGroup, setFilterRoutingGroup] = useState<string>('all')
  const [filterAvailability, setFilterAvailability] = useState<string>('all')
  const [loading, setLoading] = useState<number | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date())

  // SSE connection for real-time updates
  useEffect(() => {
    const eventSource = new EventSource('/api/admin/routing/health/events/stream')

    eventSource.addEventListener('connected', () => {
      console.log('SSE connected')
    })

    eventSource.addEventListener('route-marked-available', (e) => {
      const update = JSON.parse(e.data)
      updateRouteInList(update.assignment_id, { isAvailable: true })
      setLastUpdate(new Date())
    })

    eventSource.addEventListener('route-marked-unavailable', (e) => {
      const update = JSON.parse(e.data)
      updateRouteInList(update.assignment_id, { isAvailable: false })
      setLastUpdate(new Date())
    })

    eventSource.addEventListener('health-updated', (e) => {
      const update = JSON.parse(e.data)
      updateRouteInList(update.assignment_id, update)
      setLastUpdate(new Date())
    })

    eventSource.onerror = () => {
      console.error('SSE connection error')
      // Will automatically reconnect
    }

    return () => {
      eventSource.close()
    }
  }, [])

  const updateRouteInList = (assignmentId: number, updates: Partial<RouteHealthData>) => {
    setData((prev) =>
      prev.map((route) =>
        route.assignmentId === assignmentId
          ? { ...route, ...updates }
          : route
      )
    )
  }

  const handleMarkAvailable = async (assignmentId: number, available: boolean) => {
    setLoading(assignmentId)
    try {
      const response = await fetch(
        `/api/admin/routing/health/routes/${assignmentId}/mark-available`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ available }),
        }
      )

      if (!response.ok) {
        throw new Error('Failed to update route availability')
      }

      // Update will come through SSE
    } catch (error) {
      console.error('Error updating route availability:', error)
      alert('Failed to update route availability. Please try again.')
    } finally {
      setLoading(null)
    }
  }

  // Filter data
  const filteredData = data.filter((route) => {
    if (filterRoutingGroup !== 'all' && route.routingGroupId?.toString() !== filterRoutingGroup) {
      return false
    }
    if (filterAvailability === 'available' && !route.isAvailable) {
      return false
    }
    if (filterAvailability === 'unavailable' && route.isAvailable) {
      return false
    }
    return true
  })

  return (
    <DataCard title="" className="!p-0">
      {/* Header with filters */}
      <div className="p-4 border-b border-white/10">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="font-semibold text-white">Route Health Metrics</h2>
            <p className="text-sm text-zinc-400">
              Showing {filteredData.length} of {data.length} routes • Last update: {lastUpdate.toLocaleTimeString()}
            </p>
          </div>
          <button
            onClick={() => window.location.reload()}
            className="flex items-center gap-2 px-3 py-1.5 bg-white/5 hover:bg-white/10 rounded-lg text-zinc-400 hover:text-white transition-colors text-sm"
          >
            <RefreshCw className="h-4 w-4" />
            Refresh
          </button>
        </div>

        {/* Filters */}
        <div className="flex items-center gap-3">
          <Filter className="h-4 w-4 text-zinc-500" />
          <select
            value={filterRoutingGroup}
            onChange={(e) => setFilterRoutingGroup(e.target.value)}
            className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Routing Groups</option>
            {routingGroups.map((group) => (
              <option key={group.id} value={group.id.toString()}>
                {group.name}
              </option>
            ))}
          </select>

          <select
            value={filterAvailability}
            onChange={(e) => setFilterAvailability(e.target.value)}
            className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Status</option>
            <option value="available">Available Only</option>
            <option value="unavailable">Unavailable Only</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-white/10">
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Routing Group</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Prefix Group</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">MNO</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Health</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Success Rate</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Attempts</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Consecutive Failures</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Status</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Last Attempt</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredData.length === 0 ? (
              <tr>
                <td colSpan={10} className="px-4 py-8 text-center text-zinc-500">
                  No routes found matching the filters
                </td>
              </tr>
            ) : (
              filteredData.map((route) => {
                const successRate = route.totalAttempts > 0
                  ? ((route.successCount / route.totalAttempts) * 100).toFixed(1)
                  : '0.0'

                return (
                  <tr
                    key={route.assignmentId}
                    className="border-b border-white/5 hover:bg-white/5 transition-colors"
                  >
                    <td className="px-4 py-3 text-sm text-white">
                      {route.routingGroupName || 'Unknown'}
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {route.prefixGroupName || 'Unknown'}
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {route.mnoName || 'Unknown'}
                    </td>
                    <td className="px-4 py-3">
                      <HealthScoreBadge score={route.healthScore} />
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {successRate}%
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {route.totalAttempts}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                          route.consecutiveFailures >= 3
                            ? 'bg-red-500/20 text-red-400'
                            : route.consecutiveFailures > 0
                            ? 'bg-amber-500/20 text-amber-400'
                            : 'bg-zinc-500/20 text-zinc-400'
                        }`}
                      >
                        {route.consecutiveFailures}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <RouteStatusBadge
                        isAvailable={route.isAvailable}
                        consecutiveFailures={route.consecutiveFailures}
                      />
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-400">
                      {route.lastAttemptAt ? new Date(route.lastAttemptAt).toLocaleString() : 'Never'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        {route.isAvailable ? (
                          <button
                            onClick={() => handleMarkAvailable(route.assignmentId, false)}
                            disabled={loading === route.assignmentId}
                            className="flex items-center gap-1 px-2 py-1 bg-red-500/10 hover:bg-red-500/20 border border-red-500/20 rounded text-xs text-red-400 disabled:opacity-50 transition-colors"
                          >
                            <XCircle className="h-3 w-3" />
                            {loading === route.assignmentId ? 'Loading...' : 'Mark Unavailable'}
                          </button>
                        ) : (
                          <button
                            onClick={() => handleMarkAvailable(route.assignmentId, true)}
                            disabled={loading === route.assignmentId}
                            className="flex items-center gap-1 px-2 py-1 bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20 rounded text-xs text-emerald-400 disabled:opacity-50 transition-colors"
                          >
                            <CheckCircle className="h-3 w-3" />
                            {loading === route.assignmentId ? 'Loading...' : 'Mark Available'}
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </DataCard>
  )
}
