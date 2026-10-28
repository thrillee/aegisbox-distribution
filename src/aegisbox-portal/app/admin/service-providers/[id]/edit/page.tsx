'use client'

import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, Building2, Wallet, Route, Loader2 } from 'lucide-react'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

export default function EditServiceProviderPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id

  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(true)
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    currency: 'NGN',
    walletType: 'prepaid',
    creditLimit: '0',
    initialBalance: '0',
    routingGroup: '',
    status: 'pending',
  })

  useEffect(() => {
    if (id) {
      fetchProvider()
    }
  }, [id])

  const fetchProvider = async () => {
    try {
      const response = await fetch(`/api/admin/service-providers/${id}`)
      if (response.ok) {
        const data = await response.json()
        setFormData({
          name: data.provider?.name || '',
          email: data.provider?.email || '',
          phone: data.provider?.phone || '',
          currency: data.provider?.defaultCurrencyCode || 'NGN',
          walletType: data.provider?.walletType || 'prepaid',
          creditLimit: data.provider?.creditLimit || '0',
          initialBalance: data.provider?.initialBalance || '0',
          routingGroup: data.provider?.routingGroupId || '',
          status: data.provider?.status || 'pending',
        })
      }
    } catch (error) {
      console.error('Error fetching provider:', error)
    } finally {
      setFetching(false)
    }
  }

  const routingGroups = [
    { id: '1', name: 'MTN Nigeria Route', reference: 'MTN_NG_ROUTE' },
    { id: '2', name: 'Glo Nigeria Route', reference: 'GLO_NG_ROUTE' },
    { id: '3', name: 'Airtel Nigeria Route', reference: 'AIRTEL_NG_ROUTE' },
    { id: '4', name: 'Ghana Default Route', reference: 'GH_DEFAULT' },
    { id: '5', name: 'Egypt Default Route', reference: 'EG_DEFAULT' },
    { id: '6', name: 'System Default Route', reference: 'SYS_DEFAULT' },
  ]

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await fetch(`/api/admin/service-providers/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          email: formData.email,
          phone: formData.phone,
          currency: formData.currency,
          walletType: formData.walletType,
          creditLimit: parseFloat(formData.creditLimit) || 0,
          initialBalance: parseFloat(formData.initialBalance) || 0,
          routingGroupId: parseInt(formData.routingGroup) || null,
          status: formData.status,
        }),
      })

      if (response.ok) {
        router.push(`/admin/service-providers/${id}`)
      } else {
        const error = await response.json()
        alert(error.error || 'Failed to update service provider')
      }
    } catch (error) {
      console.error('Error updating SP:', error)
      alert('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  if (fetching) {
    return <PageLoadingSkeleton />
  }

  return (
    <div className="p-8 max-w-4xl mx-auto space-y-6">
      <div>
        <Link href={`/admin/service-providers/${id}`} className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to Service Provider
        </Link>
      </div>

      <div>
        <h1 className="text-3xl font-bold text-white">Edit Service Provider</h1>
        <p className="text-zinc-400 mt-1">Update service provider account</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Basic Information
          </h2>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-zinc-400 mb-1">Company Name *</label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
                minLength={2}
                maxLength={255}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              />
            </div>
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-zinc-400 mb-1">Email Address *</label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                required
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              />
            </div>
            <div>
              <label htmlFor="phone" className="block text-sm font-medium text-zinc-400 mb-1">Phone Number</label>
              <input
                type="tel"
                id="phone"
                name="phone"
                value={formData.phone}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              />
            </div>
            <div>
              <label htmlFor="status" className="block text-sm font-medium text-zinc-400 mb-1">Status *</label>
              <select
                id="status"
                name="status"
                value={formData.status}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              >
                <option value="pending">Pending</option>
                <option value="active">Active</option>
                <option value="suspended">Suspended</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
            <div>
              <label htmlFor="currency" className="block text-sm font-medium text-zinc-400 mb-1">Default Currency *</label>
              <select
                id="currency"
                name="currency"
                value={formData.currency}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              >
                <option value="NGN">NGN - Nigerian Naira</option>
                <option value="USD">USD - US Dollar</option>
                <option value="GHS">GHS - Ghanaian Cedi</option>
                <option value="KES">KES - Kenyan Shilling</option>
                <option value="ZAR">ZAR - South African Rand</option>
              </select>
            </div>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Wallet className="h-5 w-5" />
            Wallet Configuration
          </h2>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label htmlFor="walletType" className="block text-sm font-medium text-zinc-400 mb-1">Wallet Type *</label>
              <select
                id="walletType"
                name="walletType"
                value={formData.walletType}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              >
                <option value="prepaid">Prepaid (pay before sending)</option>
                <option value="postpaid">Postpaid (pay after sending)</option>
              </select>
            </div>
            <div>
              <label htmlFor="creditLimit" className="block text-sm font-medium text-zinc-400 mb-1">Credit Limit (for postpaid)</label>
              <input
                type="number"
                id="creditLimit"
                name="creditLimit"
                value={formData.creditLimit}
                onChange={handleChange}
                min="0"
                step="0.01"
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              />
            </div>
            <div>
              <label htmlFor="initialBalance" className="block text-sm font-medium text-zinc-400 mb-1">Add Balance</label>
              <input
                type="number"
                id="initialBalance"
                name="initialBalance"
                value={formData.initialBalance}
                onChange={handleChange}
                min="0"
                step="0.01"
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              />
            </div>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Route className="h-5 w-5" />
            Routing Configuration
          </h2>
          <div>
            <label htmlFor="routingGroup" className="block text-sm font-medium text-zinc-400 mb-1">Default Routing Group *</label>
            <select
              id="routingGroup"
              name="routingGroup"
              value={formData.routingGroup}
              onChange={handleChange}
              className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
            >
              <option value="">Select a routing group</option>
              {routingGroups.map((group) => (
                <option key={group.id} value={group.id}>{group.name} ({group.reference})</option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex items-center justify-end gap-4">
          <Link
            href={`/admin/service-providers/${id}`}
            className="px-4 py-2 border border-white/10 rounded-lg text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {loading ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  )
}
