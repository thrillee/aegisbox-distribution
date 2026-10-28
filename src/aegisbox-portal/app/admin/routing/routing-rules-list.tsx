'use client'

import { useState, Fragment } from 'react'
import Link from 'next/link'
import { ArrowRight, Edit, Trash2, X } from 'lucide-react'
import { Dialog, Transition } from '@headlessui/react'

interface Assignment {
  id: number
  routingGroupId: number
  msisdnPrefixGroupId: number
  mnoId: number
  priority: number
  status: string
  comment?: string | null
  mnoName?: string | null
  mnoCountryCode?: string | null
  prefixGroupName?: string | null
  prefixGroupReference?: string | null
  groupName?: string | null
}

interface RoutingRulesListProps {
  assignments: Assignment[]
  groupName: string
  groupId: number
}

export function RoutingRulesList({ assignments, groupName, groupId }: RoutingRulesListProps) {
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [showModal, setShowModal] = useState(false)
  const [assignmentToDelete, setAssignmentToDelete] = useState<Assignment | null>(null)

  const openDeleteModal = (assignment: Assignment, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setAssignmentToDelete(assignment)
    setShowModal(true)
  }

  const confirmDelete = async () => {
    if (!assignmentToDelete) return
    
    setDeletingId(assignmentToDelete.id)
    setShowModal(false)
    
    try {
      const res = await fetch(`/api/admin/routing-groups/${groupId}/assignments/${assignmentToDelete.id}`, {
        method: 'DELETE',
      })
      
      if (res.ok) {
        window.location.reload()
      } else {
        const data = await res.json()
        alert(data.error || 'Failed to delete routing rule')
      }
    } catch (error) {
      console.error('Error deleting assignment:', error)
      alert('Error deleting routing rule')
    } finally {
      setDeletingId(null)
      setAssignmentToDelete(null)
    }
  }

  if (assignments.length === 0) {
    return (
      <>
        <div className="p-8 text-center">
          <ArrowRight className="h-12 w-12 text-zinc-500 mx-auto mb-4" />
          <p className="text-zinc-500">
            No routing rules configured for {groupName}
          </p>
          <Link
            href={`/admin/routing/${groupId}/assignments/new`}
            className="text-white hover:underline mt-2 inline-block"
          >
            Add your first rule
          </Link>
        </div>
      </>
    )
  }

  return (
    <>
      <div className="divide-y divide-white/10">
        {assignments.map((assignment) => (
          <div
            key={assignment.id}
            className="p-4 hover:bg-white/5 transition-colors"
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="font-medium text-white">{assignment.prefixGroupName || 'N/A'}</span>
                <ArrowRight className="h-4 w-4 text-zinc-500" />
                <span className="font-medium text-white">{assignment.mnoName || 'N/A'}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  assignment.status === 'active'
                    ? 'bg-emerald-500/20 text-emerald-400'
                    : 'bg-zinc-500/20 text-zinc-400'
                }`}>
                  {assignment.status}
                </span>
              </div>
            </div>

            <div className="flex items-center justify-between text-sm text-zinc-500 mb-3">
              <span>Prefix: {assignment.prefixGroupReference || 'N/A'}</span>
              <span>Priority: {assignment.priority}</span>
            </div>

            {assignment.comment && (
              <p className="text-sm text-zinc-600 mb-3 italic">{assignment.comment}</p>
            )}

            <div className="flex items-center gap-2">
              <Link
                href={`/admin/routing/${groupId}/assignments/${assignment.id}`}
                className="flex items-center gap-1 px-2 py-1 text-xs text-zinc-400 hover:text-white hover:bg-white/10 rounded transition-colors"
              >
                <Edit className="h-3 w-3" />
                Edit
              </Link>
              <button
                onClick={(e) => openDeleteModal(assignment, e)}
                disabled={deletingId === assignment.id}
                className="flex items-center gap-1 px-2 py-1 text-xs text-zinc-400 hover:text-red-400 hover:bg-white/10 rounded transition-colors disabled:opacity-50"
              >
                <Trash2 className="h-3 w-3" />
                {deletingId === assignment.id ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        ))}
      </div>

      <Transition appear show={showModal} as={Fragment}>
        <Dialog as="div" className="relative z-50" onClose={() => setShowModal(false)}>
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
                    Delete Routing Rule
                  </Dialog.Title>
                  <div className="mt-2">
                    <p className="text-sm text-zinc-400">
                      Are you sure you want to delete this routing rule?
                    </p>
                    <p className="text-sm text-zinc-400 mt-2">
                      <strong className="text-white">{assignmentToDelete?.prefixGroupName}</strong> → <strong className="text-white">{assignmentToDelete?.mnoName}</strong>
                    </p>
                  </div>

                  <div className="mt-6 flex gap-3 justify-end">
                    <button
                      type="button"
                      className="px-4 py-2 text-sm font-medium text-zinc-400 bg-white/10 rounded-lg hover:bg-white/20 transition-colors"
                      onClick={() => setShowModal(false)}
                    >
                      Cancel
                    </button>
                    <button
                      type="button"
                      className="px-4 py-2 text-sm font-medium text-white bg-red-500 rounded-lg hover:bg-red-600 transition-colors"
                      onClick={confirmDelete}
                    >
                      Delete
                    </button>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition>
    </>
  )
}
