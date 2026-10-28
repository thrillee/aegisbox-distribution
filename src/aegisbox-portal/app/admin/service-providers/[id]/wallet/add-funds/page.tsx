'use client'

import { use, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Loader2, Plus } from 'lucide-react'

export default function AddFundsPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const resolvedParams = use(params)
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [formData, setFormData] = useState({
    amount: '',
    currency: 'NGN',
    description: '',
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setSuccess('')

    try {
      const response = await fetch(`/api/admin/service-providers/${resolvedParams.id}/wallet/fund`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })

      const data = await response.json()

      if (!response.ok) {
        setError(data.error || 'Failed to add funds')
        return
      }

      setSuccess(`Successfully added ${formData.currency} ${parseFloat(formData.amount).toLocaleString()} to wallet`)
      setFormData({ ...formData, amount: '', description: '' })

      setTimeout(() => {
        router.push(`/admin/service-providers/${resolvedParams.id}`)
      }, 2000)
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
          <h1 className="text-3xl font-bold text-white">Add Funds</h1>
          <p className="text-zinc-400">Add funds to service provider wallet</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
            {error}
          </div>
        )}

        {success && (
          <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-emerald-400">
            {success}
          </div>
        )}

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Amount *
            </label>
            <input
              type="number"
              step="0.01"
              min="0.01"
              required
              value={formData.amount}
              onChange={(e) => setFormData({ ...formData, amount: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
              placeholder="0.00"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Currency
            </label>
            <select
              value={formData.currency}
              onChange={(e) => setFormData({ ...formData, currency: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white focus:outline-none focus:ring-2 focus:ring-white/20"
            >
              <option value="NGN">NGN - Nigerian Naira</option>
              <option value="USD">USD - US Dollar</option>
              <option value="EUR">EUR - Euro</option>
              <option value="GBP">GBP - British Pound</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Description
            </label>
            <input
              type="text"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
              placeholder="e.g., Monthly top-up, Emergency credit"
            />
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
            Add Funds
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
