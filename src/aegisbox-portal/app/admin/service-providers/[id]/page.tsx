'use client'

import { use, useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowLeft, Plus, Copy, Check, X, RefreshCw, Key, Wallet, DollarSign, Settings, Loader2, MessageSquare, TrendingUp, Globe, Server, Eye, EyeOff, ToggleLeft, ToggleRight, BarChart3, Edit, Trash2, FileCheck, Users, Shuffle } from 'lucide-react'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'
import { AutoDlrSettings } from '@/components/service-providers/auto-dlr-settings'
import { PortalAccessSettings } from '@/components/service-providers/portal-access-settings'
import { RouteSenderPoolSetting } from '@/components/service-providers/route-sender-pool-setting'
import { formatMoney } from '@/lib/money'

function getStatusBadge(status: string) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-400',
    revoked: 'bg-red-500/20 text-red-400',
    connected: 'bg-emerald-500/20 text-emerald-400',
    disconnected: 'bg-red-500/20 text-red-400',
    pending: 'bg-amber-500/20 text-amber-400',
    inactive: 'bg-zinc-500/20 text-zinc-400',
  }

  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.inactive}`}>
      {status}
    </span>
  )
}

interface ServiceProvider {
  id: number
  name: string
  email: string
  status: string
  currency: string
  walletBalance: string | number
  createdAt: string
}

interface SmppCredential {
  id: number
  protocol: string
  status: string
  systemId: string | null
  bindType: string | null
  scope: string
  routingGroupId: number | null
  createdAt: string
}

interface ApiKey {
  id: number
  protocol: string
  status: string
  apiKeyIdentifier: string | null
  httpConfig: Record<string, any> | null
  scope: string
  routingGroupId: number | null
  createdAt: string
}

interface RoutingGroup {
  id: number
  name: string
  reference: string
  description: string | null
}

interface TabButtonProps {
  active: boolean
  onClick: () => void
  icon: any
  label: string
}

function TabButton({ active, onClick, icon: Icon, label }: TabButtonProps) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-4 py-2 rounded-lg transition-colors ${
        active
          ? 'bg-white text-black'
          : 'text-zinc-400 hover:text-white hover:bg-white/5'
      }`}
    >
      <Icon className="h-4 w-4" />
      {label}
    </button>
  )
}

