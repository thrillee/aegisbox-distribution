'use client'

import { useState, useEffect, useMemo, useCallback } from 'react'
import { Plus, Search, RefreshCw, FileText, Settings, Edit, Trash2, Eye, Copy, Check, X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface OtpTemplate {
  id: number
  name: string
  contentTemplate: string
  defaultBrandName: string | null
  status: string
  createdAt: Date
  updatedAt: Date
}

interface OtpTemplatesListProps {
  initialTemplates: OtpTemplate[]
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
    suspended: 'bg-red-500/20 text-red-400',
  }

  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.inactive}`}>
      {status}
    </span>
  )
}

function TemplateModal({
  isOpen,
  onClose,
  onSubmit,
  template,
}: {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: any) => void
  template?: OtpTemplate | null
}) {
  const [formData, setFormData] = useState({
    name: '',
    contentTemplate: '',
    defaultBrandName: '',
    status: 'active',
  })
  const [previewData, setPreviewData] = useState({
    code: '123456',
    brand: 'MyApp',
    expiry: '5 minutes',
  })
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (template) {
      setFormData({
        name: template.name,
        contentTemplate: template.contentTemplate,
        defaultBrandName: template.defaultBrandName || '',
        status: template.status,
      })
      setPreviewData(prev => ({
        ...prev,
        brand: template.defaultBrandName || 'MyApp',
      }))
    } else {
      setFormData({
        name: '',
        contentTemplate: '',
        defaultBrandName: '',
        status: 'active',
      })
      setPreviewData({
        code: '123456',
        brand: 'MyApp',
        expiry: '5 minutes',
      })
    }
  }, [template, isOpen])

  const previewText = useMemo(() => {
    let text = formData.contentTemplate
      .replace(/\{\{code\}\}/g, previewData.code)
      .replace(/\{\{brand\}\}/g, previewData.brand)
      .replace(/\{\{expiry\}\}/g, previewData.expiry)
    return text
  }, [formData.contentTemplate, previewData])

  const copyToClipboard = useCallback(() => {
    navigator.clipboard.writeText(previewText)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [previewText])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={onClose} />
      <div className="relative bg-zinc-900 border border-white/10 rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-xl">
        <div className="p-6 border-b border-white/10 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-white">
            {template ? 'Edit OTP Template' : 'Create OTP Template'}
          </h2>
          <button
            onClick={onClose}
            className="p-2 text-zinc-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
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
              placeholder="e.g., Default OTP"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Content Template *
            </label>
            <textarea
              value={formData.contentTemplate}
              onChange={(e) => setFormData({ ...formData, contentTemplate: e.target.value })}
              rows={4}
              className="w-full px-3 py-2 border border-white/10 rounded-lg bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20 font-mono text-sm"
              placeholder="Use {{code}} for OTP code, {{brand}} for brand name"
            />
            <p className="mt-1 text-xs text-zinc-500">
              Variables: <code className="text-blue-400">{'{{code}}'}</code> <code className="text-blue-400">{'{{brand}}'}</code> <code className="text-blue-400">{'{{expiry}}'}</code>
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">
                Default Brand Name
              </label>
              <input
                type="text"
                value={formData.defaultBrandName}
                onChange={(e) => {
                  setFormData({ ...formData, defaultBrandName: e.target.value })
                  setPreviewData({ ...previewData, brand: e.target.value || 'MyApp' })
                }}
                className="w-full px-3 py-2 border border-white/10 rounded-lg bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                placeholder="e.g., MyApp"
              />
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
              </select>
            </div>
          </div>

          <div className="bg-zinc-800/50 border border-white/10 rounded-xl p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-medium text-white flex items-center gap-2">
                <Eye className="h-4 w-4 text-zinc-400" />
                Live Preview
              </h3>
              <button
                onClick={copyToClipboard}
                className={cn(
                  'flex items-center gap-1 px-2 py-1 rounded-lg text-xs transition-colors',
                  copied ? 'bg-emerald-500/20 text-emerald-400' : 'text-zinc-400 hover:text-white hover:bg-white/10'
                )}
              >
                {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                {copied ? 'Copied!' : 'Copy'}
              </button>
            </div>
            <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-white break-all">
              {previewText || <span className="text-zinc-600 italic">Preview will appear here...</span>}
            </div>
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
            onClick={() => onSubmit(formData)}
            disabled={!formData.name || !formData.contentTemplate}
            className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {template ? 'Update Template' : 'Create Template'}
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
  template?: OtpTemplate | null
}) {
  if (!isOpen || !template) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={onClose} />
      <div className="relative bg-zinc-900 border border-white/10 rounded-2xl w-full max-w-md shadow-xl">
        <div className="p-6 border-b border-white/10 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-white">Delete Template</h2>
          <button
            onClick={onClose}
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
                Are you sure you want to delete <span className="text-white font-medium">{template.name}</span>?
                This action cannot be undone.
              </p>
            </div>
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
            onClick={onConfirm}
            className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
          >
            Delete Template
          </button>
        </div>
      </div>
    </div>
  )
}

export function OtpTemplatesList({ initialTemplates, stats }: OtpTemplatesListProps) {
  const [templates, setTemplates] = useState(initialTemplates)
  const [searchTerm, setSearchTerm] = useState('')
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<OtpTemplate | null>(null)
  const [deletingTemplate, setDeletingTemplate] = useState<OtpTemplate | null>(null)
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)

  const showToast = useCallback((message: string, type: 'success' | 'error') => {
    setToast({ message, type })
    setTimeout(() => setToast(null), 3000)
  }, [])

  const filteredTemplates = useMemo(() => {
    if (!searchTerm.trim()) return templates
    const term = searchTerm.toLowerCase()
    return templates.filter(t => 
      t.name.toLowerCase().includes(term) ||
      t.contentTemplate.toLowerCase().includes(term)
    )
  }, [templates, searchTerm])

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      const res = await fetch('/api/admin/otp-message-templates')
      if (res.ok) {
        const data = await res.json()
        setTemplates(data.templates || [])
      }
    } catch (error) {
      console.error('Error refreshing OTP templates:', error)
    } finally {
      setIsRefreshing(false)
    }
  }

  const handleSubmit = async (data: any) => {
    try {
      const url = editingTemplate
        ? `/api/admin/otp-message-templates/${editingTemplate.id}`
        : '/api/admin/otp-message-templates'
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
        showToast(editingTemplate ? 'Template updated successfully' : 'Template created successfully', 'success')
      } else {
        const error = await res.json()
        showToast(error.error || 'Failed to save template', 'error')
      }
    } catch (error) {
      console.error('Error saving template:', error)
      showToast('Failed to save template', 'error')
    }
  }

  const handleDelete = async () => {
    if (!deletingTemplate) return

    try {
      const res = await fetch(`/api/admin/otp-message-templates/${deletingTemplate.id}`, {
        method: 'DELETE',
      })

      if (res.ok) {
        handleRefresh()
        setIsDeleteModalOpen(false)
        setDeletingTemplate(null)
        showToast('Template deleted successfully', 'success')
      } else {
        showToast('Failed to delete template', 'error')
      }
    } catch (error) {
      console.error('Error deleting template:', error)
      showToast('Failed to delete template', 'error')
    }
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">OTP Message Templates</h1>
          <p className="text-zinc-400 mt-1">
            Configure reusable OTP message templates with variable support
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
            {searchTerm ? 'No templates match your search' : 'No OTP templates found'}
          </p>
          {!searchTerm && (
            <button
              onClick={() => {
                setEditingTemplate(null)
                setIsModalOpen(true)
              }}
              className="text-white hover:underline mt-2"
            >
              Create your first OTP template
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          {filteredTemplates.map((template) => (
            <div key={template.id} className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 hover:bg-white/5 transition-colors">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="font-semibold text-white">{template.name}</h3>
                  {template.defaultBrandName && (
                    <p className="text-sm text-zinc-500">{template.defaultBrandName}</p>
                  )}
                </div>
                {getStatusBadge(template.status)}
              </div>

              <div className="bg-black/50 rounded-xl p-4 mb-4 relative group">
                <p className="text-sm text-zinc-400 font-mono pr-8">
                  {template.contentTemplate}
                </p>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    navigator.clipboard.writeText(template.contentTemplate)
                  }}
                  className="absolute top-2 right-2 p-1.5 text-zinc-500 hover:text-white hover:bg-white/20 rounded transition-colors opacity-0 group-hover:opacity-100"
                  title="Copy template"
                >
                  <Copy className="h-3.5 w-3.5" />
                </button>
              </div>

              <div className="flex items-center justify-between pt-4 border-t border-white/10">
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => {
                      setEditingTemplate(template)
                      setIsModalOpen(true)
                    }}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-zinc-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                    title="Edit"
                  >
                    <Edit className="h-3.5 w-3.5" />
                    Edit
                  </button>
                  <button
                    onClick={() => {
                      setDeletingTemplate(template)
                      setIsDeleteModalOpen(true)
                    }}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-zinc-400 hover:text-red-400 hover:bg-white/10 rounded-lg transition-colors"
                    title="Delete"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    Delete
                  </button>
                </div>
                <div className="text-xs text-zinc-500">
                  {new Date(template.updatedAt).toISOString().split('T')[0]}
                </div>
              </div>
            </div>
          ))}
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

      {toast && (
        <div className={cn(
          'fixed bottom-6 right-6 px-4 py-3 rounded-xl shadow-lg flex items-center gap-3 animate-in slide-in-from-bottom-5',
          toast.type === 'success' ? 'bg-emerald-500/20 border border-emerald-500/30 text-emerald-400' : 'bg-red-500/20 border border-red-500/30 text-red-400'
        )}>
          {toast.type === 'success' ? (
            <Check className="h-4 w-4" />
          ) : (
            <X className="h-4 w-4" />
          )}
          <span className="text-sm font-medium">{toast.message}</span>
        </div>
      )}
    </div>
  )
}
