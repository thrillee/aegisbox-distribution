'use client'

import { use, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Loader2, Plus, Route } from 'lucide-react'

interface RoutingGroup {
  id: number
  name: string
  reference: string
  description: string | null
}

export default function NewSmppPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const resolvedParams = use(params)
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [fetchingRg, setFetchingRg] = useState(true)
  const [routingGroups, setRoutingGroups] = useState<RoutingGroup[]>([])
  const [error, setError] = useState('')
  const [formData, setFormData] = useState({
    systemId: '',
    password: '',
    bindType: 'trx',
    scope: 'default',
    routingGroupId: '',
    host: '',
    port: '',
  })

  useEffect(() => {
    fetchRoutingGroups()
  }, [])

  const fetchRoutingGroups = async () => {
    try {
      const response = await fetch('/api/admin/routing-groups')
      const data = await response.json()
      if (response.ok) {
        setRoutingGroups(data.routingGroups || [])
      }
    } catch (err) {
      console.error('Failed to fetch routing groups')
    } finally {
      setFetchingRg(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const response = await fetch(`/api/admin/service-providers/${resolvedParams.id}/smpp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          routingGroupId: formData.routingGroupId ? parseInt(formData.routingGroupId) : null,
        }),
      })

      const data = await response.json()

      if (!response.ok) {
        setError(data.error || 'Failed to create SMPP credential')
        return
      }

      router.push(`/admin/service-providers/${resolvedParams.id}`)
    } catch (err) {
      setError('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-8 max-w-2xl">
      <div className="flex items-center gap-4 mb-6">
        <Link
          href={`/admin/service-providers/${resolvedParams.id}`}
          className="p-2 hover:bg-white/10 rounded-xl transition-colors"
        >
          <ArrowLeft className="h-5 w-5 text-zinc-400" />
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-white">Add SMPP Credential</h1>
          <p className="text-zinc-400">Configure a new SMPP connection</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
            {error}
          </div>
        )}

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              System ID *
            </label>
            <input
              type="text"
              required
              value={formData.systemId}
              onChange={(e) => setFormData({ ...formData, systemId: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
              placeholder="Enter SMPP system ID"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Password *
            </label>
            <input
              type="password"
              required
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
              placeholder="Enter SMPP password"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">
                Bind Type
              </label>
              <select
                value={formData.bindType}
                onChange={(e) => setFormData({ ...formData, bindType: e.target.value })}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="trx">Transceiver (TRX)</option>
                <option value="tx">Transmitter (TX)</option>
                <option value="rx">Receiver (RX)</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">
                Scope
              </label>
              <input
                type="text"
                value={formData.scope}
                onChange={(e) => setFormData({ ...formData, scope: e.target.value })}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                placeholder="default"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1 flex items-center gap-2">
              <Route className="h-4 w-4" />
              Routing Group
            </label>
            <select
              value={formData.routingGroupId}
              onChange={(e) => setFormData({ ...formData, routingGroupId: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              disabled={fetchingRg}
            >
              <option value="">Select routing group</option>
              {routingGroups.map((rg) => (
                <option key={rg.id} value={rg.id}>
                  {rg.name} ({rg.reference})
                </option>
              ))}
            </select>
            <p className="text-xs text-zinc-500 mt-1">
              Messages will be routed according to this group's rules
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">
                Host
              </label>
              <input
                type="text"
                value={formData.host}
                onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                placeholder="localhost"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">
                Port
              </label>
              <input
                type="number"
                value={formData.port}
                onChange={(e) => setFormData({ ...formData, port: e.target.value })}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                placeholder="2775"
              />
            </div>
          </div>
        </div>

        <div className="flex gap-4">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-xl hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            <Plus className="h-4 w-4" />
            Create Credential
          </button>
          <Link
            href={`/admin/service-providers/${resolvedParams.id}`}
            className="px-4 py-2 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            Cancel
          </Link>
        </div>
      </form>
    </div>
  )
}
