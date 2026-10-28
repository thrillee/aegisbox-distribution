'use client'

import { useState, useEffect, useCallback } from 'react'
import { Link2, Loader2, Unlink } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface LinkedOrganization {
  clerkOrgId: string
  name: string | null
  slug: string | null
  createdAt: string
}

export function PortalAccessSettings({ providerId }: Readonly<{ providerId: number }>) {
  const [loading, setLoading] = useState(true)
  const [organizations, setOrganizations] = useState<LinkedOrganization[]>([])
  const [clerkOrgId, setClerkOrgId] = useState('')
  const [busy, setBusy] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [savedMessage, setSavedMessage] = useState<string | null>(null)

  const fetchOrganizations = useCallback(async () => {
    try {
      const res = await fetch(`/api/admin/service-providers/${providerId}/organization`)
      const data = await res.json().catch(() => ({}))
      if (res.ok) {
        setOrganizations(data.organizations || [])
      } else {
        setErrorMessage(data.error || 'Failed to load linked organizations')
      }
    } catch {
      setErrorMessage('Failed to load linked organizations')
    } finally {
      setLoading(false)
    }
  }, [providerId])

  useEffect(() => {
    fetchOrganizations()
  }, [fetchOrganizations])

  const handleLink = async (e: React.FormEvent) => {
    e.preventDefault()
    setErrorMessage(null)
    setSavedMessage(null)
    setBusy(true)
    try {
      const res = await fetch(`/api/admin/service-providers/${providerId}/organization`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clerk_org_id: clerkOrgId.trim() }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setErrorMessage(data.error || 'Failed to link organization')
        return
      }
      setClerkOrgId('')
      setSavedMessage('Organization linked')
      await fetchOrganizations()
    } catch {
      setErrorMessage('Failed to link organization')
    } finally {
      setBusy(false)
    }
  }

  const handleUnlink = async (orgId: string) => {
    if (!confirm('Members of this organization will lose access to this provider in the portal. Continue?')) return
    setErrorMessage(null)
    setSavedMessage(null)
    setBusy(true)
    try {
      const res = await fetch(
        `/api/admin/service-providers/${providerId}/organization?clerk_org_id=${encodeURIComponent(orgId)}`,
        { method: 'DELETE' }
      )
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setErrorMessage(data.error || 'Failed to unlink organization')
        return
      }
      setSavedMessage('Organization unlinked')
      await fetchOrganizations()
    } catch {
      setErrorMessage('Failed to unlink organization')
    } finally {
      setBusy(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
      </div>
    )
  }

  return (
    <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-6">
      <h3 className="text-lg font-semibold text-white">Portal Access</h3>
      <p className="text-sm text-zinc-400 mt-0.5 mb-6">
        Members of a linked Clerk organization see this provider&apos;s messages, sender IDs, templates and wallet in the provider portal.
      </p>

      {organizations.length === 0 ? (
        <p className="text-sm text-amber-400 mb-6">
          No organization is linked, so nobody can use the provider portal for this account.
        </p>
      ) : (
        <ul className="space-y-2 mb-6">
          {organizations.map((org) => (
            <li key={org.clerkOrgId} className="flex items-center justify-between rounded-lg border border-white/10 px-4 py-3">
              <div>
                <p className="font-medium text-white">{org.name || 'Unknown organization'}</p>
                <p className="text-sm text-zinc-500 font-mono">{org.clerkOrgId}</p>
              </div>
              <Button
                type="button"
                variant="outline"
                disabled={busy}
                onClick={() => handleUnlink(org.clerkOrgId)}
                className="bg-transparent border-white/10 text-zinc-400 hover:text-white hover:bg-white/5"
              >
                <Unlink className="h-4 w-4" />
                Unlink
              </Button>
            </li>
          ))}
        </ul>
      )}

      <form onSubmit={handleLink} className="space-y-4">
        <div>
          <label htmlFor={`clerk-org-${providerId}`} className="block text-sm font-medium text-zinc-300 mb-1">
            Clerk organization ID
          </label>
          <input
            id={`clerk-org-${providerId}`}
            type="text"
            placeholder="org_..."
            value={clerkOrgId}
            onChange={(e) => setClerkOrgId(e.target.value)}
            className="w-full px-3 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white placeholder:text-zinc-500 font-mono"
          />
          <p className="text-xs text-zinc-500 mt-1">
            Found in the Clerk dashboard under Organizations. Linking an organization moves it away from any other provider.
          </p>
        </div>
        {errorMessage && <p className="text-sm text-red-400">{errorMessage}</p>}
        {savedMessage && <p className="text-sm text-emerald-400">{savedMessage}</p>}
        <Button type="submit" disabled={busy || !clerkOrgId.trim()} className="bg-white text-black hover:bg-zinc-200">
          {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Link2 className="h-4 w-4" />}
          Link organization
        </Button>
      </form>
    </div>
  )
}
