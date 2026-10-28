'use client'

import { use, useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowLeft, Loader2, Route, Users, Globe, Settings, ChevronRight, Plus } from 'lucide-react'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

function getStatusBadge(status: string) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-400',
    inactive: 'bg-zinc-500/20 text-zinc-400',
    pending: 'bg-amber-500/20 text-amber-400',
  }
  
  return (
    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.inactive}`}>
      {status}
    </span>
  )
}

interface RoutingGroup {
  id: number
  name: string
  reference: string
  description: string | null
  createdAt: string
  updatedAt: string
}

interface RoutingAssignment {
  id: number
  priority: number
  status: string
  mnoId: number | null
  mnoName: string | null
  mnoCountryCode: string | null
  mnoNetworkCode: string | null
  msisdnPrefixGroupId: number | null
  prefixGroupName: string | null
  prefixGroupReference: string | null
}

interface PageProps {
  params: Promise<{ id: string }>
}

export default function RoutingGroupDetailPage({ params }: PageProps) {
  const resolvedParams = use(params)
  const [loading, setLoading] = useState(true)
  const [routingGroup, setRoutingGroup] = useState<RoutingGroup | null>(null)
  const [assignments, setAssignments] = useState<RoutingAssignment[]>([])

  useEffect(() => {
    fetchRoutingGroupData()
  }, [resolvedParams.id])

  const fetchRoutingGroupData = async () => {
    try {
      const response = await fetch(`/api/admin/routing-groups/${resolvedParams.id}`)
      if (response.ok) {
        const data = await response.json()
        setRoutingGroup(data.routingGroup)
        setAssignments(data.assignments || [])
      }
    } catch (error) {
      console.error('Error fetching routing group:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <PageLoadingSkeleton />
  }

  if (!routingGroup) {
    return (
      <div className="p-8">
        <Link href="/admin/routing-groups" className="flex items-center gap-2 text-zinc-400">
          <ArrowLeft className="h-4 w-4" />
          Back to Routing Groups
        </Link>
        <div className="mt-8 text-center">
          <h1 className="text-2xl font-bold text-white">Routing Group Not Found</h1>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/admin/routing-groups" className="p-2 hover:bg-white/10 rounded-xl transition-colors">
          <ArrowLeft className="h-5 w-5 text-zinc-400" />
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <Route className="h-8 w-8 text-white" />
            <h1 className="text-3xl font-bold text-white">{routingGroup.name}</h1>
          </div>
          <p className="text-zinc-400 mt-1">{routingGroup.reference}</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors">
          <Settings className="h-4 w-4" />
          Edit
        </button>
      </div>

      {routingGroup.description && (
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-4">
          <p className="text-zinc-400">{routingGroup.description}</p>
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-3">
        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <div className="flex items-center gap-3 mb-2">
            <Route className="h-5 w-5 text-zinc-400" />
            <h3 className="font-semibold text-white">Reference</h3>
          </div>
          <code className="text-sm font-mono text-zinc-400 bg-white/5 px-2 py-1 rounded">
            {routingGroup.reference}
          </code>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <div className="flex items-center gap-3 mb-2">
            <ChevronRight className="h-5 w-5 text-zinc-400" />
            <h3 className="font-semibold text-white">Assignments</h3>
          </div>
          <p className="text-2xl font-bold text-white">{assignments.length}</p>
          <p className="text-sm text-zinc-500">MNO routes configured</p>
        </div>

        <div className="bg-zinc-900/50 border border-white/10 rounded-xl p-6">
          <div className="flex items-center gap-3 mb-2">
            <Users className="h-5 w-5 text-zinc-400" />
            <h3 className="font-semibold text-white">Created</h3>
          </div>
          <p className="text-sm text-white">
            {new Date(routingGroup.createdAt).toLocaleDateString()}
          </p>
          <p className="text-xs text-zinc-500">
            {new Date(routingGroup.createdAt).toLocaleTimeString()}
          </p>
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-xl">
        <div className="p-6 border-b border-white/10 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Globe className="h-5 w-5 text-zinc-400" />
            <h2 className="text-lg font-semibold text-white">Routing Assignments</h2>
          </div>
          <Link
            href={`/admin/routing/${routingGroup.id}/assignments/new`}
            className="flex items-center gap-2 px-3 py-1.5 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors text-sm"
          >
            <Plus className="h-4 w-4" />
            Add Assignment
          </Link>
        </div>

        {assignments.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-white/5">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Priority</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">MNO</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Prefix Group</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {assignments.map((assignment) => (
                  <tr key={assignment.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-white/10 text-white font-medium text-sm">
                        {assignment.priority}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <p className="font-medium text-white">
                          {assignment.mnoName || 'Unknown MNO'}
                        </p>
                        {assignment.mnoCountryCode && (
                          <p className="text-xs text-zinc-500">
                            {assignment.mnoCountryCode} {assignment.mnoNetworkCode}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <p className="text-white">
                          {assignment.prefixGroupName || 'Default'}
                        </p>
                        {assignment.prefixGroupReference && (
                          <p className="text-xs text-zinc-500">
                            {assignment.prefixGroupReference}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(assignment.status)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-12 text-center">
            <Globe className="h-12 w-12 mx-auto mb-4 text-zinc-600" />
            <h3 className="text-lg font-medium text-white mb-2">No Routing Assignments</h3>
            <p className="text-zinc-500 mb-4">
              This routing group doesn't have any MNO assignments yet.
            </p>
            <Link
              href={`/admin/routing/${routingGroup.id}/assignments/new`}
              className="inline-flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors"
            >
              <Plus className="h-4 w-4" />
              Add First Assignment
            </Link>
          </div>
        )}
      </div>
    </div>
  )
}
