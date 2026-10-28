'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Search, Users, Hash, Phone, Edit2, Trash2, Eye, Plus, Loader2 } from 'lucide-react'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

export default function PrefixGroupsPage() {
  const [groups, setGroups] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [initialLoad, setInitialLoad] = useState(true)
  const [search, setSearch] = useState('')

  useEffect(() => {
    fetchGroups()
  }, [])

  const fetchGroups = async () => {
    try {
      const response = await fetch('/api/admin/msisdn-prefix-groups')
      if (response.ok) {
        const data = await response.json()
        setGroups(data.groups || [])
      }
    } catch (error) {
      console.error('Error fetching groups:', error)
    } finally {
      setLoading(false)
      setInitialLoad(false)
    }
  }

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      active: 'bg-emerald-500/20 text-emerald-400',
      inactive: 'bg-zinc-500/20 text-zinc-400',
    }
    const style = styles[status] || styles.inactive
    return <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${style}`}>{status}</span>
  }

  const getMNOColor = (mno: string) => {
    const colors: Record<string, string> = {
      MTN: 'bg-emerald-500/20 text-emerald-400',
      Airtel: 'bg-red-500/20 text-red-400',
      Glo: 'bg-blue-500/20 text-blue-400',
      '9mobile': 'bg-amber-500/20 text-amber-400',
    }
    return colors[mno] || 'bg-zinc-500/20 text-zinc-400'
  }

  const getMNOForPrefix = (prefix: string) => {
    if (prefix.startsWith('+')) return 'International'
    const prefixMap: Record<string, string> = {
      '0701': 'MTN', '0702': 'Airtel', '0703': 'Airtel', '0704': 'Airtel', '0705': 'Airtel',
      '0802': 'MTN', '0803': 'MTN', '0804': 'MTN', '0805': 'MTN', '0806': 'MTN',
      '0807': 'MTN', '0808': 'MTN', '0809': 'MTN', '0810': 'MTN', '0811': 'MTN',
      '0812': 'MTN', '0813': 'MTN', '0814': 'MTN', '0815': 'MTN', '0816': 'MTN',
      '0817': '9mobile', '0818': '9mobile', '0819': '9mobile',
      '0901': 'Glo', '0902': 'Glo', '0903': 'Glo', '0904': 'Glo', '0905': 'Glo', '0906': 'Glo',
    }
    return prefixMap[prefix] || 'Unknown'
  }

  const filteredGroups = groups.filter(group =>
    group.name.toLowerCase().includes(search.toLowerCase()) ||
    group.description?.toLowerCase().includes(search.toLowerCase())
  )

  if (initialLoad) {
    return <PageLoadingSkeleton />
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Prefix Groups</h1>
          <p className="text-zinc-400 mt-1">Manage phone number prefix groups for routing</p>
        </div>
        <Link href="/admin/routing/prefix-groups/new" className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors">
          <Plus className="h-4 w-4" />
          Create Group
        </Link>
      </div>

      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
        <input type="text" placeholder="Search prefix groups..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-10 pr-4 py-2 bg-zinc-900/50 border border-white/10 rounded-lg text-white placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-white/20" />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {filteredGroups.map((group) => (
          <div key={group.id} className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6 hover:bg-white/5 transition-colors">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="flex items-center justify-center w-12 h-12 bg-white/10 text-white rounded-xl">
                    <Users className="h-6 w-6" />
                </div>
                <div>
                  <h3 className="font-semibold text-white">{group.name}</h3>
                  <p className="text-sm text-zinc-500">{group.description}</p>
                </div>
              </div>
              {getStatusBadge(group.status || 'active')}
            </div>

            <div className="mb-4">
              <h4 className="text-sm font-medium text-zinc-400 mb-2">Prefixes</h4>
              <div className="flex flex-wrap gap-1">
                {(group.prefixes || []).slice(0, 8).map((p: any) => {
                  const mno = getMNOForPrefix(p.prefix)
                  return (
                    <span key={p.prefix} className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-mono ${getMNOColor(mno)}`}>
                      <Phone className="h-3 w-3" />
                      {p.prefix}
                    </span>
                  )
                })}
                {(group.prefixes?.length || 0) > 8 && (
                  <span className="px-2 py-0.5 bg-zinc-500/20 text-zinc-500 rounded text-xs">
                    +{(group.prefixes?.length || 0) - 8} more
                  </span>
                )}
                {(!group.prefixes || group.prefixes.length === 0) && (
                  <span className="text-sm text-zinc-500">No prefixes added</span>
                )}
              </div>
            </div>

            <div className="flex items-center justify-between pt-4 border-t border-white/10">
              <div className="flex items-center gap-4 text-sm text-zinc-500">
                <span className="flex items-center gap-1">
                  <Hash className="h-4 w-4" />
                  {(group.prefixes?.length || 0)} prefixes
                </span>
              </div>
              <div className="flex items-center gap-2">
                <button className="p-1.5 text-zinc-500 hover:text-white hover:bg-white/10 rounded transition-colors">
                  <Eye className="h-4 w-4" />
                </button>
                <button className="p-1.5 text-zinc-500 hover:text-white hover:bg-white/10 rounded transition-colors">
                  <Edit2 className="h-4 w-4" />
                </button>
                <button className="p-1.5 text-zinc-500 hover:text-red-400 hover:bg-white/10 rounded transition-colors">
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {filteredGroups.length === 0 && (
        <div className="text-center py-12">
          <Users className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
          <p className="text-zinc-500">No prefix groups found</p>
          <Link href="/admin/routing/prefix-groups/new" className="text-white hover:underline">
            Create your first group
          </Link>
        </div>
      )}

      <div className="bg-blue-500/10 border border-blue-500/20 rounded-2xl p-6">
        <h3 className="font-semibold text-white mb-2">Need Custom Routing?</h3>
        <p className="text-sm text-zinc-400 mb-4">
          Create prefix groups to route messages to specific MNOs based on phone number prefixes.
          This helps optimize delivery rates and costs.
        </p>
      </div>
    </div>
  )
}
