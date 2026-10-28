'use client'

import { useState, useEffect, useMemo } from 'react'
import Link from 'next/link'
import { Plus, Search, Radio, RefreshCw, Globe, Settings, Activity } from 'lucide-react'
import { PageHeader } from '@/components/dashboard/page-header'
import { DataCard } from '@/components/dashboard/data-card'
import { Button } from '@/components/ui/button'

interface MNO {
  id: number
  name: string
  countryCode: string
  status: string
  connectionsCount: number
  activeConnections: number
  lastActivity: string
  dailyMessages: number
  deliveryRate: number
}

interface MNOsListProps {
  initialMNOs: MNO[]
}

function getStatusBadge(status: string) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-400',
    inactive: 'bg-zinc-500/20 text-zinc-400',
  }
  
  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.inactive}`}>
      {status}
    </span>
  )
}

function getMNOBadgeColor(mno: string) {
  const colors: Record<string, string> = {
    MTN: 'bg-emerald-500/20 text-emerald-400',
    Airtel: 'bg-red-500/20 text-red-400',
    Glo: 'bg-blue-500/20 text-blue-400',
    '9mobile': 'bg-amber-500/20 text-amber-400',
    'etisalat': 'bg-purple-500/20 text-purple-400',
  }
  return colors[mno] || 'bg-zinc-500/20 text-zinc-400'
}

export function MNOsList({ initialMNOs }: MNOsListProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [mnos, setMnos] = useState(initialMNOs)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const filteredMNOs = useMemo(() => {
    if (!searchTerm.trim()) return mnos
    const term = searchTerm.toLowerCase()
    return mnos.filter(mno => 
      mno.name.toLowerCase().includes(term) ||
      mno.countryCode.includes(term)
    )
  }, [mnos, searchTerm])

  const totalConnections = mnos.reduce((acc, m) => acc + m.connectionsCount, 0)
  const activeConnections = mnos.reduce((acc, m) => acc + m.activeConnections, 0)

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      const res = await fetch('/api/admin/mnos')
      if (res.ok) {
        const data = await res.json()
        setMnos(data.mnos || [])
      }
    } catch (error) {
      console.error('Error refreshing MNOs:', error)
    } finally {
      setIsRefreshing(false)
    }
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="MNO Networks"
        description="Manage mobile network operators and their connections"
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              onClick={handleRefresh}
              disabled={isRefreshing}
              className="border-white/10 text-zinc-400 hover:text-white hover:bg-white/5"
            >
              <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
              Refresh Status
            </Button>
            <Link
              href="/admin/mnos/new"
              className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
            >
              <Plus className="h-4 w-4" />
              Add MNO
            </Link>
          </div>
        }
      />

      <div className="grid gap-4 md:grid-cols-4">
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-500/20">
              <Globe className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total MNOs</p>
              <p className="text-2xl font-bold text-white">{mnos.length}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/20">
              <Radio className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Active Connections</p>
              <p className="text-2xl font-bold text-emerald-400">{activeConnections}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-purple-500/20">
              <Activity className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Total Connections</p>
              <p className="text-2xl font-bold text-white">{totalConnections}</p>
            </div>
          </div>
        </DataCard>
        <DataCard title="">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-amber-500/20">
              <Settings className="h-5 w-5 text-amber-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Health Score</p>
              <p className="text-2xl font-bold text-white">
                {totalConnections > 0 ? Math.round((activeConnections / totalConnections) * 100) : 0}%
              </p>
            </div>
          </div>
        </DataCard>
      </div>

      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
        <input 
          type="text" 
          placeholder="Search MNOs by name or country code..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full pl-10 pr-4 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20"
        />
      </div>

      {filteredMNOs.length === 0 ? (
        <div className="text-center py-12 rounded-2xl border border-white/10 bg-zinc-900/50">
          <Globe className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
          <p className="text-zinc-500">
            {searchTerm ? 'No MNOs match your search' : 'No MNO networks found'}
          </p>
          {!searchTerm && (
            <Link href="/admin/mnos/new" className="text-white hover:underline">
              Add your first MNO
            </Link>
          )}
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          {filteredMNOs.map((mno) => (
            <DataCard key={mno.id} title="">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className={`flex items-center justify-center w-12 h-12 rounded-lg ${getMNOBadgeColor(mno.name)}`}>
                    <Globe className="h-6 w-6" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-white">{mno.name}</h3>
                    <p className="text-sm text-zinc-400">+{mno.countryCode}</p>
                  </div>
                </div>
                {getStatusBadge(mno.status)}
              </div>

              <div className="grid grid-cols-3 gap-4 mb-4">
                <div className="text-center p-3 rounded-lg bg-zinc-800/50">
                  <p className="text-2xl font-bold text-white">{mno.connectionsCount}</p>
                  <p className="text-xs text-zinc-400">Connections</p>
                </div>
                <div className="text-center p-3 rounded-lg bg-emerald-500/10">
                  <p className="text-2xl font-bold text-emerald-400">{mno.activeConnections}</p>
                  <p className="text-xs text-zinc-400">Active</p>
                </div>
                <div className="text-center p-3 rounded-lg bg-zinc-800/50">
                  <p className="text-2xl font-bold text-white">{mno.dailyMessages.toLocaleString()}</p>
                  <p className="text-xs text-zinc-400">Daily Msgs</p>
                </div>
              </div>

              <div className="flex items-center justify-between pt-4 border-t border-white/10">
                <div className="flex items-center gap-4 text-sm text-zinc-400">
                  <span className="flex items-center gap-1">
                    <Radio className="h-4 w-4" />
                    {mno.deliveryRate}% delivery
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <Link
                    href={`/admin/mnos/${mno.id}`}
                    className="p-1.5 text-zinc-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                  >
                    <Settings className="h-4 w-4" />
                  </Link>
                  <Link
                    href={`/admin/mnos/${mno.id}/connections/new`}
                    className="flex items-center gap-1 px-3 py-1.5 bg-white/10 text-white hover:bg-white/20 rounded-lg text-sm transition-colors"
                  >
                    <Plus className="h-3 w-3" />
                    Add Connection
                  </Link>
                </div>
              </div>
            </DataCard>
          ))}
        </div>
      )}
    </div>
  )
}
