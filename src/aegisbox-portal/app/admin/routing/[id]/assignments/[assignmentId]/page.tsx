'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter, useParams } from 'next/navigation'
import { ArrowLeft, Save, Route, Loader2, Trash2 } from 'lucide-react'
import { PageLoadingSkeleton } from '@/components/routing/loading-components'

interface PrefixGroup {
  id: number
  name: string
  reference: string
}

interface MNO {
  id: number
  name: string
  countryCode: string
}

interface Assignment {
  id: number
  routingGroupId: number
  msisdnPrefixGroupId: number
  mnoId: number
  priority: number
  status: string
  comment?: string | null
  prefixGroupName?: string | null
  prefixGroupReference?: string | null
  mnoName?: string | null
}

export default function EditAssignmentPage() {
  const router = useRouter()
  const params = useParams()
  const groupId = params.id as string
  const assignmentId = params.assignmentId as string

  const [loading, setLoading] = useState(false)
  const [loadingData, setLoadingData] = useState(true)
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [prefixGroups, setPrefixGroups] = useState<PrefixGroup[]>([])
  const [mnos, setMnos] = useState<MNO[]>([])
  const [assignment, setAssignment] = useState<Assignment | null>(null)
  const [formData, setFormData] = useState({
    msisdnPrefixGroupId: '',
    mnoId: '',
    priority: '0',
    comment: '',
    status: 'active',
  })

  useEffect(() => {
    const fetchData = async () => {
      try {
        const pgRes = await fetch('/api/admin/msisdn-prefix-groups')
        if (pgRes.ok) {
          const data = await pgRes.json()
          setPrefixGroups(data.groups || [])
        }

        const mnoRes = await fetch('/api/admin/mnos')
        if (mnoRes.ok) {
          const data = await mnoRes.json()
          setMnos(data.mnos || [])
        }

        const assignmentRes = await fetch(`/api/admin/routing-groups/${groupId}/assignments/${assignmentId}`)
        console.log('Assignment API response:', assignmentRes.status, assignmentRes.statusText)
        
        if (assignmentRes.ok) {
          const data = await assignmentRes.json()
          console.log('Assignment data:', data)
          const a = data.assignment
          if (a) {
            setAssignment(a)
            setFormData({
              msisdnPrefixGroupId: String(a.msisdnPrefixGroupId || ''),
              mnoId: String(a.mnoId || ''),
              priority: String(a.priority || '0'),
              comment: a.comment || '',
              status: a.status || 'active',
            })
          }
        } else {
          const errorData = await assignmentRes.json()
          console.error('Assignment API error:', errorData)
          setError(errorData.error || 'Failed to load assignment')
        }
      } catch (error) {
        console.error('Error fetching data:', error)
        setError('Failed to load data')
      } finally {
        setLoadingData(false)
      }
    }
    fetchData()
  }, [groupId, assignmentId])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await fetch(`/api/admin/routing-groups/${groupId}/assignments/${assignmentId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          msisdnPrefixGroupId: parseInt(formData.msisdnPrefixGroupId),
          mnoId: parseInt(formData.mnoId),
          priority: parseInt(formData.priority),
          comment: formData.comment,
          status: formData.status,
        }),
      })

      if (response.ok) {
        router.push(`/admin/routing/${groupId}`)
      } else {
        const error = await response.json()
        alert(error.error || 'Failed to update assignment')
      }
    } catch (error) {
      console.error('Error updating assignment:', error)
      alert('An error occurred')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this routing rule?')) {
      return
    }

    setDeleting(true)
    
    try {
      const res = await fetch(`/api/admin/routing-groups/${groupId}/assignments/${assignmentId}`, {
        method: 'DELETE',
      })
      
      if (res.ok) {
        router.push(`/admin/routing/${groupId}`)
      } else {
        const data = await res.json()
        alert(data.error || 'Failed to delete routing rule')
      }
    } catch (error) {
      console.error('Error deleting assignment:', error)
      alert('Error deleting routing rule')
    } finally {
      setDeleting(false)
    }
  }

  if (loadingData) {
    return <PageLoadingSkeleton />
  }

  if (error || !assignment) {
    return (
      <div className="p-8 text-center">
        <h1 className="text-2xl font-bold text-white mb-4">
          {error || 'Assignment Not Found'}
        </h1>
        <Link href={`/admin/routing/${groupId}`} className="text-white hover:underline">
          Back to Routing Group
        </Link>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="mb-6">
        <Link href={`/admin/routing/${groupId}`} className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to Routing Group
        </Link>
      </div>

      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Edit Routing Rule</h1>
          <p className="text-zinc-400 mt-1">Update the routing rule assignment</p>
        </div>
        <button
          onClick={handleDelete}
          disabled={deleting}
          className="flex items-center gap-2 px-4 py-2 bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30 transition-colors disabled:opacity-50"
        >
          <Trash2 className="h-4 w-4" />
          {deleting ? 'Deleting...' : 'Delete'}
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-zinc-900/50 border border-white/10 rounded-2xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <Route className="h-5 w-5" />
            Assignment Details
          </h2>

          <div className="grid gap-6 md:grid-cols-2">
            <div>
              <label htmlFor="msisdnPrefixGroupId" className="block text-sm font-medium text-zinc-400 mb-1">Prefix Group *</label>
              <select id="msisdnPrefixGroupId" name="msisdnPrefixGroupId" value={formData.msisdnPrefixGroupId} onChange={handleChange}
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20">
                <option value="">Select Prefix Group</option>
                {prefixGroups.map((pg) => (
                  <option key={pg.id} value={pg.id}>{pg.name} ({pg.reference})</option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="mnoId" className="block text-sm font-medium text-zinc-400 mb-1">MNO *</label>
              <select id="mnoId" name="mnoId" value={formData.mnoId} onChange={handleChange}
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20">
                <option value="">Select MNO</option>
                {mnos.map((mno) => (
                  <option key={mno.id} value={mno.id}>{mno.name} (+{mno.countryCode})</option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="priority" className="block text-sm font-medium text-zinc-400 mb-1">Priority</label>
              <input type="number" id="priority" name="priority" value={formData.priority} onChange={handleChange}
                min="0" step="1"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20" />
              <p className="mt-1 text-xs text-zinc-500">Lower values = higher priority</p>
            </div>

            <div>
              <label htmlFor="status" className="block text-sm font-medium text-zinc-400 mb-1">Status</label>
              <select id="status" name="status" value={formData.status} onChange={handleChange}
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/20">
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>

            <div className="md:col-span-2">
              <label htmlFor="comment" className="block text-sm font-medium text-zinc-400 mb-1">Comment</label>
              <textarea id="comment" name="comment" value={formData.comment} onChange={handleChange}
                rows={2}
                placeholder="Optional notes about this routing rule"
                className="w-full px-3 py-2 bg-zinc-800 border border-white/10 rounded-lg text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-white/20" />
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3">
          <Link href={`/admin/routing/${groupId}`}
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition-colors">
            Cancel
          </Link>
          <button type="submit" disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-50">
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {loading ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  )
}
