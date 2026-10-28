'use client'

import { useState, useEffect, useCallback } from 'react'
import Link from 'next/link'
import { Search, Filter, Download, ChevronLeft, ChevronRight, Eye, MessageSquare, BarChart2, FileCheck } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { ManualDlrForm } from '@/components/messages/manual-dlr-form'

interface Message {
  id: number
  clientMessageId: string | null
  clientRef: string | null
  originalSourceAddr: string
  originalDestinationAddr: string
  shortMessage: string
  totalSegments: number
  receivedAt: string
  completedAt: string | null
  processingStatus: string
  finalStatus: string
  cost: string
  currencyCode: string
  serviceProviderId: number
  spName: string | null
  spSystemId: string | null
  routedMnoId: number | null
  mnoName: string | null
}

interface Pagination {
  page: number
  limit: number
  total: number
  totalPages: number
}

const STATUS_COLORS: Record<string, string> = {
  received: 'bg-blue-500/20 text-blue-400',
  queued: 'bg-amber-500/20 text-amber-400',
  processing: 'bg-purple-500/20 text-purple-400',
  submitted: 'bg-indigo-500/20 text-indigo-400',
  delivered: 'bg-emerald-500/20 text-emerald-400',
  accepted: 'bg-emerald-500/20 text-emerald-400',
  forwarded: 'bg-emerald-500/20 text-emerald-400',
  failed: 'bg-red-500/20 text-red-400',
  rejected: 'bg-red-500/20 text-red-400',
  undeliverable: 'bg-red-500/20 text-red-400',
  pending: 'bg-zinc-500/20 text-zinc-400',
}

