'use client'

import { useState, useEffect, useCallback } from 'react'
import { Plus, Search, RefreshCw, Tag, CheckCircle, Edit, Trash2, X, Check, Users, Globe, Network } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface SenderId {
  id: number
  senderIdString: string
  serviceProviderId: number | null
  mnoConnectionId: number | null
  status: string
  providerId: number | null
  rateLimit: number | null
  otpTemplate: string | null
  networks: string[] | null
  maxUsageCount: number | null
  currentUsageCount: number
  resetIntervalHours: number | null
  createdAt: string
  updatedAt: string
}

interface ServiceProvider {
  id: number
  name: string
}

interface MNO {
  id: number
  name: string
  countryCode: string | null
}

interface MnoConnection {
  id: number
  mnoId: number
  protocol: string
  status: string
  host: string | null
}

interface SenderIdsPageProps {
  initialSenders: SenderId[]
  serviceProviders: ServiceProvider[]
  mnos: MNO[]
  mnoConnections: MnoConnection[]
}

const STATUS_OPTIONS = [
  { value: 'all', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'blocked', label: 'Blocked' },
]

export function SenderIdsPage({ initialSenders: _initialSenders, serviceProviders, mnos, mnoConnections }: SenderIdsPageProps) {
  const [loading, setLoading] = useState(false)
  const [senders, setSenders] = useState<SenderId[]>(_initialSenders)
  const [pagination, setPagination] = useState({ page: 1, limit: 50, total: _initialSenders.length, totalPages: 1 })
  const [filters, setFilters] = useState({
    status: 'all',
    serviceProviderId: '',
    mnoConnectionId: '',
    search: '',
  })
  const [error, setError] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [editingSender, setEditingSender] = useState<SenderId | null>(null)
  const [deletingSender, setDeletingSender] = useState<SenderId | null>(null)
  const [formData, setFormData] = useState({
    senderIdString: '',
    serviceProviderId: '',
    mnoConnectionId: '',
    status: 'pending',
    rateLimit: '100',
    otpTemplate: '',
    networks: '',
    maxUsageCount: '10000',
    resetIntervalHours: '',
  })
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)

  const showToast = useCallback((message: string, type: 'success' | 'error') => {
    setToast({ message, type })
    setTimeout(() => setToast(null), 3000)
  }, [])

  const fetchSenders = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const params = new URLSearchParams()
      params.set('page', String(pagination.page))
      params.set('limit', String(pagination.limit))
      if (filters.status !== 'all') params.set('status', filters.status)
      if (filters.serviceProviderId) params.set('serviceProviderId', filters.serviceProviderId)
      if (filters.mnoConnectionId) params.set('mnoConnectionId', filters.mnoConnectionId)
      if (filters.search) params.set('search', filters.search)

      const res = await fetch(`/api/admin/sender-ids?${params}`)
      if (res.ok) {
        const data = await res.json()
        setSenders(data.senders || [])
        setPagination(prev => ({ ...prev, ...data.pagination }))
      } else {
        setError('Failed to fetch sender IDs')
      }
    } catch (err) {
      setError('An error occurred')
    } finally {
      setLoading(false)
    }
  }, [pagination.page, pagination.limit, filters])

  useEffect(() => {
    fetchSenders()
  }, [fetchSenders])

  const openCreateModal = () => {
    setEditingSender(null)
    setFormData({
      senderIdString: '',
      serviceProviderId: '',
      mnoConnectionId: '',
      status: 'pending',
      rateLimit: '100',
      otpTemplate: '',
      networks: '',
      maxUsageCount: '10000',
      resetIntervalHours: '',
    })
    setIsModalOpen(true)
  }

  const openEditModal = (sender: SenderId) => {
    setEditingSender(sender)
    setFormData({
      senderIdString: sender.senderIdString,
      serviceProviderId: sender.serviceProviderId?.toString() || '',
      mnoConnectionId: sender.mnoConnectionId?.toString() || '',
      status: sender.status,
      rateLimit: (sender.rateLimit ?? 100).toString(),
      otpTemplate: sender.otpTemplate || '',
      networks: sender.networks?.join(', ') || '',
      maxUsageCount: (sender.maxUsageCount ?? 10000).toString(),
      resetIntervalHours: sender.resetIntervalHours?.toString() || '',
    })
    setIsModalOpen(true)
  }

  const handleSubmit = async () => {
    try {
      const url = editingSender
        ? `/api/admin/sender-ids/${editingSender.id}`
        : '/api/admin/sender-ids'
      const method = editingSender ? 'PUT' : 'POST'

      const networksArray = formData.networks
        ? formData.networks.split(',').map(n => n.trim()).filter(Boolean)
        : []

      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          senderIdString: formData.senderIdString,
          serviceProviderId: formData.serviceProviderId ? parseInt(formData.serviceProviderId) : null,
          mnoConnectionId: formData.mnoConnectionId ? parseInt(formData.mnoConnectionId) : null,
          status: formData.status,
          rateLimit: parseInt(formData.rateLimit),
          otpTemplate: formData.otpTemplate || null,
          networks: networksArray.length > 0 ? networksArray : null,
          maxUsageCount: parseInt(formData.maxUsageCount),
          resetIntervalHours: formData.resetIntervalHours ? parseInt(formData.resetIntervalHours) : null,
        }),
      })

      if (res.ok) {
        fetchSenders()
        setIsModalOpen(false)
        showToast(editingSender ? 'Sender ID updated successfully' : 'Sender ID created successfully', 'success')
      } else {
        const error = await res.json()
        showToast(error.error || 'Failed to save sender ID', 'error')
      }
    } catch (err) {
      showToast('Failed to save sender ID', 'error')
    }
  }

  const handleDelete = async () => {
    if (!deletingSender) return

    try {
      const res = await fetch(`/api/admin/sender-ids/${deletingSender.id}`, {
        method: 'DELETE',
      })

      if (res.ok) {
        fetchSenders()
        setIsDeleteModalOpen(false)
        setDeletingSender(null)
        showToast('Sender ID deleted successfully', 'success')
      } else {
        showToast('Failed to delete sender ID', 'error')
      }
    } catch (err) {
      showToast('Failed to delete sender ID', 'error')
    }
  }

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      approved: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
      rejected: 'bg-red-500/20 text-red-400 border-red-500/30',
      pending: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
      blocked: 'bg-red-500/20 text-red-400 border-red-500/30',
    }
    return colors[status] || 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30'
  }

  const getSPName = (spId: number | null) => {
    if (!spId) return 'Global'
    const sp = serviceProviders.find(sp => sp.id === spId)
    return sp?.name || 'Unknown'
  }

  const getMnoConnectionInfo = (connectionId: number | null) => {
    if (!connectionId) return 'Any Connection'
    const connection = mnoConnections.find(c => c.id === connectionId)
    if (!connection) return 'Unknown'
    const mno = mnos.find(m => m.id === connection.mnoId)
    return `${mno?.name || 'Unknown'} (${connection.protocol})`
  }

  const stats = {
    total: pagination.total,
    approved: senders.filter(s => s.status === 'approved').length,
    totalUsage: senders.reduce((sum, s) => sum + s.currentUsageCount, 0),
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Sender IDs"
        description="Manage sender IDs for SMS delivery"
        actions={
          <Button onClick={openCreateModal} className="bg-white text-black hover:bg-zinc-200">
            <Plus className="h-4 w-4 mr-2" />
            Add Sender ID
          </Button>
        }
      />

      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
          {error}
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-500/20 rounded-lg">
              <Tag className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Sender IDs</p>
              <p className="text-2xl font-bold text-white">{stats.total}</p>
            </div>
          </div>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-500/20 rounded-lg">
              <CheckCircle className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Approved</p>
              <p className="text-2xl font-bold text-emerald-400">{stats.approved}</p>
            </div>
          </div>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-amber-500/20 rounded-lg">
              <RefreshCw className="h-5 w-5 text-amber-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Usage</p>
              <p className="text-2xl font-bold text-white">{stats.totalUsage.toLocaleString()}</p>
            </div>
          </div>
        </div>
      </div>

      <DataCard title="Filters" subtitle="Filter sender IDs">
        <div className="grid gap-4 md:grid-cols-5">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Search</label>
            <input
              type="text"
              placeholder="Search sender ID..."
              value={filters.search}
              onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            />
          </div>
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
            <label className="block text-sm font-medium text-zinc-400 mb-1">Service Provider</label>
            <select
              value={filters.serviceProviderId}
              onChange={(e) => setFilters(prev => ({ ...prev, serviceProviderId: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            >
              <option value="">All SPs</option>
              {serviceProviders.map(sp => (
                <option key={sp.id} value={sp.id}>{sp.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">MNO Connection</label>
            <select
              value={filters.mnoConnectionId}
              onChange={(e) => setFilters(prev => ({ ...prev, mnoConnectionId: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            >
              <option value="">All Connections</option>
              {mnoConnections.map(conn => {
                const mno = mnos.find(m => m.id === conn.mnoId)
                return (
                  <option key={conn.id} value={conn.id}>
                    {mno?.name || 'Unknown'} ({conn.protocol})
                  </option>
                )
              })}
            </select>
          </div>
          <div className="flex items-end">
            <Button
              onClick={fetchSenders}
              disabled={loading}
              className="w-full bg-white text-black hover:bg-zinc-200"
            >
              {loading ? <RefreshCw className="h-4 w-4 animate-spin" /> : 'Apply'}
            </Button>
          </div>
        </div>
      </DataCard>

      <DataCard title="Sender IDs" subtitle={`${pagination.total} sender IDs found`}>
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <RefreshCw className="h-8 w-8 animate-spin text-zinc-500" />
          </div>
        ) : senders.length === 0 ? (
          <div className="text-center py-12 text-zinc-500">
            <Tag className="h-12 w-12 mx-auto mb-4 text-zinc-600" />
            <p>No sender IDs found</p>
            <Button onClick={openCreateModal} className="mt-4" variant="outline">
              Add your first sender ID
            </Button>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-xs font-semibold text-zinc-400 uppercase border-b border-white/10">
                    <th className="pb-3">Sender ID</th>
                    <th className="pb-3">Service Provider</th>
                    <th className="pb-3">Connection</th>
                    <th className="pb-3 text-center">Usage</th>
                    <th className="pb-3">Rate Limit</th>
                    <th className="pb-3">Status</th>
                    <th className="pb-3">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/10">
                  {senders.map((sender) => (
                    <tr key={sender.id} className="hover:bg-white/5">
                      <td className="py-3">
                        <span className="font-mono text-white">{sender.senderIdString}</span>
                      </td>
                      <td className="py-3">
                        <div className="flex items-center gap-2 text-zinc-400">
                          <Users className="h-4 w-4" />
                          <span>{getSPName(sender.serviceProviderId)}</span>
                        </div>
                      </td>
                      <td className="py-3">
                        <div className="flex items-center gap-2 text-zinc-400">
                          <Network className="h-4 w-4" />
                          <span>{getMnoConnectionInfo(sender.mnoConnectionId)}</span>
                        </div>
                      </td>
                      <td className="py-3 text-center">
                        <div className="flex flex-col items-center">
                          <span className="text-white font-medium">{sender.currentUsageCount.toLocaleString()}</span>
                          <span className="text-xs text-zinc-500">/ {(sender.maxUsageCount ?? 10000).toLocaleString()}</span>
                        </div>
                      </td>
                      <td className="py-3">
                        <span className="text-white">{(sender.rateLimit ?? 100)}/sec</span>
                      </td>
                      <td className="py-3">
                        <span className={cn('inline-flex px-2.5 py-0.5 rounded-full text-xs border', getStatusBadge(sender.status))}>
                          {sender.status}
                        </span>
                      </td>
                      <td className="py-3">
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => openEditModal(sender)}
                            className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/10 rounded transition-colors"
                            title="Edit"
                          >
                            <Edit className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => {
                              setDeletingSender(sender)
                              setIsDeleteModalOpen(true)
                            }}
                            className="p-1.5 text-zinc-400 hover:text-red-400 hover:bg-white/10 rounded transition-colors"
                            title="Delete"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
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

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={() => setIsModalOpen(false)} />
          <div className="relative bg-zinc-900 border border-white/10 rounded-2xl w-full max-w-lg shadow-xl">
            <div className="p-6 border-b border-white/10 flex items-center justify-between">
              <h2 className="text-xl font-semibold text-white">
                {editingSender ? 'Edit Sender ID' : 'Add Sender ID'}
              </h2>
              <button
                onClick={() => setIsModalOpen(false)}
                className="p-2 text-zinc-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">
                  Sender ID *
                </label>
                <input
                  type="text"
                  value={formData.senderIdString}
                  onChange={(e) => setFormData(prev => ({ ...prev, senderIdString: e.target.value.toUpperCase() }))}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white font-mono"
                  placeholder="e.g., MYBRAND"
                  required
                />
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">
                    Service Provider
                  </label>
                  <select
                    value={formData.serviceProviderId}
                    onChange={(e) => setFormData(prev => ({ ...prev, serviceProviderId: e.target.value }))}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  >
                    <option value="">Global</option>
                    {serviceProviders.map(sp => (
                      <option key={sp.id} value={sp.id}>{sp.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">
                    MNO Connection
                  </label>
                  <select
                    value={formData.mnoConnectionId}
                    onChange={(e) => setFormData(prev => ({ ...prev, mnoConnectionId: e.target.value }))}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  >
                    <option value="">Any Connection</option>
                    {mnoConnections.map(conn => {
                      const mno = mnos.find(m => m.id === conn.mnoId)
                      return (
                        <option key={conn.id} value={conn.id}>
                          {mno?.name || 'Unknown'} ({conn.protocol})
                        </option>
                      )
                    })}
                  </select>
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">
                    Rate Limit (per sec)
                  </label>
                  <input
                    type="number"
                    value={formData.rateLimit}
                    onChange={(e) => setFormData(prev => ({ ...prev, rateLimit: e.target.value }))}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                    min="1"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">
                    Max Usage Count
                  </label>
                  <input
                    type="number"
                    value={formData.maxUsageCount}
                    onChange={(e) => setFormData(prev => ({ ...prev, maxUsageCount: e.target.value }))}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                    min="1"
                  />
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">
                    Reset Interval (hours)
                  </label>
                  <input
                    type="number"
                    value={formData.resetIntervalHours}
                    onChange={(e) => setFormData(prev => ({ ...prev, resetIntervalHours: e.target.value }))}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                    placeholder="Optional"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">
                    Status
                  </label>
                  <select
                    value={formData.status}
                    onChange={(e) => setFormData(prev => ({ ...prev, status: e.target.value }))}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  >
                    <option value="pending">Pending</option>
                    <option value="approved">Approved</option>
                    <option value="rejected">Rejected</option>
                    <option value="blocked">Blocked</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">
                  Networks (comma-separated)
                </label>
                <input
                  type="text"
                  value={formData.networks}
                  onChange={(e) => setFormData(prev => ({ ...prev, networks: e.target.value }))}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  placeholder="e.g., MTN, AIRTEL, GLO"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">
                  OTP Template
                </label>
                <textarea
                  value={formData.otpTemplate}
                  onChange={(e) => setFormData(prev => ({ ...prev, otpTemplate: e.target.value }))}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  rows={2}
                  placeholder="Optional OTP template..."
                />
              </div>
            </div>

            <div className="p-6 border-t border-white/10 flex justify-end gap-3">
              <button
                onClick={() => setIsModalOpen(false)}
                className="px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSubmit}
                disabled={!formData.senderIdString}
                className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
              >
                {editingSender ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {isDeleteModalOpen && deletingSender && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={() => setIsDeleteModalOpen(false)} />
          <div className="relative bg-zinc-900 border border-white/10 rounded-2xl w-full max-w-md shadow-xl">
            <div className="p-6 border-b border-white/10 flex items-center justify-between">
              <h2 className="text-xl font-semibold text-white">Delete Sender ID</h2>
              <button
                onClick={() => setIsDeleteModalOpen(false)}
                className="p-2 text-zinc-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="p-6">
              <div className="flex items-start gap-4">
                <div className="p-3 bg-red-500/20 rounded-full">
                  <Trash2 className="h-6 w-6 text-red-400" />
                </div>
                <div>
                  <p className="text-white font-medium mb-1">Confirm Deletion</p>
                  <p className="text-zinc-400 text-sm">
                    Are you sure you want to delete sender ID{' '}
                    <span className="text-white font-mono">{deletingSender.senderIdString}</span>?
                  </p>
                </div>
              </div>
            </div>

            <div className="p-6 border-t border-white/10 flex justify-end gap-3">
              <button
                onClick={() => setIsDeleteModalOpen(false)}
                className="px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      {toast && (
        <div className={cn(
          'fixed bottom-6 right-6 px-4 py-3 rounded-xl shadow-lg flex items-center gap-3 animate-in slide-in-from-bottom-5',
          toast.type === 'success' ? 'bg-emerald-500/20 border border-emerald-500/30 text-emerald-400' : 'bg-red-500/20 border border-red-500/30 text-red-400'
        )}>
          {toast.type === 'success' ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
          <span className="text-sm font-medium">{toast.message}</span>
        </div>
      )}
    </div>
  )
}
