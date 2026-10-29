'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Save, Building2, Route, Loader2, Mail, User, Key, Globe } from 'lucide-react'
import { SmppLoginCreated, type SmppConnection } from '@/components/service-providers/smpp-login-created'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface RoutingGroup {
  id: number
  name: string
  reference: string
  description: string | null
}

export default function NewServiceProviderPage() {
  const router = useRouter()
  const searchParams = useSearchParams()

  const [loading, setLoading] = useState(false)
  const [initialLoad, setInitialLoad] = useState(true)
  const [created, setCreated] = useState<{ systemId: string; password: string; connection: SmppConnection | null } | null>(null)
  const [routingGroups, setRoutingGroups] = useState<RoutingGroup[]>([])
  const [formData, setFormData] = useState({
    name: searchParams.get('name') || '',
    email: searchParams.get('email') || '',
    currency: searchParams.get('currency') || 'NGN',
    adminFirstName: '',
    adminLastName: '',
    createSmpp: true,
    smppSystemId: '',
    smppPassword: '',
    smppBindType: 'trx',
    smppRoutingGroupId: '',
    createHttp: true,
    httpName: '',
    httpRoutingGroupId: '',
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
    } catch (error) {
      console.error('Failed to fetch routing groups:', error)
    } finally {
      setInitialLoad(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const target = e.target
    const value = target.type === 'checkbox' ? (target as HTMLInputElement).checked : target.value
    setFormData(prev => ({ ...prev, [target.name]: value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await fetch('/api/admin/service-providers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          email: formData.email,
          currency: formData.currency,
          adminFirstName: formData.adminFirstName,
          adminLastName: formData.adminLastName,
          credentials: {
            smpp: formData.createSmpp ? {
              systemId: formData.smppSystemId || undefined,
              password: formData.smppPassword || undefined,
              bindType: formData.smppBindType,
              routingGroupId: formData.smppRoutingGroupId ? parseInt(formData.smppRoutingGroupId) : undefined,
            } : null,
            http: formData.createHttp ? {
              name: formData.httpName || formData.name + ' API Key',
              routingGroupId: formData.httpRoutingGroupId ? parseInt(formData.httpRoutingGroupId) : undefined,
            } : null,
          },
        }),
      })

      const data = await response.json()

      if (response.ok) {
        const smpp = data.credentials?.find((c: { type: string }) => c.type === 'smpp')
        if (smpp?.password) {
          setCreated({ systemId: smpp.systemId, password: smpp.password, connection: data.connection ?? null })
        } else {
          router.push('/admin/service-providers')
        }
      } else {
        alert(data.error || 'Failed to create service provider')
      }
    } catch (error) {
      console.error('Error creating SP:', error)
      alert('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  if (created) {
    return (
      <div className="p-8 max-w-2xl">
        <SmppLoginCreated {...created} onDone={() => router.push('/admin/service-providers')} />
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6 max-w-4xl">
      <div className="flex items-center gap-4">
        <Link href="/admin/service-providers" className="p-2 hover:bg-white/10 rounded-xl transition-colors">
          <ArrowLeft className="h-5 w-5 text-zinc-400" />
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-white">Add Service Provider</h1>
          <p className="text-zinc-400 mt-1">Create a new service provider account</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Company Information
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
                className="w-full px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                placeholder="Your Company Name"
              />
            </div>
            <div>
              <label htmlFor="currency" className="block text-sm font-medium text-zinc-400 mb-1">Default Currency *</label>
              <select
                id="currency"
                name="currency"
                value={formData.currency}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white focus:outline-none focus:ring-2 focus:ring-white/20"
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

        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <User className="h-5 w-5" />
            Admin User (Clerk)
          </h2>
          <p className="text-sm text-zinc-400 mb-4">
            An organization will be created and an invitation sent to this user.
          </p>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label htmlFor="adminFirstName" className="block text-sm font-medium text-zinc-400 mb-1">Admin First Name *</label>
              <input
                type="text"
                id="adminFirstName"
                name="adminFirstName"
                value={formData.adminFirstName}
                onChange={handleChange}
                required
                maxLength={100}
                className="w-full px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                placeholder="John"
              />
            </div>
            <div>
              <label htmlFor="adminLastName" className="block text-sm font-medium text-zinc-400 mb-1">Admin Last Name *</label>
              <input
                type="text"
                id="adminLastName"
                name="adminLastName"
                value={formData.adminLastName}
                onChange={handleChange}
                required
                maxLength={100}
                className="w-full px-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                placeholder="Doe"
              />
            </div>
            <div className="md:col-span-2">
              <label htmlFor="email" className="block text-sm font-medium text-zinc-400 mb-1">Admin Email Address *</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
                <input
                  type="email"
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleChange}
                  required
                  className="w-full pl-10 pr-4 py-2 border border-white/10 rounded-xl bg-black/50 text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20"
                  placeholder="admin@company.com"
                />
              </div>
              <p className="mt-1 text-xs text-zinc-500">Invitation will be sent to this email address</p>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-4">
          <Link
            href="/admin/service-providers"
            className="px-4 py-2 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-xl hover:bg-zinc-200 transition-colors disabled:opacity-50"
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {loading ? 'Creating...' : 'Create Service Provider'}
          </button>
        </div>
      </form>
    </div>
  )
}
