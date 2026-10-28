'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { ArrowLeft, Save, Route, Loader2 } from 'lucide-react'

export default function NewRoutingGroupPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    reference: '',
    isDefault: false,
  })

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const target = e.target as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    const name = target.name
    if (target instanceof HTMLInputElement && target.type === 'checkbox') {
      setFormData(prev => ({ ...prev, [name]: target.checked }))
    } else {
      setFormData(prev => ({ ...prev, [name]: target.value }))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await fetch('/api/admin/routing-groups', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          description: formData.description,
          reference: formData.reference,
          isDefault: formData.isDefault,
        }),
      })

      if (response.ok) {
        router.push('/admin/routing')
      } else {
        const error = await response.json()
        alert(error.error || 'Failed to create routing group')
      }
    } catch (error) {
      console.error('Error creating routing group:', error)
      alert('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="mb-6">
        <Link href="/admin/routing" className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to Routing
        </Link>
      </div>

      <div className="mb-6">
        <h1 className="text-3xl font-bold text-white">New Routing Group</h1>
        <p className="text-zinc-400 mt-1">Create a new routing group for message delivery</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Route className="h-5 w-5" />
            Routing Group Details
          </h2>

          <div className="grid gap-6 md:grid-cols-2">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-zinc-400 mb-1">Name *</label>
              <input type="text" id="name" name="name" value={formData.name} onChange={handleChange}
                placeholder="e.g., MTN Nigeria Route" required maxLength={255}
                className="w-full px-3 py-2 bg-black/50 border border-white/10 rounded-xl text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>

            <div>
              <label htmlFor="reference" className="block text-sm font-medium text-zinc-400 mb-1">Reference *</label>
              <input type="text" id="reference" name="reference" value={formData.reference} onChange={handleChange}
                placeholder="e.g., MTN_NG_ROUTE" required maxLength={100} pattern="[A-Z0-9_]+"
                className="w-full px-3 py-2 bg-black/50 border border-white/10 rounded-xl text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" />
              <p className="mt-1 text-xs text-zinc-500">Uppercase letters, numbers, and underscores only</p>
            </div>

            <div className="md:col-span-2">
              <label htmlFor="description" className="block text-sm font-medium text-zinc-400 mb-1">Description</label>
              <textarea id="description" name="description" value={formData.description} onChange={handleChange}
                rows={3}
                placeholder="Describe the purpose of this routing group"
                className="w-full px-3 py-2 bg-black/50 border border-white/10 rounded-xl text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>

            <div>
              <label className="flex items-center gap-3 cursor-pointer">
                <input type="checkbox" name="isDefault" checked={formData.isDefault} onChange={handleChange}
                  className="w-5 h-5 rounded border-zinc-600 bg-black/50 text-white focus:ring-white/20" />
                <span className="text-sm font-medium text-zinc-400">Set as default routing group</span>
              </label>
              <p className="mt-1 text-xs text-zinc-500">Messages will use this group by default if no specific routing is configured</p>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3">
          <Link href="/admin/routing"
            className="px-4 py-2 bg-white/10 text-zinc-400 rounded-xl hover:text-white hover:bg-white/20 transition-colors">
            Cancel
          </Link>
          <button type="submit" disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-xl hover:bg-zinc-200 transition-colors disabled:opacity-50">
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {loading ? 'Creating...' : 'Create Routing Group'}
          </button>
        </div>
      </form>
    </div>
  )
}
