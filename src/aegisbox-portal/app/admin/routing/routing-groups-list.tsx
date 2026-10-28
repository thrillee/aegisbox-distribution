'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Plus, Settings, Trash2, ChevronRight } from 'lucide-react'

interface Group {
  id: number
  name: string
  reference: string
  description?: string
  assignmentCount: number
}

interface RoutingGroupsListProps {
  groups: Group[]
  selectedGroupId?: number
}

export function RoutingGroupsList({ groups, selectedGroupId }: RoutingGroupsListProps) {
  const [deletingId, setDeletingId] = useState<number | null>(null)

  const handleDelete = async (id: number, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    
    if (!confirm('Are you sure you want to delete this routing group?')) {
      return
    }

    setDeletingId(id)
    
    try {
      const res = await fetch(`/api/admin/routing-groups/${id}`, {
        method: 'DELETE',
      })
      
      if (res.ok) {
        window.location.reload()
      } else {
        alert('Failed to delete routing group')
      }
    } catch (error) {
      console.error('Error deleting group:', error)
      alert('Error deleting routing group')
    } finally {
      setDeletingId(null)
    }
  }

  if (groups.length === 0) {
    return (
      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
        <div className="p-4 border-b border-white/10 flex items-center justify-between">
          <h2 className="font-semibold text-white">Routing Groups</h2>
        </div>
        <div className="p-8 text-center">
          <Settings className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
          <p className="text-zinc-500">No routing groups found</p>
          <Link href="/admin/routing/new" className="text-white hover:underline">
            Create your first group
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
      <div className="p-4 border-b border-white/10 flex items-center justify-between">
        <h2 className="font-semibold text-white">Routing Groups</h2>
        <Link
          href="/admin/routing/new"
          className="flex items-center gap-1 px-2 py-1 bg-white text-black rounded hover:bg-zinc-200 transition-colors text-xs"
        >
          <Plus className="h-3 w-3" />
          New
        </Link>
      </div>
      <div className="divide-y divide-white/10">
        {groups.map((group) => (
          <Link
            key={group.id}
            href={`/admin/routing?selected=${group.id}`}
            className={`block p-4 hover:bg-white/5 transition-colors ${
              selectedGroupId === group.id ? 'bg-white/5 border-l-2 border-l-white' : ''
            }`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {selectedGroupId === group.id && (
                  <ChevronRight className="h-4 w-4 text-white" />
                )}
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-white">{group.name}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-zinc-500 mt-1">
                    <span>{group.reference}</span>
                    <span className="text-xs text-zinc-600">•</span>
                    <span>{group.assignmentCount} rules</span>
                  </div>
                </div>
              </div>
              <button
                onClick={(e) => handleDelete(group.id, e)}
                disabled={deletingId === group.id}
                className="p-1.5 text-zinc-500 hover:text-red-400 hover:bg-white/10 rounded transition-colors disabled:opacity-50"
                title="Delete group"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
