'use client'

import { useState, useEffect, useMemo } from 'react'
import Link from 'next/link'
import { Plus, Search, RefreshCw, FileText, Settings, Eye, Trash2, Edit } from 'lucide-react'

interface Template {
  id: number
  name: string
  content: string
  status: string
  serviceProviderId: number | null
  createdAt: Date
  updatedAt: Date
}

interface ServiceProvider {
  id: number
  name: string
}

interface TemplatesListProps {
  initialTemplates: Template[]
  serviceProviders: ServiceProvider[]
  stats: {
    total: number
    active: number
    inactive: number
  }
}

function getStatusBadge(status: string) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-400',
    inactive: 'bg-zinc-500/20 text-zinc-400',
    pending: 'bg-amber-500/20 text-amber-400',
    suspended: 'bg-red-500/20 text-red-400',
  }

  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.inactive}`}>
      {status}
    </span>
  )
}

function getServiceProviderName(spId: number | null, serviceProviders: ServiceProvider[]) {
  if (!spId) return 'Global'
  const sp = serviceProviders.find(sp => sp.id === spId)
  return sp?.name || 'Unknown'
}

function TemplateModal({
  isOpen,
  onClose,
  onSubmit,
  template,
  serviceProviders,
}: {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: any) => void
  template?: Template | null
  serviceProviders: ServiceProvider[]
}) {
  const [formData, setFormData] = useState({
    name: '',
    content: '',
    status: 'active',
    serviceProviderId: '',
  })

  useEffect(() => {
    if (template) {
      setFormData({
        name: template.name,
        content: template.content,
        status: template.status,
        serviceProviderId: template.serviceProviderId?.toString() || '',
      })
    } else {
      setFormData({
        name: '',
        content: '',
        status: 'active',
        serviceProviderId: '',
      })
    }
  }, [template, isOpen])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto backdrop-blur-xl">
        <div className="p-6 border-b border-white/10">
          <h2 className="text-xl font-semibold text-white">
            {template ? 'Edit Template' : 'Create Template'}
          </h2>
        </div>

        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Template Name *
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-lg bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
              placeholder="Enter template name"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Content *
            </label>
            <textarea
              value={formData.content}
              onChange={(e) => setFormData({ ...formData, content: e.target.value })}
              rows={5}
              className="w-full px-3 py-2 border border-white/10 rounded-lg bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
              placeholder="Enter template content"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Service Provider
            </label>
            <select
              value={formData.serviceProviderId}
              onChange={(e) => setFormData({ ...formData, serviceProviderId: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-lg bg-black/50 text-white focus:outline-none focus:ring-2 focus:ring-white/20"
            >
              <option value="">Global (All Service Providers)</option>
              {serviceProviders.map((sp) => (
                <option key={sp.id} value={sp.id.toString()}>
                  {sp.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Status
            </label>
            <select
              value={formData.status}
              onChange={(e) => setFormData({ ...formData, status: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-lg bg-black/50 text-white focus:outline-none focus:ring-2 focus:ring-white/20"
            >
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
              <option value="pending">Pending</option>
              <option value="suspended">Suspended</option>
            </select>
          </div>
        </div>

        <div className="p-6 border-t border-white/10 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              onSubmit({
                ...formData,
                serviceProviderId: formData.serviceProviderId ? parseInt(formData.serviceProviderId) : null,
              })
            }}
            className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
          >
            {template ? 'Update' : 'Create'}
          </button>
        </div>
      </div>
    </div>
  )
}

function DeleteModal({
  isOpen,
  onClose,
  onConfirm,
  template,
}: {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  template?: Template | null
}) {
  if (!isOpen || !template) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl w-full max-w-md backdrop-blur-xl">
        <div className="p-6 border-b border-white/10">
          <h2 className="text-xl font-semibold text-white">Delete Template</h2>
        </div>

        <div className="p-6">
          <p className="text-zinc-400">
            Are you sure you want to delete template <strong className="text-white">{template.name}</strong>? This action cannot be undone.
          </p>
        </div>

        <div className="p-6 border-t border-white/10 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  )
}

export function TemplatesList({ initialTemplates, serviceProviders, stats }: TemplatesListProps) {
  const [templates, setTemplates] = useState(initialTemplates)
  const [searchTerm, setSearchTerm] = useState('')
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null)
  const [deletingTemplate, setDeletingTemplate] = useState<Template | null>(null)

  const filteredTemplates = useMemo(() => {
    if (!searchTerm.trim()) return templates
    const term = searchTerm.toLowerCase()
    return templates.filter(t => 
      t.name.toLowerCase().includes(term) ||
      t.content.toLowerCase().includes(term)
    )
  }, [templates, searchTerm])

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      const res = await fetch('/api/admin/templates')
      if (res.ok) {
        const data = await res.json()
        setTemplates(data.templates || [])
      }
    } catch (error) {
      console.error('Error refreshing templates:', error)
    } finally {
      setIsRefreshing(false)
    }
  }

  const handleSubmit = async (data: any) => {
    try {
      const url = editingTemplate
        ? `/api/admin/templates/${editingTemplate.id}`
        : '/api/admin/templates'
      const method = editingTemplate ? 'PUT' : 'POST'

      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      })

      if (res.ok) {
        handleRefresh()
        setIsModalOpen(false)
        setEditingTemplate(null)
      } else {
        const error = await res.json()
        alert(error.error || 'Failed to save template')
      }
    } catch (error) {
      console.error('Error saving template:', error)
      alert('Failed to save template')
    }
  }

  const handleDelete = async () => {
    if (!deletingTemplate) return

    try {
      const res = await fetch(`/api/admin/templates/${deletingTemplate.id}`, {
        method: 'DELETE',
      })

      if (res.ok) {
        handleRefresh()
        setIsDeleteModalOpen(false)
        setDeletingTemplate(null)
      } else {
        alert('Failed to delete template')
      }
    } catch (error) {
      console.error('Error deleting template:', error)
      alert('Failed to delete template')
    }
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Message Templates</h1>
          <p className="text-zinc-400 mt-1">
            Create and manage SMS message templates
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleRefresh}
            disabled={isRefreshing}
            className="flex items-center gap-2 px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
            Refresh
          </button>
          <button
            onClick={() => {
              setEditingTemplate(null)
              setIsModalOpen(true)
            }}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
          >
            <Plus className="h-4 w-4" />
            Create Template
          </button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-500/20 rounded-lg">
              <FileText className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Templates</p>
              <p className="text-2xl font-bold text-white">{stats.total}</p>
            </div>
          </div>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-500/20 rounded-lg">
              <Settings className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Active</p>
              <p className="text-2xl font-bold text-emerald-400">{stats.active}</p>
            </div>
          </div>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-zinc-500/20 rounded-lg">
              <FileText className="h-5 w-5 text-zinc-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Inactive</p>
              <p className="text-2xl font-bold text-zinc-400">{stats.inactive}</p>
            </div>
          </div>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-500/20 rounded-lg">
              <FileText className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Global Templates</p>
              <p className="text-2xl font-bold text-white">
                {templates.filter(t => !t.serviceProviderId).length}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
        <input
          type="text"
          placeholder="Search templates..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full pl-10 pr-4 py-2 bg-zinc-900/50 border border-white/10 rounded-lg text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
        />
      </div>

      {filteredTemplates.length === 0 ? (
        <div className="text-center py-12 bg-zinc-900/50 border border-white/10 rounded-2xl">
          <FileText className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
          <p className="text-zinc-500">
            {searchTerm ? 'No templates match your search' : 'No templates found'}
          </p>
          {!searchTerm && (
            <button
              onClick={() => {
                setEditingTemplate(null)
                setIsModalOpen(true)
              }}
              className="text-white hover:underline mt-2"
            >
              Create your first template
            </button>
          )}
        </div>
      ) : (
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl overflow-hidden">
          <table className="w-full">
            <thead className="bg-black/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Content Preview</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Service Provider</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-400 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-zinc-400 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {filteredTemplates.map((template) => (
                <tr key={template.id} className="hover:bg-white/5">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-white">{template.name}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-zinc-400 max-w-xs truncate">
                      {template.content}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-zinc-400">
                      {getServiceProviderName(template.serviceProviderId, serviceProviders)}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getStatusBadge(template.status)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => {
                          setEditingTemplate(template)
                          setIsModalOpen(true)
                        }}
                        className="p-1.5 text-zinc-500 hover:text-white hover:bg-white/10 rounded transition-colors"
                        title="Edit"
                      >
                        <Edit className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => {
                          setDeletingTemplate(template)
                          setIsDeleteModalOpen(true)
                        }}
                        className="p-1.5 text-zinc-500 hover:text-red-400 hover:bg-white/10 rounded transition-colors"
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
      )}

      <TemplateModal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false)
          setEditingTemplate(null)
        }}
        onSubmit={handleSubmit}
        template={editingTemplate}
        serviceProviders={serviceProviders}
      />

      <DeleteModal
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false)
          setDeletingTemplate(null)
        }}
        onConfirm={handleDelete}
        template={deletingTemplate}
      />
    </div>
  )
}
