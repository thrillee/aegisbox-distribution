'use client'

import { useState, useEffect, useCallback } from 'react'
import { X, RefreshCw, Loader2, AlertCircle, CheckCircle, Clock, XCircle, Eye } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface QueueItem {
  id: number
  messageId: number
  status: string
  attempts: number
  maxAttempts: number
  lastAttemptAt: string | null
  nextRetryAt: string | null
  errorMessage: string | null
  createdAt: string
  lockedAt: string | null
  clientMessageId: string | null
  clientRef: string | null
}

interface PaginationData {
  page: number
  limit: number
  total: number
  totalPages: number
}

const STATUS_OPTIONS = [
  { value: 'all', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'processing', label: 'Processing' },
  { value: 'success', label: 'Delivered to customer' },
  { value: 'failed', label: 'Failed' },
]

export default function DLRQueuePage() {
  const [loading, setLoading] = useState(true)
  const [initialLoad, setInitialLoad] = useState(true)
  const [queueData, setQueueData] = useState<QueueItem[]>([])
  const [pagination, setPagination] = useState<PaginationData>({ page: 1, limit: 50, total: 0, totalPages: 0 })
  const [filters, setFilters] = useState({
    status: 'all',
    date_from: '',
    date_to: '',
  })
  const [error, setError] = useState('')
  const [selectedItem, setSelectedItem] = useState<QueueItem | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [retrying, setRetrying] = useState(false)
  const [notice, setNotice] = useState('')

  const fetchQueue = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const params = new URLSearchParams()
      params.set('page', String(pagination.page))
      params.set('limit', String(pagination.limit))
      if (filters.status !== 'all') params.set('status', filters.status)
      if (filters.date_from) params.set('date_from', filters.date_from)
      if (filters.date_to) params.set('date_to', filters.date_to)

      const res = await fetch(`/api/admin/messages/queues/dlr?${params}`)
      if (res.ok) {
        const data = await res.json()
        setQueueData(data.queue || [])
        setPagination(prev => ({ ...prev, ...data.pagination }))
      } else {
        setError('Failed to fetch queue')
      }
    } catch (err) {
      setError('An error occurred')
    } finally {
      setLoading(false)
      setInitialLoad(false)
    }
  }, [pagination.page, pagination.limit, filters])

  useEffect(() => {
    fetchQueue()
  }, [fetchQueue])

  const failedIds = queueData.filter((item) => item.status === 'failed').map((item) => item.id)

  const retryForwarding = async (ids: number[]) => {
    setRetrying(true)
    setError('')
    setNotice('')
    try {
      const res = await fetch('/api/admin/messages/queues/dlr/retry', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setError(data.error || 'Failed to retry DLR forwarding')
        return
      }
      setNotice(`${data.requeued.length} DLR(s) queued to be forwarded to the customer again.${data.skipped ? ` ${data.skipped} skipped (not failed).` : ''}`)
      closeModal()
      fetchQueue()
    } catch {
      setError('Failed to retry DLR forwarding')
    } finally {
      setRetrying(false)
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-emerald-400" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-400" />
      case 'retrying':
        return <RefreshCw className="h-4 w-4 text-amber-400" />
      case 'processing':
        return <Loader2 className="h-4 w-4 text-blue-400 animate-spin" />
      default:
        return <Clock className="h-4 w-4 text-zinc-400" />
    }
  }

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      success: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
      completed: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
      failed: 'bg-red-500/20 text-red-300 border-red-500/30',
      retrying: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
      processing: 'bg-blue-500/20 text-blue-300 border-blue-500/30',
      pending: 'bg-zinc-500/20 text-zinc-300 border-zinc-500/30',
    }
    return colors[status] || 'bg-zinc-500/20 text-zinc-300 border-zinc-500/30'
  }

  const openModal = (item: QueueItem) => {
    setSelectedItem(item)
    setIsModalOpen(true)
  }

  const closeModal = () => {
    setIsModalOpen(false)
    setSelectedItem(null)
  }

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="DLR Forwarding Queue"
        description="Delivery report forwarding queue status and management"
        actions={
          <div className="flex items-center gap-2">
            {failedIds.length > 0 && (
              <Button
                variant="outline"
                onClick={() => retryForwarding(failedIds)}
                disabled={retrying}
                className="border-white/10 text-zinc-300 hover:text-white hover:bg-white/5"
              >
                {retrying ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
                Retry {failedIds.length} failed on this page
              </Button>
            )}
            <Button onClick={fetchQueue} disabled={loading} className="bg-white text-black hover:bg-zinc-200">
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4 mr-2" />}
              Refresh
            </Button>
          </div>
        }
      />

      {notice && (
        <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-emerald-400">
          {notice}
        </div>
      )}

      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
          {error}
        </div>
      )}

      <DataCard title="Filters" subtitle="Configure queue filters">
        <div className="grid gap-4 md:grid-cols-4">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
            <select
              value={filters.status}
              onChange={(e) => setFilters(prev => ({ ...prev, status: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            >
              {STATUS_OPTIONS.map(opt => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Date From</label>
            <input
              type="datetime-local"
              value={filters.date_from}
              onChange={(e) => setFilters(prev => ({ ...prev, date_from: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Date To</label>
            <input
              type="datetime-local"
              value={filters.date_to}
              onChange={(e) => setFilters(prev => ({ ...prev, date_to: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            />
          </div>
          <div className="flex items-end">
            <Button
              onClick={fetchQueue}
              disabled={loading}
              className="w-full bg-white text-black hover:bg-zinc-200"
            >
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Apply Filters'}
            </Button>
          </div>
        </div>
      </DataCard>

      <DataCard title="Queue Items" subtitle={`${pagination.total} items found`}>
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
          </div>
        ) : queueData.length === 0 ? (
          <div className="text-center py-12 text-zinc-500">
            No queue items found
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-xs font-semibold text-zinc-400 uppercase border-b border-white/10">
                    <th className="pb-3">Status</th>
                    <th className="pb-3">Message ID</th>
                    <th className="pb-3">Client Ref</th>
                    <th className="pb-3 text-center">Attempts</th>
                    <th className="pb-3">Last Attempt</th>
                    <th className="pb-3">Next Retry</th>
                    <th className="pb-3">Error</th>
                    <th className="pb-3">Created</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/10">
                  {queueData.map((item) => (
                    <tr key={item.id} className="hover:bg-white/5 cursor-pointer" onClick={() => openModal(item)}>
                      <td className="py-3">
                        <span className={`inline-flex items-center gap-2 px-2 py-1 rounded-full text-xs border ${getStatusBadge(item.status)}`}>
                          {getStatusIcon(item.status)}
                          <span className="capitalize">{item.status}</span>
                        </span>
                      </td>
                      <td className="py-3 text-white font-mono text-sm">{item.messageId}</td>
                      <td className="py-3 text-zinc-400 text-sm">{item.clientRef || '-'}</td>
                      <td className="py-3 text-center text-zinc-400">
                        {item.attempts}/{item.maxAttempts}
                      </td>
                      <td className="py-3 text-zinc-400 text-sm">
                        {item.lastAttemptAt ? new Date(item.lastAttemptAt).toLocaleString() : '-'}
                      </td>
                      <td className="py-3 text-zinc-400 text-sm">
                        {item.nextRetryAt ? new Date(item.nextRetryAt).toLocaleString() : '-'}
                      </td>
                      <td className="py-3 text-red-400 text-sm max-w-xs truncate">
                        {item.errorMessage || '-'}
                      </td>
                      <td className="py-3 text-zinc-400 text-sm">
                        {new Date(item.createdAt).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {pagination.totalPages > 1 && (
              <div className="flex items-center justify-between mt-4 pt-4 border-t border-white/10">
                <p className="text-sm text-zinc-400">
                  Page {pagination.page} of {pagination.totalPages}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPagination(prev => ({ ...prev, page: prev.page - 1 }))}
                    disabled={pagination.page === 1}
                    className="border-white/10 text-white hover:bg-white/10"
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPagination(prev => ({ ...prev, page: prev.page + 1 }))}
                    disabled={pagination.page === pagination.totalPages}
                    className="border-white/10 text-white hover:bg-white/10"
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </DataCard>

      {isModalOpen && selectedItem && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={closeModal} />
          <div className="relative bg-zinc-900 border border-white/10 rounded-xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto shadow-xl">
            <button
              onClick={closeModal}
              className="absolute top-4 right-4 p-2 text-zinc-400 hover:text-white hover:bg-white/10 rounded-lg"
            >
              <X className="h-4 w-4" />
            </button>
            <h2 className="text-lg font-semibold text-white mb-4">Queue Item Details</h2>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-xs text-zinc-500 mb-1">ID</p>
                  <p className="text-white font-mono">{selectedItem.id}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Message ID</p>
                  <p className="text-white font-mono">{selectedItem.messageId}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Client Message ID</p>
                  <p className="text-zinc-300">{selectedItem.clientMessageId || '-'}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Client Ref</p>
                  <p className="text-zinc-300">{selectedItem.clientRef || '-'}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Status</p>
                  <span className={`inline-flex items-center gap-2 px-2 py-1 rounded-full text-xs border ${getStatusBadge(selectedItem.status)}`}>
                    {getStatusIcon(selectedItem.status)}
                    <span className="capitalize">{selectedItem.status}</span>
                  </span>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Attempts</p>
                  <p className="text-white">{selectedItem.attempts} / {selectedItem.maxAttempts}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Created</p>
                  <p className="text-zinc-300">{new Date(selectedItem.createdAt).toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Last Attempt</p>
                  <p className="text-zinc-300">{selectedItem.lastAttemptAt ? new Date(selectedItem.lastAttemptAt).toLocaleString() : '-'}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Next Retry</p>
                  <p className="text-zinc-300">{selectedItem.nextRetryAt ? new Date(selectedItem.nextRetryAt).toLocaleString() : '-'}</p>
                </div>
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Locked At</p>
                  <p className="text-zinc-300">{selectedItem.lockedAt ? new Date(selectedItem.lockedAt).toLocaleString() : '-'}</p>
                </div>
              </div>
              {selectedItem.errorMessage && (
                <div>
                  <p className="text-xs text-zinc-500 mb-1">Error Message</p>
                  <p className="text-red-400 text-sm bg-red-500/10 border border-red-500/20 rounded-lg p-3">
                    {selectedItem.errorMessage}
                  </p>
                </div>
              )}
              {selectedItem.status === 'failed' && (
                <Button
                  onClick={() => retryForwarding([selectedItem.id])}
                  disabled={retrying}
                  className="w-full bg-white text-black hover:bg-zinc-200"
                >
                  {retrying ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
                  Retry forwarding to customer
                </Button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
