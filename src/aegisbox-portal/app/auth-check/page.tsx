'use client'

import { useEffect, useState, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useUser, useOrganization, useOrganizationList } from '@clerk/nextjs'
import { Building2, ArrowRight, CheckCircle2, Loader2 } from 'lucide-react'

function AuthCheckContent() {
  const { user, isLoaded: userLoaded } = useUser()
  const { organization: currentOrg, isLoaded: orgLoaded } = useOrganization()
  const { setActive, isLoaded: orgListLoaded } = useOrganizationList()
  const router = useRouter()
  const searchParams = useSearchParams()
  const reason = searchParams.get('reason')

  const [isCreating, setIsCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!userLoaded || !orgLoaded || !orgListLoaded) return

    const adminOrgId = process.env.NEXT_PUBLIC_CLERK_ADMIN_ORG_ID
    const isAdmin = currentOrg?.id === adminOrgId || 
      user?.publicMetadata?.role === 'admin'

    if (isAdmin && currentOrg) {
      router.push('/admin/dashboard')
      return
    }

    if (currentOrg) {
      router.push('/sp/dashboard')
      return
    }
  }, [userLoaded, orgLoaded, orgListLoaded, currentOrg, user, router])

  const handleCreateOrg = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setIsCreating(true)
    setError(null)

    const formData = new FormData(e.currentTarget)
    const orgName = formData.get('orgName') as string
    const slug = formData.get('slug') as string

    try {
      const response = await fetch('/api/auth/create-org', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orgName, slug }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to create organization')
      }

      const { orgId } = await response.json()
      
      await setActive?.({ organization: orgId })
      router.push('/sp/dashboard')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create organization')
      setIsCreating(false)
    }
  }

  if (!userLoaded || !orgLoaded || !orgListLoaded) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-white" />
      </div>
    )
  }

  if (currentOrg) {
    return null
  }

  return (
    <div className="min-h-screen bg-black flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="rounded-2xl border border-white/10 bg-zinc-900/50 p-8 shadow-xl">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-white/10 mb-4">
              <Building2 className="h-8 w-8 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-white mb-2">
              Welcome to Aegisbox
            </h1>
            <p className="text-zinc-400">
              {reason === 'no_org' 
                ? 'Create your organization to get started'
                : 'Set up your account to continue'}
            </p>
          </div>

          <form onSubmit={handleCreateOrg} className="space-y-4">
            <div>
              <label htmlFor="orgName" className="block text-sm font-medium text-zinc-300 mb-1">
                Organization Name
              </label>
              <input
                type="text"
                id="orgName"
                name="orgName"
                required
                minLength={2}
                maxLength={100}
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20 focus:border-transparent"
                placeholder="Your Company Name"
              />
            </div>

            <div>
              <label htmlFor="slug" className="block text-sm font-medium text-zinc-300 mb-1">
                Organization Slug
              </label>
              <input
                type="text"
                id="slug"
                name="slug"
                required
                minLength={2}
                maxLength={50}
                pattern="[a-z0-9-]+"
                className="w-full px-4 py-2 border border-white/10 rounded-lg bg-zinc-800 text-white focus:ring-2 focus:ring-white/20 focus:border-transparent"
                placeholder="your-company"
              />
              <p className="mt-1 text-xs text-zinc-500">
                Lowercase letters, numbers, and hyphens only
              </p>
            </div>

            {error && (
              <div className="p-3 rounded-lg bg-red-500/20 text-red-400 text-sm">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={isCreating}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-white text-black rounded-lg hover:bg-zinc-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isCreating ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  Create Organization
                  <ArrowRight className="h-4 w-4" />
                </>
              )}
            </button>
          </form>

          <div className="mt-6 pt-6 border-t border-white/10">
            <div className="flex items-start gap-3">
              <CheckCircle2 className="h-5 w-5 text-emerald-500 mt-0.5" />
              <div className="text-sm text-zinc-400">
                <p className="font-medium text-white mb-1">
                  What you&apos;ll get:
                </p>
                <ul className="space-y-1">
                  <li>• Dedicated SMS gateway dashboard</li>
                  <li>• API keys for integration</li>
                  <li>• Sender ID management</li>
                  <li>• Wallet and transaction tracking</li>
                  <li>• Real-time analytics</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <p className="mt-6 text-center text-sm text-zinc-500">
          Need help?{' '}
          <a href="mailto:support@aegisbox.com" className="text-white hover:underline">
            Contact Support
          </a>
        </p>
      </div>
    </div>
  )
}

export default function AuthCheckPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    }>
      <AuthCheckContent />
    </Suspense>
  )
}
