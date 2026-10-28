'use client'

import Link from 'next/link'
import { useUser, useOrganization, SignOutButton } from '@clerk/nextjs'
import { Button } from '@/components/ui/button'
import { MessageSquare, LogOut, Loader2 } from 'lucide-react'

export function LandingNavbar() {
  const { user, isLoaded: userLoaded } = useUser()
  const { organization, isLoaded: orgLoaded } = useOrganization()

  const isLoaded = userLoaded && orgLoaded
  const isSignedIn = !!user

  const getDashboardUrl = () => {
    const adminOrgId = process.env.NEXT_PUBLIC_CLERK_ADMIN_ORG_ID
    const isAdminOrg = organization?.id === adminOrgId
    const isAdminRole = (user?.publicMetadata as Record<string, unknown>)?.role === 'admin'
    
    if (isAdminOrg || isAdminRole) {
      return '/admin/dashboard'
    }
    if (organization) {
      return '/sp/dashboard'
    }
    return '/auth-check'
  }

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 border-b border-white/10 bg-black/50 backdrop-blur-xl">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-white text-black">
            <MessageSquare className="h-5 w-5" />
          </div>
          <span className="text-lg font-semibold tracking-tight text-white">Aegisbox</span>
        </div>
        <div className="hidden md:flex items-center gap-8">
          <Link href="#features" className="text-sm text-zinc-400 hover:text-white transition-colors">
            Features
          </Link>
          <Link href="#networks" className="text-sm text-zinc-400 hover:text-white transition-colors">
            Networks
          </Link>
          <Link href="#pricing" className="text-sm text-zinc-400 hover:text-white transition-colors">
            Pricing
          </Link>
        </div>
        <div className="flex items-center gap-3">
          {!isLoaded ? (
            <Loader2 className="h-5 w-5 animate-spin text-zinc-400" />
          ) : isSignedIn ? (
            <>
              <Link href={getDashboardUrl()}>
                <Button className="bg-white text-black hover:bg-zinc-200">
                  Dashboard
                </Button>
              </Link>
              <SignOutButton>
                <Button variant="ghost" className="text-zinc-400 hover:text-white hover:bg-white/5 gap-2">
                  <LogOut className="h-4 w-4" />
                  Sign Out
                </Button>
              </SignOutButton>
            </>
          ) : (
            <>
              <Link href="/sign-in">
                <Button variant="ghost" className="text-zinc-400 hover:text-white hover:bg-white/5">
                  Sign In
                </Button>
              </Link>
              <Link href="/sign-up">
                <Button className="bg-white text-black hover:bg-zinc-200">
                  Get Started
                </Button>
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  )
}
