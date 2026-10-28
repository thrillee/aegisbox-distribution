'use client'

import { useState, useEffect, useCallback } from 'react'
import { Plus, Search, RefreshCw, Link2, Link2Off, Edit, Trash2, X, Check, ChevronDown } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface Assignment {
  id: number
  otpAlternativeSenderId: number
  otpMessageTemplateId: number
  mnoId: number | null
  priority: number
  status: string
  assignmentUsageCount: number
  createdAt: string
  updatedAt: string
  senderName: string | null
  templateName: string | null
  mnoName: string | null
  mnoCountryCode: string | null
}

interface PaginationData {
  page: number
  limit: number
  total: number
  totalPages: number
}

interface Sender {
  id: number
  senderIdString: string
  status: string
}

interface Template {
  id: number
  name: string
  status: string
}

interface MNO {
  id: number
  name: string
  countryCode: string | null
}

const STATUS_OPTIONS = [
  { value: 'all', label: 'All Statuses' },
  { value: 'active', label: 'Active' },
  { value: 'inactive', label: 'Inactive' },
]

export default function SenderTemplateAssignmentsPage() {
  const [loading, setLoading] = useState(true)
  const [initialLoad, setInitialLoad] = useState(true)
  const [assignments, setAssignments] = useState<Assignment[]>([])
  const [pagination, setPagination] = useState<PaginationData>({ page: 1, limit: 50, total: 0, totalPages: 0 })
  const [senders, setSenders] = useState<Sender[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [mnos, setMnos] = useState<MNO[]>([])
  const [filters, setFilters] = useState({
    status: 'all',
    otpAlternativeSenderId: '',
    otpMessageTemplateId: '',
    mnoId: '',
  })
  const [error, setError] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [editingAssignment, setEditingAssignment] = useState<Assignment | null>(null)
  const [deletingAssignment, setDeletingAssignment] = useState<Assignment | null>(null)
  const [formData, setFormData] = useState({
    otpAlternativeSenderId: '',
    otpMessageTemplateId: '',
    mnoId: '',
    priority: '0',
    status: 'active',
  })
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)

  const showToast = useCallback((message: string, type: 'success' | 'error') => {
    setToast({ message, type })
    setTimeout(() => setToast(null), 3000)
  }, [])

  const fetchDropdowns = useCallback(async () => {
    try {
      const [sendersRes, templatesRes, mnosRes] = await Promise.all([
        fetch('/api/admin/otp-alternative-senders'),
        fetch('/api/admin/otp-message-templates'),
        fetch('/api/admin/mnos'),
      ])
      
      if (sendersRes.ok) {
        const data = await sendersRes.json()
        setSenders(data.senders || data.data || [])
      }
      if (templatesRes.ok) {
        const data = await templatesRes.json()
        setTemplates(data.templates || data.data || [])
      }
      if (mnosRes.ok) {
        const data = await mnosRes.json()
        setMnos(data.mnos || data.data || [])
      }
    } catch (err) {
      console.error('Error fetching dropdowns:', err)
    }
  }, [])

  const fetchAssignments = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const params = new URLSearchParams()
      params.set('page', String(pagination.page))
      params.set('limit', String(pagination.limit))
      if (filters.status !== 'all') params.set('status', filters.status)
      if (filters.otpAlternativeSenderId) params.set('otpAlternativeSenderId', filters.otpAlternativeSenderId)
      if (filters.otpMessageTemplateId) params.set('otpMessageTemplateId', filters.otpMessageTemplateId)
      if (filters.mnoId) params.set('mnoId', filters.mnoId)

      const res = await fetch(`/api/admin/otp-sender-template-assignments?${params}`)
      if (res.ok) {
        const data = await res.json()
        setAssignments(data.assignments || [])
        setPagination(prev => ({ ...prev, ...data.pagination }))
      } else {
        setError('Failed to fetch assignments')
      }
    } catch (err) {
      setError('An error occurred')
    } finally {
      setLoading(false)
      setInitialLoad(false)
    }
  }, [pagination.page, pagination.limit, filters])

  useEffect(() => {
    fetchDropdowns()
  }, [fetchDropdowns])

  useEffect(() => {
    fetchAssignments()
  }, [fetchAssignments])

  const openCreateModal = () => {
    setEditingAssignment(null)
    setFormData({
      otpAlternativeSenderId: '',
      otpMessageTemplateId: '',
      mnoId: '',
      priority: '0',
      status: 'active',
    })
    setIsModalOpen(true)
  }

  const openEditModal = (assignment: Assignment) => {
    setEditingAssignment(assignment)
    setFormData({
      otpAlternativeSenderId: String(assignment.otpAlternativeSenderId),
      otpMessageTemplateId: String(assignment.otpMessageTemplateId),
      mnoId: assignment.mnoId ? String(assignment.mnoId) : '',
      priority: String(assignment.priority),
      status: assignment.status,
    })
    setIsModalOpen(true)
  }

  const handleSubmit = async () => {
    try {
      const url = editingAssignment
        ? `/api/admin/otp-sender-template-assignments/${editingAssignment.id}`
        : '/api/admin/otp-sender-template-assignments'
      const method = editingAssignment ? 'PUT' : 'POST'

      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          otpAlternativeSenderId: parseInt(formData.otpAlternativeSenderId),
          otpMessageTemplateId: parseInt(formData.otpMessageTemplateId),
          mnoId: formData.mnoId ? parseInt(formData.mnoId) : null,
          priority: parseInt(formData.priority),
        }),
      })

      if (res.ok) {
        fetchAssignments()
        setIsModalOpen(false)
        showToast(editingAssignment ? 'Assignment updated successfully' : 'Assignment created successfully', 'success')
      } else {
        const error = await res.json()
        showToast(error.error || 'Failed to save assignment', 'error')
      }
    } catch (err) {
      showToast('Failed to save assignment', 'error')
    }
  }

  const handleDelete = async () => {
    if (!deletingAssignment) return

    try {
      const res = await fetch(`/api/admin/otp-sender-template-assignments/${deletingAssignment.id}`, {
        method: 'DELETE',
      })

      if (res.ok) {
        fetchAssignments()
        setIsDeleteModalOpen(false)
        setDeletingAssignment(null)
        showToast('Assignment deleted successfully', 'success')
      } else {
        showToast('Failed to delete assignment', 'error')
      }
    } catch (err) {
      showToast('Failed to delete assignment', 'error')
    }
  }

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      active: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
      inactive: 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30',
    }
    return colors[status] || colors.inactive
  }

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Sender-Template Assignments"
        description="Link OTP senders to message templates"
        actions={
          <Button onClick={openCreateModal} className="bg-white text-black hover:bg-zinc-200">
            <Plus className="h-4 w-4 mr-2" />
            New Assignment
          </Button>
        }
      />

      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
          {error}
        </div>
      )}

      <DataCard title="Filters" subtitle="Filter assignments">
        <div className="grid gap-4 md:grid-cols-5">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Sender</label>
            <select
              value={filters.otpAlternativeSenderId}
              onChange={(e) => setFilters(prev => ({ ...prev, otpAlternativeSenderId: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            >
              <option value="">All Senders</option>
              {senders.map(s => (
                <option key={s.id} value={s.id}>{s.senderIdString}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Template</label>
            <select
              value={filters.otpMessageTemplateId}
              onChange={(e) => setFilters(prev => ({ ...prev, otpMessageTemplateId: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            >
              <option value="">All Templates</option>
              {templates.map(t => (
                <option key={t.id} value={t.id}>{t.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">MNO</label>
            <select
              value={filters.mnoId}
              onChange={(e) => setFilters(prev => ({ ...prev, mnoId: e.target.value }))}
              className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
            >
              <option value="">All MNOs</option>
              {mnos.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.countryCode})</option>
              ))}
            </select>
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
          <div className="flex items-end">
            <Button
              onClick={fetchAssignments}
              disabled={loading}
              className="w-full bg-white text-black hover:bg-zinc-200"
            >
              {loading ? <RefreshCw className="h-4 w-4 animate-spin" /> : 'Apply'}
            </Button>
          </div>
        </div>
      </DataCard>

      <DataCard title="Assignments" subtitle={`${pagination.total} assignments found`}>
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <RefreshCw className="h-8 w-8 animate-spin text-zinc-500" />
          </div>
        ) : assignments.length === 0 ? (
          <div className="text-center py-12 text-zinc-500">
            <Link2 className="h-12 w-12 mx-auto mb-4 text-zinc-600" />
            <p>No assignments found</p>
            <Button onClick={openCreateModal} className="mt-4" variant="outline">
              Create your first assignment
            </Button>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-xs font-semibold text-zinc-400 uppercase border-b border-white/10">
                    <th className="pb-3">Sender</th>
                    <th className="pb-3">Template</th>
                    <th className="pb-3">MNO</th>
                    <th className="pb-3 text-center">Priority</th>
                    <th className="pb-3 text-center">Usage</th>
                    <th className="pb-3">Status</th>
                    <th className="pb-3">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/10">
                  {assignments.map((assignment) => (
                    <tr key={assignment.id} className="hover:bg-white/5">
                      <td className="py-3">
                        <div className="flex items-center gap-2">
                          <Link2 className="h-4 w-4 text-zinc-500" />
                          <span className="font-mono text-white">{assignment.senderName || `ID: ${assignment.otpAlternativeSenderId}`}</span>
                        </div>
                      </td>
                      <td className="py-3 text-zinc-300">{assignment.templateName || `ID: ${assignment.otpMessageTemplateId}`}</td>
                      <td className="py-3 text-zinc-400">
                        {assignment.mnoName ? (
                          <span>{assignment.mnoName} {assignment.mnoCountryCode && `(${assignment.mnoCountryCode})`}</span>
                        ) : (
                          <span className="text-zinc-600">Global</span>
                        )}
                      </td>
                      <td className="py-3 text-center">
                        <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-white/10 text-white font-medium">
                          {assignment.priority}
                        </span>
                      </td>
                      <td className="py-3 text-center text-zinc-400">
                        {assignment.assignmentUsageCount.toLocaleString()}
                      </td>
                      <td className="py-3">
                        <span className={cn('inline-flex items-center px-2.5 py-0.5 rounded-full text-xs border', getStatusBadge(assignment.status))}>
                          {assignment.status}
                        </span>
                      </td>
                      <td className="py-3">
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => openEditModal(assignment)}
                            className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/10 rounded transition-colors"
                            title="Edit"
                          >
                            <Edit className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => {
                              setDeletingAssignment(assignment)
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
                {editingAssignment ? 'Edit Assignment' : 'New Assignment'}
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
                  Sender *
                </label>
                <select
                  value={formData.otpAlternativeSenderId}
                  onChange={(e) => setFormData(prev => ({ ...prev, otpAlternativeSenderId: e.target.value }))}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  required
                >
                  <option value="">Select sender...</option>
                  {senders.filter(s => s.status === 'active').map(s => (
                    <option key={s.id} value={s.id}>{s.senderIdString}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">
                  Template *
                </label>
                <select
                  value={formData.otpMessageTemplateId}
                  onChange={(e) => setFormData(prev => ({ ...prev, otpMessageTemplateId: e.target.value }))}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                  required
                >
                  <option value="">Select template...</option>
                  {templates.filter(t => t.status === 'active').map(t => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-400 mb-1">
                  MNO (Optional)
                </label>
                <select
                  value={formData.mnoId}
                  onChange={(e) => setFormData(prev => ({ ...prev, mnoId: e.target.value }))}
                  className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                >
                  <option value="">Global (All MNOs)</option>
                  {mnos.map(m => (
                    <option key={m.id} value={m.id}>{m.name} ({m.countryCode})</option>
                  ))}
                </select>
                <p className="mt-1 text-xs text-zinc-500">Leave empty to apply globally</p>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-zinc-400 mb-1">
                    Priority
                  </label>
                  <input
                    type="number"
                    value={formData.priority}
                    onChange={(e) => setFormData(prev => ({ ...prev, priority: e.target.value }))}
                    className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
                    min="0"
                  />
                  <p className="mt-1 text-xs text-zinc-500">Higher = more priority</p>
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
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                  </select>
                </div>
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
                disabled={!formData.otpAlternativeSenderId || !formData.otpMessageTemplateId}
                className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
              >
                {editingAssignment ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {isDeleteModalOpen && deletingAssignment && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={() => setIsDeleteModalOpen(false)} />
          <div className="relative bg-zinc-900 border border-white/10 rounded-2xl w-full max-w-md shadow-xl">
            <div className="p-6 border-b border-white/10 flex items-center justify-between">
              <h2 className="text-xl font-semibold text-white">Delete Assignment</h2>
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
                  <Link2Off className="h-6 w-6 text-red-400" />
                </div>
                <div>
                  <p className="text-white font-medium mb-1">Confirm Deletion</p>
                  <p className="text-zinc-400 text-sm">
                    Are you sure you want to delete this assignment? This will unlink{' '}
                    <span className="text-white">{deletingAssignment.senderName}</span> from{' '}
                    <span className="text-white">{deletingAssignment.templateName}</span>.
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