export default function ServiceProviderDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const [activeTab, setActiveTab] = useState('overview')
  const [loading, setLoading] = useState(true)
  const [provider, setProvider] = useState<ServiceProvider | null>(null)
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([])
  const [smppCreds, setSmppCreds] = useState<SmppCredential[]>([])
  const [connection, setConnection] = useState<{ host: string | null; port: string | null } | null>(null)
  const [routingGroups, setRoutingGroups] = useState<RoutingGroup[]>([])
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [togglingId, setTogglingId] = useState<number | null>(null)
  const [editingRoutingGroup, setEditingRoutingGroup] = useState<{ id: number; type: 'smpp' | 'http' } | null>(null)
  const [selectedRoutingGroup, setSelectedRoutingGroup] = useState<string>('')
  const [savingRoutingGroup, setSavingRoutingGroup] = useState(false)
  const resolvedParams = use(params)

  useEffect(() => {
    fetchProviderData()
  }, [resolvedParams.id])

  const fetchProviderData = async () => {
    try {
      const [spResponse, rgResponse] = await Promise.all([
        fetch(`/api/admin/service-providers/${resolvedParams.id}`),
        fetch('/api/admin/routing-groups'),
      ])
      
      if (spResponse.ok) {
        const data = await spResponse.json()
        setProvider(data.provider)
        setApiKeys(data.credentials?.apiKeys || [])
        setSmppCreds(data.credentials?.smpps || [])
        setConnection(data.connection ?? null)
      }
      
      if (rgResponse.ok) {
        const data = await rgResponse.json()
        setRoutingGroups(data.routingGroups || [])
      }
    } catch (error) {
      console.error('Error fetching provider:', error)
    } finally {
      setLoading(false)
    }
  }

  const getRoutingGroupInfo = (routingGroupId: number | null) => {
    if (!routingGroupId) return null
    const rg = routingGroups.find(rg => rg.id === routingGroupId)
    return rg || { id: routingGroupId, name: `Group ${routingGroupId}`, reference: `RG-${routingGroupId}`, description: null }
  }

  const copyToClipboard = async (text: string, id: string) => {
    await navigator.clipboard.writeText(text)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const toggleCredentialStatus = async (credentialId: number, currentStatus: string, type: 'smpp' | 'http') => {
    setTogglingId(credentialId)
    try {
      const newStatus = currentStatus === 'active' ? 'inactive' : 'active'
      const endpoint = type === 'smpp' 
        ? `/api/admin/service-providers/${resolvedParams.id}/smpp`
        : `/api/admin/service-providers/${resolvedParams.id}/api-keys`
      
      const response = await fetch(`${endpoint}/${credentialId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      })

      if (response.ok) {
        if (type === 'smpp') {
          setSmppCreds(prev => prev.map(c => c.id === credentialId ? { ...c, status: newStatus } : c))
        } else {
          setApiKeys(prev => prev.map(k => k.id === credentialId ? { ...k, status: newStatus } : k))
        }
      }
    } catch (error) {
      console.error('Error toggling credential status:', error)
    } finally {
      setTogglingId(null)
    }
  }

  const startEditingRoutingGroup = (credentialId: number, currentRgId: number | null, type: 'smpp' | 'http') => {
    setEditingRoutingGroup({ id: credentialId, type })
    setSelectedRoutingGroup(currentRgId?.toString() || '')
  }

  const cancelEditingRoutingGroup = () => {
    setEditingRoutingGroup(null)
    setSelectedRoutingGroup('')
  }

  const saveRoutingGroup = async () => {
    if (!editingRoutingGroup) return
    setSavingRoutingGroup(true)
    try {
      const endpoint = editingRoutingGroup.type === 'smpp'
        ? `/api/admin/service-providers/${resolvedParams.id}/smpp`
        : `/api/admin/service-providers/${resolvedParams.id}/api-keys`
      
      const response = await fetch(`${endpoint}/${editingRoutingGroup.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ routingGroupId: selectedRoutingGroup }),
      })

      if (response.ok) {
        if (editingRoutingGroup.type === 'smpp') {
          setSmppCreds(prev => prev.map(c => 
            c.id === editingRoutingGroup.id 
              ? { ...c, routingGroupId: selectedRoutingGroup ? parseInt(selectedRoutingGroup) : null } 
              : c
          ))
        } else {
          setApiKeys(prev => prev.map(k => 
            k.id === editingRoutingGroup.id 
              ? { ...k, routingGroupId: selectedRoutingGroup ? parseInt(selectedRoutingGroup) : null } 
              : k
          ))
        }
        cancelEditingRoutingGroup()
      }
    } catch (error) {
      console.error('Error saving routing group:', error)
    } finally {
      setSavingRoutingGroup(false)
    }
  }

  if (loading) {
    return <PageLoadingSkeleton />
  }

  if (!provider) {
    return (
      <div className="p-8">
        <Link href="/admin/service-providers" className="flex items-center gap-2 text-zinc-400">
          <ArrowLeft className="h-4 w-4" />
          Back to Service Providers
        </Link>
        <div className="mt-8 text-center">
          <h1 className="text-2xl font-bold text-white">Service Provider Not Found</h1>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/admin/service-providers" className="p-2 hover:bg-white/10 rounded-xl transition-colors">
          <ArrowLeft className="h-5 w-5 text-zinc-400" />
        </Link>
        <div className="flex-1">
          <h1 className="text-3xl font-bold text-white">{provider.name}</h1>
          <p className="text-zinc-400">{provider.email}</p>
        </div>
        {getStatusBadge(provider.status)}
      </div>

      <div className="flex items-center gap-4 text-sm">
        <div className="flex items-center gap-2 text-zinc-400">
          <Wallet className="h-4 w-4" />
          <span>Wallet: {formatMoney(provider.walletBalance, provider.currency)}</span>
        </div>
        <div className="flex items-center gap-2 text-zinc-400">
          <MessageSquare className="h-4 w-4" />
          <span>Created: {new Date(provider.createdAt).toLocaleDateString()}</span>
        </div>
      </div>

      <div className="border-b border-white/10">
        <nav className="flex gap-2 flex-wrap">
          <TabButton active={activeTab === 'overview'} onClick={() => setActiveTab('overview')} icon={Server} label="Overview" />
          <TabButton active={activeTab === 'analytics'} onClick={() => setActiveTab('analytics')} icon={BarChart3} label="Analytics" />
          <TabButton active={activeTab === 'smpp'} onClick={() => setActiveTab('smpp')} icon={Settings} label="SMPP" />
          <TabButton active={activeTab === 'http'} onClick={() => setActiveTab('http')} icon={Globe} label="HTTP API" />
          <TabButton active={activeTab === 'wallet'} onClick={() => setActiveTab('wallet')} icon={Wallet} label="Wallet" />
          <TabButton active={activeTab === 'http-pricing'} onClick={() => setActiveTab('http-pricing')} icon={DollarSign} label="HTTP Pricing" />
          <TabButton active={activeTab === 'wallet-pricing'} onClick={() => setActiveTab('wallet-pricing')} icon={DollarSign} label="Wallet Pricing" />
          <TabButton active={activeTab === 'dlr'} onClick={() => setActiveTab('dlr')} icon={FileCheck} label="DLR" />
          <TabButton active={activeTab === 'portal-access'} onClick={() => setActiveTab('portal-access')} icon={Users} label="Portal Access" />
          <TabButton active={activeTab === 'sender-pool'} onClick={() => setActiveTab('sender-pool')} icon={Shuffle} label="Sender Pool" />
        </nav>
      </div>

      {activeTab === 'analytics' && (
        <AnalyticsTab providerId={provider.id} />
      )}

      {activeTab === 'overview' && (
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-white">SMPP Credentials</h2>
              <Link href={`/admin/service-providers/${provider.id}/smpp/new`} className="flex items-center gap-2 px-3 py-1.5 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors text-sm">
                <Plus className="h-4 w-4" />
                Add SMPP
              </Link>
            </div>
              {smppCreds.length > 0 ? (
              <div className="space-y-3">
                {smppCreds.map((cred) => {
                  const rg = getRoutingGroupInfo(cred.routingGroupId)
                  return (
                    <div key={cred.id} className="p-4 border border-white/10 rounded-xl bg-zinc-900/30">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <Server className="h-4 w-4 text-zinc-400" />
                          <span className="font-medium text-white">{cred.systemId}</span>
                          <span className="px-2 py-0.5 bg-white/10 rounded text-xs text-zinc-400 uppercase">
                            {cred.bindType}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => toggleCredentialStatus(cred.id, cred.status, 'smpp')}
                            disabled={togglingId === cred.id}
                            className="p-1 hover:bg-white/10 rounded transition-colors"
                          >
                            {togglingId === cred.id ? (
                              <Loader2 className="h-4 w-4 animate-spin text-zinc-400" />
                            ) : cred.status === 'active' ? (
                              <ToggleRight className="h-6 w-6 text-emerald-500" />
                            ) : (
                              <ToggleLeft className="h-6 w-6 text-zinc-500" />
                            )}
                          </button>
                          {getStatusBadge(cred.status)}
                        </div>
                      </div>
                      <div className="grid grid-cols-2 gap-4 mb-3 text-sm">
                        <div>
                          <span className="text-zinc-500">Host:</span>
                          <span className="ml-2 text-zinc-300">{connection?.host ?? 'Not set'}</span>
                        </div>
                        <div>
                          <span className="text-zinc-500">Port:</span>
                          <span className="ml-2 text-zinc-300">{connection?.port ?? 'Not set'}</span>
                        </div>
                        <div>
                          <span className="text-zinc-500">System ID:</span>
                          <span className="ml-2 text-zinc-300">{cred.systemId}</span>
                        </div>
                        <div>
                          <span className="text-zinc-500">Password:</span>
                          <span className="ml-2 text-zinc-300">••••••••</span>
                        </div>
                      </div>
                      {editingRoutingGroup?.id === cred.id && editingRoutingGroup.type === 'smpp' ? (
                        <div className="flex items-center gap-2 mb-2">
                          <select
                            value={selectedRoutingGroup}
                            onChange={(e) => setSelectedRoutingGroup(e.target.value)}
                            className="flex-1 px-2 py-1 text-sm border border-white/10 rounded bg-zinc-800 text-white"
                          >
                            <option value="">Default</option>
                            {routingGroups.map((rg) => (
                              <option key={rg.id} value={rg.id}>
                                {rg.name} ({rg.reference})
                              </option>
                            ))}
                          </select>
                          <button
                            onClick={saveRoutingGroup}
                            disabled={savingRoutingGroup}
                            className="p-1 text-emerald-400 hover:bg-emerald-500/10 rounded"
                          >
                            {savingRoutingGroup ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
                          </button>
                          <button
                            onClick={cancelEditingRoutingGroup}
                            className="p-1 text-red-400 hover:bg-red-500/10 rounded"
                          >
                            <X className="h-4 w-4" />
                          </button>
                        </div>
                      ) : cred.routingGroupId ? (
                        <div className="flex items-center gap-2 mb-2 text-sm">
                          <span className="text-zinc-500">Routing:</span>
                          <Link 
                            href={`/admin/routing-groups/${cred.routingGroupId}`}
                            className="text-white hover:underline flex items-center gap-1"
                          >
                            {rg?.name || `Group ${cred.routingGroupId}`}
                            <span className="text-xs text-zinc-500">({rg?.reference})</span>
                          </Link>
                          <button
                            onClick={() => startEditingRoutingGroup(cred.id, cred.routingGroupId, 'smpp')}
                            className="p-1 text-zinc-500 hover:text-white"
                          >
                            <Settings className="h-3 w-3" />
                          </button>
                        </div>
                      ) : (
                        <div className="flex items-center gap-2 mb-2 text-sm">
                          <span className="text-zinc-500">Routing:</span>
                          <span className="text-zinc-400">Default</span>
                          <button
                            onClick={() => startEditingRoutingGroup(cred.id, null, 'smpp')}
                            className="p-1 text-zinc-500 hover:text-white"
                          >
                            <Settings className="h-3 w-3" />
                          </button>
                        </div>
                      )}
                      <div className="flex items-center justify-between text-sm text-zinc-500">
                        <span>Created: {new Date(cred.createdAt).toLocaleDateString()}</span>
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="text-center py-8 text-zinc-500">
                <Server className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>No SMPP credentials configured</p>
                <Link href={`/admin/service-providers/${provider.id}/smpp/new`} className="text-white hover:underline text-sm">
                  Add your first SMPP credential
                </Link>
              </div>
            )}
          </div>

          <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-white">HTTP API Keys</h2>
              <Link href={`/admin/service-providers/${provider.id}/api-keys/new`} className="flex items-center gap-2 px-3 py-1.5 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors text-sm">
                <Plus className="h-4 w-4" />
                Create Key
              </Link>
            </div>
            {apiKeys.length > 0 ? (
              <div className="space-y-3">
                {apiKeys.map((cred) => {
                  const rg = getRoutingGroupInfo(cred.routingGroupId)
                  return (
                    <div key={cred.id} className="p-4 border border-white/10 rounded-xl bg-black/20">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-medium text-white">
                          {cred.httpConfig?.name || 'API Key'}
                        </span>
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => toggleCredentialStatus(cred.id, cred.status, 'http')}
                            disabled={togglingId === cred.id}
                            className="p-1 hover:bg-white/10 rounded transition-colors"
                          >
                            {togglingId === cred.id ? (
                              <Loader2 className="h-4 w-4 animate-spin text-zinc-400" />
                            ) : cred.status === 'active' ? (
                              <ToggleRight className="h-6 w-6 text-emerald-500" />
                            ) : (
                              <ToggleLeft className="h-6 w-6 text-zinc-500" />
                            )}
                          </button>
                          {getStatusBadge(cred.status)}
                        </div>
                      </div>
                      <div className="flex items-center gap-2 mb-3">
                        <code className="flex-1 px-2 py-1 bg-white/10 rounded text-xs font-mono text-zinc-400">
                          {cred.apiKeyIdentifier?.substring(0, 8).toUpperCase()}...
                        </code>
                        <button
                          onClick={() => copyToClipboard(cred.apiKeyIdentifier || '', `key-${cred.id}`)}
                          className="p-1.5 hover:bg-white/10 rounded transition-colors"
                        >
                          {copiedId === `key-${cred.id}` ? (
                            <Check className="h-4 w-4 text-emerald-500" />
                          ) : (
                            <Copy className="h-4 w-4 text-zinc-400" />
                          )}
                        </button>
                      </div>
                      {editingRoutingGroup?.id === cred.id && editingRoutingGroup.type === 'http' ? (
                        <div className="flex items-center gap-2 mb-2">
                          <select
                            value={selectedRoutingGroup}
                            onChange={(e) => setSelectedRoutingGroup(e.target.value)}
                            className="flex-1 px-2 py-1 text-sm border border-white/10 rounded bg-zinc-800 text-white"
                          >
                            <option value="">Default</option>
                            {routingGroups.map((rg) => (
                              <option key={rg.id} value={rg.id}>
                                {rg.name} ({rg.reference})
                              </option>
                            ))}
                          </select>
                          <button
                            onClick={saveRoutingGroup}
                            disabled={savingRoutingGroup}
                            className="p-1 text-emerald-400 hover:bg-emerald-500/10 rounded"
                          >
                            {savingRoutingGroup ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
                          </button>
                          <button
                            onClick={cancelEditingRoutingGroup}
                            className="p-1 text-red-400 hover:bg-red-500/10 rounded"
                          >
                            <X className="h-4 w-4" />
                          </button>
                        </div>
                      ) : cred.routingGroupId ? (
                        <div className="flex items-center gap-2 mb-2 text-sm">
                          <span className="text-zinc-500">Routing:</span>
                          <Link 
                            href={`/admin/routing-groups/${cred.routingGroupId}`}
                            className="text-white hover:underline flex items-center gap-1"
                          >
                            {rg?.name || `Group ${cred.routingGroupId}`}
                            <span className="text-xs text-zinc-500">({rg?.reference})</span>
                          </Link>
                          <button
                            onClick={() => startEditingRoutingGroup(cred.id, cred.routingGroupId, 'http')}
                            className="p-1 text-zinc-500 hover:text-white"
                          >
                            <Settings className="h-3 w-3" />
                          </button>
                        </div>
                      ) : (
                        <div className="flex items-center gap-2 mb-2 text-sm">
                          <span className="text-zinc-500">Routing:</span>
                          <span className="text-zinc-400">Default</span>
                          <button
                            onClick={() => startEditingRoutingGroup(cred.id, null, 'http')}
                            className="p-1 text-zinc-500 hover:text-white"
                          >
                            <Settings className="h-3 w-3" />
                          </button>
                        </div>
                      )}
                      <div className="flex items-center justify-between text-sm text-zinc-500">
                        <span>Created: {new Date(cred.createdAt).toLocaleDateString()}</span>
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="text-center py-8 text-zinc-500">
                <Globe className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>No API keys created</p>
                <Link href={`/admin/service-providers/${provider.id}/api-keys/new`} className="text-white hover:underline text-sm">
                  Create your first API key
                </Link>
              </div>
            )}
          </div>
        </div>
      )}

      {activeTab === 'smpp' && (
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white">SMPP Credentials</h2>
            <div className="flex items-center gap-3">
              <Link href={`/admin/smpp-sessions?sp=${provider.id}`} className="px-4 py-2 border border-white/10 rounded-lg text-zinc-300 hover:text-white hover:bg-white/5 transition-colors">
                Session history
              </Link>
              <Link href={`/admin/service-providers/${provider.id}/smpp/new`} className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors">
                <Plus className="h-4 w-4" />
                Add SMPP Credential
              </Link>
            </div>
          </div>
          {smppCreds.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-zinc-800/50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">System ID</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Bind Type</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Host</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Port</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Status</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Routing Group</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Created</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/10">
                  {smppCreds.map((cred) => {
                    const rg = getRoutingGroupInfo(cred.routingGroupId)
                    return (
                      <tr key={cred.id} className="hover:bg-white/5">
                        <td className="px-4 py-3 font-mono text-sm text-white">{cred.systemId}</td>
                        <td className="px-4 py-3 text-sm text-zinc-400 uppercase">{cred.bindType}</td>
                        <td className="px-4 py-3 text-sm text-zinc-400">{connection?.host ?? 'Not set'}</td>
                        <td className="px-4 py-3 text-sm text-zinc-400 font-mono">{connection?.port ?? 'Not set'}</td>
                        <td className="px-4 py-3">{getStatusBadge(cred.status)}</td>
                        <td className="px-4 py-3">
                          {cred.routingGroupId ? (
                            <Link 
                              href={`/admin/routing-groups/${cred.routingGroupId}`}
                              className="text-white hover:underline flex items-center gap-1"
                            >
                              {rg?.name || `Group ${cred.routingGroupId}`}
                              <span className="text-xs text-zinc-500">({rg?.reference})</span>
                            </Link>
                          ) : (
                            <span className="text-zinc-500">Default</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-sm text-zinc-500">
                          {new Date(cred.createdAt).toLocaleDateString()}
                        </td>
                        <td className="px-4 py-3">
                          <button
                            onClick={() => toggleCredentialStatus(cred.id, cred.status, 'smpp')}
                            disabled={togglingId === cred.id}
                            className="p-1 hover:bg-white/10 rounded transition-colors"
                          >
                            {togglingId === cred.id ? (
                              <Loader2 className="h-4 w-4 animate-spin text-zinc-400" />
                            ) : cred.status === 'active' ? (
                              <ToggleRight className="h-6 w-6 text-emerald-500" />
                            ) : (
                              <ToggleLeft className="h-6 w-6 text-zinc-500" />
                            )}
                          </button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-12 text-zinc-500">
              <Server className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No SMPP credentials configured</p>
              <Link href={`/admin/service-providers/${provider.id}/smpp/new`} className="text-white hover:underline">
                Add your first SMPP credential
              </Link>
            </div>
          )}
        </div>
      )}

      {activeTab === 'http' && (
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white">HTTP API Keys</h2>
            <Link href={`/admin/service-providers/${provider.id}/api-keys/new`} className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors">
              <Plus className="h-4 w-4" />
              Create API Key
            </Link>
          </div>
          {apiKeys.length > 0 ? (
            <div className="space-y-4">
              {apiKeys.map((cred) => {
                const rg = getRoutingGroupInfo(cred.routingGroupId)
                return (
                  <div key={cred.id} className="p-4 border border-white/10 rounded-xl bg-black/20">
                    <div className="flex items-center justify-between mb-4">
                      <div>
                        <h3 className="font-medium text-white">
                          {cred.httpConfig?.name || 'API Key'}
                        </h3>
                        <p className="text-xs text-zinc-500">ID: {cred.id}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => toggleCredentialStatus(cred.id, cred.status, 'http')}
                          disabled={togglingId === cred.id}
                          className="p-1 hover:bg-white/10 rounded transition-colors"
                        >
                          {togglingId === cred.id ? (
                            <Loader2 className="h-4 w-4 animate-spin text-zinc-400" />
                          ) : cred.status === 'active' ? (
                            <ToggleRight className="h-6 w-6 text-emerald-500" />
                          ) : (
                            <ToggleLeft className="h-6 w-6 text-zinc-500" />
                          )}
                        </button>
                        {getStatusBadge(cred.status)}
                      </div>
                    </div>
                    <div className="flex items-center gap-2 mb-4">
                      <code className="flex-1 px-3 py-2 bg-white/10 rounded font-mono text-sm text-zinc-400 break-all">
                        {cred.apiKeyIdentifier}
                      </code>
                      <button
                        onClick={() => copyToClipboard(cred.apiKeyIdentifier || '', `full-key-${cred.id}`)}
                        className="p-2 hover:bg-white/10 rounded transition-colors"
                      >
                        {copiedId === `full-key-${cred.id}` ? (
                          <Check className="h-4 w-4 text-emerald-500" />
                        ) : (
                          <Copy className="h-4 w-4 text-zinc-400" />
                        )}
                      </button>
                    </div>
                    <div className="grid grid-cols-4 gap-4 text-sm">
                      <div>
                        <span className="text-zinc-500">Prefix:</span>
                        <span className="ml-2 font-mono text-white">
                          {cred.apiKeyIdentifier?.substring(0, 8).toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <span className="text-zinc-500">Routing:</span>
                        {cred.routingGroupId ? (
                          <Link 
                            href={`/admin/routing-groups/${cred.routingGroupId}`}
                            className="ml-2 text-white hover:underline"
                          >
                            {rg?.name || `Group ${cred.routingGroupId}`}
                          </Link>
                        ) : (
                          <span className="ml-2 text-zinc-400">Default</span>
                        )}
                      </div>
                      <div>
                        <span className="text-zinc-500">Scope:</span>
                        <span className="ml-2 text-white">{cred.scope}</span>
                      </div>
                      <div>
                        <span className="text-zinc-500">Created:</span>
                        <span className="ml-2 text-zinc-400">
                          {new Date(cred.createdAt).toLocaleDateString()}
                        </span>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="text-center py-12 text-zinc-500">
              <Globe className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No API keys created</p>
              <Link href={`/admin/service-providers/${provider.id}/api-keys/new`} className="text-white hover:underline">
                Create your first API key
              </Link>
            </div>
          )}
        </div>
      )}

      {activeTab === 'wallet' && (
        <WalletTransactionsTab
          providerId={provider.id}
          currency={provider.currency}
          walletBalance={provider.walletBalance}
        />
      )}

      {activeTab === 'http-pricing' && (
        <HttpPricingTab providerId={provider.id} />
      )}

      {activeTab === 'wallet-pricing' && (
        <WalletPricingTab providerId={provider.id} />
      )}

      {activeTab === 'dlr' && (
        <AutoDlrSettings providerId={provider.id} />
      )}

      {activeTab === 'portal-access' && (
        <PortalAccessSettings providerId={provider.id} />
      )}

      {activeTab === 'sender-pool' && (
        <RouteSenderPoolSetting providerId={provider.id} />
      )}
    </div>
  )
}

interface Transaction {
  id: number
  transactionType: 'credit' | 'debit' | 'reversal'
  amount: number
  balanceBefore: number
  balanceAfter: number
  description: string | null
  transactionDate: string
}

interface WalletTabProps {
  providerId: number
  currency: string
  walletBalance: string | number
}

function WalletTransactionsTab({ providerId, currency, walletBalance }: WalletTabProps) {
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [summary, setSummary] = useState({ totalCredits: 0, totalDebits: 0, netChange: 0 })
  const [filters, setFilters] = useState({
    type: '',
    startDate: '',
    endDate: '',
    sort: 'desc',
  })
  const [pagination, setPagination] = useState({ page: 1, limit: 20, total: 0, totalPages: 0 })

  useEffect(() => {
    fetchTransactions()
  }, [providerId, filters, pagination.page])

  const fetchTransactions = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams({
        page: pagination.page.toString(),
        limit: pagination.limit.toString(),
        sort: filters.sort,
      })
      
      if (filters.type) params.append('type', filters.type)
      if (filters.startDate) params.append('startDate', filters.startDate)
      if (filters.endDate) params.append('endDate', filters.endDate)

      const response = await fetch(`/api/admin/service-providers/${providerId}/wallet/transactions?${params}`)
      if (response.ok) {
        const data = await response.json()
        setTransactions(data.transactions || [])
        setSummary(data.summary || { totalCredits: 0, totalDebits: 0, netChange: 0 })
        setPagination(prev => ({ ...prev, ...data.pagination }))
      }
    } catch (error) {
      console.error('Error fetching transactions:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
    setPagination(prev => ({ ...prev, page: 1 }))
  }

  const handleClearFilters = () => {
    setFilters({ type: '', startDate: '', endDate: '', sort: 'desc' })
    setPagination(prev => ({ ...prev, page: 1 }))
  }

  const formatCurrency = (amount: number) => formatMoney(amount, currency)

  const getTypeBadge = (type: string) => {
    const styles: Record<string, { bg: string; text: string }> = {
      credit: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-700 dark:text-green-400' },
      debit: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-700 dark:text-red-400' },
      reversal: { bg: 'bg-yellow-100 dark:bg-yellow-900/30', text: 'text-yellow-700 dark:text-yellow-400' },
    }
    const style = styles[type] || styles.debit
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium capitalize ${style.bg} ${style.text}`}>
        {type}
      </span>
    )
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <p className="text-sm text-zinc-500">Current Balance</p>
          <p className="text-2xl font-bold text-white">{formatCurrency(Number(walletBalance))}</p>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <p className="text-sm text-zinc-500">Total Credits</p>
          <p className="text-2xl font-bold text-emerald-400">{formatCurrency(summary.totalCredits)}</p>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <p className="text-sm text-zinc-500">Charges (less refunds)</p>
          <p className="text-2xl font-bold text-red-400">{formatCurrency(summary.totalDebits)}</p>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-4">
          <p className="text-sm text-zinc-500">Net Change</p>
          <p className={`text-2xl font-bold ${summary.netChange >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
            {summary.netChange >= 0 ? '+' : ''}{formatCurrency(summary.netChange)}
          </p>
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
        <div className="p-4 border-b border-white/10">
          <div className="flex flex-wrap items-center gap-4">
            <div>
              <label className="block text-xs text-zinc-500 mb-1">Type</label>
              <select
                value={filters.type}
                onChange={(e) => handleFilterChange('type', e.target.value)}
                className="px-3 py-1.5 text-sm border border-white/10 rounded-lg bg-zinc-800 text-white"
              >
                <option value="">All Types</option>
                <option value="credit">Credit</option>
                <option value="debit">Debit</option>
                <option value="reversal">Reversal</option>
              </select>
            </div>
            <div>
              <label className="block text-xs text-zinc-500 mb-1">Start Date</label>
              <input
                type="date"
                value={filters.startDate}
                onChange={(e) => handleFilterChange('startDate', e.target.value)}
                className="px-3 py-1.5 text-sm border border-white/10 rounded-lg bg-zinc-800 text-white"
              />
            </div>
            <div>
              <label className="block text-xs text-zinc-500 mb-1">End Date</label>
              <input
                type="date"
                value={filters.endDate}
                onChange={(e) => handleFilterChange('endDate', e.target.value)}
                className="px-3 py-1.5 text-sm border border-white/10 rounded-lg bg-zinc-800 text-white"
              />
            </div>
            <div>
              <label className="block text-xs text-zinc-500 mb-1">Sort</label>
              <select
                value={filters.sort}
                onChange={(e) => handleFilterChange('sort', e.target.value)}
                className="px-3 py-1.5 text-sm border border-white/10 rounded-lg bg-zinc-800 text-white"
              >
                <option value="desc">Newest First</option>
                <option value="asc">Oldest First</option>
              </select>
            </div>
            {(filters.type || filters.startDate || filters.endDate) && (
              <button
                onClick={handleClearFilters}
                className="px-3 py-1.5 text-sm text-zinc-400 hover:text-white hover:bg-white/10 rounded-lg"
              >
                Clear Filters
              </button>
            )}
          </div>
        </div>

        {loading ? (
          <div className="p-12 flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-white" />
          </div>
        ) : transactions.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-zinc-800/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Date</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Type</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Amount</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Balance After</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Description</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {transactions.map((tx) => (
                  <tr key={tx.id} className="hover:bg-white/5">
                    <td className="px-4 py-3 text-sm text-white">
                      {new Date(tx.transactionDate).toLocaleDateString()}
                      <span className="text-xs text-zinc-500 ml-1">
                        {new Date(tx.transactionDate).toLocaleTimeString()}
                      </span>
                    </td>
                    <td className="px-4 py-3">{getTypeBadge(tx.transactionType)}</td>
                    <td className={`px-4 py-3 font-medium ${
                      tx.transactionType === 'credit' || tx.transactionType === 'reversal'
                        ? 'text-emerald-400'
                        : 'text-red-400'
                    }`}>
                      {tx.transactionType === 'credit' || tx.transactionType === 'reversal' ? '+' : '-'}
                      {formatCurrency(tx.amount)}
                    </td>
                    <td className="px-4 py-3 text-sm text-white">
                      {formatCurrency(tx.balanceAfter)}
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-500">
                      {tx.description || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-12 text-center">
            <Wallet className="h-12 w-12 mx-auto mb-4 text-zinc-600" />
            <p className="text-zinc-500">No transactions found</p>
          </div>
        )}

        {pagination.totalPages > 1 && (
          <div className="p-4 border-t border-white/10 flex items-center justify-between">
            <p className="text-sm text-zinc-500">
              Showing {((pagination.page - 1) * pagination.limit) + 1} to {Math.min(pagination.page * pagination.limit, pagination.total)} of {pagination.total}
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => setPagination(prev => ({ ...prev, page: prev.page - 1 }))}
                disabled={pagination.page === 1}
                className="px-3 py-1 text-sm border border-white/10 rounded-lg disabled:opacity-50 text-white"
              >
                Previous
              </button>
              <button
                onClick={() => setPagination(prev => ({ ...prev, page: prev.page + 1 }))}
                disabled={pagination.page >= pagination.totalPages}
                className="px-3 py-1 text-sm border border-white/10 rounded-lg disabled:opacity-50 text-white"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
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

interface PricingTabProps {
  providerId: number
}

function PricingRulesTab({ providerId }: PricingTabProps) {
  const [rules, setRules] = useState<PricingRule[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [formData, setFormData] = useState({
    mnoId: '',
    currencyCode: 'NGN',
    pricePerSms: '',
  })
  const [saving, setSaving] = useState(false)
  const [mnos, setMnos] = useState<{ id: number; name: string; countryCode: string }[]>([])

  useEffect(() => {
    fetchRules()
    fetchMnos()
  }, [providerId])

  const fetchRules = async () => {
    try {
      const response = await fetch(`/api/admin/service-providers/${providerId}/pricing`)
      if (response.ok) {
        const data = await response.json()
        setRules(data.pricingRules || [])
      }
    } catch (error) {
      console.error('Error fetching pricing rules:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchMnos = async () => {
    try {
      const response = await fetch('/api/admin/mnos')
      if (response.ok) {
        const data = await response.json()
        setMnos(data.mnos || [])
      }
    } catch (error) {
      console.error('Error fetching MNOs:', error)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const response = await fetch(`/api/admin/service-providers/${providerId}/pricing`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          mnoId: formData.mnoId || null,
        }),
      })
      if (response.ok) {
        setShowAddForm(false)
        setFormData({ mnoId: '', currencyCode: 'NGN', pricePerSms: '' })
        fetchRules()
      }
    } catch (error) {
      console.error('Error creating pricing rule:', error)
    } finally {
      setSaving(false)
    }
  }

  const getCurrencySymbol = (currencyCode: string) => {
    switch (currencyCode) {
      case 'NGN': return '₦'
      case 'USD': return '$'
      case 'EUR': return '€'
      case 'GBP': return '£'
      case 'EGP': return 'E£'
      case 'KES': return 'KSh'
      case 'GHS': return '₵'
      case 'ZAR': return 'R'
      default: return currencyCode
    }
  }

  const formatCurrency = (amount: string, currencyCode: string) => {
    const symbol = getCurrencySymbol(currencyCode)
    return `${symbol}${Number(amount).toLocaleString(undefined, { minimumFractionDigits: 4, maximumFractionDigits: 4 })}`
  }

  return (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
      <div className="p-6 border-b border-white/10 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white">Pricing Rules</h2>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
        >
          <Plus className="h-4 w-4" />
          {showAddForm ? 'Cancel' : 'Add Pricing Rule'}
        </button>
      </div>

      {showAddForm && (
        <div className="p-6 border-b border-white/10 bg-black/20">
          <form onSubmit={handleSubmit} className="grid gap-4 md:grid-cols-4">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">MNO (Optional)</label>
              <select
                value={formData.mnoId}
                onChange={(e) => setFormData(prev => ({ ...prev, mnoId: e.target.value }))}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
              >
                <option value="">All Networks (Default)</option>
                {mnos.map((mno) => (
                  <option key={mno.id} value={mno.id}>{mno.name} ({mno.countryCode})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Currency</label>
              <select
                value={formData.currencyCode}
                onChange={(e) => setFormData(prev => ({ ...prev, currencyCode: e.target.value }))}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
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
              <label className="block text-sm font-medium text-zinc-400 mb-1">Price per SMS</label>
              <input
                type="number"
                step="0.0001"
                min="0"
                value={formData.pricePerSms}
                onChange={(e) => setFormData(prev => ({ ...prev, pricePerSms: e.target.value }))}
                placeholder="0.0000"
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500"
              />
            </div>
            <div className="flex items-end">
              <button
                type="submit"
                disabled={saving}
                className="flex items-center gap-2 px-4 py-2 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 transition-colors disabled:opacity-50"
              >
                <Loader2 className={`h-4 w-4 ${saving ? 'animate-spin' : ''}`} />
                {saving ? 'Saving...' : 'Save Rule'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <div className="p-12 flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-white" />
        </div>
      ) : rules.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-800/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">MNO</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Currency</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Price/SMS</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {rules.map((rule) => (
                <tr key={rule.id} className="hover:bg-white/5">
                  <td className="px-4 py-3">
                    {rule.mnoName ? (
                      <div>
                        <span className="font-medium text-white">{rule.mnoName}</span>
                        <span className="text-xs text-zinc-500 ml-1">({rule.mnoCountryCode})</span>
                      </div>
                    ) : (
                      <span className="text-zinc-500">Default (All Networks)</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-white">{rule.currencyCode}</td>
                  <td className="px-4 py-3 font-medium text-white">{formatCurrency(rule.pricePerSms, rule.currencyCode)}</td>
                  <td className="px-4 py-3 text-sm text-zinc-500">{new Date(rule.createdAt).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="p-12 text-center">
          <DollarSign className="h-12 w-12 mx-auto mb-4 text-zinc-600" />
          <p className="text-zinc-500">No pricing rules configured</p>
          <p className="text-sm text-zinc-600 mt-1">Add a pricing rule to start billing</p>
        </div>
      )}
    </div>
  )
}

function HttpPricingTab({ providerId }: { providerId: number }) {
  const [rules, setRules] = useState<PricingRule[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [formData, setFormData] = useState({
    mnoId: '',
    currencyCode: 'NGN',
    pricePerSms: '',
  })
  const [saving, setSaving] = useState(false)
  const [mnos, setMnos] = useState<{ id: number; name: string; countryCode: string }[]>([])

  useEffect(() => {
    fetchRules()
    fetchMnos()
  }, [providerId])

  const fetchRules = async () => {
    try {
      const response = await fetch(`/api/admin/service-providers/${providerId}/pricing`)
      if (response.ok) {
        const data = await response.json()
        setRules(data.pricingRules || [])
      }
    } catch (error) {
      console.error('Error fetching HTTP pricing rules:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchMnos = async () => {
    try {
      const response = await fetch('/api/admin/mnos')
      if (response.ok) {
        const data = await response.json()
        setMnos(data.mnos || [])
      }
    } catch (error) {
      console.error('Error fetching MNOs:', error)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const response = await fetch(`/api/admin/service-providers/${providerId}/pricing`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          mnoId: formData.mnoId || null,
        }),
      })
      if (response.ok) {
        setShowAddForm(false)
        setFormData({ mnoId: '', currencyCode: 'NGN', pricePerSms: '' })
        fetchRules()
      }
    } catch (error) {
      console.error('Error creating HTTP pricing rule:', error)
    } finally {
      setSaving(false)
    }
  }

  const getCurrencySymbol = (currencyCode: string) => {
    switch (currencyCode) {
      case 'NGN': return '₦'
      case 'USD': return '$'
      case 'EUR': return '€'
      case 'GBP': return '£'
      case 'EGP': return 'E£'
      case 'KES': return 'KSh'
      case 'GHS': return '₵'
      case 'ZAR': return 'R'
      default: return currencyCode
    }
  }

  const formatCurrency = (amount: string, currencyCode: string) => {
    const symbol = getCurrencySymbol(currencyCode)
    return `${symbol}${Number(amount).toLocaleString(undefined, { minimumFractionDigits: 4, maximumFractionDigits: 4 })}`
  }

  const handleDelete = async (ruleId: number) => {
    if (!confirm('Are you sure you want to delete this HTTP pricing rule?')) return
    try {
      const response = await fetch(`/api/admin/pricing/${ruleId}`, { method: 'DELETE' })
      if (response.ok) {
        setRules(prev => prev.filter(r => r.id !== ruleId))
      }
    } catch (error) {
      console.error('Error deleting HTTP pricing rule:', error)
    }
  }

  return (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
      <div className="p-6 border-b border-white/10 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-white">HTTP API Pricing</h2>
          <p className="text-sm text-zinc-500 mt-1">Pricing rules for HTTP API messages</p>
        </div>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
        >
          <Plus className="h-4 w-4" />
          {showAddForm ? 'Cancel' : 'Add Pricing Rule'}
        </button>
      </div>

      {showAddForm && (
        <div className="p-6 border-b border-white/10 bg-black/20">
          <form onSubmit={handleSubmit} className="grid gap-4 md:grid-cols-4">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">MNO (Optional)</label>
              <select
                value={formData.mnoId}
                onChange={(e) => setFormData(prev => ({ ...prev, mnoId: e.target.value }))}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
              >
                <option value="">All Networks (Default)</option>
                {mnos.map((mno) => (
                  <option key={mno.id} value={mno.id}>{mno.name} ({mno.countryCode})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Currency</label>
              <select
                value={formData.currencyCode}
                onChange={(e) => setFormData(prev => ({ ...prev, currencyCode: e.target.value }))}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
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
              <label className="block text-sm font-medium text-zinc-400 mb-1">Price per SMS</label>
              <input
                type="number"
                step="0.0001"
                min="0"
                value={formData.pricePerSms}
                onChange={(e) => setFormData(prev => ({ ...prev, pricePerSms: e.target.value }))}
                placeholder="0.0000"
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500"
              />
            </div>
            <div className="flex items-end">
              <button
                type="submit"
                disabled={saving}
                className="flex items-center gap-2 px-4 py-2 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 transition-colors disabled:opacity-50"
              >
                <Loader2 className={`h-4 w-4 ${saving ? 'animate-spin' : ''}`} />
                {saving ? 'Saving...' : 'Save Rule'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <div className="p-12 flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-white" />
        </div>
      ) : rules.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-800/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">MNO</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Currency</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Price/SMS</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Created</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-zinc-400 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {rules.map((rule) => (
                <tr key={rule.id} className="hover:bg-white/5">
                  <td className="px-4 py-3">
                    {rule.mnoName ? (
                      <div>
                        <span className="font-medium text-white">{rule.mnoName}</span>
                        <span className="text-xs text-zinc-500 ml-1">({rule.mnoCountryCode})</span>
                      </div>
                    ) : (
                      <span className="text-zinc-500">Default (All Networks)</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-white">{rule.currencyCode}</td>
                  <td className="px-4 py-3 font-medium text-white">{formatCurrency(rule.pricePerSms, rule.currencyCode)}</td>
                  <td className="px-4 py-3 text-sm text-zinc-500">{new Date(rule.createdAt).toLocaleDateString()}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Link
                        href={`/admin/pricing/rule/${rule.id}`}
                        className="p-1.5 hover:bg-white/10 rounded transition-colors"
                        title="Edit"
                      >
                        <Edit className="h-4 w-4 text-zinc-400 hover:text-white" />
                      </Link>
                      <button
                        onClick={() => handleDelete(rule.id)}
                        className="p-1.5 hover:bg-red-500/10 rounded transition-colors"
                        title="Delete"
                      >
                        <Trash2 className="h-4 w-4 text-zinc-400 hover:text-red-400" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="p-12 text-center">
          <DollarSign className="h-12 w-12 mx-auto mb-4 text-zinc-600" />
          <p className="text-zinc-500">No HTTP pricing rules configured</p>
          <p className="text-sm text-zinc-600 mt-1">Add a pricing rule to start billing</p>
        </div>
      )}
    </div>
  )
}

function WalletPricingTab({ providerId }: { providerId: number }) {
  const [rules, setRules] = useState<PricingRule[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [formData, setFormData] = useState({
    mnoId: '',
    currencyCode: 'NGN',
    pricePerSms: '',
  })
  const [saving, setSaving] = useState(false)
  const [mnos, setMnos] = useState<{ id: number; name: string; countryCode: string }[]>([])

  useEffect(() => {
    fetchRules()
    fetchMnos()
  }, [providerId])

  const fetchRules = async () => {
    try {
      const response = await fetch(`/api/admin/service-providers/${providerId}/wallet-pricing`)
      if (response.ok) {
        const data = await response.json()
        setRules(data.pricingRules || [])
      }
    } catch (error) {
      console.error('Error fetching wallet pricing rules:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchMnos = async () => {
    try {
      const response = await fetch('/api/admin/mnos')
      if (response.ok) {
        const data = await response.json()
        setMnos(data.mnos || [])
      }
    } catch (error) {
      console.error('Error fetching MNOs:', error)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const response = await fetch(`/api/admin/service-providers/${providerId}/wallet-pricing`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          mnoId: formData.mnoId || null,
        }),
      })
      if (response.ok) {
        setShowAddForm(false)
        setFormData({ mnoId: '', currencyCode: 'NGN', pricePerSms: '' })
        fetchRules()
      }
    } catch (error) {
      console.error('Error creating wallet pricing rule:', error)
    } finally {
      setSaving(false)
    }
  }

  const getCurrencySymbol = (currencyCode: string) => {
    switch (currencyCode) {
      case 'NGN': return '₦'
      case 'USD': return '$'
      case 'EUR': return '€'
      case 'GBP': return '£'
      case 'EGP': return 'E£'
      case 'KES': return 'KSh'
      case 'GHS': return '₵'
      case 'ZAR': return 'R'
      default: return currencyCode
    }
  }

  const formatCurrency = (amount: string, currencyCode: string) => {
    const symbol = getCurrencySymbol(currencyCode)
    return `${symbol}${Number(amount).toLocaleString(undefined, { minimumFractionDigits: 4, maximumFractionDigits: 4 })}`
  }

  const handleDelete = async (ruleId: number) => {
    if (!confirm('Are you sure you want to delete this wallet pricing rule?')) return
    try {
      const response = await fetch(`/api/admin/pricing/${ruleId}`, { method: 'DELETE' })
      if (response.ok) {
        setRules(prev => prev.filter(r => r.id !== ruleId))
      }
    } catch (error) {
      console.error('Error deleting wallet pricing rule:', error)
    }
  }

  return (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
      <div className="p-6 border-b border-white/10 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-white">Wallet Pricing</h2>
          <p className="text-sm text-zinc-500 mt-1">Pricing rules for wallet-funded messages</p>
        </div>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
        >
          <Plus className="h-4 w-4" />
          {showAddForm ? 'Cancel' : 'Add Pricing Rule'}
        </button>
      </div>

      {showAddForm && (
        <div className="p-6 border-b border-white/10 bg-black/20">
          <form onSubmit={handleSubmit} className="grid gap-4 md:grid-cols-4">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">MNO (Optional)</label>
              <select
                value={formData.mnoId}
                onChange={(e) => setFormData(prev => ({ ...prev, mnoId: e.target.value }))}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
              >
                <option value="">All Networks (Default)</option>
                {mnos.map((mno) => (
                  <option key={mno.id} value={mno.id}>{mno.name} ({mno.countryCode})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Currency</label>
              <select
                value={formData.currencyCode}
                onChange={(e) => setFormData(prev => ({ ...prev, currencyCode: e.target.value }))}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white"
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
              <label className="block text-sm font-medium text-zinc-400 mb-1">Price per SMS</label>
              <input
                type="number"
                step="0.0001"
                min="0"
                value={formData.pricePerSms}
                onChange={(e) => setFormData(prev => ({ ...prev, pricePerSms: e.target.value }))}
                placeholder="0.0000"
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500"
              />
            </div>
            <div className="flex items-end">
              <button
                type="submit"
                disabled={saving}
                className="flex items-center gap-2 px-4 py-2 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 transition-colors disabled:opacity-50"
              >
                <Loader2 className={`h-4 w-4 ${saving ? 'animate-spin' : ''}`} />
                {saving ? 'Saving...' : 'Save Rule'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <div className="p-12 flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-white" />
        </div>
      ) : rules.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-800/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">MNO</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Currency</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Price/SMS</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase">Created</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-zinc-400 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/10">
              {rules.map((rule) => (
                <tr key={rule.id} className="hover:bg-white/5">
                  <td className="px-4 py-3">
                    {rule.mnoName ? (
                      <div>
                        <span className="font-medium text-white">{rule.mnoName}</span>
                        <span className="text-xs text-zinc-500 ml-1">({rule.mnoCountryCode})</span>
                      </div>
                    ) : (
                      <span className="text-zinc-500">Default (All Networks)</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-white">{rule.currencyCode}</td>
                  <td className="px-4 py-3 font-medium text-white">{formatCurrency(rule.pricePerSms, rule.currencyCode)}</td>
                  <td className="px-4 py-3 text-sm text-zinc-500">{new Date(rule.createdAt).toLocaleDateString()}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Link
                        href={`/admin/pricing/rule/${rule.id}`}
                        className="p-1.5 hover:bg-white/10 rounded transition-colors"
                        title="Edit"
                      >
                        <Edit className="h-4 w-4 text-zinc-400 hover:text-white" />
                      </Link>
                      <button
                        onClick={() => handleDelete(rule.id)}
                        className="p-1.5 hover:bg-red-500/10 rounded transition-colors"
                        title="Delete"
                      >
                        <Trash2 className="h-4 w-4 text-zinc-400 hover:text-red-400" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="p-12 text-center">
          <Wallet className="h-12 w-12 mx-auto mb-4 text-zinc-600" />
          <p className="text-zinc-500">No wallet pricing rules configured</p>
          <p className="text-sm text-zinc-600 mt-1">Add a pricing rule to start billing</p>
        </div>
      )}
    </div>
  )
}

function AnalyticsTab({ providerId }: { providerId: number }) {
  const [loading, setLoading] = useState(true)
  const [analytics, setAnalytics] = useState<any>(null)

  useEffect(() => {
    fetchAnalytics()
  }, [providerId])

  const fetchAnalytics = async () => {
    setLoading(true)
    try {
      const res = await fetch(`/api/admin/service-providers/${providerId}/analytics?days=30`)
      if (res.ok) {
        const data = await res.json()
        setAnalytics(data)
      }
    } catch (error) {
      console.error('Error fetching analytics:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-white" />
      </div>
    )
  }

  if (!analytics) {
    return (
      <div className="text-center py-12 text-zinc-500">
        <BarChart3 className="h-12 w-12 mx-auto mb-4" />
        <p>Failed to load analytics</p>
      </div>
    )
  }

  const overview = analytics.overview || {}
  const wallet = analytics.wallet || {}

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-5">
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-4">
          <p className="text-sm text-zinc-400">Total Messages</p>
          <p className="text-2xl font-bold text-white">{overview.totalMessages?.toLocaleString() || 0}</p>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-4">
          <p className="text-sm text-zinc-400">Delivery Rate</p>
          <p className="text-2xl font-bold text-emerald-400">{overview.deliveryRate || 0}%</p>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-4">
          <p className="text-sm text-zinc-400">Failure Rate</p>
          <p className="text-2xl font-bold text-red-400">{overview.failureRate || 0}%</p>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-4">
          <p className="text-sm text-zinc-400">Total Cost</p>
          <p className="text-2xl font-bold text-white">{overview.totalCost || 0}</p>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-4">
          <p className="text-sm text-zinc-400">Wallet Balance</p>
          <p className="text-2xl font-bold text-white">{wallet.balance?.toLocaleString() || 0}</p>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Status Breakdown</h3>
          <div className="grid gap-2">
            {analytics.statusBreakdown?.map((item: any) => (
              <div key={item.status} className="flex items-center justify-between p-3 bg-black/20 rounded-lg">
                <span className="capitalize text-zinc-300">{item.status}</span>
                <div className="flex items-center gap-3">
                  <span className="text-white">{item.count?.toLocaleString() || 0}</span>
                  <span className="text-xs text-zinc-500">({item.percentage}%)</span>
                </div>
              </div>
            ))}
            {(!analytics.statusBreakdown || analytics.statusBreakdown.length === 0) && (
              <p className="text-center text-zinc-500 py-4">No status data available</p>
            )}
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Top Destinations</h3>
          <div className="flex flex-col gap-2">
            {analytics.topDestinations?.slice(0, 8).map((item: any) => (
              <div key={item.prefix} className="flex items-center justify-between p-3 hover:bg-white/5 rounded-lg">
                <div className="flex items-center gap-3">
                  <Globe className="h-4 w-4 text-zinc-500 flex-shrink-0" />
                  <div>
                    <p className="font-mono text-white text-sm">{item.prefix}***</p>
                    {(item.mnoName || item.networkName) && (
                      <p className="text-xs text-zinc-500">
                        {item.mnoName && item.mnoName !== 'Unknown' && <span>{item.mnoName}</span>}
                        {item.mnoName && item.mnoName !== 'Unknown' && item.networkName && item.networkName !== 'Unknown' && <span className="mx-1">/</span>}
                        {item.networkName && item.networkName !== 'Unknown' && <span>{item.networkName}</span>}
                        {item.mnoCountryCode && <span className="ml-1 text-zinc-600">({item.mnoCountryCode})</span>}
                      </p>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-4 flex-shrink-0">
                  <span className="text-zinc-400 text-sm">{item.count?.toLocaleString() || 0}</span>
                  <span className={`text-sm font-medium min-w-[60px] text-right ${
                    parseFloat(item.deliveryRate) > 80 ? 'text-emerald-400' :
                    parseFloat(item.deliveryRate) > 50 ? 'text-amber-400' : 'text-red-400'
                  }`}>
                    {item.deliveryRate}%
                  </span>
                </div>
              </div>
            ))}
            {(!analytics.topDestinations || analytics.topDestinations.length === 0) && (
              <p className="text-center text-zinc-500 py-4">No destination data available</p>
            )}
          </div>
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Recent Daily Activity</h3>
        {analytics.dailyStats && analytics.dailyStats.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-xs uppercase text-zinc-400 border-b border-white/10">
                  <th className="pb-3">Date</th>
                  <th className="pb-3 text-right">Total</th>
                  <th className="pb-3 text-right">Delivered</th>
                  <th className="pb-3 text-right">Failed</th>
                  <th className="pb-3 text-right">Cost</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {analytics.dailyStats.slice(-14).map((day: any) => (
                  <tr key={day.date}>
                    <td className="py-3 text-white">{new Date(day.date).toLocaleDateString()}</td>
                    <td className="py-3 text-right text-zinc-400">{day.total}</td>
                    <td className="py-3 text-right text-emerald-400">{day.delivered}</td>
                    <td className="py-3 text-right text-red-400">{day.failed}</td>
                    <td className="py-3 text-right text-white">{day.cost}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-center text-zinc-500 py-4">No daily data available</p>
        )}
      </div>
    </div>
  )
}
