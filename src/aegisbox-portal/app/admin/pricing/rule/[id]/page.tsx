'use client'

import { use, useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, DollarSign, Loader2, AlertCircle, CheckCircle } from 'lucide-react'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface PricingRule {
  id: number
  serviceProviderId: number | null
  mnoId: number | null
  currencyCode: string | null
  pricePerSms: string | null
  createdAt: string | null
  updatedAt: string | null
}

export default function PricingRuleEditPage() {
  const params = useParams()
  const router = useRouter()
  const pricingId = params.id as string

  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [pricingRule, setPricingRule] = useState<PricingRule | null>(null)
  const [formData, setFormData] = useState({
    mnoId: '',
    currencyCode: 'NGN',
    pricePerSms: '',
  })

  useEffect(() => {
    fetchPricingRule()
  }, [pricingId])

  const fetchPricingRule = async () => {
    try {
      const response = await fetch(`/api/admin/pricing/${pricingId}`)
      if (response.ok) {
        const data = await response.json()
        const rule = data.pricingRule
        setPricingRule(rule)
        setFormData({
          mnoId: rule.mnoId ? String(rule.mnoId) : '',
          currencyCode: rule.currencyCode || 'NGN',
          pricePerSms: rule.pricePerSms || '',
        })
      } else {
        setError('Pricing rule not found')
      }
    } catch (err) {
      setError('Failed to fetch pricing rule')
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError('')
    setSuccess('')

    try {
      const payload = {
        mnoId: formData.mnoId ? parseInt(formData.mnoId) : null,
        currencyCode: formData.currencyCode,
        pricePerSms: formData.pricePerSms || null,
      }

      const response = await fetch(`/api/admin/pricing/${pricingId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })

      if (response.ok) {
        setSuccess('Pricing rule updated successfully')
        fetchPricingRule()
      } else {
        const data = await response.json()
        setError(data.error || 'Failed to update pricing rule')
      }
    } catch (err) {
      setError('An error occurred')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <PageLoadingSkeleton />
  }

  if (!pricingRule) {
    return (
      <div className="p-8">
        <Link href="/admin/pricing" className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to Pricing
        </Link>
        <div className="mt-8 text-center">
          <h1 className="text-2xl font-bold text-white">Pricing Rule Not Found</h1>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="mb-6">
        <Link href="/admin/pricing" className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to Pricing
        </Link>
      </div>

      <div className="mb-6">
        <h1 className="text-3xl font-bold text-white">Edit Pricing Rule</h1>
        <p className="text-zinc-400 mt-1">
          Pricing Rule #{pricingRule.id}
        </p>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 flex items-center gap-2">
          <AlertCircle className="h-5 w-5" />
          {error}
        </div>
      )}

      {success && (
        <div className="mb-6 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-emerald-400 flex items-center gap-2">
          <CheckCircle className="h-5 w-5" />
          {success}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Pricing Details
          </h2>

          <div className="grid gap-6 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">MNO Network</label>
              <select
                name="mnoId"
                value={formData.mnoId}
                onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="">Select MNO</option>
                <option value="1">MTN Nigeria</option>
                <option value="2">Glo Nigeria</option>
                <option value="3">Airtel Nigeria</option>
                <option value="4">9mobile Nigeria</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Price Per SMS</label>
              <input
                type="number"
                name="pricePerSms"
                value={formData.pricePerSms}
                onChange={handleChange}
                placeholder="e.g., 3.50"
                step="0.000001"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Currency</label>
              <select
                name="currencyCode"
                value={formData.currencyCode}
                onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20"
              >
                <option value="NGN">NGN - Nigerian Naira</option>
                <option value="USD">USD - US Dollar</option>
                <option value="GHS">GHS - Ghana Cedi</option>
              </select>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3">
          <Link
            href="/admin/pricing"
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={saving}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  )
}
