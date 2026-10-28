'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, Globe, Loader2 } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

export default function NewMNOPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    countryCode: '',
    networkCode: '',
    status: 'active',
  })

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await fetch('/api/admin/mnos', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })

      if (response.ok) {
        router.push('/admin/mnos')
      } else {
        const error = await response.json()
        alert(error.error || 'Failed to create MNO')
      }
    } catch (error) {
      console.error('Error creating MNO:', error)
      alert('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  const commonCountries = [
    { code: '234', name: 'Nigeria' },
    { code: '233', name: 'Ghana' },
    { code: '254', name: 'Kenya' },
    { code: '255', name: 'Tanzania' },
    { code: '256', name: 'Uganda' },
    { code: '27', name: 'South Africa' },
    { code: '20', name: 'Egypt' },
    { code: '1', name: 'United States/Canada' },
    { code: '44', name: 'United Kingdom' },
    { code: '971', name: 'United Arab Emirates' },
    { code: '966', name: 'Saudi Arabia' },
  ]

  return (
    <div className="p-8 space-y-6 max-w-4xl">
      <PageHeader
        title="Add MNO"
        description="Create a new mobile network operator"
        actions={
          <Link href="/admin/mnos">
            <Button variant="outline" className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">
              <ArrowLeft className="h-4 w-4" />
              Back to MNOs
            </Button>
          </Link>
        }
      />

      <form onSubmit={handleSubmit}>
        <DataCard title="MNO Information" subtitle="Basic MNO details">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-zinc-300 mb-1">MNO Name *</label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
                maxLength={100}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
                placeholder="e.g., MTN Nigeria"
              />
            </div>
            <div>
              <label htmlFor="countryCode" className="block text-sm font-medium text-zinc-300 mb-1">Country Code *</label>
              <select
                id="countryCode"
                name="countryCode"
                value={formData.countryCode}
                onChange={handleChange}
                required
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              >
                <option value="">Select Country</option>
                {commonCountries.map((country) => (
                  <option key={country.code} value={country.code}>{country.name} (+{country.code})</option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="networkCode" className="block text-sm font-medium text-zinc-300 mb-1">Network Code</label>
              <input
                type="text"
                id="networkCode"
                name="networkCode"
                value={formData.networkCode}
                onChange={handleChange}
                maxLength={10}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
                placeholder="e.g., 62120"
              />
              <p className="mt-1 text-xs text-zinc-500">MCC/MNC code for the network</p>
            </div>
            <div>
              <label htmlFor="status" className="block text-sm font-medium text-zinc-300 mb-1">Status *</label>
              <select
                id="status"
                name="status"
                value={formData.status}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
          </div>
        </DataCard>

        <div className="flex items-center justify-end gap-4 mt-6">
          <Link href="/admin/mnos">
            <Button variant="outline" className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5">
              Cancel
            </Button>
          </Link>
          <Button type="submit" disabled={loading} className="bg-white text-black hover:bg-zinc-200">
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                Create MNO
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  )
}
