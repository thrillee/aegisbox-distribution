'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, DollarSign, Hash } from 'lucide-react'

export default function NewPricingRulePage() {
  const params = useParams()
  const router = useRouter()

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    mnoId: '',
    senderType: 'regular',
    pricePerSegment: '',
    currency: 'NGN',
    minQuantity: '1',
    maxQuantity: '',
    effectiveFrom: '',
    effectiveTo: '',
    status: 'active',
  })

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    console.log('Creating pricing rule:', formData)
    router.push('/admin/pricing')
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
        <h1 className="text-3xl font-bold text-white">New Pricing Rule</h1>
        <p className="text-zinc-400 mt-1">Create a new pricing rule for SMS messages</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Rule Details
          </h2>

          <div className="grid gap-6 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Rule Name</label>
              <input type="text" name="name" value={formData.name} onChange={handleChange}
                placeholder="e.g., MTN Nigeria Regular" required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">MNO Network</label>
              <select name="mnoId" value={formData.mnoId} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20">
                <option value="">Select MNO</option>
                <option value="1">MTN Nigeria</option>
                <option value="2">Glo Nigeria</option>
                <option value="3">Airtel Nigeria</option>
                <option value="4">9mobile Nigeria</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Price Per Segment</label>
              <input type="number" name="pricePerSegment" value={formData.pricePerSegment} onChange={handleChange}
                placeholder="e.g., 3.50" required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Currency</label>
              <select name="currency" value={formData.currency} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20">
                <option value="NGN">NGN - Nigerian Naira</option>
                <option value="USD">USD - US Dollar</option>
                <option value="GHS">GHS - Ghana Cedi</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Sender Type</label>
              <select name="senderType" value={formData.senderType} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20">
                <option value="regular">Regular</option>
                <option value="otp">OTP</option>
                <option value="premium">Premium</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
              <select name="status" value={formData.status} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20">
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Hash className="h-5 w-5" />
            Quantity Limits
          </h2>

          <div className="grid gap-6 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Min Quantity</label>
              <input type="number" name="minQuantity" value={formData.minQuantity} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Max Quantity</label>
              <input type="number" name="maxQuantity" value={formData.maxQuantity} onChange={handleChange}
                placeholder="Leave empty for unlimited"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Effective From</label>
              <input type="datetime-local" name="effectiveFrom" value={formData.effectiveFrom} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Effective To</label>
              <input type="datetime-local" name="effectiveTo" value={formData.effectiveTo} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3">
          <Link href="/admin/pricing"
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors">
            Cancel
          </Link>
          <button type="submit"
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors">
            <Save className="h-4 w-4" />
            Create Rule
          </button>
        </div>
      </form>
    </div>
  )
}
