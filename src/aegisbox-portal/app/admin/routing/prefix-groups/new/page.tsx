'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, Hash, Loader2, X, Plus } from 'lucide-react'

export default function NewPrefixGroupPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [prefixInput, setPrefixInput] = useState('')
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    reference: '',
    prefixes: [] as string[],
  })

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const addPrefix = () => {
    const prefix = prefixInput.trim()
    if (prefix && !formData.prefixes.includes(prefix)) {
      setFormData(prev => ({ ...prev, prefixes: [...prev.prefixes, prefix] }))
      setPrefixInput('')
    }
  }

  const removePrefix = (prefix: string) => {
    setFormData(prev => ({ ...prev, prefixes: prev.prefixes.filter(p => p !== prefix) }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await fetch('/api/admin/msisdn-prefix-groups', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          description: formData.description,
          reference: formData.reference,
          prefixes: formData.prefixes,
        }),
      })

      if (response.ok) {
        router.push('/admin/routing/prefix-groups')
      } else {
        const error = await response.json()
        alert(error.error || 'Failed to create prefix group')
      }
    } catch (error) {
      console.error('Error creating prefix group:', error)
      alert('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  const commonPrefixes = [
    '0701', '0702', '0703', '0704', '0705',
    '0802', '0803', '0804', '0805', '0806', '0807', '0808', '0809',
    '0810', '0811', '0812', '0813', '0814', '0815', '0816', '0817', '0818', '0819',
    '0901', '0902', '0903', '0904', '0905', '0906',
    '+1', '+44', '+234', '+233', '+254',
  ]

  return (
    <div className="p-8 space-y-6 max-w-4xl">
      <div className="flex items-center gap-4">
        <Link href="/admin/routing/prefix-groups" className="p-2 hover:bg-white/10 rounded-xl transition-colors">
          <ArrowLeft className="h-5 w-5 text-zinc-400" />
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-white">Create Prefix Group</h1>
          <p className="text-zinc-400 mt-1">Create a new MSISDN prefix group for routing</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 space-y-6">
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-zinc-400 mb-1">Name *</label>
            <input type="text" id="name" name="name" value={formData.name} onChange={handleChange}
              required maxLength={255}
              className="w-full px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" placeholder="e.g., MTN Nigeria Prefixes" />
          </div>
          <div>
            <label htmlFor="reference" className="block text-sm font-medium text-zinc-400 mb-1">Reference *</label>
            <input type="text" id="reference" name="reference" value={formData.reference} onChange={handleChange}
              required maxLength={100} pattern="[A-Z0-9_]+"
              className="w-full px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" placeholder="e.g., MTN_NG_PREFIXES" />
            <p className="mt-1 text-xs text-zinc-500">Uppercase letters, numbers, and underscores only</p>
          </div>
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-zinc-400 mb-1">Description</label>
          <textarea id="description" name="description" value={formData.description} onChange={handleChange}
            rows={3}
            className="w-full px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" placeholder="Describe the purpose of this prefix group" />
        </div>

        <div className="border-t border-white/10 pt-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Hash className="h-5 w-5" />
            Prefixes
          </h2>

          <div className="flex gap-2 mb-4">
            <input type="text" value={prefixInput} onChange={(e) => setPrefixInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addPrefix())}
              placeholder="Enter prefix (e.g., 0803, +234)"
              className="flex-1 px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" />
            <button type="button" onClick={addPrefix}
              className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-xl hover:bg-zinc-200 transition-colors">
              <Plus className="h-4 w-4" />
              Add
            </button>
          </div>

          <div className="mb-4">
            <p className="text-sm text-zinc-500 mb-2">Quick add common prefixes:</p>
            <div className="flex flex-wrap gap-2">
              {commonPrefixes.map((prefix) => (
                <button type="button" key={prefix} onClick={() => setPrefixInput(prefix)}
                  className="px-3 py-1 text-sm bg-white/10 hover:bg-white/20 text-white rounded-lg transition-colors">
                  {prefix}
                </button>
              ))}
            </div>
          </div>

          {formData.prefixes.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {formData.prefixes.map((prefix) => (
                <span key={prefix} className="inline-flex items-center gap-1 px-3 py-1 bg-white/10 text-white rounded-full text-sm">
                  {prefix}
                  <button type="button" onClick={() => removePrefix(prefix)} className="hover:text-red-400">
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
            </div>
          )}

          {formData.prefixes.length === 0 && (
            <p className="text-sm text-zinc-500 italic">No prefixes added yet</p>
          )}
        </div>

        <div className="border-t border-white/10 pt-6 flex items-center justify-end gap-4">
          <Link href="/admin/routing/prefix-groups"
            className="px-4 py-2 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors">
            Cancel
          </Link>
          <button type="submit" disabled={loading || formData.prefixes.length === 0}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-xl hover:bg-zinc-200 transition-colors disabled:opacity-50">
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {loading ? 'Creating...' : 'Create Prefix Group'}
          </button>
        </div>
      </form>
    </div>
  )
}
