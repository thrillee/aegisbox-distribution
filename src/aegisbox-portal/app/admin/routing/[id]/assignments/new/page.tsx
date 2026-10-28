'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter, useParams } from 'next/navigation'
import { ArrowLeft, Save, Route, Loader2 } from 'lucide-react'

interface PrefixGroup {
  id: number
  name: string
  reference: string
}

interface MNO {
  id: number
  name: string
  countryCode: string
}

export default function NewAssignmentPage() {
  const router = useRouter()
  const params = useParams()
  const groupId = params.id

  const [loading, setLoading] = useState(false)
  const [loadingData, setLoadingData] = useState(true)
  const [prefixGroups, setPrefixGroups] = useState<PrefixGroup[]>([])
  const [mnos, setMnos] = useState<MNO[]>([])
  const [formData, setFormData] = useState({
    msisdnPrefixGroupId: '',
    mnoId: '',
    priority: '0',
    weight: '100',
    maxRetryAttempts: '2',
    isAvailable: true,
    comment: '',
    status: 'active',
  })

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [pgRes, mnoRes] = await Promise.all([
          fetch('/api/admin/msisdn-prefix-groups'),
          fetch('/api/admin/mnos'),
        ])
        if (pgRes.ok) {
          const data = await pgRes.json()
          setPrefixGroups(data.groups || [])
        }
        if (mnoRes.ok) {
          const data = await mnoRes.json()
          setMnos(data.mnos || [])
        }
      } catch (error) {
        console.error('Error fetching data:', error)
      } finally {
        setLoadingData(false)
      }
    }
    fetchData()
  }, [])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const target = e.target
    if (target instanceof HTMLInputElement && target.type === 'checkbox') {
      setFormData(prev => ({ ...prev, [target.name]: target.checked }))
    } else {
      setFormData(prev => ({ ...prev, [target.name]: target.value }))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await fetch(`/api/admin/routing-groups/${groupId}/assignments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          msisdnPrefixGroupId: parseInt(formData.msisdnPrefixGroupId),
          mnoId: parseInt(formData.mnoId),
          priority: parseInt(formData.priority),
          weight: parseInt(formData.weight),
          maxRetryAttempts: parseInt(formData.maxRetryAttempts),
          isAvailable: formData.isAvailable,
          comment: formData.comment,
          status: formData.status,
        }),
      })

      if (response.ok) {
        router.push(`/admin/routing/${groupId}`)
      } else {
        const error = await response.json()
        alert(error.error || 'Failed to create assignment')
      }
    } catch (error) {
      console.error('Error creating assignment:', error)
      alert('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  if (loadingData) {
    return (
      <div className="p-8 max-w-4xl mx-auto">
        <div className="mb-6 animate-pulse">
          <div className="h-6 bg-zinc-800 rounded w-32 mb-4"></div>
          <div className="h-8 bg-zinc-800 rounded w-64 mb-8"></div>
          <div className="h-64 bg-zinc-800 rounded-xl"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="mb-6">
        <Link href={`/admin/routing/${groupId}`} className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to Routing Group
        </Link>
      </div>

      <div className="mb-6">
        <h1 className="text-3xl font-bold text-white">Add Routing Rule</h1>
        <p className="text-zinc-400 mt-1">Assign a prefix group to an MNO for this routing group</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Route className="h-5 w-5" />
            Assignment Details
          </h2>

          <div className="grid gap-6 md:grid-cols-2">
            <div>
              <label htmlFor="msisdnPrefixGroupId" className="block text-sm font-medium text-zinc-400 mb-1">Prefix Group *</label>
              <select
                id="msisdnPrefixGroupId"
                name="msisdnPrefixGroupId"
                value={formData.msisdnPrefixGroupId}
                onChange={handleChange}
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="">Select Prefix Group</option>
                {prefixGroups.map((pg) => (
                  <option key={pg.id} value={pg.id}>{pg.name} ({pg.reference})</option>
                ))}
              </select>
              {prefixGroups.length === 0 && (
                <p className="mt-1 text-xs text-zinc-500">
                  No prefix groups available. <Link href="/admin/routing/prefix-groups/new" className="text-blue-400 hover:text-blue-300">Create one first</Link>
                </p>
              )}
            </div>

            <div>
              <label htmlFor="mnoId" className="block text-sm font-medium text-zinc-400 mb-1">MNO *</label>
              <select
                id="mnoId"
                name="mnoId"
                value={formData.mnoId}
                onChange={handleChange}
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="">Select MNO</option>
                {mnos.map((mno) => (
                  <option key={mno.id} value={mno.id}>{mno.name} (+{mno.countryCode})</option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="priority" className="block text-sm font-medium text-zinc-400 mb-1">Priority</label>
              <input
                type="number"
                id="priority"
                name="priority"
                value={formData.priority}
                onChange={handleChange}
                min="0"
                step="1"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              />
              <p className="mt-1 text-xs text-zinc-500">Lower values = higher priority</p>
            </div>

            <div>
              <label htmlFor="weight" className="block text-sm font-medium text-zinc-400 mb-1">Weight</label>
              <input
                type="number"
                id="weight"
                name="weight"
                value={formData.weight}
                onChange={handleChange}
                min="0"
                max="100"
                step="1"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              />
              <p className="mt-1 text-xs text-zinc-500">Routing weight (0-100, higher = more traffic)</p>
            </div>

            <div>
              <label htmlFor="maxRetryAttempts" className="block text-sm font-medium text-zinc-400 mb-1">Max Retry Attempts</label>
              <input
                type="number"
                id="maxRetryAttempts"
                name="maxRetryAttempts"
                value={formData.maxRetryAttempts}
                onChange={handleChange}
                min="0"
                max="5"
                step="1"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              />
              <p className="mt-1 text-xs text-zinc-500">Number of retries before marking as failed (0-5)</p>
            </div>

            <div>
              <label htmlFor="status" className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
              <select
                id="status"
                name="status"
                value={formData.status}
                onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>

            <div className="flex items-center pt-2">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  id="isAvailable"
                  name="isAvailable"
                  checked={formData.isAvailable}
                  onChange={handleChange}
                  className="w-5 h-5 rounded border-zinc-600 bg-black/50 text-white focus:ring-white/20"
                />
                <div>
                  <span className="text-sm font-medium text-zinc-300">Initially Available</span>
                  <p className="text-xs text-zinc-500">Route will be used for message routing</p>
                </div>
              </label>
            </div>

            <div className="md:col-span-2">
              <label htmlFor="comment" className="block text-sm font-medium text-zinc-400 mb-1">Comment</label>
              <textarea
                id="comment"
                name="comment"
                value={formData.comment}
                onChange={handleChange}
                rows={2}
                placeholder="Optional notes about this routing rule"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
              />
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3">
          <Link
            href={`/admin/routing/${groupId}`}
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={loading || !formData.msisdnPrefixGroupId || !formData.mnoId}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {loading ? 'Creating...' : 'Add Rule'}
          </button>
        </div>
      </form>
    </div>
  )
}
