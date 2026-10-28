import { redirect } from 'next/navigation'
import { auth } from '@clerk/nextjs/server'
import { DashboardSidebar } from '@/components/dashboard/sidebar'

export default async function SPLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { userId, orgId, sessionClaims } = await auth()

  if (!userId) {
    redirect('/sign-in')
  }

  const adminOrgId = process.env.CLERK_ADMIN_ORG_ID
  const metadata = sessionClaims?.metadata as Record<string, unknown> | undefined
  const isAdmin = orgId === adminOrgId || metadata?.role === 'admin'

  if (isAdmin) {
    redirect('/admin/dashboard')
  }

  if (!orgId) {
    redirect('/auth-check')
  }

  return (
    <div className="min-h-screen bg-black">
      <DashboardSidebar variant="sp" />
      <main className="lg:ml-64 min-h-screen">
        {children}
      </main>
    </div>
  )
}
