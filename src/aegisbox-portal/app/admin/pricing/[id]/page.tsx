'use client'

import { use, useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowLeft, Plus, Search, DollarSign, Loader2, Edit, Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

interface Mno {
  id: number
  name: string
  countryCode: string
}

interface PricingRule {
  id: number
  mnoId: number | null
  mnoName: string | null
  mnoCountryCode: string | null
  currencyCode: string
  pricePerSms: string
  createdAt: string
}

export default function ServiceProviderPricingPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const resolvedParams = use(params)
  const [loading, setLoading] = useState(true)
  const [serviceProvider, setServiceProvider] = useState<{ id: number; name: string } | null>(null)
  const [httpPricing, setHttpPricing] = useState<PricingRule[]>([])
  const [walletPricing, setWalletPricing] = useState<PricingRule[]>([])
  const [mnos, setMnos] = useState<Mno[]>([])
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    fetchData()
  }, [resolvedParams.id])

  const fetchData = async () => {
    try {
      const [spRes, httpRes, walletRes, mnosRes] = await Promise.all([
        fetch(`/api/admin/service-providers/${resolvedParams.id}`),
        fetch(`/api/admin/service-providers/${resolvedParams.id}/pricing`),
        fetch(`/api/admin/service-providers/${resolvedParams.id}/wallet-pricing`),
        fetch('/api/admin/mnos'),
      ])

      if (spRes.ok) {
        const data = await spRes.json()
        setServiceProvider(data.provider)
      }

      if (httpRes.ok) {
        const data = await httpRes.json()
        setHttpPricing(data.pricingRules || [])
      }

      if (walletRes.ok) {
        const data = await walletRes.json()
        setWalletPricing(data.pricingRules || [])
      }

      if (mnosRes.ok) {
        const data = await mnosRes.json()
        setMnos(data.mnos || [])
      }
    } catch (error) {
      console.error('Error fetching data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (ruleId: number, type: 'http' | 'wallet') => {
    if (!confirm('Are you sure you want to delete this pricing rule?')) return

    try {
      const response = await fetch(`/api/admin/pricing/${ruleId}`, { method: 'DELETE' })
      if (response.ok) {
        if (type === 'http') {
          setHttpPricing(prev => prev.filter(r => r.id !== ruleId))
        } else {
          setWalletPricing(prev => prev.filter(r => r.id !== ruleId))
        }
      }
    } catch (error) {
      console.error('Error deleting rule:', error)
    }
  }

  const filterRules = (rules: PricingRule[]) => {
    if (!searchTerm) return rules
    const term = searchTerm.toLowerCase()
    return rules.filter(r =>
      r.mnoName?.toLowerCase().includes(term) ||
      r.currencyCode.toLowerCase().includes(term)
    )
  }

  const getCurrencySymbol = (currency: string) => {
    switch (currency) {
      case 'NGN': return '₦'
      case 'USD': return '$'
      case 'GHS': return '₵'
      case 'KES': return 'KSh'
      case 'ZAR': return 'R'
      default: return currency
    }
  }

  const PricingTable = ({ rules, type }: { rules: PricingRule[], type: 'http' | 'wallet' }) => (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead className="bg-zinc-800/50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">MNO</th>
            <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Currency</th>
            <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Price/SMS</th>
            <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Created</th>
            <th className="px-6 py-3 text-right text-xs font-semibold text-zinc-400 uppercase">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-white/10">
          {filterRules(rules).length > 0 ? filterRules(rules).map((rule) => (
            <tr key={rule.id} className="hover:bg-white/5 transition-colors">
              <td className="px-6 py-4">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-white">
                    {rule.mnoName || 'Default (All Networks)'}
                  </span>
                  {rule.mnoCountryCode && (
                    <span className="text-xs text-zinc-500">({rule.mnoCountryCode})</span>
                  )}
                </div>
              </td>
              <td className="px-6 py-4 text-white">{rule.currencyCode}</td>
              <td className="px-6 py-4 font-medium text-white">
                {getCurrencySymbol(rule.currencyCode)}{parseFloat(rule.pricePerSms).toFixed(4)}
              </td>
              <td className="px-6 py-4 text-zinc-400">
                {new Date(rule.createdAt).toLocaleDateString()}
              </td>
              <td className="px-6 py-4 text-right">
                <div className="flex items-center justify-end gap-2">
                  <Link
                    href={`/admin/pricing/rule/${rule.id}`}
                    className="p-1.5 hover:bg-white/10 rounded transition-colors"
                    title="Edit"
                  >
                    <Edit className="h-4 w-4 text-zinc-400" />
                  </Link>
                  <button
                    onClick={() => handleDelete(rule.id, type)}
                    className="p-1.5 hover:bg-red-500/10 rounded transition-colors"
                    title="Delete"
                  >
                    <Trash2 className="h-4 w-4 text-zinc-400 hover:text-red-400" />
                  </button>
                </div>
              </td>
            </tr>
          )) : (
            <tr>
              <td colSpan={5} className="px-6 py-12 text-center text-zinc-500">
                No pricing rules configured
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )

  if (loading) {
    return (
      <div className="p-8 space-y-6">
        <div className="animate-pulse">
          <div className="h-6 bg-zinc-800 rounded w-32 mb-4"></div>
          <div className="h-8 bg-zinc-800 rounded w-64 mb-8"></div>
        </div>
        <div className="h-64 bg-zinc-800 rounded-xl"></div>
        <div className="h-64 bg-zinc-800 rounded-xl"></div>
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/admin/pricing" className="p-2 hover:bg-white/10 rounded-xl transition-colors">
          <ArrowLeft className="h-5 w-5 text-zinc-400" />
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-white">
            {serviceProvider?.name || 'Service Provider'} Pricing
          </h1>
          <p className="text-zinc-400">Manage pricing rules for HTTP API and wallet-funded messages</p>
        </div>
      </div>

      <div className="relative flex-1 max-w-xs">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
        <input
          type="text"
          placeholder="Search pricing rules..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full pl-10 pr-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:ring-2 focus:ring-white/20"
        />
      </div>

      <DataCard
        title="HTTP API Pricing"
        subtitle="Pricing rules for HTTP API messages"
        action={{
          label: 'Add Pricing Rule',
          href: `/admin/pricing/${resolvedParams.id}/http-pricing/new`,
        }}
      >
        <PricingTable rules={httpPricing} type="http" />
      </DataCard>

      <DataCard
        title="Wallet Pricing"
        subtitle="Pricing rules for wallet-funded messages"
        action={{
          label: 'Add Pricing Rule',
          href: `/admin/pricing/${resolvedParams.id}/wallet-pricing/new`,
        }}
      >
        <PricingTable rules={walletPricing} type="wallet" />
      </DataCard>
    </div>
  )
}
