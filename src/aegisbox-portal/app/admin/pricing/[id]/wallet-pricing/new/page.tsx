'use client'

import { use, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Plus, DollarSign, Loader2 } from 'lucide-react'

interface Mno {
  id: number
  name: string
  countryCode: string
}

export default function NewWalletPricingPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const resolvedParams = use(params)
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [mnos, setMnos] = useState<Mno[]>([])
  const [formData, setFormData] = useState({
    mnoId: '',
    currencyCode: 'NGN',
    pricePerSms: '',
  })

  useEffect(() => {
    fetchMnos()
  }, [])

  const fetchMnos = async () => {
    try {
      const response = await fetch('/api/admin/mnos')
      const data = await response.json()
      if (response.ok) {
        setMnos(data.mnos || [])
      }
    } catch (err) {
      console.error('Failed to fetch MNOs')
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const response = await fetch(`/api/admin/service-providers/${resolvedParams.id}/wallet-pricing`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })

      const data = await response.json()

      if (!response.ok) {
        setError(data.error || 'Failed to create pricing rule')
        return
      }

      router.push(`/admin/pricing/${resolvedParams.id}`)
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
          href={`/admin/pricing/${resolvedParams.id}`}
          className="p-2 hover:bg-white/10 rounded-xl transition-colors"
        >
          <ArrowLeft className="h-5 w-5 text-zinc-400" />
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-white">Add Wallet Pricing</h1>
          <p className="text-zinc-400">Configure SMS pricing for wallet-funded messages</p>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Mobile Network Operator *
            </label>
            <select
              required
              value={formData.mnoId}
              onChange={(e) => setFormData({ ...formData, mnoId: e.target.value })}
              className="w-full px-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:ring-2 focus:ring-white/20"
            >
              <option value="">Select MNO</option>
              <option value="">All Networks (Default Rate)</option>
              {mnos.map((mno) => (
                <option key={mno.id} value={mno.id}>
                  {mno.name} ({mno.countryCode})
                </option>
              ))}
            </select>
            <p className="text-xs text-zinc-500 mt-1">Leave MNO empty to set a default rate for all networks</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Currency
            </label>
            <select
              value={formData.currencyCode}
              onChange={(e) => setFormData({ ...formData, currencyCode: e.target.value })}
              className="w-full px-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:ring-2 focus:ring-white/20"
            >
              <option value="NGN">NGN - Nigerian Naira (₦)</option>
              <option value="USD">USD - US Dollar ($)</option>
              <option value="GHS">GHS - Ghanaian Cedi (₵)</option>
              <option value="KES">KES - Kenyan Shilling (KSh)</option>
              <option value="ZAR">ZAR - South African Rand (R)</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Price per SMS *
            </label>
              <input
              type="number"
              step="0.0001"
              min="0"
              required
              value={formData.pricePerSms}
              onChange={(e) => setFormData({ ...formData, pricePerSms: e.target.value })}
              className="w-full px-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20"
              placeholder="0.00"
            />
            <p className="text-xs text-zinc-500 mt-1">Price per 160-character SMS segment</p>
          </div>
        </div>

        <div className="flex gap-4">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
            {loading ? 'Creating...' : 'Create Pricing Rule'}
          </button>
          <Link
            href={`/admin/pricing/${resolvedParams.id}`}
            className="px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            Cancel
          </Link>
        </div>
      </form>
    </div>
  )
}
