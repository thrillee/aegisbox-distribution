'use client'

import { use, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Loader2, Plus } from 'lucide-react'

interface Mno {
  id: number
  name: string
  countryCode: string
}

export default function NewPricingPage({
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
      const response = await fetch(`/api/admin/service-providers/${resolvedParams.id}/pricing`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })

      const data = await response.json()

      if (!response.ok) {
        setError(data.error || 'Failed to create pricing rule')
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
          className="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
        >
          <ArrowLeft className="h-5 w-5 text-slate-600 dark:text-slate-400" />
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">Add Pricing Rule</h1>
          <p className="text-slate-600 dark:text-slate-400">Configure SMS pricing for MNO routes</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-600 dark:text-red-400">
            {error}
          </div>
        )}

        <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Mobile Network Operator *
            </label>
            <select
              required
              value={formData.mnoId}
              onChange={(e) => setFormData({ ...formData, mnoId: e.target.value })}
              className="w-full px-3 py-2 border border-slate-200 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white"
            >
              <option value="">Select MNO</option>
              <option value="">All Networks (Default Rate)</option>
              {mnos.map((mno) => (
                <option key={mno.id} value={mno.id}>
                  {mno.name} ({mno.countryCode})
                </option>
              ))}
            </select>
            <p className="text-xs text-slate-500 mt-1">Leave MNO empty to set a default rate for all networks</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Currency
            </label>
            <select
              value={formData.currencyCode}
              onChange={(e) => setFormData({ ...formData, currencyCode: e.target.value })}
              className="w-full px-3 py-2 border border-slate-200 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white"
            >
              <option value="NGN">NGN - Nigerian Naira (₦)</option>
              <option value="USD">USD - US Dollar ($)</option>
              <option value="EUR">EUR - Euro (€)</option>
              <option value="GBP">GBP - British Pound (£)</option>
              <option value="EGP">EGP - Egyptian Pound (E£)</option>
              <option value="KES">KES - Kenyan Shilling (KSh)</option>
              <option value="GHS">GHS - Ghanaian Cedi (₵)</option>
              <option value="ZAR">ZAR - South African Rand (R)</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Price per SMS *
            </label>
            <input
              type="number"
              step="0.0001"
              min="0"
              required
              value={formData.pricePerSms}
              onChange={(e) => setFormData({ ...formData, pricePerSms: e.target.value })}
              className="w-full px-3 py-2 border border-slate-200 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white"
              placeholder="0.00"
            />
            <p className="text-xs text-slate-500 mt-1">Price per 160-character SMS segment</p>
          </div>
        </div>

        <div className="flex gap-4">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            <Plus className="h-4 w-4" />
            Create Pricing Rule
          </button>
          <Link
            href={`/admin/service-providers/${resolvedParams.id}`}
            className="px-4 py-2 border border-slate-200 dark:border-slate-600 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
          >
            Cancel
          </Link>
        </div>
      </form>
    </div>
  )
}
