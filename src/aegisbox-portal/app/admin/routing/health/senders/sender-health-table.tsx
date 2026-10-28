'use client'

import * as React from 'react'
import { useState, useEffect } from 'react'
import { HealthScoreBadge } from '@/components/routing/health-score-badge'
import { SenderStatusBadge } from '@/components/routing/sender-status-badge'
import { DataCard } from '@/components/dashboard/data-card'
import { ShieldOff, RefreshCw, Filter, Search } from 'lucide-react'

interface SenderHealthData {
  id: number
  senderId: number
  mnoId: number
  mnoName: string | null
  deliveryRate: string | null
  dlrResponseRate: string | null
  healthScore: string | null
  totalMessages: number
  isBlocked: boolean
  blockedAt: Date | null
  blockedReason: string | null
  lastMessageAt: Date | null
  updatedAt: Date
}

interface SenderHealthTableProps {
  initialData: SenderHealthData[]
  mnos: { id: number; name: string }[]
}

export function SenderHealthTable({ initialData, mnos }: SenderHealthTableProps) {
  const [data, setData] = useState<SenderHealthData[]>(initialData)
  const [filterMNO, setFilterMNO] = useState<string>('all')
  const [filterBlocked, setFilterBlocked] = useState<string>('all')
  const [searchSender, setSearchSender] = useState<string>('')
  const [loading, setLoading] = useState<number | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date())

  // SSE connection for real-time updates
  useEffect(() => {
    const eventSource = new EventSource('/api/admin/routing/health/events/stream')

    eventSource.addEventListener('connected', () => {
      console.log('SSE connected')
    })

    eventSource.addEventListener('sender-blocked', (e) => {
      const update = JSON.parse(e.data)
      updateSenderInList(update.sender_id, update.mno_id, {
        isBlocked: true,
        blockedAt: new Date(update.blocked_at),
        blockedReason: update.reason,
      })
      setLastUpdate(new Date())
    })

    eventSource.addEventListener('sender-unblocked', (e) => {
      const update = JSON.parse(e.data)
      updateSenderInList(update.sender_id, update.mno_id, {
        isBlocked: false,
        blockedAt: null,
        blockedReason: null,
      })
      setLastUpdate(new Date())
    })

    eventSource.addEventListener('health-updated', (e) => {
      const update = JSON.parse(e.data)
      if (update.sender_id && update.mno_id) {
        updateSenderInList(update.sender_id, update.mno_id, update)
        setLastUpdate(new Date())
      }
    })

    eventSource.onerror = () => {
      console.error('SSE connection error')
      // Will automatically reconnect
    }

    return () => {
      eventSource.close()
    }
  }, [])

  const updateSenderInList = (senderId: string, mnoId: number, updates: Partial<SenderHealthData>) => {
    const senderIdNum = parseInt(senderId, 10)
    setData((prev) =>
      prev.map((sender) =>
        sender.senderId === senderIdNum && sender.mnoId === mnoId
          ? { ...sender, ...updates }
          : sender
      )
    )
  }

  const handleUnblock = async (senderId: number, mnoId: number) => {
    const sender = data.find((s) => s.senderId === senderId && s.mnoId === mnoId)
    if (!sender) return

    setLoading(sender.id)
    try {
      const response = await fetch(
        `/api/admin/routing/health/senders/${senderId}/${mnoId}/unblock`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ reason: 'Manually unblocked via admin UI' }),
        }
      )

      if (!response.ok) {
        throw new Error('Failed to unblock sender')
      }

      // Update will come through SSE
    } catch (error) {
      console.error('Error unblocking sender:', error)
      alert('Failed to unblock sender. Please try again.')
    } finally {
      setLoading(null)
    }
  }

  // Filter data
  const filteredData = data.filter((sender) => {
    if (filterMNO !== 'all' && sender.mnoId?.toString() !== filterMNO) {
      return false
    }
    if (filterBlocked === 'blocked' && !sender.isBlocked) {
      return false
    }
    if (filterBlocked === 'active' && sender.isBlocked) {
      return false
    }
    if (searchSender && !sender.senderId.toString().toLowerCase().includes(searchSender.toLowerCase())) {
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
            <h2 className="font-semibold text-white">Sender ID Performance Metrics</h2>
            <p className="text-sm text-zinc-400">
              Showing {filteredData.length} of {data.length} senders • Last update: {lastUpdate.toLocaleTimeString()}
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
        <div className="flex items-center gap-3 flex-wrap">
          <Filter className="h-4 w-4 text-zinc-500" />
          
          <div className="relative flex-1 max-w-xs">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
            <input
              type="text"
              placeholder="Search sender ID..."
              value={searchSender}
              onChange={(e) => setSearchSender(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-sm text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <select
            value={filterMNO}
            onChange={(e) => setFilterMNO(e.target.value)}
            className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All MNOs</option>
            {mnos.map((mno) => (
              <option key={mno.id} value={mno.id.toString()}>
                {mno.name}
              </option>
            ))}
          </select>

          <select
            value={filterBlocked}
            onChange={(e) => setFilterBlocked(e.target.value)}
            className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Status</option>
            <option value="active">Active Only</option>
            <option value="blocked">Blocked Only</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-white/10">
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Sender ID</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">MNO</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Health</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Delivery Rate</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">DLR Response</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Total Messages</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Status</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Blocked At</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Blocked Reason</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-zinc-400 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredData.length === 0 ? (
              <tr>
                <td colSpan={10} className="px-4 py-8 text-center text-zinc-500">
                  No senders found matching the filters
                </td>
              </tr>
            ) : (
              filteredData.map((sender) => {
                const deliveryRate = sender.deliveryRate
                  ? parseFloat(sender.deliveryRate).toFixed(1)
                  : '0.0'
                const dlrResponseRate = sender.dlrResponseRate
                  ? parseFloat(sender.dlrResponseRate).toFixed(1)
                  : '0.0'

                return (
                  <tr
                    key={`${sender.senderId}-${sender.mnoId}`}
                    className="border-b border-white/5 hover:bg-white/5 transition-colors"
                  >
                    <td className="px-4 py-3 text-sm font-medium text-white">
                      {sender.senderId}
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {sender.mnoName || 'Unknown'}
                    </td>
                    <td className="px-4 py-3">
                      <HealthScoreBadge score={sender.healthScore} />
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {deliveryRate}%
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {dlrResponseRate}%
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-300">
                      {sender.totalMessages.toLocaleString()}
                    </td>
                    <td className="px-4 py-3">
                      <SenderStatusBadge
                        deliveryRate={sender.deliveryRate}
                        isBlocked={sender.isBlocked}
                      />
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-400">
                      {sender.blockedAt ? new Date(sender.blockedAt).toLocaleString() : '—'}
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-400 max-w-xs truncate">
                      {sender.blockedReason || '—'}
                    </td>
                    <td className="px-4 py-3">
                      {sender.isBlocked ? (
                        <button
                          onClick={() => handleUnblock(sender.senderId, sender.mnoId)}
                          disabled={loading === sender.id}
                          className="flex items-center gap-1 px-2 py-1 bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20 rounded text-xs text-emerald-400 disabled:opacity-50 transition-colors"
                        >
                          <ShieldOff className="h-3 w-3" />
                          {loading === sender.id ? 'Loading...' : 'Unblock'}
                        </button>
                      ) : (
                        <span className="text-xs text-zinc-500">—</span>
                      )}
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
