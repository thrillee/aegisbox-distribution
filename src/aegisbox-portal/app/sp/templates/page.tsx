'use client'

import { useState, useEffect } from 'react'
import { Plus, Search, Edit2, Trash2, Eye, Copy, FileText, Hash } from 'lucide-react'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface Template {
  id: number
  name: string | null
  content: string | null
  status: string | null
  createdAt: Date | null
}

const categories = ['All', 'Authentication', 'Transactional', 'Financial', 'Marketing', 'Security']

export default function TemplatesPage() {
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [initialLoad, setInitialLoad] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('All')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null)
  const [viewingTemplate, setViewingTemplate] = useState<Template | null>(null)

  useEffect(() => {
    fetchTemplates()
  }, [])

  async function fetchTemplates() {
    try {
      const response = await fetch('/api/sp/templates')
      const data = await response.json()
      if (data.templates) {
        setTemplates(data.templates)
      }
    } catch (error) {
      console.error('Error fetching templates:', error)
    } finally {
      setLoading(false)
      setInitialLoad(false)
    }
  }

  const filteredTemplates = templates.filter((template) => {
    const matchesSearch = (template.name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
      (template.content || '').toLowerCase().includes(searchQuery.toLowerCase())
    const matchesCategory = selectedCategory === 'All'
    return matchesSearch && matchesCategory
  })

  function extractVariables(content: string): string[] {
    const matches = content.match(/\{([^}]+)\}/g)
    return matches ? matches.map(v => v.slice(1, -1)) : []
  }

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Message Templates</h1>
          <p className="text-zinc-400 mt-1">
            Create and manage reusable message templates
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
        >
          <Plus className="h-4 w-4" />
          Create Template
        </button>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
          <input
            type="text"
            placeholder="Search templates..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
          />
        </div>
        <div className="flex items-center gap-2 bg-zinc-800 border border-white/10 rounded-lg p-1">
          {categories.map((category) => (
            <button
              key={category}
              onClick={() => setSelectedCategory(category)}
              className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                selectedCategory === category
                  ? 'bg-white text-black'
                  : 'text-zinc-400 hover:text-white hover:bg-white/5'
              }`}
            >
              {category}
            </button>
          ))}
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-zinc-500">Loading templates...</div>
        ) : filteredTemplates.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-zinc-800/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Template Name</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Variables</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Created</th>
                  <th className="px-6 py-3 text-right text-xs font-semibold text-zinc-400 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {filteredTemplates.map((template) => {
                  const variables = template.content ? extractVariables(template.content) : []
                  return (
                    <tr key={template.id} className="hover:bg-white/5 transition-colors">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-3">
                          <div className="flex items-center justify-center w-10 h-10 bg-white/10 text-white rounded-lg">
                            <FileText className="h-5 w-5" />
                          </div>
                          <div>
                            <div className="font-medium text-white">{template.name || 'Untitled'}</div>
                            <div className="text-sm text-zinc-500 truncate max-w-xs">
                              {template.content?.substring(0, 50)}...
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          template.status === 'active'
                            ? 'bg-emerald-500/20 text-emerald-400'
                            : 'bg-zinc-500/20 text-zinc-400'
                        }`}>
                          {template.status || 'Unknown'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-1">
                          {variables.slice(0, 3).map((variable) => (
                            <span
                              key={variable}
                              className="inline-flex items-center gap-1 px-2 py-0.5 bg-blue-500/20 text-blue-400 rounded text-xs font-mono"
                            >
                              <Hash className="h-3 w-3" />
                              {variable}
                            </span>
                          ))}
                          {variables.length > 3 && (
                            <span className="text-xs text-zinc-500">+{variables.length - 3} more</span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-zinc-400">
                        {template.createdAt ? new Date(template.createdAt).toLocaleDateString() : 'Unknown'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => setViewingTemplate(template)}
                            className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/5 rounded transition-colors"
                            title="View"
                          >
                            <Eye className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => setEditingTemplate(template)}
                            className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/5 rounded transition-colors"
                            title="Edit"
                          >
                            <Edit2 className="h-4 w-4" />
                          </button>
                          <button
                            className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/5 rounded transition-colors"
                            title="Duplicate"
                          >
                            <Copy className="h-4 w-4" />
                          </button>
                          <button
                            className="p-1.5 text-red-400 hover:text-red-300 hover:bg-red-500/10 rounded transition-colors"
                            title="Delete"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-8 text-center text-zinc-500">No templates found</div>
        )}
      </div>

      {showCreateModal && (
        <TemplateModal
          onClose={() => setShowCreateModal(false)}
          onSave={async (templateData) => {
            try {
              const response = await fetch('/api/sp/templates', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(templateData),
              })
              if (response.ok) {
                await fetchTemplates()
              }
            } catch (error) {
              console.error('Error creating template:', error)
            }
            setShowCreateModal(false)
          }}
        />
      )}

      {editingTemplate && (
        <TemplateModal
          template={editingTemplate}
          onClose={() => setEditingTemplate(null)}
          onSave={async (templateData) => {
            console.log('Updating template:', templateData)
            setEditingTemplate(null)
          }}
        />
      )}

      {viewingTemplate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-zinc-900 border border-white/10 rounded-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-white/10">
              <h2 className="text-xl font-semibold text-white">{viewingTemplate.name}</h2>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <h3 className="text-sm font-medium text-zinc-400 mb-2">Template Content</h3>
                <div className="p-4 bg-black/50 rounded-xl border border-white/10">
                  <p className="text-white whitespace-pre-wrap">
                    {viewingTemplate.content?.split(/(\{[^}]+\})/).map((part, index) => {
                      if (part.startsWith('{') && part.endsWith('}')) {
                        return (
                          <span
                            key={index}
                            className="inline-flex items-center px-1.5 py-0.5 bg-blue-500/20 text-blue-400 rounded font-mono text-sm mx-0.5"
                          >
                            {part}
                          </span>
                        )
                      }
                      return part
                    })}
                  </p>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h3 className="text-sm font-medium text-zinc-400 mb-1">Status</h3>
                  <p className="text-white">{viewingTemplate.status}</p>
                </div>
                <div>
                  <h3 className="text-sm font-medium text-zinc-400 mb-1">Created</h3>
                  <p className="text-white">
                    {viewingTemplate.createdAt ? new Date(viewingTemplate.createdAt).toLocaleString() : 'Unknown'}
                  </p>
                </div>
              </div>
            </div>
            <div className="p-6 border-t border-white/10 flex justify-end">
              <button
                onClick={() => setViewingTemplate(null)}
                className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function TemplateModal({
  template,
  onClose,
  onSave,
}: {
  template?: Template
  onClose: () => void
  onSave: (template: any) => void
}) {
  const [name, setName] = useState(template?.name || '')
  const [templateContent, setTemplateContent] = useState(template?.content || '')

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-zinc-900 border border-white/10 rounded-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6 border-b border-white/10">
          <h2 className="text-xl font-semibold text-white">
            {template ? 'Edit Template' : 'Create New Template'}
          </h2>
        </div>
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Template Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., OTP Verification"
              className="w-full px-3 py-2 bg-black/50 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">Template Content</label>
            <textarea
              value={templateContent}
              onChange={(e) => setTemplateContent(e.target.value)}
              placeholder="Enter your message template. Use {variable_name} for dynamic content."
              rows={6}
              className="w-full px-3 py-2 bg-black/50 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20 resize-none"
            />
            <p className="text-xs text-zinc-500 mt-1">
              Use curly braces around variables, e.g., <code className="px-1 bg-white/10 rounded">{'{code}'}</code>
            </p>
          </div>
        </div>
        <div className="p-6 border-t border-white/10 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => onSave({ name, content: templateContent })}
            className="px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
          >
            {template ? 'Save Changes' : 'Create Template'}
          </button>
        </div>
      </div>
    </div>
  )
}