function getStatusBadge(status: string) {
  const style = STATUS_COLORS[status] || STATUS_COLORS.pending
  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${style}`}>
      {status}
    </span>
  )
}

function formatCurrency(amount: string, currency: string) {
  const symbol = currency === 'NGN' ? '₦' : currency === 'USD' ? '$' : currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : currency
  return `${symbol}${Number(amount || 0).toFixed(2)}`
}

export default function MessagesPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [pagination, setPagination] = useState<Pagination>({ page: 1, limit: 50, total: 0, totalPages: 0 })
  const [loading, setLoading] = useState(true)
  const [filters, setFilters] = useState({
    status: 'all',
    sp_id: '',
    mno_id: '',
    client_ref: '',
    sender: '',
    destination: '',
    date_from: '',
    date_to: '',
  })
  const [showFilters, setShowFilters] = useState(false)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [showDlrForm, setShowDlrForm] = useState(false)
  const [dlrNotice, setDlrNotice] = useState<string | null>(null)

  const fetchMessages = useCallback(async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams({
        page: String(pagination.page),
        limit: String(pagination.limit),
      })
      
      if (filters.status !== 'all') params.set('status', filters.status)
      if (filters.sp_id) params.set('sp_id', filters.sp_id)
      if (filters.mno_id) params.set('mno_id', filters.mno_id)
      if (filters.client_ref) params.set('client_ref', filters.client_ref)
      if (filters.sender) params.set('sender', filters.sender)
      if (filters.destination) params.set('destination', filters.destination)
      if (filters.date_from) params.set('date_from', filters.date_from)
      if (filters.date_to) params.set('date_to', filters.date_to)

      const res = await fetch(`/api/admin/messages?${params}`)
      if (res.ok) {
        const data = await res.json()
        setMessages(data.messages || [])
        setPagination(data.pagination)
      }
    } catch (error) {
      console.error('Error fetching messages:', error)
    } finally {
      setLoading(false)
    }
  }, [pagination.page, pagination.limit, filters])

  useEffect(() => {
    fetchMessages()
  }, [fetchMessages])

  useEffect(() => {
    setSelectedIds([])
    setShowDlrForm(false)
  }, [pagination.page, filters])

  const allSelected = messages.length > 0 && messages.every((msg) => selectedIds.includes(msg.id))

  const toggleSelected = (id: number) => {
    setSelectedIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }

  const toggleAll = () => {
    setSelectedIds(allSelected ? [] : Array.from(new Set(messages.map((msg) => msg.id))))
  }

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
    setPagination(prev => ({ ...prev, page: 1 }))
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Messages"
        description="View and manage all SMS messages"
        actions={
          <div className="flex items-center gap-2">
            <Link
              href="/admin/messages/daily-report"
              className="flex items-center gap-2 px-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
            >
              <BarChart2 className="h-4 w-4" />
              Daily Report
            </Link>
            <Link
              href="/admin/messages/analytics"
              className="flex items-center gap-2 px-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
            >
              <BarChart2 className="h-4 w-4" />
              Analytics
            </Link>
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5"
            >
              <Filter className="h-4 w-4" />
              Filters
            </Button>
          </div>
        }
      />

      {showFilters && (
        <DataCard title="Filters" className="mb-6">
          <div className="grid gap-4 md:grid-cols-6">
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Status</label>
              <select
                value={filters.status}
                onChange={(e) => handleFilterChange('status', e.target.value)}
                className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
              >
                <option value="all">All Statuses</option>
                <option value="received">Received</option>
                <option value="queued">Queued</option>
                <option value="processing">Processing</option>
                <option value="submitted">Submitted</option>
                <option value="delivered">Delivered</option>
                <option value="failed">Failed</option>
                <option value="pending">Pending</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Sender</label>
              <input
                type="text"
                placeholder="Search sender..."
                value={filters.sender}
                onChange={(e) => handleFilterChange('sender', e.target.value)}
                className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">MSISDN</label>
              <input
                type="text"
                placeholder="Search destination..."
                value={filters.destination}
                onChange={(e) => handleFilterChange('destination', e.target.value)}
                className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Client Ref</label>
              <input
                type="text"
                placeholder="Search client ref..."
                value={filters.client_ref}
                onChange={(e) => handleFilterChange('client_ref', e.target.value)}
                className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Date From</label>
              <input
                type="date"
                value={filters.date_from}
                onChange={(e) => handleFilterChange('date_from', e.target.value)}
                className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Date To</label>
              <input
                type="date"
                value={filters.date_to}
                onChange={(e) => handleFilterChange('date_to', e.target.value)}
                className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white"
              />
            </div>
          </div>
        </DataCard>
      )}

      {dlrNotice && (
        <div className="rounded-2xl border border-emerald-500/20 bg-emerald-500/10 p-4 text-sm text-emerald-400">
          {dlrNotice}
        </div>
      )}

      {selectedIds.length > 0 && !showDlrForm && (
        <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-zinc-900/50 px-6 py-3">
          <span className="text-sm text-zinc-300">{selectedIds.length} message(s) selected</span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              onClick={() => setSelectedIds([])}
              className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5"
            >
              Clear
            </Button>
            <Button
              onClick={() => {
                setDlrNotice(null)
                setShowDlrForm(true)
              }}
              className="bg-white text-black hover:bg-zinc-200"
            >
              <FileCheck className="h-4 w-4" />
              Send DLR
            </Button>
          </div>
        </div>
      )}

      {showDlrForm && selectedIds.length > 0 && (
        <DataCard title="Send Manual DLR" subtitle={`For ${selectedIds.length} selected message(s) that did not get a DLR from the MNO`}>
          <ManualDlrForm
            messageIds={selectedIds}
            onCancel={() => setShowDlrForm(false)}
            onDone={(summary) => {
              setShowDlrForm(false)
              setSelectedIds([])
              setDlrNotice(summary)
              fetchMessages()
            }}
          />
        </DataCard>
      )}

      <DataCard title="" className="!p-0">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-800/50">
              <tr>
                <th className="px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    aria-label="Select all messages on this page"
                    checked={allSelected}
                    onChange={toggleAll}
                    className="h-4 w-4 accent-white"
                  />
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Message ID</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Client Ref</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">From → To</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Message</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Service Provider</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">MNO</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Cost</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Received</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {loading ? (
                <tr>
                  <td colSpan={11} className="px-4 py-12 text-center">
                    <div className="flex items-center justify-center">
                      <MessageSquare className="h-8 w-8 animate-pulse text-zinc-500" />
                    </div>
                    <p className="text-zinc-500 mt-2">Loading messages...</p>
                  </td>
                </tr>
              ) : messages.length === 0 ? (
                <tr>
                  <td colSpan={11} className="px-4 py-12 text-center">
                    <MessageSquare className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
                    <p className="text-zinc-500">No messages found</p>
                  </td>
                </tr>
              ) : (
                messages.map((msg) => (
                  <tr key={msg.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        aria-label={`Select message ${msg.id}`}
                        checked={selectedIds.includes(msg.id)}
                        onChange={() => toggleSelected(msg.id)}
                        className="h-4 w-4 accent-white"
                      />
                    </td>
                    <td className="px-4 py-3 text-sm font-mono text-white">{msg.id}</td>
                    <td className="px-4 py-3 text-sm text-zinc-400">{msg.clientRef || '-'}</td>
                    <td className="px-4 py-3 text-sm">
                      <div className="flex items-center gap-1">
                        <span className="text-white">{msg.originalSourceAddr}</span>
                        <span className="text-zinc-500">→</span>
                        <span className="text-white">{msg.originalDestinationAddr}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm max-w-xs truncate text-zinc-400">
                      {msg.shortMessage}
                    </td>
                    <td className="px-4 py-3 text-sm text-white">{msg.spName || '-'}</td>
                    <td className="px-4 py-3 text-sm text-white">{msg.mnoName || '-'}</td>
                    <td className="px-4 py-3">{getStatusBadge(msg.processingStatus)}</td>
                    <td className="px-4 py-3 text-sm font-medium text-white">
                      {formatCurrency(msg.cost, msg.currencyCode)}
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-400">
                      {new Date(msg.receivedAt).toLocaleString()}
                    </td>
                    <td className="px-4 py-3">
                      <Link
                        href={`/admin/messages/${msg.id}`}
                        className="text-white hover:underline text-sm"
                      >
                        View
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {pagination.totalPages > 1 && (
          <div className="p-4 border-t border-white/10 flex items-center justify-between">
            <p className="text-sm text-zinc-400">
              Showing {(pagination.page - 1) * pagination.limit + 1} to {Math.min(pagination.page * pagination.limit, pagination.total)} of {pagination.total} messages
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPagination(prev => ({ ...prev, page: prev.page - 1 }))}
                disabled={pagination.page === 1}
                className="p-2 border border-white/10 rounded-lg hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed text-zinc-400"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <span className="text-sm text-zinc-400">
                Page {pagination.page} of {pagination.totalPages}
              </span>
              <button
                onClick={() => setPagination(prev => ({ ...prev, page: prev.page + 1 }))}
                disabled={pagination.page === pagination.totalPages}
                className="p-2 border border-white/10 rounded-lg hover:bg-white/5 disabled:opacity-50 disabled:cursor-not-allowed text-zinc-400"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </div>
        )}
      </DataCard>
    </div>
  )
}
