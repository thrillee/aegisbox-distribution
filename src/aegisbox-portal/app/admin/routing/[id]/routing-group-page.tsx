'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { ArrowLeft, Plus, MapPin, ArrowRight, Settings, Edit, Trash2, Loader2 } from 'lucide-react'
import { Dialog, Transition } from '@headlessui/react'
import { Fragment } from 'react'

interface Assignment {
  id: number
  msisdnPrefixGroupId: number
  mnoId: number
  status: string
  priority: number
  comment?: string | null
  createdAt: string
  msisdnPrefixGroupName: string
  mnoName: string
  prefixes: { id: number; prefix: string }[]
}

interface RoutingGroup {
  id: number
  name: string
  reference: string
  description?: string | null
  status?: string
  isDefault?: boolean
  assignments?: Assignment[]
}

interface RoutingGroupPageProps {
  initialGroup: RoutingGroup
}

function getStatusBadge(status: string) {
  const styles: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-400',
    inactive: 'bg-zinc-500/20 text-zinc-400',
  }
  return <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.inactive}`}>{status}</span>
}

export function RoutingGroupPage({ initialGroup }: RoutingGroupPageProps) {
  const router = useRouter()
  const [group, setGroup] = useState(initialGroup)
  const [isEditing, setIsEditing] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [loading, setLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: group.name || '',
    reference: group.reference || '',
    description: group.description || '',
    status: group.status || 'active',
  })

  const handleSave = async () => {
    setLoading(true)
    try {
      const res = await fetch(`/api/admin/routing-groups/${group.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })
      
      if (res.ok) {
        const data = await res.json()
        setGroup({ ...group, ...data.routingGroup })
        setIsEditing(false)
      } else {
        const error = await res.json()
        alert(error.error || 'Failed to update routing group')
      }
    } catch (error) {
      console.error('Error updating group:', error)
      alert('Error updating routing group')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    try {
      const res = await fetch(`/api/admin/routing-groups/${group.id}`, {
        method: 'DELETE',
      })
      
      if (res.ok) {
        router.push('/admin/routing')
      } else {
        const error = await res.json()
        alert(error.error || 'Failed to delete routing group')
      }
    } catch (error) {
      console.error('Error deleting group:', error)
      alert('Error deleting routing group')
    } finally {
      setIsDeleting(false)
      setShowDeleteModal(false)
    }
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/admin/routing" className="p-2 hover:bg-white/10 rounded-xl transition-colors">
            <ArrowLeft className="h-5 w-5 text-zinc-400" />
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-3xl font-bold text-white">{group.name}</h1>
              {group.isDefault && (
                <span className="px-2 py-0.5 bg-white/10 text-white text-xs font-medium rounded-full">Default</span>
              )}
            </div>
            <p className="text-zinc-400 mt-1">{group.reference}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {isEditing ? (
            <>
              <button
                onClick={() => setIsEditing(false)}
                className="px-4 py-2 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={loading}
                className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-xl hover:bg-zinc-200 transition-colors disabled:opacity-50"
              >
                {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Settings className="h-4 w-4" />}
                {loading ? 'Saving...' : 'Save'}
              </button>
            </>
          ) : (
            <>
              <button
                onClick={() => setIsEditing(true)}
                className="flex items-center gap-2 px-4 py-2 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors"
              >
                <Edit className="h-4 w-4" />
                Edit
              </button>
              <button
                onClick={() => setShowDeleteModal(true)}
                className="flex items-center gap-2 px-4 py-2 border border-red-500/20 rounded-xl text-red-400 hover:bg-red-500/10 transition-colors"
              >
                <Trash2 className="h-4 w-4" />
                Delete
              </button>
            </>
          )}
        </div>
      </div>

      {isEditing && (
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="font-semibold text-white mb-4">Edit Routing Group</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Reference *</label>
              <input
                type="text"
                value={formData.reference}
                onChange={(e) => setFormData({ ...formData, reference: e.target.value })}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white"
                required
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-zinc-400 mb-1">Description</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                rows={2}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                className="w-full px-3 py-2 border border-white/10 rounded-xl bg-black/50 text-white"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
          </div>
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-500/20 rounded-lg">
              <Settings className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Status</p>
              <p className="text-xl font-bold">{getStatusBadge(group.status || 'active')}</p>
            </div>
          </div>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-500/20 rounded-lg">
              <MapPin className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Routing Rules</p>
              <p className="text-xl font-bold text-white">{group.assignments?.length || 0}</p>
            </div>
          </div>
        </div>
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-500/20 rounded-lg">
              <ArrowRight className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-zinc-400">Prefix Groups</p>
              <p className="text-xl font-bold text-white">{group.assignments?.length || 0}</p>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-zinc-900/50 border border-white/10 rounded-2xl">
        <div className="p-4 border-b border-white/10 flex items-center justify-between">
          <h2 className="font-semibold text-white">Routing Rules</h2>
          <Link href={`/admin/routing/${group.id}/assignments/new`} className="flex items-center gap-2 px-3 py-1.5 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors text-sm">
            <Plus className="h-4 w-4" />
            Add Rule
          </Link>
        </div>

        {(!group.assignments || group.assignments.length === 0) ? (
          <div className="p-8 text-center">
            <MapPin className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
            <p className="text-zinc-500">No routing rules configured</p>
            <Link href={`/admin/routing/${group.id}/assignments/new`} className="text-white hover:underline">
              Add your first rule
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-black/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Priority</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Prefix Group</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">MNO</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Prefixes</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {group.assignments.map((assignment) => (
                  <tr key={assignment.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap font-medium text-white">{assignment.priority}</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <MapPin className="h-4 w-4 text-zinc-500" />
                        <span className="font-medium text-white">{assignment.msisdnPrefixGroupName}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-white">{assignment.mnoName}</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex flex-wrap gap-1">
                        {(assignment.prefixes || []).slice(0, 3).map((p) => (
                          <span key={p.prefix} className="px-2 py-0.5 bg-white/10 rounded text-xs text-white">{p.prefix}</span>
                        ))}
                        {(assignment.prefixes?.length || 0) > 3 && (
                          <span className="px-2 py-0.5 bg-white/10 rounded text-xs text-zinc-400">+{(assignment.prefixes?.length || 0) - 3}</span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">{getStatusBadge(assignment.status)}</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <Link
                        href={`/admin/routing/${group.id}/assignments/${assignment.id}`}
                        className="text-white hover:underline text-sm"
                      >
                        Edit
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <Transition appear show={showDeleteModal} as={Fragment}>
        <Dialog as="div" className="relative z-50" onClose={() => setShowDeleteModal(false)}>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-black/50" />
          </Transition.Child>

          <div className="fixed inset-0 overflow-y-auto">
            <div className="flex min-h-full items-center justify-center p-4">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 scale-95"
                enterTo="opacity-100 scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 scale-100"
                leaveTo="opacity-0 scale-95"
              >
                <Dialog.Panel className="w-full max-w-md transform overflow-hidden rounded-2xl bg-zinc-900/50 border border-white/10 p-6 shadow-xl transition-all backdrop-blur-xl">
                  <Dialog.Title as="h3" className="text-lg font-medium text-white mb-4">
                    Delete Routing Group
                  </Dialog.Title>
                  <div className="mt-2">
                    <p className="text-sm text-zinc-400">
                      Are you sure you want to delete <strong className="text-white">{group.name}</strong>?
                    </p>
                    <p className="text-sm text-zinc-400 mt-2">
                      This will also delete all {group.assignments?.length || 0} routing rules associated with this group.
                    </p>
                  </div>

                  <div className="mt-6 flex gap-3 justify-end">
                    <button
                      type="button"
                      className="px-4 py-2 text-sm font-medium text-zinc-400 bg-white/10 rounded-lg hover:bg-white/20 transition-colors"
                      onClick={() => setShowDeleteModal(false)}
                    >
                      Cancel
                    </button>
                    <button
                      type="button"
                      className="px-4 py-2 text-sm font-medium text-white bg-red-500 rounded-lg hover:bg-red-600 transition-colors"
                      onClick={handleDelete}
                      disabled={isDeleting}
                    >
                      {isDeleting ? 'Deleting...' : 'Delete'}
                    </button>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition>
    </div>
  )
}
